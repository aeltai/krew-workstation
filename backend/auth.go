package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

type RancherUser struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

type UserCapabilities struct {
	Backups          bool `json:"backups"`
	SyncKubeconfig   bool `json:"syncKubeconfig"`
	ManagePlugins    bool `json:"managePlugins"`
	Terminal         bool `json:"terminal"`
	BackupOperatorUp bool `json:"backupOperatorUp"`
}

type AuthMeResponse struct {
	User         RancherUser      `json:"user"`
	AuthMode     string           `json:"authMode"`
	Capabilities UserCapabilities `json:"capabilities"`
	Error        string           `json:"error,omitempty"`
}

type requestUser struct {
	Token      string
	AuthMode   string
	User       RancherUser
	Kubeconfig string
}

const ctxRequestUser = "krewRequestUser"

func tokenFromRequest(c *gin.Context) string {
	if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if t := c.GetHeader("X-Rancher-Token"); t != "" {
		return t
	}
	if t := c.Query("token"); t != "" {
		return t
	}
	return ""
}

func allowServiceTokenFallback() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("ALLOW_SERVICE_TOKEN")))
	return v == "" || v == "true" || v == "1" || v == "yes"
}

func resolveRequestToken(c *gin.Context) (string, string, error) {
	if tok := tokenFromRequest(c); tok != "" {
		return tok, "session", nil
	}
	if allowServiceTokenFallback() {
		if tok := rancherToken(); tok != "" {
			return tok, "service", nil
		}
	}
	return "", "", fmt.Errorf("authentication required: log into Rancher or pass Authorization: Bearer <token>")
}

func fetchRancherUser(token string) (RancherUser, error) {
	body, err := rancherRequestWithToken("GET", "/v3/users?me=true", token)
	if err != nil {
		return RancherUser{}, err
	}
	var result struct {
		Data []struct {
			ID          string `json:"id"`
			Username    string `json:"username"`
			Name        string `json:"name"`
			DisplayName string `json:"displayName"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return RancherUser{}, err
	}
	if len(result.Data) == 0 {
		return RancherUser{}, fmt.Errorf("no user returned from Rancher")
	}
	u := result.Data[0]
	name := u.DisplayName
	if name == "" {
		name = u.Name
	}
	if name == "" {
		name = u.Username
	}
	return RancherUser{ID: u.ID, Username: u.Username, DisplayName: name}, nil
}

func userKubeconfigDir(userID string) string {
	safe := strings.NewReplacer("/", "_", ":", "_", " ", "_").Replace(userID)
	if safe == "" {
		safe = "anonymous"
	}
	return filepath.Join("/root/.kube/users", safe)
}

func kubeconfigPathForUser(userID string) string {
	return filepath.Join(userKubeconfigDir(userID), "config")
}

func requestUserFromContext(c *gin.Context) (*requestUser, bool) {
	v, ok := c.Get(ctxRequestUser)
	if !ok {
		return nil, false
	}
	ru, ok := v.(*requestUser)
	return ru, ok
}

func loadRequestUser(c *gin.Context) (*requestUser, error) {
	if ru, ok := requestUserFromContext(c); ok {
		return ru, nil
	}
	token, authMode, err := resolveRequestToken(c)
	if err != nil {
		return nil, err
	}
	user, err := fetchRancherUser(token)
	if err != nil {
		return nil, fmt.Errorf("invalid Rancher token: %w", err)
	}
	ru := &requestUser{
		Token:      token,
		AuthMode:   authMode,
		User:       user,
		Kubeconfig: kubeconfigPathForUser(user.ID),
	}
	c.Set(ctxRequestUser, ru)
	return ru, nil
}

func requireAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := loadRequestUser(c); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}
		c.Next()
	}
}

func kubectlAuthCanI(kubeCfg, resource, verb string) bool {
	out, err := runKubectlWithConfig(kubeCfg, "auth", "can-i", verb, resource)
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) == "yes"
}

func kubectlAuthCanINamespace(kubeCfg, resource, verb, ns string) bool {
	out, err := runKubectlWithConfig(kubeCfg, "auth", "can-i", verb, resource, "-n", ns)
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) == "yes"
}

func evaluateCapabilities(token, kubeCfg string) UserCapabilities {
	caps := UserCapabilities{
		SyncKubeconfig: true,
		ManagePlugins:  true,
		Terminal:       true,
	}

	if _, err := os.Stat(kubeCfg); err == nil {
		caps.SyncKubeconfig = kubectlAuthCanINamespace(kubeCfg, "secrets", "get", "default") ||
			kubectlAuthCanI(kubeCfg, "nodes", "list")
	}

	backupStatus := fetchBackupOperatorStatusForKubeconfig(kubeCfg)
	caps.BackupOperatorUp = backupStatus.Installed

	canListBackups := kubectlAuthCanI(kubeCfg, "backups.resources.cattle.io", "list")
	canListRestores := kubectlAuthCanI(kubeCfg, "restores.resources.cattle.io", "list")
	caps.Backups = backupStatus.Installed && (canListBackups || canListRestores)

	return caps
}

func syncKubeconfigForUser(token string, ru *requestUser) (int, error) {
	clusters, err := fetchClustersWithToken(token)
	if err != nil {
		return 0, err
	}
	if len(clusters) == 0 {
		return 0, nil
	}
	var configs []string
	for _, cl := range clusters {
		cfg, err := fetchKubeconfigWithToken(cl.ID, token)
		if err != nil {
			return 0, fmt.Errorf("cluster %s: %w", cl.Name, err)
		}
		configs = append(configs, cfg)
	}
	merged, err := mergeKubeconfigs(configs)
	if err != nil {
		return 0, err
	}
	dir := filepath.Dir(ru.Kubeconfig)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return 0, err
	}
	if err := os.WriteFile(ru.Kubeconfig, merged, 0600); err != nil {
		return 0, err
	}
	_ = os.MkdirAll("/root/backups", 0700)
	_ = ensurePolymorphConfigForKubeconfig(ru.Kubeconfig)
	return len(clusters), nil
}
