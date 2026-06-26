# Krew Workstation

A Rancher UI extension that lets you manage kubectl plugins (via [Krew](https://krew.sigs.k8s.io/)) from the Rancher Dashboard. Terminal, plugin catalog, kubeconfig sync, k9s, stern, backups, and more.

**Related:** [Developer Portal](https://github.com/aeltai/rancher-devportal) is a separate extension for self-service environments (Fleet, operators, virtual clusters).

## Documentation

- [Extension development](docs/extension.md)
- [Local development](docs/development.md)
- [Publishing / GitHub Pages](docs/publishing.md)
- [Helm chart](helm/README.md)

Extension bundle (GitHub Pages):  
`https://aeltai.github.io/krew-workstation/extensions/krew/0.1.0/plugin/`

## Architecture

```
 Rancher Dashboard (browser)
      |
      v
 KrewPage.vue  ──HTTP──>  krew-backend (Go)  ──> Rancher API (clusters, kubeconfigs)
                                              ──> kubectl krew (install/uninstall/upgrade/list)
```

- **krew-backend** — a single Go binary that connects to the Rancher API, fetches kubeconfigs per cluster, and runs `kubectl krew` commands locally.
- **KrewPage.vue** — a Rancher UI extension page that renders the plugin list and sends commands to the backend.

## Quick Start

### 1. Start Rancher + krew backend

```bash
cd pkg/krew
docker compose up -d
```

Wait a minute or two for Rancher to boot. It will be at **https://localhost:8449**.

### 2. Rancher authentication (RBAC)

When embedded in Rancher Dashboard, the extension **does not need a separate API key**. It mints a short-lived bearer token from your logged-in Rancher session (`ext.cattle.io/Token`).

Each user gets:
- Their own kubeconfig (scoped by Rancher RBAC via `generateKubeconfig`)
- Cluster list filtered to clusters they can access
- **Backups** tab only if the backup operator is installed **and** they can `list` Backup/Restore CRs
- Terminal shell using **their** kubeconfig (token passed on WebSocket connect)

**Local dev only:** `RANCHER_TOKEN` in `.env` is a service-account fallback (admin-level). For RBAC testing, set `ALLOW_SERVICE_TOKEN=false` in docker-compose and log in as different Rancher users.

Optional: create API keys under **User Avatar → Account & API Keys** for CI/automation — not required for normal UI use.

### 3. (Optional dev) Service token

```bash
# .env — fallback when UI session token is unavailable during local dev
RANCHER_TOKEN="token-xxxxx:yyyyyyyyyyyyyy"
docker compose up -d --build krew-backend
```

### 4. Run the UI dev server

```bash
cd ../..   # repo root
yarn install
API=http://localhost:8089 yarn dev
```

### 5. Open the extension

Go to **https://localhost:8005**, log in, and navigate to **Tools → Krew Plugin Manager**.

## Backend API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/clusters` | List Rancher clusters |
| GET | `/api/clusters/:id/plugins` | List krew plugins for a cluster |
| POST | `/api/clusters/:id/plugins/:name/install` | Install a plugin |
| DELETE | `/api/clusters/:id/plugins/:name` | Uninstall a plugin |
| POST | `/api/clusters/:id/plugins/:name/upgrade` | Upgrade a plugin |

## Helm Chart

Deploy to Kubernetes:

```bash
helm install krew-workstation ./helm/krew-workstation \
  --set rancher.url=https://rancher.cattle-system.svc \
  -n krew-workstation --create-namespace
```

See [helm/README.md](helm/README.md) for full options.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `RANCHER_URL` | `https://rancher:443` | Rancher API URL |
| `RANCHER_TOKEN` | (optional) | Dev-only service token fallback when UI session token unavailable |
| `ALLOW_SERVICE_TOKEN` | `true` | Set `false` to require per-user Authorization on all `/api/*` routes |
| `PORT` | `3000` | Backend listen port |
