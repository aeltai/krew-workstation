<template>
  <div class="krew-page" :class="themeClass" :style="themeVars">
    <div v-show="activeTab !== 'terminal'" v-if="clusters.length" class="cluster-info">
      Clusters:
      <span v-for="c in clusters" :key="c.id" :class="['cluster-badge', c.state]">{{ c.name }}</span>
    </div>

    <div v-if="error" class="banner error">
      {{ error }}
      <button class="dismiss" @click="error = ''">&times;</button>
    </div>
    <div v-if="message" class="banner success">
      {{ message }}
      <button class="dismiss" @click="message = ''">&times;</button>
    </div>

    <div v-if="loading" class="loading-bar" />

    <!-- Tabs: Plugins | Files | Terminal -->
    <div class="panels">
      <div v-show="activeTab === 'plugins'" class="frame-toolbar">
        <div class="toolbar-left">
          <button class="btn role-primary xs" :disabled="loading" @click="loadPlugins">
            <i class="icon icon-refresh" /> Refresh
          </button>
          <button class="btn role-secondary xs" :disabled="loading" @click="updateIndex">Update index</button>
        </div>
        <label class="search-label">Search plugins</label>
        <input v-model="search" type="text" class="search-input" placeholder="by name or description…" />
      </div>
      <div class="tabs">
        <button :class="{ active: activeTab === 'terminal' }" @click="activeTab = 'terminal'">Terminal</button>
        <button :class="{ active: activeTab === 'plugins' }" @click="activeTab = 'plugins'">Plugins</button>
        <button :class="{ active: activeTab === 'files' }" @click="activeTab = 'files'">Files</button>
      </div>

      <div v-show="activeTab === 'plugins'" class="panel plugins-panel">
        <table v-if="paginatedPlugins.length" class="plugin-table">
          <thead>
            <tr>
              <th>Plugin</th>
              <th>Version</th>
              <th>Description</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in paginatedPlugins" :key="p.name">
              <td class="name">{{ p.name }}</td>
              <td>{{ p.version || '-' }}</td>
              <td class="desc">{{ p.description }}</td>
              <td>
                <span :class="['badge', p.installed ? 'installed' : 'available']">
                  {{ p.installed ? 'Installed' : 'Available' }}
                </span>
              </td>
              <td class="actions">
                <button v-if="!p.installed" class="btn role-primary sm" :disabled="busy === p.name" @click="installPlugin(p)">Install</button>
                <button v-if="p.installed" class="btn role-secondary sm" :disabled="busy === p.name" @click="upgradePlugin(p)">Upgrade</button>
                <button v-if="p.installed" class="btn role-tertiary sm" :disabled="busy === p.name" @click="uninstallPlugin(p)">Uninstall</button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-if="paginatedPlugins.length" class="pagination">
          <span class="pagination-info">{{ pluginPageStart }}-{{ pluginPageEnd }} of {{ filteredPlugins.length }}</span>
          <button class="btn role-tertiary sm" :disabled="pluginPage <= 1" @click="pluginPage = Math.max(1, pluginPage - 1)">Prev</button>
          <button class="btn role-tertiary sm" :disabled="pluginPage >= pluginPageCount" @click="pluginPage = Math.min(pluginPageCount, pluginPage + 1)">Next</button>
        </div>
        <div v-else-if="!loading" class="empty-state">
          <p v-if="search">No plugins matching "{{ search }}".</p>
          <p v-else>Click <strong>Refresh plugins</strong> to load the list.</p>
        </div>
      </div>

      <div v-show="activeTab === 'files'" class="panel files-panel">
        <div class="files-toolbar">
          <button class="btn role-secondary sm" :disabled="fsPath === '/root'" @click="fsNavigate('/root')">Root</button>
          <button class="btn role-secondary sm" :disabled="!fsPath || fsPath === '/root'" @click="fsNavigate(parentPath)">↑ Up</button>
          <span class="fs-path">{{ fsPath }}</span>
        </div>
        <div v-if="fsError" class="fs-error">{{ fsError }}</div>
        <div v-else class="fs-tree">
          <div class="fs-header">
            <span class="fs-col-name">Name</span>
            <span class="fs-col-size">Size</span>
            <span class="fs-col-mode">Mode</span>
            <span class="fs-col-date">Modified</span>
          </div>
          <div
            v-for="e in fsEntries"
            :key="e.path"
            :class="['fs-row', e.dir ? 'dir' : 'file']"
            @click="e.dir ? fsNavigate(e.path) : null"
          >
            <span class="fs-col-name">
              <span :class="['fs-icon', e.dir ? 'dir' : 'file']">{{ e.dir ? '📁' : '📄' }}</span>
              {{ e.name }}
            </span>
            <span class="fs-col-size">{{ formatSize(e.size) }}</span>
            <span class="fs-col-mode">{{ e.mode }}</span>
            <span class="fs-col-date">{{ e.modTime }}</span>
          </div>
        </div>
      </div>

      <div v-show="activeTab === 'terminal'" class="panel terminal-panel">
        <div class="shell-header">
          <div class="shell-header-left">
            <span v-if="containerInfo" class="shell-meta">{{ containerInfo.baseImage }} · {{ containerInfo.goVersion }} · {{ containerInfo.hostname || '' }}{{ clusters.length ? ' · ' + clusters.length + ' cluster' + (clusters.length !== 1 ? 's' : '') : '' }}</span>
            <span v-if="currentContext" class="shell-context" :title="'Current kubectl context'">{{ currentContext }}</span>
            <button class="btn role-tertiary xs theme-btn" :title="darkMode ? 'Switch to light mode' : 'Switch to dark mode'" @click="toggleTheme">{{ darkMode ? '☀' : '☽' }}</button>
            <button class="btn role-tertiary xs about-btn" :class="{ active: showAboutKrew }" title="About Krew" @click="showAboutKrew = !showAboutKrew">?</button>
            <div v-if="showAboutKrew" class="about-krew-card">
              <h3>What is Krew?</h3>
              <p>Plugin manager for kubectl. <a href="https://krew.sigs.k8s.io" target="_blank" rel="noopener">krew.sigs.k8s.io</a></p>
            </div>
          </div>
          <div class="shell-header-right">
            <span v-if="shellConnected" class="shell-status connected">Connected</span>
            <span v-else class="shell-status">Disconnected</span>
            <button
              v-if="terminalReady"
              class="btn role-secondary sm connect-btn"
              :class="{ 'role-primary': shellConnected }"
              @click="shellConnected ? disconnectShell() : connectShell()"
            >
              {{ shellConnected ? 'Disconnect' : 'Connect' }}
            </button>
            <button
              class="btn role-tertiary sm cheatsheet-btn"
              :class="{ active: showCheatsheet }"
              title="Cheatsheet"
              @click="showCheatsheet = !showCheatsheet"
            >
              Cheatsheet
            </button>
          </div>
          <div v-if="showCheatsheet" class="cheatsheet-panel">
            <div class="cheatsheet-title">Quick reference</div>
            <div class="cheatsheet-section">Aliases</div>
            <code>k</code> = kubectl · <code>kk</code> = kubectl krew
            <div class="cheatsheet-section">Krew</div>
            <code>kk list</code> · <code>kk search</code> · <code>kk install &lt;name&gt;</code><br>
            <code>kk uninstall</code> · <code>kk upgrade</code> · <code>kk update</code>
            <div class="cheatsheet-section">Plugins (run with k)</div>
            <code>k stern . -n &lt;ns&gt;</code> · <code>k get-all -n &lt;ns&gt;</code><br>
            <code>k lineage &lt;res&gt;</code> · <code>k9s</code>
            <div class="cheatsheet-section">CLIs</div>
            <code>zellij</code> · <code>crictl</code> · <code>etcdctl</code> · <code>runc</code>
            <div class="cheatsheet-section">SSH to nodes</div>
            <code>k ssh-jump</code> — kubectl plugin (needs ssh, ssh-agent)
          </div>
        </div>
        <div ref="terminalContainer" class="terminal-container" />
        <div v-if="!terminalReady" class="terminal-placeholder">Loading terminal…</div>
      </div>
    </div>
  </div>
