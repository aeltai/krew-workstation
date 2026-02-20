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

## Rancher public host (fix 401 / "Sync failed" when connecting to the cluster)

If the plugin shows **401 Unauthorized** or **"Sync failed: undefined"** when you click "Connect to the cluster" / sync kubeconfig, the backend is calling Rancher at the internal URL while the token was issued for the URL you use in the browser. Set the public host to match your Rancher UI URL:

```bash
helm upgrade krew-workstation ./helm/krew-workstation -n krew-workstation \
  --set rancher.publicHost=rancher.35.157.202.39.sslip.io
```

(Use your actual Rancher host, without `https://`.)

## Backend via Rancher meta proxy (no Ingress for API)

The extension uses **Rancher’s meta proxy** (same as node-driver): requests go to `/meta/proxy/krew-workstation.krew-workstation.svc:3000/api/...` so Rancher’s backend proxies to the krew-workstation Service. No Ingress path is required for REST API (Sync, clusters, plugins, etc.).

**Allow list:** The chart creates a **ProxyEndpoint** (management.cattle.io) so the meta proxy allow list is updated automatically—same mechanism as other UI extensions. No manual Settings step. If you disabled it, set `metaProxy.enabled: true` or create a ProxyEndpoint with route `krew-workstation.krew-workstation.svc`. If the proxy returns 502/503, the host is not allow-listed (check that the ProxyEndpoint exists).

**Terminal (WebSocket):** The terminal tab needs `wss://<rancher-host>/krew-api`. Enable the chart’s Ingress so that path is routed to the backend:

```bash
helm upgrade krew-workstation ./helm/krew-workstation -n krew-workstation \
  --set ingress.enabled=true \
  --set ingress.host=rancher.35.157.202.39.sslip.io
```

Use your Rancher host (same as in the UI). If you already set `rancher.publicHost` in values, the Ingress can use it; otherwise set `ingress.host`.

**Alternative: patch the Rancher Ingress** so `/krew-api` is on the same Ingress (one resource). The Ingress backend must be in the same namespace as the Ingress (e.g. `cattle-system`). Create a proxy Service and Endpoints there, then patch:

```bash
# 1) Proxy Service in cattle-system (no selector)
kubectl apply -f - <<'EOF'
apiVersion: v1
kind: Service
metadata:
  name: krew-workstation
  namespace: cattle-system
spec:
  ports:
    - port: 3000
      targetPort: 3000
      protocol: TCP
      name: http
EOF

# 2) Copy Endpoints from krew-workstation so the Service has backends
kubectl get endpoints krew-workstation -n krew-workstation -o json \
  | jq '.metadata |= {"name": "krew-workstation", "namespace": "cattle-system"} | del(.metadata.resourceVersion, .metadata.uid, .metadata.selfLink, .metadata.creationTimestamp)' \
  | kubectl apply -f -

# 3) Add path /krew-api to Rancher Ingress (adjust namespace/name if yours differ)
kubectl patch ingress rancher -n cattle-system --type=json -p='[
  {"op": "add", "path": "/spec/rules/0/http/paths/-", "value": {
    "path": "/krew-api",
    "pathType": "Prefix",
    "backend": {"service": {"name": "krew-workstation", "port": {"number": 3000}}}
  }}
]'
```

After that you can disable the chart’s own Ingress: `--set ingress.enabled=false`. Re-run step 2 if krew-workstation pods change (Endpoints are not updated automatically across namespaces).

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
