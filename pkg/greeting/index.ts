import { importTypes } from '@rancher/auto-import';
import { IPlugin } from '@shell/core/types';
import GreetingPage from './GreetingPage.vue';
import ToolsPage from './pages/tools.vue';

// Init the package
export default function(plugin: IPlugin) {
  // Auto-import model, detail, edit from the folders
  importTypes(plugin);

  // Provide plugin metadata
  plugin.metadata = {
    name: 'krew-plugin-manager',
    version: '0.1.0',
    description: 'Manage kubectl plugins using Krew',
    icon: 'icon-download'
  };

  // Register the routes
  plugin.addRoute({
    name: 'tools',
    path: '/tools',
    component: ToolsPage
  });

  plugin.addRoute({
    name: 'krew-manager',
    path: '/tools/krew',
    component: GreetingPage
  });

  // Load the product
  plugin.addProduct(require('./product'));
} 