import { test, expect } from '@playwright/test';

// Test configuration
const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';
const API_URL = process.env.API_URL || 'http://localhost:8080';

// Mock data for testing
const mockQueueMessages = [
  {
    id: '1a2b3c-4d5e6f-7g8h9i',
    sender: 'test@example.com',
    recipients: ['recipient1@example.com', 'recipient2@example.com'],
    size: 2048,
    age: '2h 30m',
    status: 'queued',
    retry_count: 0,
  },
  {
    id: '2b3c4d-5e6f7g-8h9i0j',
    sender: 'sender2@example.com',
    recipients: ['recipient3@example.com'],
    size: 1024,
    age: '1h 15m',
    status: 'deferred',
    retry_count: 2,
  },
  {
    id: '3c4d5e-6f7g8h-9i0j1k',
    sender: 'sender3@example.com',
    recipients: ['recipient4@example.com'],
    size: 4096,
    age: '30m',
    status: 'frozen',
    retry_count: 0,
  },
];

const mockQueueResponse = {
  success: true,
  data: {
    messages: mockQueueMessages,
    total_messages: 150,
    deferred_messages: 25,
    frozen_messages: 5,
    oldest_message_age: '2h 30m',
  },
  meta: {
    page: 1,
    per_page: 25,
    total: 150,
    total_pages: 6,
  },
};

