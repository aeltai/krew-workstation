<template>
  <div class="krew-manager-page">
    <header>
      <h1>Kubectl Plugin Manager</h1>
      <div class="auth-controls">
        <input 
          type="text" 
          v-model="rancherUrl" 
          placeholder="Rancher URL"
          class="input-field"
        />
        <input 
          type="password" 
          v-model="apiKey" 
          placeholder="Rancher API Key"
          class="input-field"
        />
        <button 
          class="btn role-primary" 
          @click="authenticate"
          :disabled="!rancherUrl || !apiKey"
        >
          Connect to Rancher
        </button>
      </div>
    </header>

    <div class="content">
      <!-- Cluster Selection -->
      <div v-if="isRancherConnected" class="cluster-section">
        <select 
          v-model="selectedClusterId"
          class="cluster-select"
          @change="onClusterChange"
        >
          <option value="">Select a cluster</option>
          <option 
            v-for="cluster in clusters" 
            :key="cluster.id" 
            :value="cluster.id"
          >
            {{ cluster.name }}
          </option>
        </select>

        <!-- Kubeconfig Display -->
        <div v-if="kubeConfig" class="kubeconfig-section">
          <div class="section-header">
            <h3>Kubeconfig</h3>
            <button 
              class="btn role-secondary sm" 
              @click="copyKubeConfig"
            >
              Copy to Clipboard
            </button>
          </div>
          <pre class="kubeconfig-display">{{ kubeConfig }}</pre>
        </div>
      </div>

      <!-- Plugin List -->
      <div v-if="selectedClusterId" class="plugin-list">
        <div class="controls">
          <button 
            class="btn role-primary" 
            @click="refreshPlugins"
            :disabled="isLoading"
          >
            Refresh Plugins
          </button>
          <input 
            type="text" 
            v-model="searchQuery" 
            placeholder="Search plugins..."
            class="search-input"
          />
        </div>

        <div v-if="isLoading" class="loading">
          Loading plugins...
        </div>

        <!-- Terminal Output -->
        <div v-if="terminalOutput" class="terminal-output">
          <div class="terminal-header">
            <h3>Terminal Output</h3>
            <button @click="clearTerminal" class="btn role-secondary sm">Clear</button>
          </div>
          <pre>{{ terminalOutput }}</pre>
        </div>
        
        <table v-else>
          <thead>
            <tr>
              <th>Name</th>
              <th>Version</th>
              <th>Description</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="plugin in filteredPlugins" :key="plugin.name">
              <td>{{ plugin.name }}</td>
              <td>{{ plugin.version || 'N/A' }}</td>
              <td>{{ plugin.description }}</td>
              <td>
                <span :class="['status', plugin.installed ? 'installed' : 'not-installed']">
                  {{ plugin.installed ? 'Installed' : 'Not Installed' }}
                </span>
              </td>
              <td>
                <button 
                  class="btn role-secondary sm"
                  @click="togglePlugin(plugin)"
                  :disabled="isLoading"
                >
                  {{ plugin.installed ? 'Uninstall' : 'Install' }}
                </button>
                <button 
                  v-if="plugin.installed"
                  class="btn role-secondary sm"
                  @click="upgradePlugin(plugin)"
                  :disabled="isLoading"
                >
                  Upgrade
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<script>
import { KrewService } from './krewService';

