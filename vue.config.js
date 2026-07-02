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

// SPA fallback — avoid 404 when opening /platform/c/_/portal or /tools/c/_/krew directly
vueConfig.devServer = {
  ...(vueConfig.devServer || {}),
  historyApiFallback: true,
};

const existingChainWebpack = vueConfig.chainWebpack;

vueConfig.chainWebpack = (webpackConfig) => {
  if (existingChainWebpack) {
    existingChainWebpack(webpackConfig);
  }

  webpackConfig.plugins.delete('eslint');

  // Follow symlinks so pkg/devportal (symlinked from rancher-devportal) is watched
  webpackConfig.resolve.symlinks(true);

  // Poll for changes — necessary for symlinked directories
  webpackConfig.plugin('watchman-fix').use(
    class WatchSymlinksPlugin {
      apply(compiler) {
        compiler.options.watchOptions = {
          ...compiler.options.watchOptions,
          followSymlinks: true,
          poll: 500,
        };
      }
    }
  );
};

module.exports = vueConfig;
