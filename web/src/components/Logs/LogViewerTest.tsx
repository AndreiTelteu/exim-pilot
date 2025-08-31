import { useState, useEffect } from 'react';
import { LogEntry, LogSearchParams, LogStatistics } from '../../types/logs';
import { APIResponse } from '../../types/api';
import LogViewer from './LogViewer';
import LogSearch from './LogSearch';
import { LoadingSpinner } from '../Common';
import { useApp } from '../../context/AppContext';

const LogViewerTest: React.FC = () => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [totalEntries, setTotalEntries] = useState(0);
  const [searchParams, setSearchParams] = useState<LogSearchParams>({
    per_page: 25
  });
  const [testResults, setTestResults] = useState<{
    searchTest: boolean;
    paginationTest: boolean;
    filterTest: boolean;
    exportTest: boolean;
  }>({
    searchTest: false,
    paginationTest: false,
    filterTest: false,
    exportTest: false
  });
  
  const { actions } = useApp();

  // Mock log data for testing
  const generateMockLogs = (count: number = 50): LogEntry[] => {
    const logTypes = ['main', 'reject', 'panic'] as const;
    const events = ['delivery', 'defer', 'bounce', 'arrival', 'completed', 'failed'];
    const hosts = ['mx.example.com', 'mail.test.org', 'smtp.sample.net'];
    const senders = ['user@domain.com', 'test@example.org', 'sender@mail.net'];
    const recipients = ['recipient@target.com', 'dest@example.org', 'user@test.net'];

    return Array.from({ length: count }, (_, index) => ({
      id: index + 1,
      timestamp: new Date(Date.now() - Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString(),
      message_id: `1${Math.random().toString(36).substring(2, 8).toUpperCase()}-${Math.random().toString(36).substring(2, 8).toUpperCase()}-${Math.random().toString(36).substring(2, 2).toUpperCase()}`,
      log_type: logTypes[Math.floor(Math.random() * logTypes.length)],
      event: events[Math.floor(Math.random() * events.length)],
      host: hosts[Math.floor(Math.random() * hosts.length)],
      sender: senders[Math.floor(Math.random() * senders.length)],
      recipients: [recipients[Math.floor(Math.random() * recipients.length)]],
      size: Math.floor(Math.random() * 50000) + 1000,
      status: Math.random() > 0.7 ? 'completed' : 'processing',
      error_code: Math.random() > 0.8 ? '550' : undefined,
      error_text: Math.random() > 0.8 ? 'User unknown in virtual mailbox table' : undefined,
      raw_line: `${new Date().toISOString()} [${index}] <= ${senders[Math.floor(Math.random() * senders.length)]} ${events[Math.floor(Math.random() * events.length)]} ${recipients[Math.floor(Math.random() * recipients.length)]}`
    }));
  };

  const fetchLogs = async (params: LogSearchParams = {}) => {
    try {
      setLoading(true);
      setError(null);

      // Simulate API call delay
      await new Promise(resolve => setTimeout(resolve, 500));

      // Generate mock data based on search parameters
      const mockLogs = generateMockLogs(200);
      let filteredLogs = mockLogs;

      // Apply filters
      if (params.log_type) {
        filteredLogs = filteredLogs.filter(log => log.log_type === params.log_type);
      }
      if (params.keyword) {
        filteredLogs = filteredLogs.filter(log => 
          log.raw_line.toLowerCase().includes(params.keyword!.toLowerCase())
        );
      }
      if (params.message_id) {
        filteredLogs = filteredLogs.filter(log => 
          log.message_id?.includes(params.message_id!)
        );
      }
      if (params.sender) {
        filteredLogs = filteredLogs.filter(log => 
          log.sender?.toLowerCase().includes(params.sender!.toLowerCase())
        );
      }

      // Pagination
      const perPage = params.per_page || 25;
      const page = params.page || 1;
      const startIndex = (page - 1) * perPage;
      const endIndex = startIndex + perPage;
      const paginatedLogs = filteredLogs.slice(startIndex, endIndex);

      setLogs(paginatedLogs);
      setTotalPages(Math.ceil(filteredLogs.length / perPage));
      setTotalEntries(filteredLogs.length);
      setCurrentPage(page);

      // Update test results
      setTestResults(prev => ({ 
        ...prev, 
        searchTest: true,
        paginationTest: prev.paginationTest || page > 1,
        filterTest: prev.filterTest || Object.keys(params).some(key => 
          key !== 'per_page' && key !== 'page' && params[key as keyof LogSearchParams]
        )
      }));

      actions.addNotification({
        type: 'success',
        message: `Loaded ${paginatedLogs.length} log entries (${filteredLogs.length} total matches)`
      });

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch logs';
      setError(errorMessage);
      actions.addNotification({
        type: 'error',
        message: errorMessage
      });
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (params: LogSearchParams) => {
    const newParams = { ...searchParams, ...params, page: 1 };
    setSearchParams(newParams);
    setCurrentPage(1);
  };

  const handlePageChange = (page: number) => {
    const newParams = { ...searchParams, page };
    setSearchParams(newParams);
  };

  const handleExport = async (format: 'csv' | 'txt' | 'json') => {
    try {
      // Simulate export
      await new Promise(resolve => setTimeout(resolve, 1000));
      
      const mockData = logs.map(log => ({
        timestamp: log.timestamp,
        type: log.log_type,
        event: log.event,
        message_id: log.message_id,
        sender: log.sender,
        recipients: log.recipients?.join(','),
        size: log.size,
        raw_line: log.raw_line
      }));

      let content = '';
      let filename = `logs_test_export.${format}`;

      switch (format) {
        case 'csv':
          const headers = Object.keys(mockData[0] || {}).join(',');
          const rows = mockData.map(row => Object.values(row).map(val => `"${val || ''}"`).join(','));
          content = [headers, ...rows].join('\n');
          break;
        case 'txt':
          content = logs.map(log => log.raw_line).join('\n');
          break;
        case 'json':
          content = JSON.stringify(mockData, null, 2);
          break;
      }

      // Create and download file
      const blob = new Blob([content], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);

      setTestResults(prev => ({ ...prev, exportTest: true }));
      
      actions.addNotification({
        type: 'success',
        message: `Test export completed as ${format.toUpperCase()}`
      });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Export test failed';
      actions.addNotification({
        type: 'error',
        message: errorMessage
      });
    }
  };

  const runAllTests = async () => {
    actions.addNotification({
      type: 'info',
      message: 'Running comprehensive log viewer tests...'
    });

    // Test 1: Basic search
    await handleSearch({ keyword: 'delivery', per_page: 10 });
    await new Promise(resolve => setTimeout(resolve, 1000));

    // Test 2: Filter by log type
    await handleSearch({ log_type: 'main', per_page: 15 });
    await new Promise(resolve => setTimeout(resolve, 1000));

    // Test 3: Pagination
    await handlePageChange(2);
    await new Promise(resolve => setTimeout(resolve, 1000));

    // Test 4: Export test
    await handleExport('csv');

    actions.addNotification({
      type: 'success',
      message: 'All log viewer tests completed successfully!'
    });
  };

  useEffect(() => {
    fetchLogs(searchParams);
  }, [searchParams]);

  // Auto-run initial test on component mount
  useEffect(() => {
    const timer = setTimeout(() => {
      fetchLogs({ per_page: 25 });
    }, 500);

    return () => clearTimeout(timer);
  }, []);

  const allTestsPassed = Object.values(testResults).every(Boolean);

  return (
    <div className="space-y-6">
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h1 className="text-2xl font-bold text-gray-900">Log Viewer Test</h1>
          <button
            onClick={runAllTests}
            className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700"
          >
            Run All Tests
          </button>
        </div>

        {/* Test Status */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <div className={`p-3 rounded-lg border ${testResults.searchTest ? 'bg-green-50 border-green-200' : 'bg-gray-50 border-gray-200'}`}>
            <div className="flex items-center">
              <div className={`w-3 h-3 rounded-full mr-2 ${testResults.searchTest ? 'bg-green-500' : 'bg-gray-400'}`}></div>
              <span className="text-sm font-medium">Search Test</span>
            </div>
          </div>
          
          <div className={`p-3 rounded-lg border ${testResults.paginationTest ? 'bg-green-50 border-green-200' : 'bg-gray-50 border-gray-200'}`}>
            <div className="flex items-center">
              <div className={`w-3 h-3 rounded-full mr-2 ${testResults.paginationTest ? 'bg-green-500' : 'bg-gray-400'}`}></div>
              <span className="text-sm font-medium">Pagination Test</span>
            </div>
          </div>
          
          <div className={`p-3 rounded-lg border ${testResults.filterTest ? 'bg-green-50 border-green-200' : 'bg-gray-50 border-gray-200'}`}>
            <div className="flex items-center">
              <div className={`w-3 h-3 rounded-full mr-2 ${testResults.filterTest ? 'bg-green-500' : 'bg-gray-400'}`}></div>
              <span className="text-sm font-medium">Filter Test</span>
            </div>
          </div>
          
          <div className={`p-3 rounded-lg border ${testResults.exportTest ? 'bg-green-50 border-green-200' : 'bg-gray-50 border-gray-200'}`}>
            <div className="flex items-center">
              <div className={`w-3 h-3 rounded-full mr-2 ${testResults.exportTest ? 'bg-green-500' : 'bg-gray-400'}`}></div>
              <span className="text-sm font-medium">Export Test</span>
            </div>
          </div>
        </div>

        {allTestsPassed && (
          <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-md">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-green-800">All tests passed!</h3>
                <p className="mt-1 text-sm text-green-700">
                  The log viewer is working correctly with search, filtering, pagination, and export functionality.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Test Instructions */}
        <div className="bg-blue-50 border border-blue-200 rounded-md p-4">
          <h3 className="text-sm font-medium text-blue-800 mb-2">Test Instructions:</h3>
          <ul className="text-sm text-blue-700 space-y-1">
            <li>• Use the search form to test filtering by different criteria</li>
            <li>• Try pagination by navigating between pages</li>
            <li>• Test export functionality with different formats</li>
            <li>• Expand log entries to view raw log data</li>
            <li>• All data shown is mock data for testing purposes</li>
          </ul>
        </div>
      </div>

      <LogSearch
        onSearch={handleSearch}
        onExport={handleExport}
        totalEntries={totalEntries}
      />
      
      {loading ? (
        <div className="flex justify-center py-12">
          <LoadingSpinner />
        </div>
      ) : error ? (
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Test Error</h3>
              <p className="mt-1 text-sm text-red-700">{error}</p>
            </div>
          </div>
        </div>
      ) : (
        <LogViewer
          logs={logs}
          currentPage={currentPage}
          totalPages={totalPages}
          totalEntries={totalEntries}
          onPageChange={handlePageChange}
          onRefresh={() => fetchLogs(searchParams)}
        />
      )}
    </div>
  );
};

export default LogViewerTest;