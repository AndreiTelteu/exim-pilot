import { FullConfig } from '@playwright/test';

async function globalTeardown(config: FullConfig) {
  console.log('🧹 Starting global test teardown...');

  try {
    // Clean up test data
    console.log('Cleaning up test data...');

    // Remove temporary test files
    const fs = require('fs');
    const path = require('path');
    
    const testDataPath = './test_results/test-data.json';
    if (fs.existsSync(testDataPath)) {
      fs.unlinkSync(testDataPath);
      console.log('✅ Test data file cleaned up');
    }

    // Clean up any temporary databases or files created during tests
    const tempFiles = [
      './test_results/temp_*.db',
      './test_results/temp_*.log',
    ];

    for (const pattern of tempFiles) {
      const glob = require('glob');
      const files = glob.sync(pattern);
      for (const file of files) {
        try {
          fs.unlinkSync(file);
          console.log(`✅ Cleaned up temporary file: ${file}`);
        } catch (error) {
          console.warn(`⚠️  Could not clean up file ${file}:`, error.message);
        }
      }
    }

    // Stop backend server if we started it
    if (process.env.START_BACKEND === 'true') {
      console.log('Stopping backend server...');
      // In a real implementation, you would stop your Go backend here
    }

    // Generate test summary
    const summaryPath = './test_results/test-summary.json';
    const summary = {
      teardownTime: new Date().toISOString(),
      testRun: {
        completed: true,
        duration: Date.now() - (global.testStartTime || Date.now()),
      },
      cleanup: {
        tempFilesRemoved: true,
        testDataCleaned: true,
      },
    };

    fs.writeFileSync(summaryPath, JSON.stringify(summary, null, 2));
    console.log('✅ Test summary saved');

  } catch (error) {
    console.error('❌ Global teardown failed:', error);
    // Don't throw error in teardown to avoid masking test failures
  }

  console.log('✅ Global test teardown completed');
}

export default globalTeardown;