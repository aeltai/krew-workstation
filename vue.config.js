const config = require('@rancher/shell/vue.config'); // eslint-disable-line @typescript-eslint/no-var-requires

const vueConfig = config(__dirname, {
  excludes: [],
  proxy: {
    // Devportal backend — avoid HTTPS→HTTP mixed-content from Shell dev UI (:8005)
    '/devportal-api': {
      target:       'http://localhost:9010',
      changeOrigin: true,
      pathRewrite:  { '^/devportal-api': '' },
    },
    // Krew backend (same pattern)
    '/krew-api': {
      target:       'http://localhost:9000',
      changeOrigin: true,
      pathRewrite:  { '^/krew-api': '' },
    },
  },
});

vueConfig.lintOnSave = false;

const existingChainWebpack = vueConfig.chainWebpack;

vueConfig.chainWebpack = (webpackConfig) => {
  if (existingChainWebpack) {
    existingChainWebpack(webpackConfig);
  }

  webpackConfig.plugins.delete('eslint');
};

module.exports = vueConfig;
