const BLANK_CLUSTER = '_';

import KrewPage from '../KrewPage.vue';

const YOUR_PRODUCT_NAME = 'tools';
const TOOLS_HUB_PAGE = 'hub';
const KREW_PAGE = 'krew';

const routes = [
  {
    path: `/${YOUR_PRODUCT_NAME}/c/:cluster/${TOOLS_HUB_PAGE}`,
    redirect: (to) => ({
      name: `${YOUR_PRODUCT_NAME}-c-cluster-${KREW_PAGE}`,
      params: { cluster: to.params.cluster },
    }),
  },
  {
    name: `${YOUR_PRODUCT_NAME}-c-cluster-${KREW_PAGE}`,
    path: `/${YOUR_PRODUCT_NAME}/c/:cluster/${KREW_PAGE}`,
    component: KrewPage,
    meta: {
      product: YOUR_PRODUCT_NAME,
      cluster: BLANK_CLUSTER,
    },
  },
];

export default routes;
