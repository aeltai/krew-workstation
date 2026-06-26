# Krew Workstation — UI extension

## Layout

```
pkg/krew/
├── index.ts           # Plugin entry
├── product.ts         # Tools → Krew Workstation sidebar
├── package.json       # name: krew (must match UIPlugin)
├── KrewPage.vue       # Terminal, plugins, files, backups tabs
└── routing/
    └── extension-routing.ts
```

## Plugin name

The extension **`name`** in `pkg/krew/package.json` must be `krew`. This matches:

- Helm `uiPlugin.pluginName`
- UIPlugin CR `spec.plugin.name`
- Bundle file `krew-0.1.0.umd.min.js`

## Build

```bash
yarn install
yarn build-pkg
```

Artifacts land in `dist-pkg/krew-0.1.0/`. Copy to `extensions/krew/0.1.0/plugin/` for commit or CI publish.

## Rancher Shell version

Uses `@rancher/shell` **3.0.4** with Vue **3.2.13**. Run `yarn postinstall` (automatic) to patch Shell components for Vue 3 `$refs` array handling — see `scripts/patch-shell-vue3-refs.js`.

## Backend API

The extension talks to **krew-backend** (Go) on port 9000 (dev) or the in-cluster Service. All `/api/*` routes require a Rancher bearer token except when `ALLOW_SERVICE_TOKEN=true` (dev only).

See the main [README](../README.md) for the API table.

## Tabs and RBAC

| Tab | Gate |
|-----|------|
| Terminal | Always (with auth) |
| Plugins | `capabilities.managePlugins` |
| Files | Authenticated user |
| Backups | `capabilities.backups` (operator installed + CR list permission) |

Capabilities come from `GET /api/auth/me`.
