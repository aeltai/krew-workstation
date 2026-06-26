package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	backupOperatorNS    = "cattle-resources-system"
	backupPodLabel      = "app.kubernetes.io/name=rancher-backup"
	polymorphConfigPath = "/root/.config/rancher-polymorph/rancher-polymorph.yaml"
)

type BackupOperatorStatus struct {
	Installed   bool   `json:"installed"`
	Namespace   string `json:"namespace"`
	NamespaceOK bool   `json:"namespaceOk"`
	Deployment  string `json:"deployment,omitempty"`
	Replicas    string `json:"replicas,omitempty"`
	PodName     string `json:"podName,omitempty"`
	PodPhase    string `json:"podPhase,omitempty"`
	PodReady    bool   `json:"podReady"`
	Message     string `json:"message,omitempty"`
	Context     string `json:"context,omitempty"`
}

type BackupResource struct {
	Name      string `json:"name"`
	Created   string `json:"created,omitempty"`
	Ready     string `json:"ready,omitempty"`
	Message   string `json:"message,omitempty"`
	Storage   string `json:"storage,omitempty"`
	Filename  string `json:"filename,omitempty"`
}

type RestoreResource struct {
	Name           string `json:"name"`
	Created        string `json:"created,omitempty"`
	Ready          string `json:"ready,omitempty"`
	Message        string `json:"message,omitempty"`
	BackupFilename string `json:"backupFilename,omitempty"`
	Prune          bool   `json:"prune"`
}

type BackupsStatusResponse struct {
	Operator BackupOperatorStatus `json:"operator"`
	Backups  []BackupResource     `json:"backups"`
	Restores []RestoreResource    `json:"restores"`
	Error    string               `json:"error,omitempty"`
}

func runKubectlWithConfig(kubeCfg string, args ...string) (string, error) {
	if kubeCfg == "" {
		kubeCfg = kubeConfigPath()
	}
	cmd := exec.Command("kubectl", args...)
	cmd.Env = append(os.Environ(), "KUBECONFIG="+kubeCfg)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	output := strings.TrimSpace(stdout.String())
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = output
		}
		if output == "" {
			return "", fmt.Errorf("%w: %s", err, msg)
		}
		return output, fmt.Errorf("%w: %s", err, msg)
	}
	return output, nil
}

func runKubectl(args ...string) (string, error) {
	return runKubectlWithConfig(kubeConfigPath(), args...)
}

func kubectlJSONWithConfig(kubeCfg string, args ...string) (map[string]interface{}, error) {
	jsonArgs := append(args, "-o", "json")
	out, err := runKubectlWithConfig(kubeCfg, jsonArgs...)
	if err != nil {
		return nil, err
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(out), &obj); err != nil {
		return nil, fmt.Errorf("parse kubectl json: %w", err)
	}
	return obj, nil
}

func kubectlJSON(args ...string) (map[string]interface{}, error) {
	return kubectlJSONWithConfig(kubeConfigPath(), args...)
}

func currentKubeContextForConfig(kubeCfg string) string {
	out, err := runKubectlConfigPath(kubeCfg, "current-context")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

func currentKubeContext() string {
	return currentKubeContextForConfig(kubeConfigPath())
}

func fetchBackupOperatorStatusForKubeconfig(kubeCfg string) BackupOperatorStatus {
	status := BackupOperatorStatus{
		Namespace: backupOperatorNS,
		Context:   currentKubeContextForConfig(kubeCfg),
	}
	if kubeCfg == "" {
		status.Message = "kubeconfig not synced"
		return status
	}
	if _, err := os.Stat(kubeCfg); err != nil {
		status.Message = "kubeconfig not synced for this user"
		return status
	}

	_, err := runKubectlWithConfig(kubeCfg, "get", "namespace", backupOperatorNS)
	if err != nil {
		status.Message = "Rancher Backup operator not installed — install the Rancher Backup app on the local cluster"
		return status
	}
	status.NamespaceOK = true

	deployOut, err := runKubectlWithConfig(kubeCfg, "get", "deployment", "-n", backupOperatorNS,
		"-l", backupPodLabel,
		"-o", "jsonpath={.items[0].metadata.name},{.items[0].status.readyReplicas}/{.items[0].spec.replicas}")
	if err == nil && deployOut != "" {
		parts := strings.SplitN(deployOut, ",", 2)
		if len(parts) >= 1 {
			status.Deployment = parts[0]
		}
		if len(parts) >= 2 {
			status.Replicas = parts[1]
		}
	}

	podObj, err := kubectlJSONWithConfig(kubeCfg, "get", "pods", "-n", backupOperatorNS, "-l", backupPodLabel)
	if err == nil {
		items, _ := podObj["items"].([]interface{})
		if len(items) > 0 {
			pod, _ := items[0].(map[string]interface{})
			meta, _ := pod["metadata"].(map[string]interface{})
			podStatus, _ := pod["status"].(map[string]interface{})
			if name, _ := meta["name"].(string); name != "" {
				status.PodName = name
			}
			if phase, _ := podStatus["phase"].(string); phase != "" {
				status.PodPhase = phase
			}
			if conds, ok := podStatus["conditions"].([]interface{}); ok {
				for _, c := range conds {
					cond, _ := c.(map[string]interface{})
					if cond["type"] == "Ready" && cond["status"] == "True" {
						status.PodReady = true
					}
				}
			}
		}
	}

	status.Installed = status.NamespaceOK && status.PodReady
	if status.Installed {
		status.Message = "Connected to rancher-backup operator"
	} else if status.NamespaceOK {
		status.Message = "Namespace exists but operator pod is not ready yet"
	}
	return status
}

func conditionStatus(conditions []interface{}, condType string) (status, message string) {
	for _, c := range conditions {
		cond, _ := c.(map[string]interface{})
		if cond["type"] != condType {
			continue
		}
		if s, _ := cond["status"].(string); s != "" {
			status = s
		}
		if m, _ := cond["message"].(string); m != "" {
			message = m
		}
	}
	return status, message
}

func fetchBackupCRsForKubeconfig(kubeCfg string) ([]BackupResource, error) {
	obj, err := kubectlJSONWithConfig(kubeCfg, "get", "backups", "-n", backupOperatorNS)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "no matches") {
			return nil, nil
		}
		return nil, err
	}
	items, _ := obj["items"].([]interface{})
	var backups []BackupResource
	for _, item := range items {
		cr, _ := item.(map[string]interface{})
		meta, _ := cr["metadata"].(map[string]interface{})
		spec, _ := cr["spec"].(map[string]interface{})
		st, _ := cr["status"].(map[string]interface{})

		b := BackupResource{Name: stringField(meta, "name")}
		if ts := stringField(meta, "creationTimestamp"); ts != "" {
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				b.Created = t.Format("2006-01-02 15:04")
			} else {
				b.Created = ts
			}
		}
		if conds, ok := st["conditions"].([]interface{}); ok {
			b.Ready, b.Message = conditionStatus(conds, "Ready")
		}
		if loc, ok := spec["storageLocation"].(map[string]interface{}); ok {
			if s3, ok := loc["s3"].(map[string]interface{}); ok {
				b.Storage = "s3:" + stringField(s3, "bucketName")
			}
		}
		if b.Storage == "" {
			b.Storage = "local"
		}
		backups = append(backups, b)
	}
	return backups, nil
}

