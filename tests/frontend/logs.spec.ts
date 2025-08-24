import { test, expect } from '@playwright/test';

// Test configuration
const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';
const API_URL = process.env.API_URL || 'http://localhost:8080';

// Mock log data
const mockLogEntries = [
  {
    id: 1,
    timestamp: '2024-01-01T12:00:00Z',
    message_id: '1a2b3c-4d5e6f-7g8h9i',
    log_type: 'main',
    event: 'arrival',
    host: 'localhost',
    sender: 'test@example.com',
    recipients: ['recipient1@example.com'],
    size: 2048,
    status: 'received',
    error_code: null,
    error_text: null,
    raw_line: '2024-01-01 12:00:00 1a2b3c-4d5e6f-7g8h9i <= test@example.com H=localhost [127.0.0.1] P=esmtp S=2048',
  },
  {
    id: 2,
    timestamp: '2024-01-01T12:01:00Z',
    message_id: '1a2b3c-4d5e6f-7g8h9i',
    log_type: 'main',
    event: 'delivery',
    host: 'remote.example.com',
    sender: 'test@example.com',
    recipients: ['recipient1@example.com'],
    size: 2048,
    status: 'delivered',
    error_code: null,
    error_text: null,
    raw_line: '2024-01-01 12:01:00 1a2b3c-4d5e6f-7g8h9i => recipient1@example.com R=remote_smtp T=remote_smtp',
  },
  {
    id: 3,
    timestamp: '2024-01-01T12:02:00Z',
    message_id: '2b3c4d-5e6f7g-8h9i0j',
    log_type: 'main',
    event: 'defer',
    host: 'remote.example.com',
    sender: 'sender2@example.com',
    recipients: ['recipient2@example.com'],
    size: 1024,
    status: 'deferred',
    error_code: '451',
    error_text: 'Temporary failure',
    raw_line: '2024-01-01 12:02:00 2b3c4d-5e6f7g-8h9i0j == recipient2@example.com R=remote_smtp defer (-44): SMTP error',
  },
  {
    id: 4,
    timestamp: '2024-01-01T12:03:00Z',
    message_id: '3c4d5e-6f7g8h-9i0j1k',
    log_type: 'reject',
    event: 'reject',
    host: 'spammer.example.com',
    sender: 'spam@spammer.example.com',
    recipients: ['victim@example.com'],
    size: 512,
    status: 'rejected',
    error_code: '550',
    error_text: 'Rejected by policy',
    raw_line: '2024-01-01 12:03:00 rejected SMTP connection from spammer.example.com [192.168.1.100]',
  },
];

const mockLogsResponse = {
  success: true,
  data: {
    entries: mockLogEntries,
    search_time: '0.025s',
    aggregations: {
      by_log_type: { main: 3, reject: 1 },
      by_event: { arrival: 1, delivery: 1, defer: 1, reject: 1 },
    },
  },
  meta: {
    page: 1,
    per_page: 50,
    total: 4,
    total_pages: 1,
  },
};

