# Krew Plugin Manager for Rancher Dashboard

This is a Kubectl Plugin Manager that integrates with Rancher Dashboard, allowing you to manage kubectl plugins using Krew.

## Features
- List available kubectl plugins
- Install and uninstall plugins
- Upgrade existing plugins
- SSH connection to backend for direct plugin management
- Search functionality for plugins

## Development
To develop the plugin:
1. Run `yarn install` in the root directory
2. Run `yarn dev` to start development mode
3. Access the plugin at `https://127.0.0.1:8005` after logging in to Rancher 