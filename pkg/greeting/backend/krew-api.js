const { execSync } = require('child_process');
const express = require('express');
const router = express.Router();
const fs = require('fs');
const https = require('https');
const fetch = require('node-fetch');

// Create HTTPS agent that ignores self-signed certificates
const httpsAgent = new https.Agent({
  rejectUnauthorized: false
});

// Helper function to run Krew command with specific kubeconfig
function runKrewCommand(command, clusterId) {
  try {
    // Ensure PATH includes Krew binaries
    const path = process.env.PATH || '';
    const home = process.env.HOME || process.env.USERPROFILE;
    const krewPath = `${home}/.krew/bin`;
    process.env.PATH = `${path}:${krewPath}`;

    // Get kubeconfig for the cluster
    const kubeconfig = getClusterKubeconfig(clusterId);
    
    // Run the command with specific kubeconfig
    const result = execSync(command, { 
      encoding: 'utf8',
      env: {
        ...process.env,
        KUBECONFIG: kubeconfig
      }
    });

    return {
      success: true,
      output: result,
      command: command
    };
  } catch (error) {
    console.error(`Error running command: ${command}`, error);
    return {
      success: false,
      error: error.message,
      command: command
    };
  }
}

// Helper to get kubeconfig for a specific cluster
async function getClusterKubeconfig(clusterId) {
  const rancherUrl = process.env.RANCHER_URL || 'https://rancher:443';
  const token = process.env.RANCHER_TOKEN;

  try {
    console.log(`Fetching kubeconfig from ${rancherUrl} for cluster ${clusterId}`);
    const response = await fetch(
      `${rancherUrl}/v3/clusters/${clusterId}?action=generateKubeconfig`,
      {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Accept': 'application/json'
        },
        agent: httpsAgent // Use the HTTPS agent that ignores self-signed certs
      }
    );

    if (!response.ok) {
      const error = await response.text();
      console.error('Rancher API error:', error);
      throw new Error(`Failed to get kubeconfig: ${response.status} ${response.statusText}`);
    }

    const data = await response.json();
    const kubeconfigPath = `/app/kubeconfig/${clusterId}`;
    
    // Ensure directory exists
    fs.mkdirSync('/app/kubeconfig', { recursive: true });
    
    // Save kubeconfig to file
    fs.writeFileSync(kubeconfigPath, data.config);
    console.log(`Saved kubeconfig to ${kubeconfigPath}`);
    return kubeconfigPath;
  } catch (error) {
    console.error('Error getting kubeconfig:', error);
    throw error;
  }
}

// Test connection endpoint
router.get('/test', async (req, res) => {
  try {
    const response = await fetch(`${process.env.RANCHER_URL}/v3/clusters`, {
      headers: {
        'Authorization': `Bearer ${process.env.RANCHER_TOKEN}`,
        'Accept': 'application/json'
      },
      agent: httpsAgent
    });

    const data = await response.json();
    res.json({
      status: 'ok',
      rancher_url: process.env.RANCHER_URL,
      clusters: data
    });
  } catch (error) {
    res.status(500).json({
      status: 'error',
      message: error.message,
      stack: error.stack
    });
  }
});

// List installed plugins for a specific cluster
router.get('/clusters/:clusterId/plugins', async (req, res) => {
  try {
    const { clusterId } = req.params;
    const result = await runKrewCommand('kubectl krew list', clusterId);
    
    if (!result.success) {
      throw new Error(result.error);
    }

    const installedPlugins = result.output
      .split('\n')
      .filter(line => line.trim())
      .map(line => {
        const [name, version] = line.split(/\s+/);
        return { name, version, installed: true };
      });

    // Get available plugins
    const availableResult = await runKrewCommand('kubectl krew search', clusterId);
    const availablePlugins = availableResult.success ? 
      availableResult.output
        .split('\n')
        .slice(1) // Skip header
        .filter(line => line.trim())
        .map(line => {
          const [name, description] = line.split(/\s{2,}/);
          return {
            name,
            description,
            installed: installedPlugins.some(p => p.name === name)
          };
        }) : [];

    res.json({
      plugins: availablePlugins,
      terminalOutput: result.output
    });
  } catch (error) {
    console.error('Error listing plugins:', error);
    res.status(500).json({ 
      error: 'Failed to list plugins', 
      details: error.message,
      stack: error.stack
    });
  }
});

// Install a plugin for a specific cluster
router.post('/clusters/:clusterId/plugins/:name/install', async (req, res) => {
  const { clusterId, name } = req.params;
  try {
    // First update the index
    const updateResult = runKrewCommand('kubectl krew update', clusterId);
    if (!updateResult.success) {
      throw new Error(updateResult.error);
    }

    // Then install the plugin
    const installResult = runKrewCommand(`kubectl krew install ${name}`, clusterId);
    if (!installResult.success) {
      throw new Error(installResult.error);
    }

    res.json({
      success: true,
      terminalOutput: installResult.output
    });
  } catch (error) {
    console.error(`Error installing plugin ${name}:`, error);
    res.status(500).json({ 
      error: `Failed to install plugin ${name}`, 
      details: error.message 
    });
  }
});

// Uninstall a plugin from a specific cluster
router.delete('/clusters/:clusterId/plugins/:name', async (req, res) => {
  const { clusterId, name } = req.params;
  try {
    const result = runKrewCommand(`kubectl krew uninstall ${name}`, clusterId);
    if (!result.success) {
      throw new Error(result.error);
    }

    res.json({
      success: true,
      terminalOutput: result.output
    });
  } catch (error) {
    console.error(`Error uninstalling plugin ${name}:`, error);
    res.status(500).json({ 
      error: `Failed to uninstall plugin ${name}`, 
      details: error.message 
    });
  }
});

// Upgrade a plugin for a specific cluster
router.post('/clusters/:clusterId/plugins/:name/upgrade', async (req, res) => {
  const { clusterId, name } = req.params;
  try {
    // First update the index
    const updateResult = runKrewCommand('kubectl krew update', clusterId);
    if (!updateResult.success) {
      throw new Error(updateResult.error);
    }

    // Then upgrade the plugin
    const upgradeResult = runKrewCommand(`kubectl krew upgrade ${name}`, clusterId);
    if (!upgradeResult.success) {
      throw new Error(upgradeResult.error);
    }

    res.json({
      success: true,
      terminalOutput: upgradeResult.output
    });
  } catch (error) {
    console.error(`Error upgrading plugin ${name}:`, error);
    res.status(500).json({ 
      error: `Failed to upgrade plugin ${name}`, 
      details: error.message 
    });
  }
});

module.exports = router; 