test.describe('Log Viewer', () => {
  test.beforeEach(async ({ page }) => {
    // Mock API responses
    await page.route(`${API_URL}/api/v1/logs*`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockLogsResponse),
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

    // Navigate to logs page
    await page.goto(`${BASE_URL}/logs`);
  });

  test('should display log entries', async ({ page }) => {
    // Wait for the log viewer to load
    await expect(page.locator('[data-testid="log-viewer"]')).toBeVisible();

    // Check if log entries are displayed
    for (const entry of mockLogEntries) {
      await expect(page.locator(`text=${entry.message_id}`)).toBeVisible();
      await expect(page.locator(`text=${entry.log_type}`)).toBeVisible();
      await expect(page.locator(`text=${entry.event}`)).toBeVisible();
    }
  });

  test('should display log entry details correctly', async ({ page }) => {
    // Check timestamp formatting
    await expect(page.locator('text=1/1/2024, 12:00:00 PM')).toBeVisible();
    
    // Check log type badges with correct colors
    await expect(page.locator('[data-testid="log-type-main"]')).toHaveClass(/bg-blue-100/);
    await expect(page.locator('[data-testid="log-type-reject"]')).toHaveClass(/bg-red-100/);
    
    // Check event badges with correct colors
    await expect(page.locator('[data-testid="event-arrival"]')).toHaveClass(/bg-green-100/);
    await expect(page.locator('[data-testid="event-delivery"]')).toHaveClass(/bg-blue-100/);
    await expect(page.locator('[data-testid="event-defer"]')).toHaveClass(/bg-yellow-100/);
    await expect(page.locator('[data-testid="event-reject"]')).toHaveClass(/bg-red-100/);
  });

  test('should allow log entry selection', async ({ page }) => {
    // Select first log entry
    await page.check(`[data-testid="select-log-${mockLogEntries[0].id}"]`);
    
    // Verify selection state
    await expect(page.locator(`[data-testid="select-log-${mockLogEntries[0].id}"]`)).toBeChecked();
    
    // Check if export button shows count
    await expect(page.locator('text=Export Selected (1)')).toBeVisible();
  });

  test('should allow select all functionality', async ({ page }) => {
    // Click select all checkbox
    await page.check('[data-testid="select-all-logs"]');
    
    // Verify all log entries are selected
    for (const entry of mockLogEntries) {
      await expect(page.locator(`[data-testid="select-log-${entry.id}"]`)).toBeChecked();
    }
    
    // Check export button shows total count
    await expect(page.locator(`text=Export Selected (${mockLogEntries.length})`)).toBeVisible();
    
    // Uncheck select all
    await page.uncheck('[data-testid="select-all-logs"]');
    
    // Verify all entries are unselected
    for (const entry of mockLogEntries) {
      await expect(page.locator(`[data-testid="select-log-${entry.id}"]`)).not.toBeChecked();
    }
  });

  test('should handle log search functionality', async ({ page }) => {
    // Mock search response
    await page.route(`${API_URL}/api/v1/logs/search`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            entries: [mockLogEntries[0]], // Return filtered results
            search_time: '0.015s',
            aggregations: { by_log_type: { main: 1 } },
          },
          meta: { page: 1, per_page: 50, total: 1, total_pages: 1 },
        }),
      });
    });

    // Open search form
    await page.click('[data-testid="search-toggle"]');
    
    // Fill search criteria
    await page.fill('[data-testid="search-message-id"]', '1a2b3c-4d5e6f-7g8h9i');
    await page.selectOption('[data-testid="search-log-type"]', 'main');
    await page.selectOption('[data-testid="search-event"]', 'arrival');
    
    // Submit search
    await page.click('[data-testid="search-submit"]');
    
    // Verify search results
    await expect(page.locator(`text=${mockLogEntries[0].message_id}`)).toBeVisible();
    await expect(page.locator(`text=${mockLogEntries[1].message_id}`)).toHaveCount(0);
  });

  test('should handle date range filtering', async ({ page }) => {
    // Mock date range search response
    await page.route(`${API_URL}/api/v1/logs*start_time*`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            entries: mockLogEntries.slice(0, 2), // Return subset
            search_time: '0.020s',
            aggregations: { by_log_type: { main: 2 } },
          },
          meta: { page: 1, per_page: 50, total: 2, total_pages: 1 },
        }),
      });
    });

    // Open search form
    await page.click('[data-testid="search-toggle"]');
    
    // Set date range
    await page.fill('[data-testid="search-start-date"]', '2024-01-01T12:00');
    await page.fill('[data-testid="search-end-date"]', '2024-01-01T12:01');
    
    // Submit search
    await page.click('[data-testid="search-submit"]');
    
    // Verify filtered results
    await expect(page.locator('text=2 log entries')).toBeVisible();
  });

  test('should export selected logs', async ({ page }) => {
    // Select some log entries
    await page.check(`[data-testid="select-log-${mockLogEntries[0].id}"]`);
    await page.check(`[data-testid="select-log-${mockLogEntries[1].id}"]`);
    
    // Mock download
    const downloadPromise = page.waitForEvent('download');
    
    // Click export button
    await page.click('[data-testid="export-selected"]');
    
    // Verify download starts
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/exim-logs-\d{4}-\d{2}-\d{2}\.csv/);
  });

  test('should handle real-time log tail', async ({ page }) => {
    // Mock WebSocket connection for real-time logs
    await page.addInitScript(() => {
      window.mockWebSocket = {
        send: () => {},
        close: () => {},
        addEventListener: (event: string, callback: Function) => {
          if (event === 'message') {
            // Simulate new log entries
            setTimeout(() => {
              callback({
                data: JSON.stringify({
                  type: 'log_entry',
                  data: {
                    id: 5,
                    timestamp: '2024-01-01T12:04:00Z',
                    message_id: '4d5e6f-7g8h9i-0j1k2l',
                    log_type: 'main',
                    event: 'arrival',
                    sender: 'new@example.com',
                    raw_line: '2024-01-01 12:04:00 new log entry',
                  }
                })
              });
            }, 1000);
          }
        }
      };
    });

    // Navigate to real-time tail
    await page.click('[data-testid="real-time-tail"]');
    
    // Verify real-time tail is active
    await expect(page.locator('[data-testid="tail-status"]')).toContainText('Connected');
    
    // Wait for new log entry to appear
    await expect(page.locator('text=4d5e6f-7g8h9i-0j1k2l')).toBeVisible({ timeout: 2000 });
    await expect(page.locator('text=new@example.com')).toBeVisible();
  });

  test('should handle error states', async ({ page }) => {
    // Mock error response
    await page.route(`${API_URL}/api/v1/logs*`, async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: 'Database connection failed',
        }),
      });
    });

    // Reload page to trigger error
    await page.reload();
    
    // Verify error message is displayed
    await expect(page.locator('text=Error loading logs')).toBeVisible();
    await expect(page.locator('text=Database connection failed')).toBeVisible();
    
    // Test retry functionality
    await page.route(`${API_URL}/api/v1/logs*`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockLogsResponse),
      });
    });
    
    await page.click('[data-testid="retry-button"]');
    
    // Verify data loads after retry
    await expect(page.locator(`text=${mockLogEntries[0].message_id}`)).toBeVisible();
  });

  test('should handle pagination', async ({ page }) => {
    // Mock paginated response
    const paginatedResponse = {
      ...mockLogsResponse,
      data: {
        ...mockLogsResponse.data,
        entries: [
          {
            id: 5,
            timestamp: '2024-01-01T12:05:00Z',
            message_id: '5e6f7g-8h9i0j-1k2l3m',
            log_type: 'main',
            event: 'delivery',
            sender: 'page2@example.com',
            recipients: ['page2recipient@example.com'],
            raw_line: '2024-01-01 12:05:00 page 2 log entry',
          },
        ],
      },
      meta: { page: 2, per_page: 50, total: 100, total_pages: 2 },
    };

    await page.route(`${API_URL}/api/v1/logs*page=2*`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(paginatedResponse),
      });
    });

    // Click next page
    await page.click('[data-testid="pagination-next"]');
    
    // Verify page 2 content
    await expect(page.locator('text=5e6f7g-8h9i0j-1k2l3m')).toBeVisible();
    await expect(page.locator('text=page2@example.com')).toBeVisible();
  });

  test('should display log statistics', async ({ page }) => {
    // Verify aggregation statistics are displayed
    await expect(page.locator('text=Search completed in 0.025s')).toBeVisible();
    
    // Check aggregation data if displayed
    await expect(page.locator('[data-testid="log-stats"]')).toBeVisible();
  });
});