</template>

<script>
// Backend URL: same pattern as other ui-plugin-examples (node-driver uses relative /meta/proxy/*, VncViewer uses window.location). Same origin + /krew-api so no port-forward; dev uses localhost:9000.
function getBackendUrl() {
  const o = window.location.origin;
  if (o.startsWith('http://localhost') || o.startsWith('http://127.0.0.1')) return 'http://localhost:9000';
  return o + '/krew-api';
}
const BACKEND_URL = getBackendUrl();
const WS_URL = BACKEND_URL.replace(/^http/, 'ws');

// Get Rancher token from current session (cookie sent automatically to same origin)
let _tokenCache = { token: null, expires: 0 };
async function getRancherToken() {
  if (_tokenCache.token && Date.now() < _tokenCache.expires) return _tokenCache.token;
  const base = window.location.origin;
  // Try Steve API (management cluster may be "local" or have a custom ID)
  const paths = [
    '/k8s/clusters/local/apis/ext.cattle.io/v1/tokens',
    '/v1/tokens.ext.cattle.io',
  ];
  let lastErr;
  for (const apiPath of paths) {
    try {
      const resp = await fetch(`${base}${apiPath}`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          apiVersion: 'ext.cattle.io/v1',
          kind:       'Token',
          metadata:   { generateName: 'krew-' },
          spec:       { description: 'Krew Workstation', ttl: 3600000 },
        }),
      });
      if (!resp.ok) {
        const err = await resp.json().catch(() => ({}));
        throw new Error(err.message || `Token API ${resp.status}`);
      }
      const data = await resp.json();
      const token = data.status?.bearerToken || data.status?.value || data.token;
      if (token) _tokenCache = { token, expires: Date.now() + 50 * 60 * 1000 };
      return token;
    } catch (e) {
      lastErr = e;
    }
  }
  throw lastErr || new Error('Could not get Rancher token');
}

