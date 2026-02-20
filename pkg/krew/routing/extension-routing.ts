// definition of a "blank cluster" in Rancher Dashboard
const BLANK_CLUSTER = '_';

import KrewPage from '../KrewPage.vue';

// to achieve naming consistency throughout the extension
const YOUR_PRODUCT_NAME = 'vishnu';
const TOOLS_HUB_PAGE = 'hub';
const KREW_PAGE = 'krew';

const routes = [
  // Redirect hub to krew (legacy)
  {
    path: `/${YOUR_PRODUCT_NAME}/c/:cluster/${TOOLS_HUB_PAGE}`,
    redirect: (to: { params: { cluster: string } }) => ({ name: `${YOUR_PRODUCT_NAME}-c-cluster-${KREW_PAGE}`, params: { cluster: to.params.cluster } }),
  },
  // Krew Workstation
  {
    name: `${YOUR_PRODUCT_NAME}-c-cluster-${KREW_PAGE}`,
    path: `/${YOUR_PRODUCT_NAME}/c/:cluster/${KREW_PAGE}`,
    component: KrewPage,
    meta: {
      product: YOUR_PRODUCT_NAME,
      cluster: BLANK_CLUSTER
    },
  }
];

export default routes;
