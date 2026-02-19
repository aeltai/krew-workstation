// definition of a "blank cluster" in Rancher Dashboard
const BLANK_CLUSTER = '_';

export function init($plugin: any, store: any) {
  const YOUR_PRODUCT_NAME = 'tools';
  const TOOLS_HUB_PAGE = 'hub';
  const KREW_PAGE = 'krew';

  const {
    product,
    basicType,
    virtualType,
  } = $plugin.DSL(store, YOUR_PRODUCT_NAME);

  // Registering a top-level product — goes directly to Krew Workstation
  product({
    icon: 'globe',
    inStore: 'management',
    weight: 100,
    to: {
      name: `${YOUR_PRODUCT_NAME}-c-cluster-${KREW_PAGE}`,
      params: {
        product: YOUR_PRODUCT_NAME,
        cluster: BLANK_CLUSTER
      }
    }
  });

  // Creating custom pages
  virtualType({
    label: 'Krew Workstation',
    name: KREW_PAGE,
    route: {
      name: `${YOUR_PRODUCT_NAME}-c-cluster-${KREW_PAGE}`,
      params: {
        product: YOUR_PRODUCT_NAME,
        cluster: BLANK_CLUSTER
      }
    },
    icon: 'icon-download'
  });

  // Registering the defined pages as side-menu entries
  basicType([KREW_PAGE]);
} 