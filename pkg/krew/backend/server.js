const express = require('express');
const cors = require('cors');
const krewApi = require('./krew-api');

const app = express();
const port = process.env.PORT || 3000;

// Enable CORS
app.use(cors());
app.use(express.json());

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({ status: 'healthy' });
});

// Use krew API routes
app.use('/', krewApi);

// Error handling middleware
app.use((err, req, res, next) => {
  console.error('Error:', err);
  res.status(500).json({
    error: 'Internal Server Error',
    details: err.message
  });
});

// Start server
app.listen(port, () => {
  console.log(`Server running on port ${port}`);
  console.log('Environment:', {
    NODE_ENV: process.env.NODE_ENV,
    RANCHER_URL: process.env.RANCHER_URL,
    RANCHER_TOKEN: process.env.RANCHER_TOKEN ? '***' : 'not set'
  });
}); 