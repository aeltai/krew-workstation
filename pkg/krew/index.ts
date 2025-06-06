import { importTypes } from '@rancher/auto-import';
import { IPlugin } from '@shell/core/types';

import ToolsPage from './pages/tools.vue';
import KrewPage from './KrewPage.vue';

// Init the package
export default function(plugin: IPlugin) {
  // Auto-import model, detail, edit from the folders
  importTypes(plugin);

  // Provide plugin metadata from package.json
  plugin.metadata = require('./package.json');

  // Register the routes
  plugin.addRoute({
    name: 'tools',
    path: '/tools',
    component: ToolsPage
  });

  // Load the product
  plugin.addProduct(require('./product'));

  // Add a route to the product
  plugin.addRoute({
    name: 'krew-page',
    path: '/krew',
    component: KrewPage
  });

}