test.describe('Queue Management', () => {
  test.beforeEach(async ({ page }) => {
    // Mock API responses
    await page.route(`${API_URL}/api/v1/queue*`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockQueueResponse),
      });
    });

    await page.route(`${API_URL}/api/v1/auth/me`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: { id: 1, username: 'testuser', email: 'test@example.com' },
        }),
      });
    });

    // Navigate to queue page
    await page.goto(`${BASE_URL}/queue`);
  });

  test('should display queue messages', async ({ page }) => {
    // Wait for the queue list to load
    await expect(page.locator('[data-testid="queue-list"]')).toBeVisible();

    // Check if messages are displayed
    for (const message of mockQueueMessages) {
      await expect(page.locator(`text=${message.id}`)).toBeVisible();
      await expect(page.locator(`text=${message.sender}`)).toBeVisible();
      await expect(page.locator(`text=${message.status}`)).toBeVisible();
    }
  });

  test('should display queue statistics', async ({ page }) => {
    // Check queue statistics
    await expect(page.locator('text=150 messages in queue')).toBeVisible();
    
    // Check individual statistics if displayed
    await expect(page.locator('text=25').first()).toBeVisible(); // deferred messages
    await expect(page.locator('text=5').first()).toBeVisible();  // frozen messages
  });

  test('should allow sorting by different columns', async ({ page }) => {
    // Click on sender column header to sort
    await page.click('[data-testid="sort-sender"]');
    
    // Verify API call was made with sort parameters
    await page.waitForRequest(request => 
      request.url().includes('sort_field=sender') && 
      request.url().includes('sort_direction=asc')
    );

    // Click again to reverse sort
    await page.click('[data-testid="sort-sender"]');
    
    await page.waitForRequest(request => 
      request.url().includes('sort_field=sender') && 
      request.url().includes('sort_direction=desc')
    );
  });

  test('should allow message selection', async ({ page }) => {
    // Select first message
    await page.check(`[data-testid="select-${mockQueueMessages[0].id}"]`);
    
    // Verify selection state
    await expect(page.locator(`[data-testid="select-${mockQueueMessages[0].id}"]`)).toBeChecked();
    
    // Check if bulk actions are enabled
    await expect(page.locator('[data-testid="bulk-actions"]')).toBeVisible();
  });

  test('should allow select all functionality', async ({ page }) => {
    // Click select all checkbox
    await page.check('[data-testid="select-all"]');
    
    // Verify all messages are selected
    for (const message of mockQueueMessages) {
      await expect(page.locator(`[data-testid="select-${message.id}"]`)).toBeChecked();
    }
    
    // Uncheck select all
    await page.uncheck('[data-testid="select-all"]');
    
    // Verify all messages are unselected
    for (const message of mockQueueMessages) {
      await expect(page.locator(`[data-testid="select-${message.id}"]`)).not.toBeChecked();
    }
  });

  test('should open message details', async ({ page }) => {
    // Mock message details response
    await page.route(`${API_URL}/api/v1/queue/${mockQueueMessages[0].id}`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            ...mockQueueMessages[0],
            headers: { 'From': 'test@example.com', 'To': 'recipient1@example.com' },
            envelope: { sender: 'test@example.com', recipients: ['recipient1@example.com'] },
          },
        }),
      });
    });

    // Click view button for first message
    await page.click(`[data-testid="view-${mockQueueMessages[0].id}"]`);
    
    // Verify message details modal/page opens
    await expect(page.locator('[data-testid="message-details"]')).toBeVisible();
    await expect(page.locator(`text=${mockQueueMessages[0].id}`)).toBeVisible();
  });

  test('should handle queue operations', async ({ page }) => {
    // Mock operation responses
    await page.route(`${API_URL}/api/v1/queue/*/deliver`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ success: true, message: 'Message delivered successfully' }),
      });
    });

    await page.route(`${API_URL}/api/v1/queue/*/freeze`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ success: true, message: 'Message frozen successfully' }),
      });
    });

    // Select a message
    await page.check(`[data-testid="select-${mockQueueMessages[0].id}"]`);
    
    // Test deliver operation
    await page.click('[data-testid="bulk-deliver"]');
    await expect(page.locator('text=Message delivered successfully')).toBeVisible();
    
    // Test freeze operation
    await page.click('[data-testid="bulk-freeze"]');
    await expect(page.locator('text=Message frozen successfully')).toBeVisible();
  });

  test('should handle search functionality', async ({ page }) => {
    // Mock search response
    await page.route(`${API_URL}/api/v1/queue/search`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            messages: [mockQueueMessages[0]], // Return filtered results
          },
          meta: { page: 1, per_page: 25, total: 1, total_pages: 1 },
        }),
      });
    });

    // Open search form
    await page.click('[data-testid="search-toggle"]');
    
    // Fill search criteria
    await page.fill('[data-testid="search-sender"]', 'test@example.com');
    await page.selectOption('[data-testid="search-status"]', 'queued');
    
    // Submit search
    await page.click('[data-testid="search-submit"]');
    
    // Verify search results
    await expect(page.locator(`text=${mockQueueMessages[0].id}`)).toBeVisible();
    await expect(page.locator(`text=${mockQueueMessages[1].id}`)).not.toBeVisible();
  });

  test('should handle pagination', async ({ page }) => {
    // Mock page 2 response
    await page.route(`${API_URL}/api/v1/queue*page=2*`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          ...mockQueueResponse,
          data: {
            ...mockQueueResponse.data,
            messages: [
              {
                id: '4d5e6f-7g8h9i-0j1k2l',
                sender: 'page2@example.com',
                recipients: ['page2recipient@example.com'],
                size: 1500,
                age: '45m',
                status: 'queued',
                retry_count: 1,
              },
            ],
          },
          meta: { page: 2, per_page: 25, total: 150, total_pages: 6 },
        }),
      });
    });

    // Click next page
    await page.click('[data-testid="pagination-next"]');
    
    // Verify page 2 content
    await expect(page.locator('text=4d5e6f-7g8h9i-0j1k2l')).toBeVisible();
    await expect(page.locator('text=page2@example.com')).toBeVisible();
  });

  test('should handle error states', async ({ page }) => {
    // Mock error response
    await page.route(`${API_URL}/api/v1/queue*`, async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: 'Internal server error',
        }),
      });
    });

    // Reload page to trigger error
    await page.reload();
    
    // Verify error message is displayed
    await expect(page.locator('text=Error loading queue')).toBeVisible();
    await expect(page.locator('text=Internal server error')).toBeVisible();
    
    // Test retry functionality
    await page.route(`${API_URL}/api/v1/queue*`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockQueueResponse),
      });
    });
    
    await page.click('[data-testid="retry-button"]');
    
    // Verify data loads after retry
    await expect(page.locator(`text=${mockQueueMessages[0].id}`)).toBeVisible();
  });

  test('should handle loading states', async ({ page }) => {
    // Mock slow response
    await page.route(`${API_URL}/api/v1/queue*`, async (route) => {
      await new Promise(resolve => setTimeout(resolve, 1000)); // 1 second delay
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockQueueResponse),
      });
    });

    // Reload page
    await page.reload();
    
    // Verify loading spinner is shown
    await expect(page.locator('[data-testid="loading-spinner"]')).toBeVisible();
    
    // Wait for data to load
    await expect(page.locator(`text=${mockQueueMessages[0].id}`)).toBeVisible();
    
    // Verify loading spinner is hidden
    await expect(page.locator('[data-testid="loading-spinner"]')).not.toBeVisible();
  });
});

