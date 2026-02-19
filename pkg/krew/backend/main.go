package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/creack/pty/v2"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gopkg.in/yaml.v3"
)

type Plugin struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
}

type PluginsResponse struct {
	Plugins        []Plugin `json:"plugins"`
	TerminalOutput string   `json:"terminalOutput,omitempty"`
	Error          string   `json:"error,omitempty"`
}

type Cluster struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

type ClustersResponse struct {
	Clusters []Cluster `json:"clusters"`
}

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
	Timeout: 30 * time.Second,
}

func rancherURL() string {
	if u := os.Getenv("RANCHER_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "https://rancher:443"
}

func rancherToken() string {
	return os.Getenv("RANCHER_TOKEN")
}

func rancherRequestWithToken(method, path, token string) ([]byte, error) {
	tok := token
	if tok == "" {
		tok = rancherToken()
	}
	if tok == "" {
		return nil, fmt.Errorf("no Rancher token: set RANCHER_TOKEN or pass Authorization header from logged-in session")
	}
	url := rancherURL() + path
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to %s failed: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("rancher API %s returned %d: %s", path, resp.StatusCode, string(body))
	}
	return body, nil
}

func rancherRequest(method, path string) ([]byte, error) {
	return rancherRequestWithToken(method, path, "")
}

func fetchClustersWithToken(token string) ([]Cluster, error) {
	body, err := rancherRequestWithToken("GET", "/v3/clusters", token)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []struct {
			ID    string `json:"id"`
			Name  string `json:"name"`
			State string `json:"state"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	clusters := make([]Cluster, len(result.Data))
	for i, c := range result.Data {
		clusters[i] = Cluster{ID: c.ID, Name: c.Name, State: c.State}
	}
	return clusters, nil
}

func fetchClusters() ([]Cluster, error) {
	return fetchClustersWithToken("")
}

func fetchKubeconfigWithToken(clusterID, token string) (string, error) {
	url := rancherURL() + "/v3/clusters/" + clusterID + "?action=generateKubeconfig"
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}
	tok := token
	if tok == "" {
		tok = rancherToken()
	}
	if tok == "" {
		return "", fmt.Errorf("no Rancher token")
	}
	req.Header.Set("Authorization", "Bearer "+tok)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("kubeconfig request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("kubeconfig API returned %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Config string `json:"config"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse kubeconfig response: %w", err)
	}
	return result.Config, nil
}

func krewRoot() string {
	if r := os.Getenv("KREW_ROOT"); r != "" {
		return r
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".krew")
}

// runKrew runs a krew command. Krew plugins are installed globally on the
// machine, they are NOT per-cluster. A kubeconfig is only needed when
// *running* a plugin, not when installing/uninstalling/searching.
func runKrew(args ...string) (string, error) {
	root := krewRoot()
	cmdArgs := append([]string{"krew"}, args...)
	cmd := exec.Command("kubectl", cmdArgs...)
	cmd.Env = append(os.Environ(),
		"KREW_ROOT="+root,
		fmt.Sprintf("PATH=%s:%s", filepath.Join(root, "bin"), os.Getenv("PATH")),
	)

	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		return output, fmt.Errorf("krew %s failed: %w\n%s", strings.Join(args, " "), err, output)
	}
	return output, nil
}

func parseInstalledPlugins(output string) map[string]string {
	installed := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] != "PLUGIN" {
			installed[fields[0]] = fields[1]
		}
	}
	return installed
}

func parseAvailablePlugins(output string) []Plugin {
	var plugins []Plugin
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] == "NAME" {
			continue
		}
		name := fields[0]
		desc := strings.Join(fields[1:], " ")
		plugins = append(plugins, Plugin{Name: name, Description: desc})
	}
	return plugins
}

// kubeconfig structures for merging
type kubeConfig struct {
	APIVersion     string                 `yaml:"apiVersion"`
	Kind           string                 `yaml:"kind"`
	Clusters       []namedCluster         `yaml:"clusters"`
	Contexts       []namedContext         `yaml:"contexts"`
	CurrentContext string                 `yaml:"current-context"`
	Users          []namedUser            `yaml:"users"`
	Preferences    map[string]interface{} `yaml:"preferences,omitempty"`
}

type namedCluster struct {
	Name    string                 `yaml:"name"`
	Cluster map[string]interface{} `yaml:"cluster"`
}

type namedContext struct {
	Name    string                 `yaml:"name"`
	Context map[string]interface{} `yaml:"context"`
}

type namedUser struct {
	Name string                 `yaml:"name"`
	User map[string]interface{} `yaml:"user"`
}

