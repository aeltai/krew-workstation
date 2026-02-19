# Krew Workstation Helm Chart

Deploy the Krew Workstation backend to Kubernetes (e.g. Rancher management cluster).

## Prerequisites

- Kubernetes cluster with Rancher installed
- Rancher URL reachable from the cluster (e.g. `https://rancher.cattle-system.svc`)

## Install

```bash
# Add your values
helm install krew-workstation ./krew-workstation \
  --set rancher.url=https://rancher.cattle-system.svc \
  --set rancher.token="token-xxx:yyy" \
  --create-namespace \
  -n krew-workstation
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `rancher.url` | Rancher API URL (from within cluster) | `https://rancher.cattle-system.svc` |
| `rancher.token` | Optional Rancher bearer token | `""` (UI passes token per-request) |
| `persistence.enabled` | Persist krew plugins across restarts | `true` |
| `persistence.size` | PVC size for krew data | `1Gi` |
| `ingress.enabled` | Expose via Ingress | `false` |

## Rancher URL

When running inside the same cluster as Rancher:
- Use `https://rancher.cattle-system.svc` (ClusterIP)
- Or your Rancher ingress host if different

## Uninstall

```bash
helm uninstall krew-workstation -n krew-workstation
```
