import GreetingPage from '../GreetingPage.vue';
import ToolsPage from '../pages/tools.vue';

const routes = [
  {
    name: 'tools',
    path: '/tools',
    component: ToolsPage,
    meta: {
      product: 'tools'
    },
  },
  {
    name: 'krew-manager',
    path: '/tools/krew',
    component: GreetingPage,
    meta: {
      product: 'tools'
    },
  }
];

export default routes; 