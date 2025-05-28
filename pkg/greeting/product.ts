import { DSL } from '@shell/store/type-map';

export function init($plugin: any, store: any) {
  const {
    product,
    basicType,
    headers,
    configureType,
    virtualType,
  } = $plugin.DSL(store, 'tools');

  // Create the top-level Tools product
  product({
    icon: 'globe',
    inStore: 'management',
    weight: 100,
    label: 'Tools',
    category: 'global',
    to: { name: 'tools' }
  });

  // Create the Tools type
  virtualType({
    label: 'Tools',
    name: 'tools',
    route: {
      name: 'tools'
    },
    icon: 'icon-tools'
  });

  // Create the Krew type
  virtualType({
    label: 'Krew Plugin Manager',
    name: 'krew',
    route: {
      name: 'krew-manager'
    },
    icon: 'icon-download'
  });

  // Add the navigation
  store.dispatch('type-map/addNavigation', {
    name: 'tools',
    label: 'Tools',
    icon: 'icon-tools',
    children: [
      {
        name: 'krew',
        label: 'Krew Plugin Manager',
        route: { name: 'krew-manager' },
        icon: 'icon-download'
      }
    ]
  });
} 