export default {
  name: 'KrewManager',
  layout: 'plain',
  
  data() {
    return {
      rancherUrl: localStorage.getItem('rancher_url') || window.location.origin,
      apiKey: '',
      clusters: [],
      selectedClusterId: '',
      kubeConfig: '',
      plugins: [],
      searchQuery: '',
      isLoading: false,
      terminalOutput: '',
      krewService: new KrewService()
    };
  },

  computed: {
    isRancherConnected() {
      return this.rancherUrl && this.clusters.length > 0;
    },

    filteredPlugins() {
      if (!this.searchQuery) return this.plugins;
      const query = this.searchQuery.toLowerCase();
      return this.plugins.filter(plugin => 
        plugin.name.toLowerCase().includes(query) || 
        plugin.description.toLowerCase().includes(query)
      );
    }
  },

  async mounted() {
    await this.refreshPlugins();
  },

  methods: {
    async authenticate() {
      try {
        // Store Rancher URL for future use
        localStorage.setItem('rancher_url', this.rancherUrl);
        
        // Make API call directly to Rancher
        const response = await fetch(`${this.rancherUrl}/v3/clusters`, {
          headers: {
            'Authorization': `Bearer ${this.apiKey}`,
          }
        });

        if (!response.ok) {
          throw new Error(`Failed to authenticate: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();
        this.clusters = data.data.map(cluster => ({
          id: cluster.id,
          name: cluster.name
        }));

        this.apiKey = ''; // Clear API key from memory
      } catch (error) {
        console.error('Authentication failed:', error);
        alert('Failed to authenticate with Rancher. Please check your credentials and ensure you have access to the cluster.');
      }
    },

    async onClusterChange() {
      if (this.selectedClusterId) {
        try {
          const response = await fetch(
            `${this.rancherUrl}/v3/clusters/${this.selectedClusterId}?action=generateKubeconfig`,
            {
              method: 'POST',
              headers: {
                'Authorization': `Bearer ${this.apiKey}`
              }
            }
          );

          if (response.ok) {
            const data = await response.json();
            this.kubeConfig = data.config;
          } else {
            throw new Error(`Failed to get kubeconfig: ${response.status} ${response.statusText}`);
          }
        } catch (error) {
          console.error('Error fetching kubeconfig:', error);
          alert('Failed to fetch kubeconfig. Please check your permissions and try again.');
        }
      }
    },

    async copyKubeConfig() {
      try {
        await navigator.clipboard.writeText(this.kubeConfig);
        alert('Kubeconfig copied to clipboard!');
      } catch (error) {
        console.error('Failed to copy:', error);
        alert('Failed to copy kubeconfig. Please try again.');
      }
    },

    clearTerminal() {
      this.terminalOutput = '';
    },

    async refreshPlugins() {
      if (!this.selectedClusterId) return;

      this.isLoading = true;
      try {
        const response = await fetch(
          `https://krew-manager-backend:9000/clusters/${this.selectedClusterId}/plugins`,
          {
            headers: {
              'Authorization': `Bearer ${this.apiKey}`
            }
          }
        );

        if (!response.ok) {
          throw new Error('Failed to fetch plugins');
        }

        const data = await response.json();
        this.plugins = data.plugins;
        this.terminalOutput = data.terminalOutput;
      } catch (error) {
        console.error('Error fetching plugins:', error);
        alert('Failed to fetch plugins. Please try again.');
      } finally {
        this.isLoading = false;
      }
    },

    async togglePlugin(plugin) {
      this.isLoading = true;
      try {
        const method = plugin.installed ? 'DELETE' : 'POST';
        const endpoint = plugin.installed 
          ? `clusters/${this.selectedClusterId}/plugins/${plugin.name}`
          : `clusters/${this.selectedClusterId}/plugins/${plugin.name}/install`;

        const response = await fetch(
          `https://krew-manager-backend:9000/${endpoint}`,
          {
            method,
            headers: {
              'Authorization': `Bearer ${this.apiKey}`
            }
          }
        );

        if (!response.ok) {
          throw new Error(`Failed to ${plugin.installed ? 'uninstall' : 'install'} plugin`);
        }

        const data = await response.json();
        this.terminalOutput = data.terminalOutput;
        await this.refreshPlugins();
      } catch (error) {
        console.error('Error toggling plugin:', error);
        alert('Failed to toggle plugin. Please try again.');
      } finally {
        this.isLoading = false;
      }
    },

    async upgradePlugin(plugin) {
      this.isLoading = true;
      try {
        const response = await fetch(
          `https://krew-manager-backend:9000/clusters/${this.selectedClusterId}/plugins/${plugin.name}/upgrade`,
          {
            method: 'POST',
            headers: {
              'Authorization': `Bearer ${this.apiKey}`
            }
          }
        );

        if (!response.ok) {
          throw new Error('Failed to upgrade plugin');
        }

        const data = await response.json();
        this.terminalOutput = data.terminalOutput;
        await this.refreshPlugins();
      } catch (error) {
        console.error('Error upgrading plugin:', error);
        alert('Failed to upgrade plugin. Please try again.');
      } finally {
        this.isLoading = false;
      }
    }
  }
};
</script>

<style lang="scss" scoped>
.krew-manager-page {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;

  header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;

    h1 {
      color: var(--primary);
      margin: 0;
      font-size: 2em;
    }

    .auth-controls {
      display: flex;
      gap: 10px;
      align-items: center;

      .input-field {
        padding: 8px;
        border: 1px solid var(--border);
        border-radius: 4px;
        width: 200px;
      }
    }
  }

  .content {
    background: var(--box-bg);
    border-radius: 4px;
    padding: 20px;

    .cluster-section {
      margin-bottom: 20px;

      .cluster-select {
        width: 100%;
        padding: 8px;
        border: 1px solid var(--border);
        border-radius: 4px;
        margin-bottom: 15px;
      }
    }

    .kubeconfig-section {
      margin-bottom: 20px;

      .section-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 10px;

        h3 {
          margin: 0;
          color: var(--primary);
        }
      }

      .kubeconfig-display {
        background: var(--body-bg);
        padding: 15px;
        border-radius: 4px;
        overflow-x: auto;
        font-family: monospace;
        font-size: 0.9em;
      }
    }
  }

  .plugin-list {
    .controls {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;

      .search-input {
        padding: 8px;
        border: 1px solid var(--border);
        border-radius: 4px;
        width: 300px;
      }
    }

    table {
      width: 100%;
      border-collapse: collapse;

      th, td {
        padding: 12px;
        text-align: left;
        border-bottom: 1px solid var(--border);
      }

      th {
        font-weight: bold;
        color: var(--primary);
      }

      .status {
        padding: 4px 8px;
        border-radius: 4px;
        font-size: 0.9em;

        &.installed {
          background: var(--success);
          color: white;
        }

        &.not-installed {
          background: var(--warning);
          color: white;
        }
      }

      .btn {
        margin-right: 8px;

        &:last-child {
          margin-right: 0;
        }
      }
    }
  }

  .terminal-output {
    background: var(--body-bg);
    border-radius: 4px;
    margin: 20px 0;
    overflow: hidden;

    .terminal-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 10px 15px;
      background: var(--primary);
      color: white;

      h3 {
        margin: 0;
        font-size: 1.1em;
      }
    }

    pre {
      margin: 0;
      padding: 15px;
      font-family: monospace;
      white-space: pre-wrap;
      max-height: 300px;
      overflow-y: auto;
    }
  }
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
  font-style: italic;
  color: var(--secondary);
}

.btn {
  &.sm {
    padding: 4px 8px;
    font-size: 0.9em;
  }
}
</style> 