test.describe('Log Viewer - Virtual Scrolling', () => {
  test('should handle large log datasets with virtual scrolling', async ({ page }) => {
    // Generate large dataset
    const largeLogDataset = Array.from({ length: 10000 }, (_, i) => ({
      id: i + 1,
      timestamp: new Date(Date.now() - i * 1000).toISOString(),
      message_id: `msg-${i.toString().padStart(6, '0')}`,
      log_type: ['main', 'reject', 'panic'][i % 3],
      event: ['arrival', 'delivery', 'defer', 'bounce'][i % 4],
      sender: `sender${i}@example.com`,
      recipients: [`recipient${i}@example.com`],
      raw_line: `2024-01-01 12:00:${(i % 60).toString().padStart(2, '0')} log entry ${i}`,
    }));

    // Mock large dataset response
    await page.route(`${API_URL}/api/v1/logs*`, async (route) => {
      const url = new URL(route.request().url());
      const page_num = parseInt(url.searchParams.get('page') || '1');
      const per_page = parseInt(url.searchParams.get('per_page') || '100');
      
      const start = (page_num - 1) * per_page;
      const end = start + per_page;
      const pageData = largeLogDataset.slice(start, end);

      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            entries: pageData,
            search_time: '0.050s',
            aggregations: { by_log_type: { main: 3333, reject: 3333, panic: 3334 } },
          },
          meta: {
            page: page_num,
            per_page,
            total: largeLogDataset.length,
            total_pages: Math.ceil(largeLogDataset.length / per_page),
          },
        }),
      });
    });

    // Navigate to logs with virtualization enabled
    await page.goto(`${BASE_URL}/logs?virtualization=true`);
    
    // Verify virtual scrolling container is present
    await expect(page.locator('[data-testid="virtual-log-list"]')).toBeVisible();
    
    // Verify only visible items are rendered
    const renderedItems = await page.locator('[data-testid^="log-entry-"]').count();
    expect(renderedItems).toBeLessThan(100); // Should render much fewer than total
    
    // Test scrolling behavior
    await page.evaluate(() => {
      const virtualList = document.querySelector('[data-testid="virtual-log-list"]');
      if (virtualList) {
        virtualList.scrollTop = 10000; // Scroll down significantly
      }
    });
    
    // Wait for new items to be rendered
    await page.waitForTimeout(500);
    
    // Verify different items are now visible
    await expect(page.locator('text=msg-000200')).toBeVisible();
  });
});

