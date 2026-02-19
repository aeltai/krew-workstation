# Install Krew Workstation on a Rancher Management Cluster

**Repo:** `https://github.com/aeltai/krew-workstation`  
**Or from this monorepo:** `ui-plugin-examples/pkg/krew`

## What to install

Krew Workstation = Rancher UI extension + backend. Deploy to the **management cluster** (where Rancher runs).

## Prerequisites

- Rancher installed on the cluster
- `kubectl` and `helm` available
- Cluster has outbound access to GitHub (for image pulls)

## Install command

```bash
# Clone or use the chart
git clone https://github.com/aeltai/krew-workstation.git
cd krew-workstation

# Install on management cluster
helm install krew-workstation ./helm/krew-workstation \
  --set rancher.url=https://rancher.cattle-system.svc \
  --create-namespace \
  -n krew-workstation
```

If Rancher uses a different URL from inside the cluster, set it:

```bash
--set rancher.url=https://YOUR_RANCHER_INGRESS_HOST
```

## Optional: Rancher token

If the backend needs to call Rancher API without the UI token:

```bash
--set rancher.token="token-xxx:yyy"
```

## After install

1. Rebuild the backend image (includes k9s): `docker build -t ghcr.io/aeltai/krew-workstation:latest ./backend`
2. Push to your registry and set: `--set image.repository=YOUR_REGISTRY/krew-workstation`
3. The extension appears in Rancher Dashboard under "Krew Workstation" (terminal icon)

## Uninstall

```bash
helm uninstall krew-workstation -n krew-workstation
```