func mergeKubeconfigs(configs []string) ([]byte, error) {
	var merged kubeConfig
	merged.APIVersion = "v1"
	merged.Kind = "Config"
	seenClusters := make(map[string]bool)
	seenContexts := make(map[string]bool)
	seenUsers := make(map[string]bool)

	for i, cfgYaml := range configs {
		if strings.TrimSpace(cfgYaml) == "" {
			continue
		}
		var cfg kubeConfig
		if err := yaml.Unmarshal([]byte(cfgYaml), &cfg); err != nil {
			return nil, fmt.Errorf("parse config %d: %w", i, err)
		}
		for _, c := range cfg.Clusters {
			if !seenClusters[c.Name] {
				merged.Clusters = append(merged.Clusters, c)
				seenClusters[c.Name] = true
			}
		}
		for _, c := range cfg.Contexts {
			if !seenContexts[c.Name] {
				merged.Contexts = append(merged.Contexts, c)
				seenContexts[c.Name] = true
			}
		}
		for _, u := range cfg.Users {
			if !seenUsers[u.Name] {
				merged.Users = append(merged.Users, u)
				seenUsers[u.Name] = true
			}
		}
		if merged.CurrentContext == "" && cfg.CurrentContext != "" {
			merged.CurrentContext = cfg.CurrentContext
		}
	}

	rewriteKubeconfigServerURLs(&merged)
	return yaml.Marshal(merged)
}

// rewriteKubeconfigServerURLs replaces 127.0.0.1/localhost in cluster server URLs
// with the Rancher host from RANCHER_URL, so kubectl inside the container can reach
// the API (Rancher proxies it) instead of connecting to the container's own loopback.
// Also sets insecure-skip-tls-verify for all clusters (Rancher cert often doesn't match "rancher" host).
func rewriteKubeconfigServerURLs(cfg *kubeConfig) {
	ru := rancherURL()
	rancherU, err := url.Parse(ru)
	if err != nil {
		return
	}
	rancherHost := rancherU.Host
	if rancherU.Port() == "" && rancherU.Scheme == "https" {
		rancherHost = rancherU.Hostname() + ":443"
	} else if rancherU.Port() == "" && rancherU.Scheme == "http" {
		rancherHost = rancherU.Hostname() + ":80"
	}
	for i := range cfg.Clusters {
		cluster := cfg.Clusters[i].Cluster
		server, _ := cluster["server"].(string)
		if server == "" {
			continue
		}
		su, err := url.Parse(server)
		if err != nil {
			continue
		}
		host := su.Hostname()
		if host == "127.0.0.1" || host == "localhost" || host == "::1" {
			su.Scheme = rancherU.Scheme
			su.Host = rancherHost
			cluster["server"] = su.String()
		}
		// Rancher cert is for localhost/rancher.cattle-system, not "rancher" — skip TLS verify for all
		cluster["insecure-skip-tls-verify"] = true
	}
}

func kubeConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kube", "config")
}

// detectCLIs returns which container/runtime CLIs are installed.
func detectCLIs() []string {
	clis := []string{"crictl", "runc", "etcdctl", "zellij", "ssh"}
	var found []string
	for _, name := range clis {
		if _, err := exec.LookPath(name); err == nil {
			found = append(found, name)
		}
	}
	return found
}

// fetchWelcome runs kk list in background and returns formatted output.
func fetchWelcome() string {
	var listOut string
	done := make(chan struct{})
	go func() {
		listOut, _ = runKrew("list")
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}

	clis := detectCLIs()

	var b strings.Builder
	b.WriteString("\r\n")
	b.WriteString("  ╔══════════════════════════════════════════════════════╗\r\n")
	b.WriteString("  ║              Krew Workstation                        ║\r\n")
	b.WriteString("  ║   k=kubectl   kk='kubectl krew'   tab completion ✓    ║\r\n")
	b.WriteString("  ╚══════════════════════════════════════════════════════╝\r\n\r\n")

	if len(clis) > 0 {
		b.WriteString("  CLIs: ")
		b.WriteString(strings.Join(clis, ", "))
		b.WriteString("\r\n")
		b.WriteString("  k ssh-jump — SSH to nodes via jump host\r\n\r\n")
	}

	if listOut != "" {
		plugins := parsePluginNames(listOut)
		if len(plugins) > 0 {
			b.WriteString("  Installed plugins: ")
			b.WriteString(strings.Join(plugins, ", "))
			b.WriteString("\r\n\r\n")
		}
	}

	b.WriteString("  Ready. Try: kk list | k9s | zellij | k ssh-jump\r\n\r\n")
	return b.String()
}

func parsePluginNames(listOut string) []string {
	var names []string
	for _, line := range strings.Split(listOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "PLUGIN") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			names = append(names, fields[0])
		}
	}
	return names
}

