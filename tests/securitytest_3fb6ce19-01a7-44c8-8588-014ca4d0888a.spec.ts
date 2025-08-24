
import { test } from '@playwright/test';
import { expect } from '@playwright/test';

test('SecurityTest_2025-08-24', async ({ page, context }) => {
  
    // Navigate to URL
    await page.goto('http://localhost:8080/api/v1/health');
});