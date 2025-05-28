import { execSync } from 'child_process';

interface KrewPlugin {
  name: string;
  version: string;
  description: string;
  installed: boolean;
  repository?: string;
}

interface Cluster {
  id: string;
  name: string;
  kubeConfig: string;
}

export class KrewService {
  private readonly KREW_INDEX_URL = 'https://krew.sigs.k8s.io/plugins/';
  private readonly KREW_API_URL = 'https://api.github.com/repos/kubernetes-sigs/krew-index/contents/plugins';
  private readonly BACKEND_API = 'http://localhost:3000/v1/krew'; // Updated backend URL
  private apiKey: string = '';

  constructor() {
    this.apiKey = localStorage.getItem('krew_github_token') || '';
  }

  setApiKey(key: string) {
    this.apiKey = key;
    localStorage.setItem('krew_github_token', key);
  }

  private async fetchWithAuth(url: string): Promise<Response> {
    const headers: HeadersInit = {
      'Accept': 'application/vnd.github.v3+json'
    };
    
    if (this.apiKey) {
      headers['Authorization'] = `token ${this.apiKey}`;
    }

    return fetch(url, { headers });
  }

  private async backendRequest(endpoint: string, method: string = 'GET', body?: any): Promise<any> {
    try {
      const response = await fetch(`${this.BACKEND_API}${endpoint}`, {
        method,
        headers: {
          'Content-Type': 'application/json'
        },
        body: body ? JSON.stringify(body) : undefined
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('Backend request failed:', error);
      throw error;
    }
  }

  async listPlugins(): Promise<KrewPlugin[]> {
    try {
      // Fetch plugin list from GitHub API
      const response = await this.fetchWithAuth(this.KREW_API_URL);
      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const files = await response.json();
      const plugins: KrewPlugin[] = [];

      // Process each plugin manifest
      for (const file of files) {
        if (file.name.endsWith('.yaml')) {
          const manifestResponse = await this.fetchWithAuth(file.download_url);
          if (manifestResponse.ok) {
            const manifest = await manifestResponse.text();
            try {
              // Basic YAML parsing for plugin info
              const name = file.name.replace('.yaml', '');
              const description = manifest.match(/description:\s*(.+)/)?.[1] || '';
              const version = manifest.match(/version:\s*(.+)/)?.[1] || '';
              const repository = manifest.match(/homepage:\s*(.+)/)?.[1] || '';
              
              plugins.push({
                name,
                version,
                description,
                repository,
                installed: false // We'll check installation status separately
              });
            } catch (e) {
              console.error(`Error parsing manifest for ${file.name}:`, e);
            }
          }
        }
      }

      // Check installed plugins from backend
      try {
        const installedPlugins = await this.backendRequest('/installed');
        plugins.forEach(plugin => {
          plugin.installed = installedPlugins.includes(plugin.name);
        });
      } catch (e) {
        console.warn('Could not check installed plugins:', e);
      }

      return plugins;
    } catch (error) {
      console.error('Error listing plugins:', error);
      return [];
    }
  }

  async installPlugin(name: string): Promise<boolean> {
    try {
      await this.backendRequest(`/plugins/${name}/install`, 'POST');
      return true;
    } catch (error) {
      console.error('Error installing plugin:', error);
      return false;
    }
  }

  async uninstallPlugin(name: string): Promise<boolean> {
    try {
      await this.backendRequest(`/plugins/${name}/uninstall`, 'DELETE');
      return true;
    } catch (error) {
      console.error('Error uninstalling plugin:', error);
      return false;
    }
  }

  async upgradePlugin(name: string): Promise<boolean> {
    try {
      await this.backendRequest(`/plugins/${name}/upgrade`, 'POST');
      return true;
    } catch (error) {
      console.error('Error upgrading plugin:', error);
      return false;
    }
  }

  isAuthenticated(): boolean {
    return true; // No authentication required for basic functionality
  }
} 