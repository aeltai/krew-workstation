# Local development

## Prerequisites

- Node.js 20+
- Docker / Docker Compose
- Yarn

## Stack

| Service | URL | Notes |
|---------|-----|-------|
| Rancher | https://localhost:8449 | Bootstrap password `admin` |
| krew-backend | http://localhost:9000 | Mapped from container :3000 |
| UI dev server | https://localhost:8005 | `yarn dev` |

## Steps

### 1. Start Rancher and backend

```bash
docker compose up -d
```

Optional `.env` for dev service token fallback:

```
RANCHER_TOKEN=token-xxxxx:yyyy
```

### 2. Install UI dependencies

```bash
yarn install
```

`postinstall` patches `@rancher/shell` for Vue 3.

### 3. Run the extension dev server

```bash
API=http://localhost:9000 yarn dev
```

`API` points the Shell dev proxy at krew-backend.

### 4. Open Rancher UI

Navigate to **Tools → Krew Workstation** at https://localhost:8005 (log in with Rancher credentials).

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
