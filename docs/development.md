# Local development

## Both extensions in one Rancher UI

Krew Workstation and [Developer Portal](https://github.com/aeltai/rancher-devportal) are **separate repos** but both show in the **same Rancher Dashboard sidebar**. GitHub Pages is only CDN for the extension JS — not a user-facing app.

### Link Developer Portal for local dev

```bash
./scripts/link-devportal.sh
```

This symlinks `../rancher-devportal/pkg/devportal` into `pkg/devportal`. One `yarn dev` loads **both** extensions.

### Start stack

```bash
docker compose up -d
docker compose -f ../rancher-devportal/docker-compose.local.yml up -d   # devportal API :9010
./scripts/link-devportal.sh
yarn install
API=http://localhost:8089 yarn dev
```

Open **https://localhost:8005** only (not a separate port):

| Sidebar | Extension |
|---------|-----------|
| **Tools → Krew Workstation** | Terminal, plugins, backups |
| **Platform → Developer Portal** | Self-service environments |

## RBAC testing

Set in `docker-compose.yml` for `krew-backend`:

```yaml
- ALLOW_SERVICE_TOKEN=false
```

Then only per-user Rancher session tokens work — log in as different users to test cluster scoping and Backups tab visibility.

## Rebuild backend

```bash
docker compose up -d --build krew-backend
```

## Install local CLIs in backend container

Drop binaries or krew plugin manifests in `.local-plugins/` (see `backend/install-local-plugins.sh`).

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Infinite loading in sidebar | Check UIPlugin `pluginName` matches extension `name` |
| `$refs` errors in console | Re-run `yarn postinstall` |
| Backend 401 | Log into Rancher UI or set `RANCHER_TOKEN` |
| WebSocket shell fails | Pass token; check backend logs |
