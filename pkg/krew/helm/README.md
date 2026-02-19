# Krew Workstation Helm Chart

Deploy Krew Workstation (backend + UI extension) to the **Rancher management cluster**.

## Target: Management Cluster

This chart deploys to the **management cluster** (local), where Rancher UI runs. The backend needs Rancher API access; the UI extension loads in the Dashboard.

## Prerequisites

- Rancher management cluster
- Rancher URL reachable from the cluster (e.g. `https://rancher.cattle-system.svc`)
- Extension JS bundle published (for UIPlugin endpoint)

## Install

### Option 1: Direct install (backend + UIPlugin)

```bash
helm install krew-workstation ./krew-workstation \
  --set rancher.url=https://rancher.cattle-system.svc \
  --set rancher.token="token-xxx" \
  --set uiPlugin.endpoint="https://raw.githubusercontent.com/YOUR_ORG/ui-plugin-examples/main/extensions/krew-plugin-manager/0.1.0" \
  --create-namespace \
  -n krew-workstation
```

### Option 2: With bootstrap (register extension catalog)

To auto-register the extension repo so Krew appears in Extensions UI:

```bash
helm install krew-workstation ./krew-workstation \
  --set rancher.url=https://rancher.cattle-system.svc \
  --set bootstrap.enabled=true \
  --set catalog.url="https://YOUR_ORG.github.io/krew-workstation-charts" \
  --create-namespace \
  -n krew-workstation
```

Or with Git repo:

```bash
--set catalog.gitRepo="https://github.com/YOUR_ORG/krew-workstation" \
--set catalog.gitBranch="main"
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `rancher.url` | Rancher API URL (from within cluster) | `https://rancher.cattle-system.svc` |
| `rancher.token` | Optional Rancher bearer token | `""` |
| `persistence.enabled` | Persist krew plugins across restarts | `true` |
| `persistence.size` | PVC size for krew data | `1Gi` |
| `uiPlugin.enabled` | Deploy UIPlugin (UI extension) | `true` |
| `uiPlugin.endpoint` | URL to extension JS bundle | (see values.yaml) |
| `catalog.createClusterRepo` | Create ClusterRepo (add repo to catalog) | `false` |
| `bootstrap.enabled` | Run Job to create ClusterRepo on install | `false` |
| `catalog.url` | Helm/GitHub Pages repo URL | `""` |
| `catalog.gitRepo` | Git repo URL (alternative to url) | `""` |

## Resources Deployed

- **Deployment** – backend with krew + shell
- **PVC** – persistent storage for krew plugins
- **Service** – ClusterIP for backend
- **UIPlugin** – registers UI extension in Rancher
- **ClusterRepo** (optional) – adds extension catalog
- **Bootstrap Job** (optional) – creates ClusterRepo on install

## Uninstall

```bash
helm uninstall krew-workstation -n krew-workstation
```
