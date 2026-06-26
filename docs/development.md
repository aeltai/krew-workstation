# Local development

## Both extensions in one Rancher UI

Krew Workstation and [Developer Portal](https://github.com/aeltai/rancher-devportal) are **separate repos** but both appear in the **same Rancher Dashboard sidebar**. GitHub Pages hosts only the extension JS bundle — not a user-facing app.

### Link Developer Portal for local dev

```bash
./scripts/link-devportal.sh
```

This symlinks `../rancher-devportal/pkg/devportal` → `pkg/devportal`. One `yarn dev` loads **both** extensions. The webpack config uses `poll: 500` to watch through symlinks so hot-reload works on edits in either repo.

### Start stack

```bash
# 1 — Rancher + krew backend
docker compose up -d

# 2 — Developer Portal backend on the same Docker network
export RANCHER_TOKEN=$(grep RANCHER_TOKEN .env | cut -d= -f2-)
cd ../rancher-devportal
docker compose -f docker-compose.local.yml up -d --build
cd -

# 3 — Link + UI
./scripts/link-devportal.sh
API=http://localhost:8089 yarn dev
```

Open **https://localhost:8005**:

| Sidebar | Extension |
|---------|-----------|
| **Tools → Krew Workstation** | Terminal, kubectl plugins, backups |
| **Platform → Developer Portal** | Self-service environments + manifest preview |

## Krew backend — included CLI tools

The `krew-backend` Docker image ships these tools pre-installed:

| Tool | Purpose |
|------|---------|
| `kubectl` | Cluster management |
| `helm` | Helm chart operations |
| `fleet` | Rancher Fleet CLI |
| `rancher` | Rancher CLI |
| `kwctl` | Kubewarden policy tool |
| `virtctl` | KubeVirt VM management |
| `longhornctl` | Longhorn storage diagnostics |
| `etcdctl` | etcd key/value operations |
| `k9s` | Cluster TUI |
| `crictl` | Container runtime CLI |
| `zellij` | Terminal multiplexer |
| `runc` | OCI container runtime |

Tools are shown in the welcome banner when you open a terminal tab in Krew Workstation:

```
  === CLI Tools ===
  ✓ kubectl       Client Version: v1.x
  ✓ helm          v3.x
  ✓ fleet         version v0.x
  ...
```

Additional local CLIs can be dropped in `.local-plugins/<name>/bin/` (see `backend/install-local-binaries.sh`).

## RBAC testing

```yaml
# docker-compose.yml krew-backend env:
- ALLOW_SERVICE_TOKEN=false   # production: per-user token only
```

Log in as different Rancher users to test cluster scoping and Backups tab visibility.

## Rebuild backend

```bash
docker compose up -d --build krew-backend
```

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Infinite loading in sidebar | Check UIPlugin `pluginName` matches extension `name` |
| `$refs` errors in console | Re-run `node scripts/patch-shell-vue3-refs.js` |
| Backend 401 | Log into Rancher UI or set `RANCHER_TOKEN` in `.env` |
| WebSocket shell fails | Pass token via Rancher login; check backend logs |
| Dev server not hot-reloading devportal changes | `vue.config.js` uses poll watching — changes take ~1s |
| Cards/text overlapping in Developer Portal | Hard-refresh (Cmd+Shift+R) after backend restart |
