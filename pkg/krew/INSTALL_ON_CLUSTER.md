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

## Rancher public host (fix 401 when opening the plugin)

If the plugin page loads but shows "must authenticate" or 401 when syncing kubeconfig, the backend is calling Rancher at the internal URL while the token was issued for the URL you use in the browser. Set the public host to match your Rancher UI URL:

```bash
helm upgrade krew-workstation ./helm/krew-workstation -n krew-workstation \
  --set rancher.publicHost=rancher.35.157.202.39.sslip.io
```

(Use your actual Rancher host, without `https://`.)

## Backend URL (same cluster, no port-forward)

The extension calls the backend at **same host as Rancher** + `/krew-api` (e.g. `https://rancher.35.157.202.39.sslip.io/krew-api`). So whatever serves your Rancher UI (Ingress or LB) must route path `/krew-api` to the `krew-workstation` service in `krew-workstation` namespace (port 3000). Add that path to your existing Rancher ingress; no separate Ingress is required.

## Rancher API token (required for Sync / list all clusters)

The backend must call Rancher’s API (`/v3/clusters`, generateKubeconfig). Tokens sent from the browser often return **401** when used from the backend (session-bound). Set a **Rancher API token** so the backend can list clusters and sync kubeconfig.

1. In Rancher: **User (avatar) → Account & API Keys → Create API Key**. Name it e.g. `krew-workstation`, copy the Bearer token (`token-xxxx:yyyy`).
2. Upgrade the release with the token (use a secret in production instead of `--set`):

```bash
helm upgrade krew-workstation ./helm/krew-workstation -n krew-workstation \
  --set rancher.publicHost=rancher.35.157.202.39.sslip.io \
  --set rancher.token="token-xxxx:yyyy"
kubectl rollout restart deployment krew-workstation -n krew-workstation
```

After this, **Sync** in the plugin should succeed and you’ll see all clusters (not only `local`) in the merged kubeconfig.

### Mount token and public host from a Secret (recommended)

Avoid putting the token (or host) in Helm values or CLI. Create a Kubernetes Secret and point the chart at it:

```bash
# Create the secret in the same namespace as the release
kubectl create secret generic krew-workstation-rancher -n krew-workstation \
  --from-literal=token='token-xxxx:yyyy' \
  --from-literal=publicHost='rancher.35.157.202.39.sslip.io'
```

Then install or upgrade with the existing secret (no token/host in values):

```bash
helm upgrade krew-workstation ./helm/krew-workstation -n krew-workstation \
  --set rancher.existingSecret.name=krew-workstation-rancher \
  --set rancher.existingSecret.tokenKey=token \
  --set rancher.existingSecret.publicHostKey=publicHost
kubectl rollout restart deployment krew-workstation -n krew-workstation
```

Default key names are `token` and `publicHost`; you can omit `tokenKey`/`publicHostKey` if your secret uses those. The chart does **not** create its own rancher secret when `rancher.existingSecret.name` is set.

## Allow root in this namespace (Pod Security)

If the cluster enforces non-root and the pod fails with "runAsUser breaks non-root policy", label the namespace so this one can run as root:

```bash
kubectl label namespace krew-workstation pod-security.kubernetes.io/enforce=privileged --overwrite
```

Then restart the deployment: `kubectl rollout restart deployment krew-workstation -n krew-workstation`

## After install

1. Rebuild the backend image (includes k9s): `docker build -t ghcr.io/aeltai/krew-workstation:latest ./backend`
2. Push to your registry and set: `--set image.repository=YOUR_REGISTRY/krew-workstation`
3. The extension appears in Rancher Dashboard under "Krew Workstation" (terminal icon)

## Troubleshooting

### Helm upgrade conflict with Rancher (`conflict with "rancher" ... imagePullPolicy`)

If the Deployment was created or modified by Rancher, Helm can conflict on fields like `imagePullPolicy`. The chart now uses `image.pullPolicy: Always` to align with Rancher. If upgrade still fails, force Helm to take over (brief rollout):

```bash
helm upgrade krew-workstation ./helm/krew-workstation -n krew-workstation --force --reuse-values
```

## Uninstall

```bash
helm uninstall krew-workstation -n krew-workstation
```
