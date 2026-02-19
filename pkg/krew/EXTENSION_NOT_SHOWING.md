# Why Krew Workstation Doesn't Appear in Rancher

## Why it works locally but not with the Helm chart

**Locally** (`yarn dev` from ui-plugin-examples): The extension is compiled into the Rancher shell, so the sidebar entry is always there.

**With Helm**: The extension is loaded from the **UIPlugin** endpoint. For the sidebar to show you must:

1. **Use the correct plugin name** – The chart sets `spec.plugin.name` to `krew` (must match `plugin/package.json` `"name"`). Upgrade so this is applied:
   ```bash
   helm upgrade krew-workstation ./helm/krew-workstation -n krew-workstation --reuse-values
   ```

2. **Be on the management cluster** – The product is registered with `inStore: 'management'`. In the cluster dropdown (top-left), select the **local** / **management** cluster. If a downstream cluster is selected, "Krew Workstation" will not appear.

3. **Enable the extension** – In Rancher go to **Extensions**. Find "Krew Workstation" and ensure it is **Enabled**. If it’s disabled, enable it and hard-refresh (Ctrl+Shift+R / Cmd+Shift+R).

4. **Avoid cached broken JS** – The chart sets `noCache: true` so the UI fetches fresh JS. After upgrading, do a hard refresh.

5. **Check the browser console** – Open DevTools (F12) → Console. Reload and look for errors loading the extension (e.g. 404 for the plugin URL or JS errors). The extension is loaded from `endpoint`/plugin/`main` (e.g. `.../extensions/krew/0.1.0/plugin/krew-0.1.0.umd.min.js`).

6. **UIPlugin namespace** – The UIPlugin must be in `cattle-ui-plugin-system`. The chart deploys it there by default.

---

## Root cause (endpoint 404)

The **UIPlugin** points to an extension URL that **does not exist**:

```
https://raw.githubusercontent.com/aeltai/ui-plugin-examples/main/extensions/krew-plugin-manager/0.1.0
```

Rancher loads the UI extension from this URL. When it returns 404, the extension fails to load and does not appear in the sidebar.

## Fix: Build and publish the extension

### 1. Build the extension

From `ui-plugin-examples` root:

```bash
cd /path/to/ui-plugin-examples
yarn build-pkg krew
```

Output: `dist-pkg/krew-0.1.0/` with `krew-0.1.0.umd.min.js` and related files.

If the build fails (e.g. TypeScript errors in node_modules), try:

```bash
cd pkg/krew
npm run build   # if the package has a build script
# or use the parent's vue-cli-service
```

### 2. Publish the built files

**Option A: Push to krew-workstation repo**

```bash
# From ui-plugin-examples root
yarn build-pkg krew

# Prepare extensions layout (package.json must be inside plugin/)
KREW_EXT="extensions/krew/0.1.0"
mkdir -p ${KREW_EXT}/plugin
cp dist-pkg/krew-0.1.0/krew-0.1.0.umd.min.js ${KREW_EXT}/plugin/
cp pkg/krew/extensions/krew/0.1.0/plugin/package.json ${KREW_EXT}/plugin/

# Push to krew-workstation
cd /path/to/krew-workstation  # your clone
cp -r /path/to/ui-plugin-examples/extensions .
git add extensions/
git commit -m "Add extension bundle with plugin/package.json"
git push origin main
```

Then upgrade Helm with the correct endpoint:

```bash
helm upgrade krew-workstation ./helm/krew-workstation -n krew-workstation \
  --set uiPlugin.endpoint="https://raw.githubusercontent.com/aeltai/krew-workstation/main/extensions/krew/0.1.0" \
  --set uiPlugin.version="0.1.0"
```

**Option B: GitHub Pages**

1. Enable GitHub Pages on the krew-workstation repo
2. Publish the `dist-pkg` output to the `gh-pages` branch
3. Endpoint: `https://aeltai.github.io/krew-workstation/krew/0.1.0`

### 3. Update UIPlugin name/version

The UIPlugin `name` and `version` must match the built plugin filename. The build outputs `krew-0.1.0.umd.min.js`, so ensure the UIPlugin uses a compatible name. Rancher typically looks for `plugin/<name>-<version>.umd.min.js` under the endpoint.

### 4. UIPlugin namespace

Rancher expects extensions in `cattle-ui-plugin-system`. The Helm chart now deploys the UIPlugin there by default. If you upgraded from an older install, delete the old UIPlugin:

```bash
kubectl delete uiplugin -n krew-workstation krew-workstation
```

### 5. Restart / clear cache

After publishing:

```bash
kubectl rollout restart deployment -n krew-workstation
```

Hard-refresh the browser (Ctrl+Shift+R or Cmd+Shift+R) to clear cached extension assets.

### 6. UIPlugin spec.plugin.name must match extension package.json

The Helm chart now sets `spec.plugin.name` to **`krew`** (matching `extensions/.../plugin/package.json` `"name": "krew"`). If the UIPlugin used a different name (e.g. `krew-workstation`), Rancher may not register the sidebar entry. Upgrade with:

```bash
helm upgrade krew-workstation ./helm/krew-workstation -n krew-workstation --reuse-values
```

(Chart default is now `pluginName: krew`.)

### 7. Where to find it

- Select the **management cluster** (local) in the cluster dropdown (top-left).
- Look for **"Krew Workstation"** in the left sidebar (terminal icon).
- If you don’t see it: hard refresh (Ctrl+Shift+R / Cmd+Shift+R), clear site data for the Rancher URL, or try an incognito window.
