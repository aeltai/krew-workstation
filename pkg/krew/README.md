# Krew Workstation

A Rancher UI extension that lets you manage kubectl plugins (via [Krew](https://krew.sigs.k8s.io/)) from the Rancher Dashboard. Terminal, plugin catalog, kubeconfig sync, rk9s, stern, and more.

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

### 2. Create a Rancher API token

Log into Rancher at https://localhost:8449 (password: `admin`), then go to **User Avatar → Account & API Keys → Create API Key**. Copy the bearer token.

### 3. Set the token and restart the backend

```bash
export RANCHER_TOKEN="token-xxxxx:yyyyyyyyyyyyyy"
docker compose up -d --build krew-backend
```

### 4. Run the UI dev server

```bash
cd ../..   # ui-plugin-examples root
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
| `RANCHER_TOKEN` | (optional) | Rancher API bearer token; UI passes per-request |
| `PORT` | `3000` | Backend listen port |
