import { chromium, FullConfig } from '@playwright/test';

async function globalSetup(config: FullConfig) {
  console.log('🚀 Starting global test setup...');

  // Start backend server if needed
  if (process.env.START_BACKEND === 'true') {
    console.log('Starting backend server...');
    // In a real implementation, you would start your Go backend here
    // For now, we assume it's already running or will be mocked
  }

  // Create a browser instance for setup tasks
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();

  try {
    // Perform any global setup tasks
    console.log('Performing global setup tasks...');

    // Example: Create test user, seed data, etc.
    // This would typically involve API calls to set up test data

    // Health check - verify services are running
    const baseURL = process.env.BASE_URL || 'http://localhost:3000';
    const apiURL = process.env.API_URL || 'http://localhost:8080';

    try {
      // Check if frontend is accessible
      await page.goto(baseURL, { timeout: 10000 });
      console.log('✅ Frontend server is accessible');
    } catch (error) {
      console.warn('⚠️  Frontend server not accessible, tests will use mocked responses');
    }

    try {
      // Check if backend API is accessible
      const response = await page.request.get(`${apiURL}/api/v1/health`);
      if (response.ok()) {
        console.log('✅ Backend API is accessible');
      } else {
        console.warn('⚠️  Backend API not accessible, tests will use mocked responses');
      }
    } catch (error) {
      console.warn('⚠️  Backend API not accessible, tests will use mocked responses');
    }

    // Set up test data storage
    const testDataPath = './test_results/test-data.json';
    const testData = {
      setupTime: new Date().toISOString(),
      testUsers: [
        {
          username: 'testuser',
          email: 'test@example.com',
          password: 'testpass123',
        },
      ],
      mockData: {
        queueMessages: 150,
        logEntries: 10000,
        users: 5,
      },
    };

    // Save test data for use in tests
    const fs = require('fs');
    const path = require('path');
    
    // Ensure directory exists
    const dir = path.dirname(testDataPath);
    if (!fs.existsSync(dir)) {
      fs.mkdirSync(dir, { recursive: true });
    }
    
    fs.writeFileSync(testDataPath, JSON.stringify(testData, null, 2));
    console.log('✅ Test data configuration saved');

  } catch (error) {
    console.error('❌ Global setup failed:', error);
    throw error;
  } finally {
    await context.close();
    await browser.close();
  }

  console.log('✅ Global test setup completed');
}

export default globalSetup;