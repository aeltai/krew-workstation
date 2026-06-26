# Publishing the Krew extension

## GitHub Pages (recommended)

Workflow: `.github/workflows/build-extension-pages.yml`

On push to `main`:

1. `yarn build-pkg`
2. Copy to `_site/extensions/krew/0.1.0/plugin/`
3. Deploy via GitHub Actions → Pages

Enable **Settings → Pages → Build and deployment → GitHub Actions** on first use.

### UIPlugin endpoint

```
https://aeltai.github.io/krew-workstation/extensions/krew/0.1.0/plugin
```

Helm default (`helm/krew-workstation/values.yaml`):

```yaml
uiPlugin:
  pluginName: krew
  version: "0.1.0"
  endpoint: "https://aeltai.github.io/krew-workstation/extensions/krew/0.1.0/plugin"
```

The endpoint is the **directory** containing `package.json` and `krew-0.1.0.umd.min.js`, not the JS file itself.

### Verify hosting

```bash
curl -sI "https://aeltai.github.io/krew-workstation/extensions/krew/0.1.0/plugin/krew-0.1.0.umd.min.js" | head -1
```

## Raw GitHub (alternative)

```
https://raw.githubusercontent.com/aeltai/krew-workstation/main/extensions/krew/0.1.0/plugin
```

Commit built files under `extensions/` if you do not use Pages.

## Container image

Backend image: `ghcr.io/aeltai/krew-workstation`. Build and push separately from the UI bundle.

## Version bump checklist

1. Bump `version` in `pkg/krew/package.json` and root `package.json`
2. Update `helm/krew-workstation/values.yaml` (`uiPlugin.version`, path segment)
3. Re-run Pages workflow
4. `helm upgrade` or patch UIPlugin CR

## Extension catalog

Optional: set `bootstrap.enabled=true` and `catalog.gitRepo=https://github.com/aeltai/krew-workstation` to register in Rancher **Extensions** UI.
