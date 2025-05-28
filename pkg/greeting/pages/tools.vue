<template>
  <div class="tools-page">
    <header class="hero">
      <div class="hero-content">
        <h1>Rancher Tools Hub</h1>
        <p class="subtitle">Enhance your Kubernetes experience with powerful tools and utilities</p>
      </div>
    </header>
    
    <section class="features">
      <div class="feature">
        <i class="icon icon-gear" />
        <h3>Streamlined Management</h3>
        <p>Manage your cluster tools and plugins from a centralized dashboard</p>
      </div>
      <div class="feature">
        <i class="icon icon-download" />
        <h3>Easy Installation</h3>
        <p>Install and update kubectl plugins with just one click</p>
      </div>
      <div class="feature">
        <i class="icon icon-shield" />
        <h3>Secure Integration</h3>
        <p>Seamlessly integrated with Rancher's security and authentication</p>
      </div>
    </section>

    <section class="tools-section">
      <h2>Available Tools</h2>
      <div class="tools-grid">
        <RouterLink 
          class="tool-card" 
          :to="{ name: 'krew-manager' }"
        >
          <div class="card-header">
            <i class="icon icon-download" />
            <h3>Krew Plugin Manager</h3>
          </div>
          <div class="card-content">
            <p class="description">Discover and manage kubectl plugins using Krew</p>
            <ul class="features-list">
              <li><i class="icon icon-checkmark" /> Browse official plugin repository</li>
              <li><i class="icon icon-checkmark" /> One-click installations</li>
              <li><i class="icon icon-checkmark" /> Automatic updates</li>
              <li><i class="icon icon-checkmark" /> Plugin version management</li>
            </ul>
          </div>
          <div class="card-footer">
            <span class="learn-more">Get Started <i class="icon icon-chevron-right" /></span>
          </div>
        </RouterLink>
        <!-- Add more tool cards here as needed -->
      </div>
    </section>

    <section class="deployment-guide">
      <h2>Backend Deployment Guide</h2>
      <div class="deployment-steps">
        <div class="step">
          <h3>1. Build the Backend Image</h3>
          <pre class="code-block">
cd backend
docker build -t krew-manager-backend:latest .
          </pre>
        </div>
        <div class="step">
          <h3>2. Deploy to Kubernetes</h3>
          <pre class="code-block">
# Apply the deployment and service
kubectl apply -f k8s/deployment.yaml
          </pre>
        </div>
        <div class="step">
          <h3>3. Verify Deployment</h3>
          <pre class="code-block">
# Check deployment status
kubectl get deployment krew-manager-backend

# Check service status
kubectl get service krew-manager-backend
          </pre>
        </div>
        <div class="step">
          <h3>4. Access the Service</h3>
          <p>The backend service will be available at:</p>
          <pre class="code-block">
http://krew-manager-backend:3000
          </pre>
          <p>Note: Update the frontend configuration to use this service endpoint.</p>
        </div>
      </div>
    </section>

    <section class="cta-section">
      <div class="cta-content">
        <h2>Ready to Enhance Your Cluster?</h2>
        <p>Start managing your Kubernetes tools with our intuitive interface</p>
      </div>
    </section>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'ToolsPage',
  layout: 'plain'
});
</script>

<style lang="scss" scoped>
.tools-page {
  .hero {
    background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
    padding: 60px 20px;
    text-align: center;
    margin-bottom: 40px;
    color: white;

    .hero-content {
      max-width: 800px;
      margin: 0 auto;

      h1 {
        font-size: 3em;
        margin: 0 0 20px;
        font-weight: bold;
      }

      .subtitle {
        font-size: 1.5em;
        margin: 0;
        opacity: 0.9;
      }
    }
  }

  .features {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 30px;
    max-width: 1200px;
    margin: 0 auto 60px;
    padding: 0 20px;

    .feature {
      text-align: center;
      padding: 30px;
      background: var(--box-bg);
      border-radius: 8px;
      transition: transform 0.2s;

      &:hover {
        transform: translateY(-5px);
      }

      i {
        font-size: 40px;
        color: var(--primary);
        margin-bottom: 20px;
      }

      h3 {
        margin: 0 0 15px;
        color: var(--primary);
        font-size: 1.4em;
      }

      p {
        margin: 0;
        color: var(--secondary);
        line-height: 1.5;
      }
    }
  }

  .tools-section {
    max-width: 1200px;
    margin: 0 auto 60px;
    padding: 0 20px;

    h2 {
      text-align: center;
      color: var(--primary);
      font-size: 2em;
      margin-bottom: 40px;
    }

    .tools-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
      gap: 30px;
    }

    .tool-card {
      display: flex;
      flex-direction: column;
      text-decoration: none;
      background: var(--box-bg);
      border: 1px solid var(--border);
      border-radius: 8px;
      overflow: hidden;
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-5px);
        box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
      }

      .card-header {
        padding: 20px;
        background: var(--primary);
        color: white;
        display: flex;
        align-items: center;
        gap: 15px;

        i {
          font-size: 24px;
        }

        h3 {
          margin: 0;
          font-size: 1.4em;
        }
      }

      .card-content {
        padding: 20px;
        flex-grow: 1;

        .description {
          color: var(--secondary);
          margin: 0 0 20px;
          font-size: 1.1em;
        }

        .features-list {
          list-style: none;
          padding: 0;
          margin: 0;

          li {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 10px;
            color: var(--secondary);

            i {
              color: var(--success);
              font-size: 16px;
            }
          }
        }
      }

      .card-footer {
        padding: 15px 20px;
        border-top: 1px solid var(--border);

        .learn-more {
          display: flex;
          align-items: center;
          justify-content: flex-end;
          gap: 5px;
          color: var(--primary);
          font-weight: bold;

          i {
            font-size: 14px;
            transition: transform 0.2s;
          }
        }

        &:hover {
          .learn-more i {
            transform: translateX(5px);
          }
        }
      }
    }
  }

  .deployment-guide {
    max-width: 1200px;
    margin: 0 auto 60px;
    padding: 0 20px;

    h2 {
      text-align: center;
      color: var(--primary);
      font-size: 2em;
      margin-bottom: 40px;
    }

    .deployment-steps {
      display: grid;
      gap: 30px;

      .step {
        background: var(--box-bg);
        border-radius: 8px;
        padding: 20px;
        border: 1px solid var(--border);

        h3 {
          color: var(--primary);
          margin: 0 0 15px;
        }

        .code-block {
          background: var(--body-bg);
          padding: 15px;
          border-radius: 4px;
          font-family: monospace;
          white-space: pre;
          overflow-x: auto;
        }

        p {
          margin: 15px 0;
          color: var(--secondary);
        }
      }
    }
  }

  .cta-section {
    background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
    padding: 60px 20px;
    text-align: center;
    color: white;

    .cta-content {
      max-width: 600px;
      margin: 0 auto;

      h2 {
        font-size: 2em;
        margin: 0 0 20px;
      }

      p {
        font-size: 1.2em;
        margin: 0;
        opacity: 0.9;
      }
    }
  }
}
</style> 