const XTERM_CDN = 'https://cdn.jsdelivr.net/npm';
const XTERM_VER = '5.3.0';
const FIT_VER = '0.8.0';

function loadScript(src) {
  return new Promise((resolve, reject) => {
    const s = document.createElement('script');
    s.src = src;
    s.onload = resolve;
    s.onerror = reject;
    document.head.appendChild(s);
  });
}

function loadCss(href) {
  return new Promise((resolve, reject) => {
    const l = document.createElement('link');
    l.rel = 'stylesheet';
    l.href = href;
    l.onload = resolve;
    l.onerror = reject;
    document.head.appendChild(l);
  });
}

export default {
  name: 'KrewPage',
  layout: 'plain',

  data() {
    return {
      showAboutKrew:    false,
      showCheatsheet:   false,
      darkMode:         true,
      pendingSyncMessage: '',
      clusters:         [],
      plugins:        [],
      search:         '',
      loading:        false,
      busy:           '',
      error:          '',
      message:        '',
      activeTab:      'terminal',
      terminalReady:  false,
      shellConnected: false,
      term:           null,
      fitAddon:       null,
      ws:             null,
      fsPath:         '/root',
      fsEntries:      [],
      fsError:        '',
      currentContext: '',
      installedPlugins: [],
      syncingKubeconfig: false,
      containerInfo: null,
      pluginPage: 1,
      pluginsPerPage: 25,
    };
  },

  computed: {
    themeClass() {
      return this.darkMode ? 'theme-dark' : 'theme-light';
    },
    themeVars() {
      return this.darkMode
        ? {
            '--krew-bg': '#0d0d0d',
            '--krew-panel': '#1a1a1a',
            '--krew-panel-border': '#333',
            '--krew-tabs': '#252525',
            '--krew-toolbar': '#252525',
            '--krew-shell-header': '#252525',
            '--krew-text': '#e0e0e0',
            '--krew-muted': '#888',
          }
        : {
            '--krew-bg': '#f0f0f0',
            '--krew-panel': '#fff',
            '--krew-panel-border': '#ddd',
            '--krew-tabs': '#f0f0f0',
            '--krew-toolbar': '#f5f5f5',
            '--krew-shell-header': '#e8e8e8',
            '--krew-text': '#333',
            '--krew-muted': '#666',
          };
    },
    filteredPlugins() {
      let list = this.plugins;
      if (this.search) {
        const q = this.search.toLowerCase();
        list = list.filter(
          (p) => p.name.toLowerCase().includes(q) || (p.description || '').toLowerCase().includes(q)
        );
      }
      return [...list].sort((a, b) => (a.installed === b.installed ? 0 : a.installed ? -1 : 1));
    },
    pluginPageCount() {
      return Math.max(1, Math.ceil(this.filteredPlugins.length / this.pluginsPerPage));
    },
    paginatedPlugins() {
      const start = (this.pluginPage - 1) * this.pluginsPerPage;
      return this.filteredPlugins.slice(start, start + this.pluginsPerPage);
    },
    pluginPageStart() {
      return this.filteredPlugins.length ? (this.pluginPage - 1) * this.pluginsPerPage + 1 : 0;
    },
    pluginPageEnd() {
      return Math.min(this.pluginPage * this.pluginsPerPage, this.filteredPlugins.length);
    },
    parentPath() {
      if (!this.fsPath || this.fsPath === '/') return '/';
      const parts = this.fsPath.split('/').filter(Boolean);
      parts.pop();
      return parts.length ? '/' + parts.join('/') : '/';
    },
  },

  async mounted() {
    const saved = localStorage.getItem('krew-darkMode');
    if (saved !== null) this.darkMode = saved === 'true';
    await Promise.all([this.fetchClusters(), this.loadPlugins(), this.fetchContainerInfo()]);
    this.loadFs(this.fsPath);
    await this.syncKubeconfig(); // Must complete before terminal — kubectl needs kubeconfig
    this.initTerminal();
  },

  watch: {
    darkMode(v) {
      localStorage.setItem('krew-darkMode', String(v));
      this.$nextTick(() => this.applyTerminalTheme());
    },
    activeTab(tab) {
      if (tab === 'terminal' && this.fitAddon) {
        this.$nextTick(() => this.fitAddon.fit());
      }
    },
  },

  beforeDestroy() {
    this.disconnectShell();
    if (this.term) this.term.dispose();
  },

  methods: {
    toggleTheme() {
      this.darkMode = !this.darkMode;
      this.$nextTick(() => this.applyTerminalTheme());
    },
    applyTerminalTheme() {
      if (!this.term) return;
      const theme = this.darkMode
        ? { background: '#1a1a1a', foreground: '#e0e0e0' }
        : { background: '#f5f5f5', foreground: '#333' };
      this.term.options.theme = { ...theme };
      if (typeof this.term.refresh === 'function') {
        this.term.refresh(0, this.term.rows - 1);
      }
    },
    async api(method, path, opts = {}) {
      const headers = { ...opts.headers };
      try {
        const token = await getRancherToken();
        if (token) headers['Authorization'] = `Bearer ${token}`;
      } catch (_) {}
      const resp = await fetch(`${BACKEND_URL}${path}`, {
        method,
        headers: { 'Content-Type': 'application/json', ...headers },
        ...opts,
      });
      const data = await resp.json();
      if (!resp.ok) throw new Error(data.error || `HTTP ${resp.status}`);
      return data;
    },

    async fetchClusters() {
      try {
        const data = await this.api('GET', '/api/clusters');
        this.clusters = data.clusters || [];
      } catch (e) {}
    },

    async syncKubeconfig() {
      this.syncingKubeconfig = true;
      this.error = '';
      try {
        const data = await this.api('POST', '/api/kubeconfig/sync');
        await this.fetchContext();
        const n = data.clusters ?? 0;
        const msg = n > 0 ? `Kubeconfig synced for ${n} cluster(s)` : 'No clusters to sync';
        this.pendingSyncMessage = msg;
      } catch (e) {
        this.error = `Sync failed: ${e.message}`;
        await this.fetchContext();
      } finally {
        this.syncingKubeconfig = false;
      }
    },

    async fetchContext() {
      try {
        const data = await this.api('GET', '/api/context');
        this.currentContext = data.context || '';
      } catch (e) {
        this.currentContext = '';
      }
    },

    async fetchInstalledPlugins() {
      try {
        const data = await this.api('GET', '/api/plugins/installed');
        this.installedPlugins = data.plugins || [];
      } catch (e) {
        this.installedPlugins = [];
      }
    },

    async fetchContainerInfo() {
      try {
        const data = await this.api('GET', '/api/info');
        this.containerInfo = data;
      } catch (e) {
        this.containerInfo = null;
      }
    },

    formatSize(bytes) {
      if (bytes < 0) return '-';
      if (bytes < 1024) return bytes + ' B';
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
      return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
    },

    async loadPlugins() {
      this.loading = true;
      this.error = '';
      try {
        const data = await this.api('GET', '/api/plugins');
        this.plugins = data.plugins || [];
      } catch (e) {
        this.error = `Backend unreachable at ${BACKEND_URL} — ${e.message}`;
      } finally {
        this.loading = false;
      }
    },

    async updateIndex() {
      this.loading = true;
      this.message = '';
      try {
        await this.api('POST', '/api/plugins/update');
        this.message = 'Plugin index updated';
        await this.loadPlugins();
      } catch (e) {
        this.error = `Update failed: ${e.message}`;
      } finally {
        this.loading = false;
      }
    },

    async installPlugin(p) {
      this.busy = p.name;
      this.message = '';
      try {
        await this.api('POST', `/api/plugins/${p.name}/install`);
        this.message = `Installed ${p.name}`;
        await this.loadPlugins();
      } catch (e) {
        this.error = `Install failed: ${e.message}`;
      } finally {
        this.busy = '';
      }
    },

    async uninstallPlugin(p) {
      this.busy = p.name;
      this.message = '';
      try {
        await this.api('DELETE', `/api/plugins/${p.name}`);
        this.message = `Uninstalled ${p.name}`;
        await this.loadPlugins();
      } catch (e) {
        this.error = `Uninstall failed: ${e.message}`;
      } finally {
        this.busy = '';
      }
    },

    async upgradePlugin(p) {
      this.busy = p.name;
      this.message = '';
      try {
        await this.api('POST', `/api/plugins/${p.name}/upgrade`);
        this.message = `Upgraded ${p.name}`;
        await this.loadPlugins();
      } catch (e) {
        this.error = `Upgrade failed: ${e.message}`;
      } finally {
        this.busy = '';
      }
    },

    async loadFs(path) {
      this.fsError = '';
      try {
        const data = await this.api('GET', `/api/fs?path=${encodeURIComponent(path)}`);
        this.fsPath = data.path;
        this.fsEntries = data.entries || [];
      } catch (e) {
        this.fsError = e.message;
      }
    },

    fsNavigate(path) {
      this.loadFs(path);
    },

    async initTerminal() {
      const container = this.$refs.terminalContainer;
      if (!container) return;
      try {
        await loadCss(`${XTERM_CDN}/xterm@${XTERM_VER}/css/xterm.css`);
        await loadScript(`${XTERM_CDN}/xterm@${XTERM_VER}/lib/xterm.js`);
        await loadScript(`${XTERM_CDN}/xterm-addon-fit@${FIT_VER}/lib/xterm-addon-fit.js`);
      } catch (e) {
        container.innerHTML = `<p class="fs-error">Failed to load terminal: ${e.message}</p>`;
        return;
      }

      const { Terminal } = window;
      const FitAddon = window.FitAddon?.FitAddon || window.FitAddon;
      if (!Terminal || !FitAddon) {
        container.innerHTML = '<p class="fs-error">Terminal not available</p>';
        return;
      }

      this.term = new Terminal({
        cursorBlink: true,
        theme: this.darkMode ? { background: '#1a1a1a', foreground: '#e0e0e0' } : { background: '#f5f5f5', foreground: '#333' },
        fontSize: 13,
      });
      this.fitAddon = new FitAddon();
      this.term.loadAddon(this.fitAddon);
      this.term.open(container);
      this.fitAddon.fit();
      this.terminalReady = true;

      let resizeScheduled = false;
      const resizeObserver = new ResizeObserver(() => {
        if (resizeScheduled) return;
        resizeScheduled = true;
        setTimeout(() => {
          resizeScheduled = false;
          if (this.fitAddon) {
            this.fitAddon.fit();
            this.$nextTick(() => this.sendResize());
          }
        }, 0);
      });
      resizeObserver.observe(container);

      this.term.writeln('');
      this.term.writeln('  Krew Workstation — connecting…');
      this.term.writeln('');
      this.connectShell();
    },

    connectShell() {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) return;
      const ws = new WebSocket(`${WS_URL}/api/ws/shell`);
      this.ws = ws;

      ws.binaryType = 'arraybuffer';
      ws.onopen = () => {
        this.shellConnected = true;
        this.fetchInstalledPlugins();
        this.sendResize();
        if (this.pendingSyncMessage && this.term) {
          this.term.writeln('\r\n  ✓ ' + this.pendingSyncMessage + '\r\n');
          this.pendingSyncMessage = '';
        }
      };
      ws.onclose = () => {
        this.shellConnected = false;
        this.term?.writeln('\r\nDisconnected.');
      };
      ws.onerror = () => {
        this.term?.writeln('\r\nWebSocket error. Is the backend running on ' + BACKEND_URL + '?');
      };
      ws.onmessage = (ev) => {
        if (ev.data instanceof ArrayBuffer && this.term) {
          const buf = new Uint8Array(ev.data);
          this.term.write(buf);
        }
      };

      if (this.term) {
        this.term.onData((data) => {
          if (ws.readyState === WebSocket.OPEN) {
            const bytes = new TextEncoder().encode(data);
            ws.send(bytes.buffer);
          }
        });
      }
    },

    sendResize() {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN || !this.term) return;
      this.ws.send(JSON.stringify({ type: 'resize', cols: this.term.cols, rows: this.term.rows }));
    },

    disconnectShell() {
      if (this.ws) {
        this.ws.close();
        this.ws = null;
      }
      this.shellConnected = false;
    },
  },
};
</script>

