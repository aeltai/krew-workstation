const { defineConfig } = require('@vue/cli-service');

module.exports = defineConfig({
  devServer: {
    port: 8007,
    host: 'localhost',
    https: false,
    hot: true,
    headers: {
      'Access-Control-Allow-Origin': '*'
    },
    proxy: {
      '/v1': {
        target: 'http://localhost:80',
        secure: false,
        ws: true,
        changeOrigin: true
      },
      '/v3': {
        target: 'http://localhost:80',
        secure: false,
        ws: true,
        changeOrigin: true
      },
      '/api': {
        target: 'http://localhost:80',
        secure: false,
        ws: true,
        changeOrigin: true
      },
      '/k8s': {
        target: 'http://localhost:80',
        secure: false,
        ws: true,
        changeOrigin: true
      }
    }
  },
  transpileDependencies: true,
  publicPath: './',
  outputDir: 'dist',
  configureWebpack: {
    resolve: {
      alias: {
        '@': require('path').resolve(__dirname, 'src')
      }
    }
  }
});