test.describe('Log Viewer - Performance', () => {
  test('should maintain performance with frequent log updates', async ({ page }) => {
    let updateCount = 0;
    
    // Mock frequent WebSocket updates
    await page.addInitScript(() => {
      window.mockWebSocket = {
        send: () => {},
        close: () => {},
        addEventListener: (event: string, callback: Function) => {
          if (event === 'message') {
            // Simulate frequent log updates
            const interval = setInterval(() => {
              callback({
                data: JSON.stringify({
                  type: 'log_entry',
                  data: {
                    id: Date.now(),
                    timestamp: new Date().toISOString(),
                    message_id: `realtime-${Date.now()}`,
                    log_type: 'main',
                    event: 'arrival',
                    sender: 'realtime@example.com',
                    raw_line: `${new Date().toISOString()} realtime log entry`,
                  }
                })
              });
            }, 100); // Very frequent updates
            
            // Stop after 5 seconds
            setTimeout(() => clearInterval(interval), 5000);
          }
        }
      };
    });

    await page.goto(`${BASE_URL}/logs`);
    
    // Enable real-time tail
    await page.click('[data-testid="real-time-tail"]');
    
    // Measure performance during updates
    const startTime = Date.now();
    
    // Wait for updates to process
    await page.waitForTimeout(6000);
    
    const endTime = Date.now();
    const duration = endTime - startTime;
    
    // Verify UI remains responsive
    await expect(page.locator('[data-testid="log-viewer"]')).toBeVisible();
    
    // Check that the page didn't freeze (should complete within reasonable time)
    expect(duration).toBeLessThan(7000); // Should not take much longer than expected
  });

  test('should handle rapid search operations efficiently', async ({ page }) => {
    await page.goto(`${BASE_URL}/logs`);
    
    // Open search form
    await page.click('[data-testid="search-toggle"]');
    
    // Measure time for rapid search operations
    const startTime = Date.now();
    
    // Rapidly change search criteria
    const searchTerms = ['test', 'example', 'arrival', 'delivery', 'defer'];
    
    for (const term of searchTerms) {
      await page.fill('[data-testid="search-message-id"]', term);
      await page.waitForTimeout(100); // Small delay between changes
    }
    
    const endTime = Date.now();
    const duration = endTime - startTime;
    
    // Should handle rapid changes efficiently
    expect(duration).toBeLessThan(2000);
    
    // Verify UI remains responsive
    await expect(page.locator('[data-testid="search-form"]')).toBeVisible();
  });
});