<style lang="scss" scoped>
.krew-page {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 60px);
  max-width: none;
  margin: 0;
  padding: 6px 12px;
  overflow: hidden;
  background: var(--krew-bg, #0d0d0d);
}

.frame-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  padding: 6px 12px;
  background: var(--krew-toolbar, #252525);
  border-bottom: 1px solid var(--krew-panel-border, #333);
  .toolbar-left { display: flex; gap: 6px; align-items: center; }
  .toolbar-left .btn.xs {
    padding: 2px 8px;
    font-size: 0.7em;
    min-height: 22px;
  }
  .search-label {
    font-size: 0.8em;
    color: var(--krew-text, #b0b0b0);
    margin-right: 4px;
  }
  .search-input {
    padding: 4px 8px;
    font-size: 0.8em;
    border: 1px solid var(--krew-panel-border, #444);
    border-radius: 4px;
    width: 200px;
    background: var(--krew-panel, #1a1a1a);
    color: var(--krew-text, #e0e0e0);
    &::placeholder { color: var(--krew-muted, #666); }
  }
}

.cluster-info {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
  margin-top: 2px;
  font-size: 0.75em;
  color: var(--muted);
  .cluster-badge {
    padding: 2px 6px;
    border-radius: 4px;
    font-size: 0.8em;
    background: #e3f2fd;
    color: #1565c0;
    &.active { background: #c8e6c9; color: #2e7d32; }
  }
}

.banner {
  flex-shrink: 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 8px;
  border-radius: 4px;
  margin-bottom: 4px;
  font-size: 0.75em;
  &.error { background: #fdecea; color: #b71c1c; }
  &.success { background: #e8f5e9; color: #1b5e20; }
  .dismiss { background: none; border: none; cursor: pointer; font-size: 1.1em; padding: 0 4px; }
}

.loading-bar {
  flex-shrink: 0;
  height: 2px;
  background: var(--primary);
  margin-bottom: 4px;
}

  .panels {
  flex: 1;
  min-height: 300px;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--krew-panel-border, #333);
  border-radius: 6px;
  overflow: hidden;
  background: var(--krew-panel, #1a1a1a);

  .tabs {
    display: flex;
    background: var(--krew-tabs, #252525);
    border-bottom: 1px solid var(--krew-panel-border, #333);
    flex-shrink: 0;
    button {
      padding: 4px 12px;
      font-size: 0.8em;
      border: none;
      background: none;
      cursor: pointer;
      font-weight: 600;
      color: #888;
      &.active { color: #4caf50; border-bottom: 2px solid #4caf50; margin-bottom: -1px; }
    }
  }

  .panel {
    flex: 1;
    overflow: auto;
    padding: 12px;
  }

  .plugins-panel {
    min-height: 300px;
    background: var(--krew-panel, #1a1a1a);
    color: var(--krew-text, #e0e0e0);
    .plugin-table {
      width: 100%;
      border-collapse: collapse;
      font-size: 0.85em;
      font-family: monospace;
      th, td { padding: 6px 10px; text-align: left; border-bottom: 1px solid var(--krew-panel-border, #333); }
      th { font-weight: 600; color: #4caf50; background: var(--krew-tabs, #252525); }
      .name { font-weight: 600; color: #64b5f6; }
      .desc { color: var(--krew-muted, #888); max-width: 280px; }
      .badge {
        padding: 2px 6px;
        border-radius: 4px;
        font-size: 0.75em;
        font-weight: 600;
        &.installed { background: #2e7d32; color: #a5d6a7; }
        &.available { background: #1565c0; color: #90caf9; }
      }
      .actions .btn { margin-right: 4px; }
    }
    .pagination {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 8px 0;
      font-size: 0.8em;
      color: var(--krew-muted, #888);
      .pagination-info { margin-right: 8px; }
    }
  }

  .terminal-panel {
    display: flex;
    flex-direction: column;
    padding: 0;
    background: var(--krew-panel, #1a1a1a);
    min-height: 300px;
    overflow: hidden;

    .shell-header {
      flex-shrink: 0;
      position: relative;
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 6px 12px;
      background: var(--krew-shell-header, #252525);
      color: var(--krew-text, #b0b0b0);
      font-size: 0.75em;
      .shell-header-left {
        display: flex;
        align-items: center;
        gap: 10px;
        position: relative;
        .shell-meta {
          font-size: 0.9em;
          font-family: monospace;
          color: var(--krew-text, #b0b0b0);
        }
        .shell-context {
          padding: 2px 6px;
          border-radius: 4px;
          font-size: 0.9em;
          font-family: monospace;
          background: #3d5a80;
          color: #90caf9;
          max-width: 120px;
          overflow: hidden;
          text-overflow: ellipsis;
        }
        .theme-light & .shell-context { background: #e3f2fd; color: #1565c0; }
        .btn.xs {
          padding: 0 6px;
          font-size: 0.9em;
          min-height: 20px;
          min-width: 20px;
          line-height: 1;
          &.active { background: var(--primary); color: #fff; }
        }
        .about-krew-card {
          position: absolute;
          top: 100%;
          left: 0;
          z-index: 20;
          margin-top: 4px;
          padding: 8px 12px;
          background: var(--krew-panel, #2d2d2d);
          border: 1px solid var(--krew-panel-border, #444);
          border-radius: 6px;
          font-size: 0.8em;
          color: var(--krew-text, #ccc);
          h3 { margin: 0 0 4px; font-size: 0.95em; }
          p { margin: 0; }
          a { color: var(--primary); }
        }
      }
      .shell-header-right {
        display: flex;
        align-items: center;
        gap: 12px;
        > *:not(:last-child)::after {
          content: ' · ';
          margin-left: 4px;
          color: #555;
          font-weight: normal;
        }
      }
      .shell-status { color: #888; }
      .shell-status.connected { color: #4caf50; }
      .connect-btn { min-width: 72px; }
      .cheatsheet-btn { &.active { background: var(--primary); color: #fff; } }
      .cheatsheet-panel {
        position: absolute;
        top: 100%;
        right: 0;
        z-index: 20;
        margin-top: 4px;
        padding: 12px 14px;
        min-width: 280px;
        background: #2d2d2d;
        border: 1px solid #444;
        border-radius: 6px;
        font-size: 0.8em;
        line-height: 1.6;
        color: #ccc;
        box-shadow: 0 4px 12px rgba(0,0,0,0.4);
        .cheatsheet-title { font-weight: 600; color: #fff; margin-bottom: 8px; }
        .cheatsheet-section { font-weight: 600; color: #888; margin-top: 10px; margin-bottom: 4px; font-size: 0.9em; }
        code { background: #1a1a1a; padding: 1px 4px; border-radius: 3px; font-size: 0.9em; }
      }
    }

    .terminal-container {
      flex: 1;
      min-height: 200px;
      padding: 6px;
      overflow: hidden;
    }

    .terminal-placeholder {
      padding: 24px;
      color: #666;
      text-align: center;
    }
  }

  .files-panel {
    background: #1a1a1a;
    color: #e0e0e0;
    .theme-light & {
      background: #fff;
      color: #333;
    }
    .files-toolbar {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 10px;
      .fs-path { font-family: monospace; font-size: 0.9em; color: #4caf50; }
    }
    .fs-error { color: #ef5350; font-size: 0.9em; }
    .fs-tree {
      font-size: 0.85em;
      font-family: monospace;
      .fs-header {
        display: grid;
        grid-template-columns: 1fr 80px 120px 140px;
        gap: 12px;
        padding: 6px 8px;
        font-weight: 600;
        color: #4caf50;
        border-bottom: 1px solid var(--krew-panel-border, #333);
        background: var(--krew-tabs, #252525);
      }
      .fs-row {
        display: grid;
        grid-template-columns: 1fr 80px 120px 140px;
        gap: 12px;
        padding: 6px 8px;
        cursor: pointer;
        border-radius: 4px;
        align-items: center;
        &.dir:hover { background: var(--krew-tabs, #252525); }
        &.file { cursor: default; }
        .fs-col-name {
          display: flex;
          align-items: center;
          gap: 8px;
          .fs-icon { font-size: 1.1em; }
        }
        .fs-col-size, .fs-col-mode { font-size: 0.9em; color: var(--krew-muted, #888); }
        .fs-col-date { font-size: 0.9em; color: var(--krew-muted, #888); }
      }
    }
  }
}

.empty-state {
  padding: 24px;
  text-align: center;
  color: var(--muted);
  font-size: 0.9em;
}
.panels .plugins-panel .empty-state,
.panels .files-panel .empty-state {
  color: var(--krew-muted, #888);
}

::v-deep .xterm {
  padding: 4px;
  background-color: var(--krew-panel, #1a1a1a) !important;
}
::v-deep .xterm-viewport {
  overflow-y: auto !important;
}
</style>
