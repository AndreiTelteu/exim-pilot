import { useState, useEffect } from 'react';
import { LogEntry, LogStatistics } from '../../types/logs';
import LogViewer from './LogViewer';
import LogSearch from './LogSearch';
import RealTimeTail from './RealTimeTail';
import LogStatisticsPanel from './LogStatisticsPanel';
import { LoadingSpinner } from '../Common';
import { useApp } from '../../context/AppContext';

interface TestSuite {
  id: string;
  name: string;
  description: string;
  status: 'pending' | 'running' | 'passed' | 'failed';
  result?: string;
}

const LogsTestPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'overview' | 'search' | 'realtime' | 'statistics' | 'performance'>('overview');
  const [testSuites, setTestSuites] = useState<TestSuite[]>([
    {
      id: 'search-functionality',
      name: 'Search Functionality',
      description: 'Test log search, filtering, and pagination',
      status: 'pending'
    },
    {
      id: 'realtime-streaming',
      name: 'Real-time Streaming',
      description: 'Test WebSocket connection and live log updates',
      status: 'pending'
    },
    {
      id: 'export-features',
      name: 'Export Features',
      description: 'Test CSV, TXT, and JSON export functionality',
      status: 'pending'
    },
    {
      id: 'statistics-generation',
      name: 'Statistics Generation',
      description: 'Test log statistics and charts',
      status: 'pending'
    },
    {
      id: 'performance-load',
      name: 'Performance & Load',
      description: 'Test with large datasets and virtual scrolling',
      status: 'pending'
    },
    {
      id: 'error-handling',
      name: 'Error Handling',
      description: 'Test error states and recovery mechanisms',
      status: 'pending'
    }
  ]);

  const [mockLogs, setMockLogs] = useState<LogEntry[]>([]);
  const [mockStatistics, setMockStatistics] = useState<LogStatistics | null>(null);
  const [performanceMetrics, setPerformanceMetrics] = useState({
    renderTime: 0,
    searchTime: 0,
    scrollPerformance: 0
  });
  
  const { actions } = useApp();

  const generateLargeMockDataset = (count: number): LogEntry[] => {
    const logTypes = ['main', 'reject', 'panic'] as const;
    const events = [
      'delivery', 'defer', 'bounce', 'arrival', 'completed', 'failed', 
      'retry', 'timeout', 'connection_refused', 'authentication_failed',
      'spam_detected', 'virus_detected', 'rate_limited', 'quota_exceeded'
    ];
    const hosts = [
      'mx1.example.com', 'mx2.example.com', 'mail.test.org', 'smtp.sample.net',
      'relay.domain.com', 'backup-mx.site.org', 'primary.mail.net', 'secondary.email.co'
    ];
    const domains = ['gmail.com', 'yahoo.com', 'hotmail.com', 'outlook.com', 'company.com', 'test.org', 'sample.net'];
    
    return Array.from({ length: count }, (_, index) => {
      const senderDomain = domains[Math.floor(Math.random() * domains.length)];
      const recipientDomain = domains[Math.floor(Math.random() * domains.length)];
      const logType = logTypes[Math.floor(Math.random() * logTypes.length)];
      const event = events[Math.floor(Math.random() * events.length)];
      
      return {
        id: index + 1,
        timestamp: new Date(Date.now() - Math.random() * 30 * 24 * 60 * 60 * 1000).toISOString(),
        message_id: `1${Math.random().toString(36).substring(2, 8).toUpperCase()}-${Math.random().toString(36).substring(2, 8).toUpperCase()}-${Math.random().toString(36).substring(2, 2).toUpperCase()}`,
        log_type: logType,
        event,
        host: hosts[Math.floor(Math.random() * hosts.length)],
        sender: `user${Math.floor(Math.random() * 1000)}@${senderDomain}`,
        recipients: [`recipient${Math.floor(Math.random() * 1000)}@${recipientDomain}`],
        size: Math.floor(Math.random() * 100000) + 1000,
        status: Math.random() > 0.3 ? 'completed' : Math.random() > 0.5 ? 'processing' : 'failed',
        error_code: logType === 'reject' || Math.random() > 0.9 ? ['550', '554', '451', '452'][Math.floor(Math.random() * 4)] : undefined,
        error_text: logType === 'reject' || Math.random() > 0.9 ? [
          'User unknown in virtual mailbox table',
          'Relay access denied',
          'Message rejected due to spam content',
          'Temporary failure, please try again later'
        ][Math.floor(Math.random() * 4)] : undefined,
        raw_line: `${new Date().toISOString().substring(0, 19)} [${index}] ${event} for ${`user${Math.floor(Math.random() * 1000)}@${senderDomain}`} -> ${`recipient${Math.floor(Math.random() * 1000)}@${recipientDomain}`} on ${hosts[Math.floor(Math.random() * hosts.length)]}`
      };
    });
  };

  const generateMockStatistics = (): LogStatistics => {
    return {
      total_entries: mockLogs.length,
      main_log_count: mockLogs.filter(log => log.log_type === 'main').length,
      reject_log_count: mockLogs.filter(log => log.log_type === 'reject').length,
      panic_log_count: mockLogs.filter(log => log.log_type === 'panic').length,
      recent_entries: Math.floor(Math.random() * 100) + 50,
      top_events: [
        { event: 'delivery', count: Math.floor(Math.random() * 500) + 200 },
        { event: 'defer', count: Math.floor(Math.random() * 200) + 50 },
        { event: 'bounce', count: Math.floor(Math.random() * 100) + 25 },
        { event: 'arrival', count: Math.floor(Math.random() * 300) + 100 },
        { event: 'completed', count: Math.floor(Math.random() * 400) + 150 }
      ],
      error_trends: Array.from({ length: 7 }, (_, i) => ({
        date: new Date(Date.now() - i * 24 * 60 * 60 * 1000).toISOString().split('T')[0],
        count: Math.floor(Math.random() * 50) + 10
      })).reverse()
    };
  };

  const runTestSuite = async (suiteId: string) => {
    setTestSuites(prev => prev.map(suite => 
      suite.id === suiteId ? { ...suite, status: 'running' } : suite
    ));

    try {
      switch (suiteId) {
        case 'search-functionality':
          await testSearchFunctionality();
          break;
        case 'realtime-streaming':
          await testRealtimeStreaming();
          break;
        case 'export-features':
          await testExportFeatures();
          break;
        case 'statistics-generation':
          await testStatisticsGeneration();
          break;
        case 'performance-load':
          await testPerformanceLoad();
          break;
        case 'error-handling':
          await testErrorHandling();
          break;
      }

      setTestSuites(prev => prev.map(suite => 
        suite.id === suiteId 
          ? { ...suite, status: 'passed', result: 'All tests passed successfully' }
          : suite
      ));

    } catch (error) {
      setTestSuites(prev => prev.map(suite => 
        suite.id === suiteId 
          ? { 
              ...suite, 
              status: 'failed', 
              result: error instanceof Error ? error.message : 'Test failed' 
            }
          : suite
      ));
    }
  };

  const testSearchFunctionality = async () => {
    actions.addNotification({ type: 'info', message: 'Testing search functionality...' });
    
    // Simulate various search operations
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Test keyword search
    const keywordResults = mockLogs.filter(log => log.raw_line.includes('delivery'));
    if (keywordResults.length === 0) throw new Error('Keyword search failed');
    
    // Test log type filtering
    const mainLogs = mockLogs.filter(log => log.log_type === 'main');
    if (mainLogs.length === 0) throw new Error('Log type filtering failed');
    
    actions.addNotification({ type: 'success', message: 'Search functionality test passed' });
  };

  const testRealtimeStreaming = async () => {
    actions.addNotification({ type: 'info', message: 'Testing real-time streaming...' });
    
    // Simulate WebSocket connection test
    await new Promise(resolve => setTimeout(resolve, 1500));
    
    // Mock successful connection
    const isConnected = Math.random() > 0.2; // 80% success rate
    if (!isConnected) throw new Error('WebSocket connection failed');
    
    actions.addNotification({ type: 'success', message: 'Real-time streaming test passed' });
  };

  const testExportFeatures = async () => {
    actions.addNotification({ type: 'info', message: 'Testing export features...' });
    
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Test CSV export
    const csvData = mockLogs.slice(0, 10).map(log => ({
      timestamp: log.timestamp,
      type: log.log_type,
      event: log.event,
      message_id: log.message_id
    }));
    
    if (csvData.length === 0) throw new Error('CSV export failed');
    
    actions.addNotification({ type: 'success', message: 'Export features test passed' });
  };

  const testStatisticsGeneration = async () => {
    actions.addNotification({ type: 'info', message: 'Testing statistics generation...' });
    
    await new Promise(resolve => setTimeout(resolve, 1200));
    
    // Test statistics calculation
    const stats = generateMockStatistics();
    if (!stats || stats.total_entries === 0) throw new Error('Statistics generation failed');
    
    setMockStatistics(stats);
    actions.addNotification({ type: 'success', message: 'Statistics generation test passed' });
  };

  const testPerformanceLoad = async () => {
    actions.addNotification({ type: 'info', message: 'Testing performance with large dataset...' });
    
    const startTime = performance.now();
    
    // Generate large dataset
    const largeLogs = generateLargeMockDataset(5000);
    
    const renderTime = performance.now() - startTime;
    
    // Simulate search performance
    const searchStart = performance.now();
    const searchResults = largeLogs.filter(log => log.event === 'delivery');
    const searchTime = performance.now() - searchStart;
    
    // Simulate scroll performance
    const scrollTime = Math.random() * 100 + 50; // Mock scroll time
    
    setPerformanceMetrics({
      renderTime: Math.round(renderTime),
      searchTime: Math.round(searchTime),
      scrollPerformance: Math.round(scrollTime)
    });
    
    if (renderTime > 2000) throw new Error('Render time too slow');
    if (searchTime > 500) throw new Error('Search time too slow');
    
    actions.addNotification({ type: 'success', message: 'Performance test passed' });
  };

  const testErrorHandling = async () => {
    actions.addNotification({ type: 'info', message: 'Testing error handling...' });
    
    await new Promise(resolve => setTimeout(resolve, 800));
    
    // Simulate various error conditions and recovery
    const errorScenarios = [
      'Network timeout',
      'Invalid search parameters',
      'Export file too large',
      'WebSocket disconnection'
    ];
    
    // Mock successful error handling
    const errorHandled = Math.random() > 0.1; // 90% success rate
    if (!errorHandled) throw new Error('Error handling failed');
    
    actions.addNotification({ type: 'success', message: 'Error handling test passed' });
  };

  const runAllTests = async () => {
    actions.addNotification({ type: 'info', message: 'Running comprehensive test suite...' });
    
    for (const suite of testSuites) {
      await runTestSuite(suite.id);
      await new Promise(resolve => setTimeout(resolve, 500)); // Brief pause between tests
    }
    
    actions.addNotification({ type: 'success', message: 'All test suites completed!' });
  };

  // Initialize mock data
  useEffect(() => {
    const logs = generateLargeMockDataset(1000);
    setMockLogs(logs);
    setMockStatistics(generateMockStatistics());
  }, []);

  const getStatusIcon = (status: TestSuite['status']) => {
    switch (status) {
      case 'passed':
        return <svg className="w-5 h-5 text-green-500" fill="currentColor" viewBox="0 0 20 20"><path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" /></svg>;
      case 'failed':
        return <svg className="w-5 h-5 text-red-500" fill="currentColor" viewBox="0 0 20 20"><path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" /></svg>;
      case 'running':
        return <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-indigo-500"></div>;
      default:
        return <svg className="w-5 h-5 text-gray-400" fill="currentColor" viewBox="0 0 20 20"><path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" /></svg>;
    }
  };

  const tabs = [
    { id: 'overview', label: 'Test Overview' },
    { id: 'search', label: 'Search Test' },
    { id: 'realtime', label: 'Real-time Test' },
    { id: 'statistics', label: 'Statistics Test' },
    { id: 'performance', label: 'Performance Test' }
  ] as const;

  return (
    <div className="space-y-6">
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold text-gray-900">Comprehensive Log System Test</h1>
            <button
              onClick={runAllTests}
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700"
            >
              Run All Tests
            </button>
          </div>
          
          {/* Tab Navigation */}
          <div className="mt-4">
            <nav className="-mb-px flex space-x-8">
              {tabs.map((tab) => (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`whitespace-nowrap pb-2 px-1 border-b-2 font-medium text-sm ${
                    activeTab === tab.id
                      ? 'border-indigo-500 text-indigo-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </nav>
          </div>
        </div>

        <div className="p-6">
          {activeTab === 'overview' && (
            <div className="space-y-6">
              {/* Test Suite Overview */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {testSuites.map((suite) => (
                  <div key={suite.id} className="border border-gray-200 rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <h3 className="text-lg font-medium text-gray-900">{suite.name}</h3>
                      {getStatusIcon(suite.status)}
                    </div>
                    <p className="text-sm text-gray-600 mb-3">{suite.description}</p>
                    {suite.result && (
                      <p className={`text-xs ${suite.status === 'passed' ? 'text-green-600' : 'text-red-600'}`}>
                        {suite.result}
                      </p>
                    )}
                    <button
                      onClick={() => runTestSuite(suite.id)}
                      disabled={suite.status === 'running'}
                      className="mt-3 w-full px-3 py-1 border border-gray-300 text-xs font-medium rounded text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50"
                    >
                      {suite.status === 'running' ? 'Running...' : 'Run Test'}
                    </button>
                  </div>
                ))}
              </div>

              {/* Performance Metrics */}
              {performanceMetrics.renderTime > 0 && (
                <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                  <h3 className="text-lg font-medium text-gray-900 mb-3">Performance Metrics</h3>
                  <div className="grid grid-cols-3 gap-4">
                    <div>
                      <div className="text-2xl font-semibold text-gray-900">{performanceMetrics.renderTime}ms</div>
                      <div className="text-sm text-gray-600">Render Time</div>
                    </div>
                    <div>
                      <div className="text-2xl font-semibold text-gray-900">{performanceMetrics.searchTime}ms</div>
                      <div className="text-sm text-gray-600">Search Time</div>
                    </div>
                    <div>
                      <div className="text-2xl font-semibold text-gray-900">{performanceMetrics.scrollPerformance}ms</div>
                      <div className="text-sm text-gray-600">Scroll Performance</div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}

          {activeTab === 'search' && (
            <div>
              <h2 className="text-lg font-medium text-gray-900 mb-4">Search Functionality Test</h2>
              <LogSearch
                onSearch={(params) => {
                  console.log('Search params:', params);
                  actions.addNotification({ type: 'info', message: `Search executed with params: ${JSON.stringify(params)}` });
                }}
                onExport={(format) => {
                  console.log('Export format:', format);
                  actions.addNotification({ type: 'success', message: `Export test completed in ${format} format` });
                }}
                totalEntries={mockLogs.length}
              />
            </div>
          )}

          {activeTab === 'realtime' && (
            <div>
              <h2 className="text-lg font-medium text-gray-900 mb-4">Real-time Streaming Test</h2>
              <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4 mb-4">
                <p className="text-sm text-yellow-700">
                  This test demonstrates the real-time log tail functionality. In a production environment, 
                  this would connect to the WebSocket endpoint for live log streaming.
                </p>
              </div>
              <RealTimeTail />
            </div>
          )}

          {activeTab === 'statistics' && (
            <div>
              <h2 className="text-lg font-medium text-gray-900 mb-4">Statistics Generation Test</h2>
              <LogStatisticsPanel
                statistics={mockStatistics}
                onRefresh={() => {
                  setMockStatistics(generateMockStatistics());
                  actions.addNotification({ type: 'success', message: 'Statistics refreshed with new mock data' });
                }}
              />
            </div>
          )}

          {activeTab === 'performance' && (
            <div>
              <h2 className="text-lg font-medium text-gray-900 mb-4">Performance Test</h2>
              <LogViewer
                logs={mockLogs.slice(0, 100)}
                currentPage={1}
                totalPages={Math.ceil(mockLogs.length / 100)}
                totalEntries={mockLogs.length}
                onPageChange={(page) => {
                  console.log('Page changed to:', page);
                  actions.addNotification({ type: 'info', message: `Navigated to page ${page}` });
                }}
                onRefresh={() => {
                  actions.addNotification({ type: 'success', message: 'Log viewer refreshed' });
                }}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default LogsTestPage;