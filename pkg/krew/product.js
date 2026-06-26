const BLANK_CLUSTER = '_';

export function init($plugin, store) {
  const YOUR_PRODUCT_NAME = 'tools';
  const TOOLS_HUB_PAGE = 'hub';
  const KREW_PAGE = 'krew';

  const {
    product,
    basicType,
    virtualType,
  } = $plugin.DSL(store, YOUR_PRODUCT_NAME);

  product({
    icon:  'terminal',
    label: 'Krew Workstation',
    inStore: 'management',
    weight: 100,
    to: {
      name: `${YOUR_PRODUCT_NAME}-c-cluster-${KREW_PAGE}`,
      params: {
        product: YOUR_PRODUCT_NAME,
        cluster: BLANK_CLUSTER,
      },
    },
  });

  virtualType({
    label: 'Krew Workstation',
    name: KREW_PAGE,
    route: {
      name: `${YOUR_PRODUCT_NAME}-c-cluster-${KREW_PAGE}`,
      params: {
        product: YOUR_PRODUCT_NAME,
        cluster: BLANK_CLUSTER,
      },
    },
    icon: 'terminal',
  });

  basicType([KREW_PAGE]);
}
