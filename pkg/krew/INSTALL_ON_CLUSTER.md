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

The extension uses **Rancher’s meta proxy** (same as node-driver): requests go to `/meta/proxy/http:/krew-workstation.krew-workstation.svc:3000/api/...` so Rancher’s backend proxies to the krew-workstation Service. No Ingress path is required for REST API (Sync, clusters, plugins, etc.).

**Allow list:** The chart creates a **ProxyEndpoint** (management.cattle.io) so the meta proxy allow list is updated automatically—same mechanism as other UI extensions. No manual Settings step. If you disabled it, set `metaProxy.enabled: true` or create a ProxyEndpoint with route `krew-workstation.krew-workstation.svc`. If the proxy returns **502/503** and the ProxyEndpoint exists, see **Troubleshooting → 502/503 from meta proxy** below.

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

1. Rebuild the backend image (includes rk9s): `docker build -t ghcr.io/aeltai/krew-workstation:latest ./backend`
2. Push to your registry and set: `--set image.repository=YOUR_REGISTRY/krew-workstation`
3. The extension appears in Rancher Dashboard under "Krew Workstation" (terminal icon)

## Troubleshooting

### 502/503 from meta proxy (Backend unreachable)

Rancher’s **meta proxy** allow list (who can be called via `/meta/proxy/...`) is built from:

**502 with whitelist correct:** The meta proxy assumes **HTTPS** by default. If your backend is plain HTTP (like krew-workstation on port 3000), the extension must use `/meta/proxy/http:/host:port` (one slash after `http:`). Without this, Rancher tries HTTPS and gets a connection error → 502.

1. **ProxyEndpoint** resources (management.cattle.io) – the chart creates one so the backend is allow-listed automatically.
2. The **whitelist-domain** global setting – a comma-separated list of hostnames.

The controller that reconciles ProxyEndpoint into the allow list is only registered when **Multi-Cluster Management (MCM)** is enabled. If MCM is disabled, ProxyEndpoint is ignored and you get 502/503 even though the CR exists.

**Manual fallback (works with or without MCM):** Add the backend host to **whitelist-domain** so the meta proxy allows it:

- In Rancher UI: **Global Settings** (or **Settings** in the API) → find **whitelist-domain** → set or append (comma-separated) the hostname **without** port, e.g. `krew-workstation.krew-workstation.svc`. Save.
- Or set the env var on the Rancher server (see below).

**Option B – Add to Rancher deployment on the cluster**

So the meta proxy allow list includes `krew-workstation.krew-workstation.svc` without using the UI:

**If Rancher is installed with Helm** (e.g. `helm install rancher rancher-stable/rancher`), add the whitelist domain to the Helm values and upgrade. The official chart supports `extraEnv`:

```yaml
# values.yaml (or --set on CLI)
extraEnv:
  - name: CATTLE_WHITELIST_DOMAIN
    value: "forums.rancher.com,krew-workstation.krew-workstation.svc"
```

If you already have `extraEnv`, append the `CATTLE_WHITELIST_DOMAIN` entry. Then:

```bash
helm upgrade rancher rancher-stable/rancher -n cattle-system -f values.yaml
```

(Use your actual Helm release name, namespace, and repo if different.)

**If you prefer to patch the running deployment** (works for any install method):

```bash
# Add or update CATTLE_WHITELIST_DOMAIN on the Rancher deployment (default: cattle-system)
kubectl set env deployment/rancher -n cattle-system \
  CATTLE_WHITELIST_DOMAIN=forums.rancher.com,krew-workstation.krew-workstation.svc
```

If the deployment has a different name or namespace, adjust. To **append** to an existing value instead of replacing, get the current value first:

```bash
kubectl set env deployment/rancher -n cattle-system \
  CATTLE_WHITELIST_DOMAIN="$(kubectl get deployment rancher -n cattle-system -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="CATTLE_WHITELIST_DOMAIN")].value}'),krew-workstation.krew-workstation.svc"
```

If `CATTLE_WHITELIST_DOMAIN` is not set yet, the above may leave a leading comma; in that case set explicitly:

```bash
kubectl set env deployment/rancher -n cattle-system \
  CATTLE_WHITELIST_DOMAIN=forums.rancher.com,krew-workstation.krew-workstation.svc
```

Rancher will roll out the new env; after the pod is ready, retry the Krew Workstation UI.

After changing the setting (UI or deployment), the meta proxy uses the new list on the next request.

The meta proxy validates the request hostname (port is stripped) against the allow list; it supports exact match, `*suffix` wildcards, and `%.` patterns (see Rancher `pkg/httpproxy/proxy.go`). Use the exact hostname above.

*(Note: This is the meta proxy **domain** allow list. It is separate from HTTP_PROXY/NO_PROXY and CATTLE_WHITELIST_ENVVARS, which control outbound proxy and env var passthrough.)*

### Helm upgrade conflict with Rancher (`conflict with "rancher" ... imagePullPolicy`)

If the Deployment was created or modified by Rancher, Helm can conflict on fields like `imagePullPolicy`. The chart now uses `image.pullPolicy: Always` to align with Rancher. If upgrade still fails, force Helm to take over (brief rollout):

```bash
helm upgrade krew-workstation ./helm/krew-workstation -n krew-workstation --force --reuse-values
```

## Uninstall

```bash
helm uninstall krew-workstation -n krew-workstation
```