func runKubectlConfig(args ...string) (string, error) {
	kubeCfg := kubeConfigPath()
	cmd := exec.Command("kubectl", append([]string{"config"}, args...)...)
	cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeCfg)
	out, err := cmd.CombinedOutput()
	return string(out), err
}


func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, X-Rancher-Token")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.GET("/api/info", func(c *gin.Context) {
		hostname, _ := os.Hostname()
		c.JSON(200, gin.H{
			"baseImage":   "alpine:latest",
			"goVersion":  runtime.Version(),
			"hostname":   hostname,
			"krewRoot":   krewRoot(),
		})
	})

	// ── Rancher clusters (for the UI dropdown) ──

	tokenFromRequest := func(c *gin.Context) string {
		if h := c.GetHeader("Authorization"); strings.HasPrefix(h, "Bearer ") {
			return strings.TrimPrefix(h, "Bearer ")
		}
		return c.GetHeader("X-Rancher-Token")
	}

	r.GET("/api/clusters", func(c *gin.Context) {
		token := tokenFromRequest(c)
		clusters, err := fetchClustersWithToken(token)
		if err != nil {
			c.JSON(502, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, ClustersResponse{Clusters: clusters})
	})

	// ── Kubeconfig: sync from Rancher, get current context ──

	r.POST("/api/kubeconfig/sync", func(c *gin.Context) {
		token := tokenFromRequest(c)
		clusters, err := fetchClustersWithToken(token)
		if err != nil {
			c.JSON(502, gin.H{"error": err.Error()})
			return
		}
		if len(clusters) == 0 {
			c.JSON(200, gin.H{"message": "no clusters to sync", "clusters": 0})
			return
		}
		var configs []string
		for _, cl := range clusters {
			cfg, err := fetchKubeconfigWithToken(cl.ID, token)
			if err != nil {
				c.JSON(502, gin.H{"error": fmt.Sprintf("cluster %s: %v", cl.Name, err)})
				return
			}
			configs = append(configs, cfg)
		}
		merged, err := mergeKubeconfigs(configs)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		kubeDir := filepath.Dir(kubeConfigPath())
		if err := os.MkdirAll(kubeDir, 0700); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		if err := os.WriteFile(kubeConfigPath(), merged, 0600); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"message": "kubeconfig synced", "clusters": len(clusters)})
	})

	r.GET("/api/kubeconfig", func(c *gin.Context) {
		data, err := os.ReadFile(kubeConfigPath())
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(404, gin.H{"error": "kubeconfig not found; sync from Rancher first"})
				return
			}
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.Header("Content-Disposition", "attachment; filename=config")
		c.Data(200, "application/x-yaml", data)
	})

	r.GET("/api/context", func(c *gin.Context) {
		out, err := runKubectlConfig("current-context")
		ctx := strings.TrimSpace(out)
		if err != nil || ctx == "" {
			c.JSON(200, gin.H{"context": ""})
			return
		}
		c.JSON(200, gin.H{"context": ctx})
	})

	// ── Global plugin management (not per-cluster) ──

	r.GET("/api/plugins", func(c *gin.Context) {
		installedOutput, _ := runKrew("list")
		installed := parseInstalledPlugins(installedOutput)

		searchOutput, err := runKrew("search")
		if err != nil {
			c.JSON(500, PluginsResponse{Error: err.Error(), TerminalOutput: searchOutput})
			return
		}

		plugins := parseAvailablePlugins(searchOutput)
		for i := range plugins {
			if ver, ok := installed[plugins[i].Name]; ok {
				plugins[i].Installed = true
				plugins[i].Version = ver
			}
		}

		c.JSON(200, PluginsResponse{
			Plugins:        plugins,
			TerminalOutput: installedOutput + "\n" + searchOutput,
		})
	})

	r.POST("/api/plugins/:name/install", func(c *gin.Context) {
		name := c.Param("name")

		updateOut, _ := runKrew("update")
		installOut, err := runKrew("install", name)
		output := updateOut + "\n" + installOut
		if err != nil {
			c.JSON(500, PluginsResponse{Error: err.Error(), TerminalOutput: output})
			return
		}
		c.JSON(200, PluginsResponse{TerminalOutput: output})
	})

	r.DELETE("/api/plugins/:name", func(c *gin.Context) {
		name := c.Param("name")

		output, err := runKrew("uninstall", name)
		if err != nil {
			c.JSON(500, PluginsResponse{Error: err.Error(), TerminalOutput: output})
			return
		}
		c.JSON(200, PluginsResponse{TerminalOutput: output})
	})

	r.POST("/api/plugins/:name/upgrade", func(c *gin.Context) {
		name := c.Param("name")

		updateOut, _ := runKrew("update")
		upgradeOut, err := runKrew("upgrade", name)
		output := updateOut + "\n" + upgradeOut
		if err != nil {
			c.JSON(500, PluginsResponse{Error: err.Error(), TerminalOutput: output})
			return
		}
		c.JSON(200, PluginsResponse{TerminalOutput: output})
	})

	r.POST("/api/plugins/update", func(c *gin.Context) {
		output, err := runKrew("update")
		if err != nil {
			c.JSON(500, PluginsResponse{Error: err.Error(), TerminalOutput: output})
			return
		}
		c.JSON(200, PluginsResponse{TerminalOutput: output})
	})

	r.GET("/api/plugins/installed", func(c *gin.Context) {
		output, err := runKrew("list")
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		installed := parseInstalledPlugins(output)
		names := make([]string, 0, len(installed))
		for n := range installed {
			names = append(names, n)
		}
		c.JSON(200, gin.H{"plugins": names})
	})

	// ── WebSocket PTY shell (real bash session in the container) ──

	wsUpgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	r.GET("/api/ws/shell", func(c *gin.Context) {
		conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Fetch welcome in background (kk list, version, update, info) — blocks up to 12s
		welcome := fetchWelcome()

		shell := "/bin/bash"
		if _, err := os.Stat(shell); os.IsNotExist(err) {
			shell = "/bin/sh"
		}

		cmd := exec.Command(shell, "-i")
		cmd.Env = append(os.Environ(),
			"TERM=xterm-256color",
			"KREW_ROOT="+krewRoot(),
			fmt.Sprintf("PATH=%s:%s", filepath.Join(krewRoot(), "bin"), os.Getenv("PATH")),
		)

		ptmx, err := pty.Start(cmd)
		if err != nil {
			conn.WriteMessage(websocket.TextMessage, []byte("pty start failed: "+err.Error()+"\r\n"))
			return
		}
		defer ptmx.Close()

		// Send custom welcome (fetched in background)
		conn.WriteMessage(websocket.BinaryMessage, []byte(welcome))

		// Inject aliases
		ptmx.Write([]byte("alias k=kubectl; alias kk='kubectl krew'\n"))

		go func() {
			for {
				mt, data, err := conn.ReadMessage()
				if err != nil {
					return
				}
				if mt == websocket.TextMessage && len(data) > 8 && string(data[:7]) == `{"type"` {
					var msg struct {
						Type string `json:"type"`
						Cols int    `json:"cols"`
						Rows int    `json:"rows"`
					}
					if json.Unmarshal(data, &msg) == nil && msg.Type == "resize" && msg.Cols > 0 && msg.Rows > 0 {
						pty.Setsize(ptmx, &pty.Winsize{Cols: uint16(msg.Cols), Rows: uint16(msg.Rows)})
						continue
					}
				}
				if mt == websocket.BinaryMessage && len(data) > 0 {
					ptmx.Write(data)
				}
			}
		}()

		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				break
			}
			if n > 0 {
				if err := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); err != nil {
					break
				}
			}
		}
	})

	// ── Filesystem browser (safe, restricted paths) ──

	allowedDirs := map[string]bool{
		"/root": true, "/app": true, "/tmp": true,
	}

	r.GET("/api/fs", func(c *gin.Context) {
		rawPath := c.Query("path")
		if rawPath == "" {
			rawPath = "/root"
		}
		clean := filepath.Clean(rawPath)
		if !filepath.IsAbs(clean) {
			clean = filepath.Join("/root", clean)
		}
		if !allowedDirs[clean] {
			allowed := false
			for dir := range allowedDirs {
				if strings.HasPrefix(clean, dir+"/") || clean == dir {
					allowed = true
					break
				}
			}
			if !allowed {
				c.JSON(403, gin.H{"error": "path not allowed"})
				return
			}
		}
		entries, err := os.ReadDir(clean)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		type entry struct {
			Name    string `json:"name"`
			Dir     bool   `json:"dir"`
			Path    string `json:"path"`
			Size    int64  `json:"size"`
			Mode    string `json:"mode"`
			ModTime string `json:"modTime"`
		}
		var list []entry
		for _, e := range entries {
			info, err := e.Info()
			size := int64(0)
			mode := ""
			modTime := ""
			if err == nil && info != nil {
				size = info.Size()
				mode = info.Mode().String()
				modTime = info.ModTime().Format("2006-01-02 15:04")
			}
			if e.IsDir() {
				size = -1
			}
			list = append(list, entry{
				Name:    e.Name(),
				Dir:     e.IsDir(),
				Path:    filepath.Join(clean, e.Name()),
				Size:    size,
				Mode:    mode,
				ModTime: modTime,
			})
		}
		c.JSON(200, gin.H{"path": clean, "entries": list})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Printf("krew-manager listening on :%s  RANCHER_URL=%s\n", port, rancherURL())
	if err := r.Run(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start: %v\n", err)
		os.Exit(1)
	}
}
