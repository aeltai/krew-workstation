package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Plugin struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Installed   bool   `json:"installed"`
}

type Response struct {
	Plugins        []Plugin `json:"plugins,omitempty"`
	TerminalOutput string   `json:"terminalOutput,omitempty"`
	Error          string   `json:"error,omitempty"`
}

func createHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
		Timeout: 10 * time.Second,
	}
}

func execInKrewPod(command string) (string, error) {
	// Get the Krew manager pod name
	getPodCmd := exec.Command("kubectl", "get", "pods", "-l", "app=krew-manager", "-o", "jsonpath={.items[0].metadata.name}")
	podName, err := getPodCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get krew-manager pod: %v", err)
	}

	// Execute command in the pod
	execCmd := exec.Command("kubectl", "exec", "-i", string(podName), "--", "sh", "-c",
		fmt.Sprintf("export PATH=\"${KREW_ROOT:-$HOME/.krew}/bin:$PATH\" && %s", command))
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr
	err = execCmd.Run()

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\nErrors:\n" + stderr.String()
	}

	return output, err
}

func getInstalledPlugins() ([]Plugin, error) {
	output, err := execInKrewPod("kubectl krew list")
	if err != nil {
		return nil, err
	}

	var plugins []Plugin
	lines := strings.Split(output, "\n")
	for _, line := range lines[1:] { // Skip header line
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			plugins = append(plugins, Plugin{
				Name:      fields[0],
				Version:   fields[1],
				Installed: true,
			})
		}
	}
	return plugins, nil
}

func getAvailablePlugins() ([]Plugin, error) {
	output, err := execInKrewPod("kubectl krew search")
	if err != nil {
		return nil, err
	}

	var plugins []Plugin
	lines := strings.Split(output, "\n")
	for _, line := range lines[1:] { // Skip header line
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			description := strings.Join(fields[1:], " ")
			plugins = append(plugins, Plugin{
				Name:        fields[0],
				Description: description,
				Installed:   false,
			})
		}
	}
	return plugins, nil
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Enable CORS
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Authorization, Content-Type, x-rancher-url")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// List Krew plugins
	r.GET("/clusters/:clusterId/plugins", func(c *gin.Context) {
		installed, err := getInstalledPlugins()
		if err != nil {
			c.JSON(500, Response{
				Error: fmt.Sprintf("Failed to get installed plugins: %v", err),
			})
			return
		}

		available, err := getAvailablePlugins()
		if err != nil {
			c.JSON(500, Response{
				Error: fmt.Sprintf("Failed to get available plugins: %v", err),
			})
			return
		}

		// Create a map of installed plugins
		installedMap := make(map[string]string)
		for _, p := range installed {
			installedMap[p.Name] = p.Version
		}

		// Merge available and installed plugins
		var allPlugins []Plugin
		for _, p := range available {
			if version, ok := installedMap[p.Name]; ok {
				p.Installed = true
				p.Version = version
			}
			allPlugins = append(allPlugins, p)
		}

		c.JSON(200, Response{
			Plugins: allPlugins,
		})
	})

	// Install a plugin
	r.POST("/clusters/:clusterId/plugins/:pluginName/install", func(c *gin.Context) {
		pluginName := c.Param("pluginName")
		output, err := execInKrewPod(fmt.Sprintf("kubectl krew install %s", pluginName))
		if err != nil {
			c.JSON(500, Response{
				Error:          err.Error(),
				TerminalOutput: output,
			})
			return
		}

		c.JSON(200, Response{TerminalOutput: output})
	})

	// Uninstall a plugin
	r.DELETE("/clusters/:clusterId/plugins/:pluginName", func(c *gin.Context) {
		pluginName := c.Param("pluginName")
		output, err := execInKrewPod(fmt.Sprintf("kubectl krew uninstall %s", pluginName))
		if err != nil {
			c.JSON(500, Response{
				Error:          err.Error(),
				TerminalOutput: output,
			})
			return
		}

		c.JSON(200, Response{TerminalOutput: output})
	})

	// Upgrade a plugin
	r.POST("/clusters/:clusterId/plugins/:pluginName/upgrade", func(c *gin.Context) {
		pluginName := c.Param("pluginName")
		output, err := execInKrewPod(fmt.Sprintf("kubectl krew upgrade %s", pluginName))
		if err != nil {
			c.JSON(500, Response{
				Error:          err.Error(),
				TerminalOutput: output,
			})
			return
		}

		c.JSON(200, Response{TerminalOutput: output})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Starting server on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}
}
