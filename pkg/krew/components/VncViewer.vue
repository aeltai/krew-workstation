<template>
  <div class="vnc-container">
    <iframe
      ref="vncIframe"
      :src="vncUrl"
      class="vnc-iframe"
      @load="handleIframeLoad"
    ></iframe>
  </div>
</template>

<script>
export default {
  name: 'VncViewer',

  data() {
    return {
      isLoaded: false
    };
  },

  computed: {
    vncUrl() {
      const host = window.location.hostname;
      const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const httpProtocol = window.location.protocol === 'https:' ? 'https:' : 'http:';
      
      // Construct the noVNC URL with auto-connect and scaling enabled
      return `${httpProtocol}//${host}:8080/vnc.html?autoconnect=true&resize=scale&password=rancher&path=websockify&show_dot=true`;
    }
  },

  methods: {
    handleIframeLoad() {
      this.isLoaded = true;
      this.$emit('connected');
    }
  },

  beforeDestroy() {
    if (this.isLoaded) {
      this.$emit('disconnected');
    }
  }
};
</script>

<style lang="scss" scoped>
.vnc-container {
  width: 100%;
  height: 100%;
  min-height: 600px;
  background: #000;
  position: relative;
  overflow: hidden;
}

.vnc-iframe {
  width: 100%;
  height: 100%;
  border: none;
  position: absolute;
  top: 0;
  left: 0;
}
</style> 