test.describe('Queue Management - Virtual Scrolling', () => {
  test('should handle large datasets with virtual scrolling', async ({ page }) => {
    // Generate large dataset
    const largeDataset = Array.from({ length: 1000 }, (_, i) => ({
      id: `msg-${i.toString().padStart(6, '0')}`,
      sender: `sender${i}@example.com`,
      recipients: [`recipient${i}@example.com`],
      size: 1000 + i,
      age: `${Math.floor(i / 60)}h ${i % 60}m`,
      status: ['queued', 'deferred', 'frozen'][i % 3],
      retry_count: i % 5,
    }));

    // Mock large dataset response
    await page.route(`${API_URL}/api/v1/queue*`, async (route) => {
      const url = new URL(route.request().url());
      const page_num = parseInt(url.searchParams.get('page') || '1');
      const per_page = parseInt(url.searchParams.get('per_page') || '100');
      
      const start = (page_num - 1) * per_page;
      const end = start + per_page;
      const pageData = largeDataset.slice(start, end);

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            messages: pageData,
            total_messages: largeDataset.length,
            deferred_messages: Math.floor(largeDataset.length / 3),
            frozen_messages: Math.floor(largeDataset.length / 3),
            oldest_message_age: '24h',
          },
          meta: {
            page: page_num,
            per_page,
            total: largeDataset.length,
            total_pages: Math.ceil(largeDataset.length / per_page),
          },
        }),
      });
    });

    // Navigate to queue with virtualization enabled
    await page.goto(`${BASE_URL}/queue?virtualization=true`);
    
    // Verify virtual scrolling container is present
    await expect(page.locator('[data-testid="virtual-list"]')).toBeVisible();
    
    // Verify only visible items are rendered (not all 1000)
    const renderedItems = await page.locator('[data-testid^="queue-item-"]').count();
    expect(renderedItems).toBeLessThan(50); // Should render much fewer than total
    
    // Test scrolling behavior
    await page.evaluate(() => {
      const virtualList = document.querySelector('[data-testid="virtual-list"]');
      if (virtualList) {
        virtualList.scrollTop = 5000; // Scroll down significantly
      }
    });
    
    // Wait for new items to be rendered
    await page.waitForTimeout(500);
    
    // Verify different items are now visible
    await expect(page.locator('text=msg-000100')).toBeVisible();
  });
});

test.describe('Queue Management - Performance', () => {
  test('should maintain performance with frequent updates', async ({ page }) => {
    let requestCount = 0;
    
    // Mock WebSocket updates
    await page.addInitScript(() => {
      // Mock WebSocket for real-time updates
      window.mockWebSocket = {
        send: () => {},
        close: () => {},
        addEventListener: (event: string, callback: Function) => {
          if (event === 'message') {
            // Simulate frequent updates
            setInterval(() => {
              callback({
                data: JSON.stringify({
                  type: 'queue_update',
                  data: { updated_messages: 1 }
                })
              });
            }, 1000);
          }
        }
      };
    });

    // Track API requests
    await page.route(`${API_URL}/api/v1/queue*`, async (route) => {
      requestCount++;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockQueueResponse),
      });
    });

    await page.goto(`${BASE_URL}/queue`);
    
    // Wait for initial load
    await expect(page.locator('[data-testid="queue-list"]')).toBeVisible();
    
    const initialRequestCount = requestCount;
    
    // Wait for several update cycles
    await page.waitForTimeout(5000);
    
    // Verify requests are not excessive (should be throttled/debounced)
    const finalRequestCount = requestCount;
    const additionalRequests = finalRequestCount - initialRequestCount;
    
    // Should not make more than 1 request per second on average
    expect(additionalRequests).toBeLessThan(10);
  });

  test('should handle rapid user interactions efficiently', async ({ page }) => {
    await page.goto(`${BASE_URL}/queue`);
    
    // Measure time for rapid selections
    const startTime = Date.now();
    
    // Rapidly select/deselect messages
    for (let i = 0; i < mockQueueMessages.length; i++) {
      await page.check(`[data-testid="select-${mockQueueMessages[i].id}"]`);
      await page.uncheck(`[data-testid="select-${mockQueueMessages[i].id}"]`);
    }
    
    const endTime = Date.now();
    const duration = endTime - startTime;
    
    // Should complete rapid interactions within reasonable time
    expect(duration).toBeLessThan(2000); // 2 seconds for all interactions
    
    // Verify UI remains responsive
    await expect(page.locator('[data-testid="queue-list"]')).toBeVisible();
  });
});