func fetchRestoreCRsForKubeconfig(kubeCfg string) ([]RestoreResource, error) {
	obj, err := kubectlJSONWithConfig(kubeCfg, "get", "restores", "-n", backupOperatorNS)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "no matches") {
			return nil, nil
		}
		return nil, err
	}
	items, _ := obj["items"].([]interface{})
	var restores []RestoreResource
	for _, item := range items {
		cr, _ := item.(map[string]interface{})
		meta, _ := cr["metadata"].(map[string]interface{})
		spec, _ := cr["spec"].(map[string]interface{})
		st, _ := cr["status"].(map[string]interface{})

		r := RestoreResource{
			Name:           stringField(meta, "name"),
			BackupFilename: stringField(spec, "backupFilename"),
			Prune:          boolField(spec, "prune"),
		}
		if ts := stringField(meta, "creationTimestamp"); ts != "" {
			if t, err := time.Parse(time.RFC3339, ts); err == nil {
				r.Created = t.Format("2006-01-02 15:04")
			} else {
				r.Created = ts
			}
		}
		if conds, ok := st["conditions"].([]interface{}); ok {
			r.Ready, r.Message = conditionStatus(conds, "Ready")
		}
		restores = append(restores, r)
	}
	return restores, nil
}

func fetchBackupsStatusForKubeconfig(kubeCfg string) BackupsStatusResponse {
	resp := BackupsStatusResponse{
		Operator: fetchBackupOperatorStatusForKubeconfig(kubeCfg),
		Backups:  []BackupResource{},
		Restores: []RestoreResource{},
	}
	if !resp.Operator.NamespaceOK {
		return resp
	}
	if backups, err := fetchBackupCRsForKubeconfig(kubeCfg); err != nil {
		resp.Error = err.Error()
	} else if backups != nil {
		resp.Backups = backups
	}
	if restores, err := fetchRestoreCRsForKubeconfig(kubeCfg); err != nil {
		if resp.Error == "" {
			resp.Error = err.Error()
		}
	} else if restores != nil {
		resp.Restores = restores
	}
	return resp
}

func fetchBackupOperatorStatus() BackupOperatorStatus {
	return fetchBackupOperatorStatusForKubeconfig(kubeConfigPath())
}

func fetchBackupsStatus() BackupsStatusResponse {
	return fetchBackupsStatusForKubeconfig(kubeConfigPath())
}

func stringField(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func boolField(m map[string]interface{}, key string) bool {
	if m == nil {
		return false
	}
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

type polymorphConfig struct {
	Restore struct {
		Kubeconfig         string `yaml:"kubeconfig"`
		Context            string `yaml:"context"`
		Namespace          string `yaml:"namespace"`
		OperatorNamespace  string `yaml:"operator_namespace"`
		BackupPodLabel     string `yaml:"backup_pod_label"`
		BackupContainerPath string `yaml:"backup_container_path"`
		RestoreName        string `yaml:"restore_name"`
		WatchTimeout       string `yaml:"watch_timeout"`
	} `yaml:"restore"`
	Defaults struct {
		OutputDir string `yaml:"output_dir"`
	} `yaml:"defaults"`
}

func ensurePolymorphConfigForKubeconfig(kubeCfg string) error {
	ctx := currentKubeContextForConfig(kubeCfg)
	if ctx == "" {
		ctx = "local"
	}
	cfg := polymorphConfig{}
	cfg.Restore.Kubeconfig = kubeCfg
	cfg.Restore.Context = ctx
	cfg.Restore.Namespace = backupOperatorNS
	cfg.Restore.OperatorNamespace = backupOperatorNS
	cfg.Restore.BackupPodLabel = backupPodLabel
	cfg.Restore.BackupContainerPath = "/var/lib/rancher-backup"
	cfg.Restore.RestoreName = "rancher-restore"
	cfg.Restore.WatchTimeout = "30m"
	cfg.Defaults.OutputDir = "/root/backups"

	dir := filepath.Dir(polymorphConfigPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(polymorphConfigPath, data, 0600)
}

func ensurePolymorphConfig() error {
	return ensurePolymorphConfigForKubeconfig(kubeConfigPath())
}
