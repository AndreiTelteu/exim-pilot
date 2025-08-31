import { useState, useEffect, useRef, useCallback } from 'react';
import { LogEntry } from '../../types/logs';
import { useApp } from '../../context/AppContext';
import { webSocketService } from '../../services/websocket';
import { formatDistanceToNow } from 'date-fns';

interface RealTimeTailProps {}

interface TailFilters {
  log_type?: string;
  keyword?: string;
  message_id?: string;
  sender?: string;
}

const RealTimeTail: React.FC<RealTimeTailProps> = () => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [autoScroll, setAutoScroll] = useState(true);
  const [filters, setFilters] = useState<TailFilters>({});
  const [maxLines, setMaxLines] = useState(500);
  const [showFilters, setShowFilters] = useState(false);
  
  const logsContainerRef = useRef<HTMLDivElement>(null);
  const { actions } = useApp();

  const scrollToBottom = useCallback(() => {
    if (autoScroll && logsContainerRef.current) {
      logsContainerRef.current.scrollTop = logsContainerRef.current.scrollHeight;
    }
  }, [autoScroll]);

  const filterLogEntry = useCallback((entry: LogEntry): boolean => {
    if (filters.log_type && entry.log_type !== filters.log_type) {
      return false;
    }
    if (filters.keyword && !entry.raw_line.toLowerCase().includes(filters.keyword.toLowerCase())) {
      return false;
    }
    if (filters.message_id && entry.message_id !== filters.message_id) {
      return false;
    }
    if (filters.sender && (!entry.sender || !entry.sender.toLowerCase().includes(filters.sender.toLowerCase()))) {
      return false;
    }
    return true;
  }, [filters]);

  const addLogEntry = useCallback((entry: LogEntry) => {
    if (!isPaused && filterLogEntry(entry)) {
      setLogs(prevLogs => {
        const newLogs = [entry, ...prevLogs].slice(0, maxLines);
        return newLogs;
      });
    }
  }, [isPaused, filterLogEntry, maxLines]);

  const handleWebSocketMessage = useCallback((message: any) => {
    if (message.type === 'log_entry' && message.data) {
      addLogEntry(message.data);
    }
  }, [addLogEntry]);

  useEffect(() => {
    const subscribe = async () => {
      try {
        await webSocketService.subscribe('/api/v1/logs/tail', handleWebSocketMessage);
        setIsConnected(true);
        
        actions.addNotification({
          type: 'info',
          message: 'Connected to real-time log stream'
        });
      } catch (error) {
        console.error('Failed to subscribe to log tail:', error);
        setIsConnected(false);
        
        actions.addNotification({
          type: 'error',
          message: 'Failed to connect to real-time log stream'
        });
      }
    };

    subscribe();

    return () => {
      webSocketService.unsubscribe('/api/v1/logs/tail');
    };
  }, [handleWebSocketMessage, actions]);

  useEffect(() => {
    if (autoScroll && logs.length > 0) {
      setTimeout(scrollToBottom, 100);
    }
  }, [logs, scrollToBottom, autoScroll]);

  const handleClearLogs = () => {
    setLogs([]);
    actions.addNotification({
      type: 'info',
      message: 'Log buffer cleared'
    });
  };

  const handleFilterChange = (key: string, value: string) => {
    setFilters(prev => ({
      ...prev,
      [key]: value === '' ? undefined : value
    }));
  };

  const getLogTypeColor = (logType: string) => {
    switch (logType) {
      case 'main':
        return 'text-blue-600';
      case 'reject':
        return 'text-red-600';
      case 'panic':
        return 'text-yellow-600';
      default:
        return 'text-gray-600';
    }
  };

  const getEventColor = (event: string) => {
    if (event.includes('delivery') || event.includes('completed')) {
      return 'text-green-600';
    } else if (event.includes('defer') || event.includes('failed')) {
      return 'text-yellow-600';
    } else if (event.includes('bounce') || event.includes('reject')) {
      return 'text-red-600';
    }
    return 'text-gray-600';
  };

  const formatTimestamp = (timestamp: string) => {
    try {
      return new Date(timestamp).toLocaleTimeString();
    } catch {
      return timestamp;
    }
  };

  return (
    <div className="space-y-4">
      {/* Controls */}
      <div className="bg-white shadow rounded-lg p-4">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2">
              <div className={`w-3 h-3 rounded-full ${isConnected ? 'bg-green-400' : 'bg-red-400'}`}></div>
              <span className="text-sm font-medium text-gray-700">
                {isConnected ? 'Connected' : 'Disconnected'}
              </span>
            </div>
            
            <div className="text-sm text-gray-500">
              {logs.length} entries
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <button
              onClick={() => setShowFilters(!showFilters)}
              className={`px-3 py-1 text-sm rounded-md ${
                showFilters 
                  ? 'bg-indigo-100 text-indigo-700' 
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              Filters
            </button>
            
            <button
              onClick={() => setIsPaused(!isPaused)}
              className={`px-3 py-1 text-sm rounded-md ${
                isPaused 
                  ? 'bg-yellow-100 text-yellow-700' 
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              {isPaused ? 'Resume' : 'Pause'}
            </button>
            
            <button
              onClick={() => setAutoScroll(!autoScroll)}
              className={`px-3 py-1 text-sm rounded-md ${
                autoScroll 
                  ? 'bg-green-100 text-green-700' 
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              Auto-scroll: {autoScroll ? 'On' : 'Off'}
            </button>
            
            <button
              onClick={handleClearLogs}
              className="px-3 py-1 text-sm rounded-md bg-red-100 text-red-700 hover:bg-red-200"
            >
              Clear
            </button>
          </div>
        </div>

        {/* Filters */}
        {showFilters && (
          <div className="mt-4 pt-4 border-t border-gray-200">
            <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Log Type</label>
                <select
                  value={filters.log_type || ''}
                  onChange={(e) => handleFilterChange('log_type', e.target.value)}
                  className="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-indigo-500"
                >
                  <option value="">All Types</option>
                  <option value="main">Main</option>
                  <option value="reject">Reject</option>
                  <option value="panic">Panic</option>
                </select>
              </div>
              
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Keyword</label>
                <input
                  type="text"
                  value={filters.keyword || ''}
                  onChange={(e) => handleFilterChange('keyword', e.target.value)}
                  placeholder="Search..."
                  className="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-indigo-500"
                />
              </div>
              
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Message ID</label>
                <input
                  type="text"
                  value={filters.message_id || ''}
                  onChange={(e) => handleFilterChange('message_id', e.target.value)}
                  placeholder="Message ID..."
                  className="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-indigo-500 font-mono"
                />
              </div>
              
              <div>
                <label className="block text-xs font-medium text-gray-700 mb-1">Max Lines</label>
                <select
                  value={maxLines}
                  onChange={(e) => setMaxLines(parseInt(e.target.value))}
                  className="w-full px-2 py-1 text-sm border border-gray-300 rounded focus:outline-none focus:ring-1 focus:ring-indigo-500"
                >
                  <option value={100}>100</option>
                  <option value={250}>250</option>
                  <option value={500}>500</option>
                  <option value={1000}>1000</option>
                </select>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Log Display */}
      <div className="bg-white shadow rounded-lg">
        <div
          ref={logsContainerRef}
          className="h-96 overflow-y-auto p-4 space-y-2"
          onScroll={(e) => {
            const target = e.target as HTMLDivElement;
            const isAtBottom = target.scrollHeight - target.scrollTop === target.clientHeight;
            if (autoScroll && !isAtBottom) {
              setAutoScroll(false);
            }
          }}
        >
          {logs.length === 0 ? (
            <div className="text-center py-12 text-gray-500">
              {isPaused ? (
                <div>
                  <svg className="mx-auto h-8 w-8 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 9v6m4-6v6m7-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <p>Log streaming is paused</p>
                </div>
              ) : !isConnected ? (
                <div>
                  <svg className="mx-auto h-8 w-8 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v3m0 0v3m0-3h3m-3 0H9m12 0a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                  <p>Connecting to log stream...</p>
                </div>
              ) : (
                <div>
                  <svg className="mx-auto h-8 w-8 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  <p>Waiting for log entries...</p>
                </div>
              )}
            </div>
          ) : (
            logs.map((log, index) => (
              <div
                key={`${log.id}-${index}`}
                className="text-sm font-mono border-l-4 border-gray-200 pl-3 py-1 hover:bg-gray-50 transition-colors"
              >
                <div className="flex items-start space-x-3">
                  <span className="text-gray-500 text-xs whitespace-nowrap">
                    {formatTimestamp(log.timestamp)}
                  </span>
                  
                  <span className={`font-medium text-xs uppercase whitespace-nowrap ${getLogTypeColor(log.log_type)}`}>
                    {log.log_type}
                  </span>
                  
                  <span className={`font-medium text-xs whitespace-nowrap ${getEventColor(log.event)}`}>
                    {log.event}
                  </span>
                  
                  {log.message_id && (
                    <span className="text-indigo-600 text-xs whitespace-nowrap">
                      {log.message_id.length > 12 ? `${log.message_id.substring(0, 12)}...` : log.message_id}
                    </span>
                  )}
                  
                  <span className="text-gray-600 text-xs flex-1 break-all">
                    {log.raw_line.length > 200 ? `${log.raw_line.substring(0, 200)}...` : log.raw_line}
                  </span>
                </div>
                
                {(log.error_code || log.error_text) && (
                  <div className="mt-1 pl-16">
                    <span className="text-red-600 text-xs">
                      {log.error_code && `[${log.error_code}] `}
                      {log.error_text}
                    </span>
                  </div>
                )}
              </div>
            ))
          )}
        </div>
        
        {/* Scroll to bottom button */}
        {!autoScroll && logs.length > 0 && (
          <div className="absolute bottom-4 right-4">
            <button
              onClick={() => {
                setAutoScroll(true);
                scrollToBottom();
              }}
              className="bg-indigo-600 text-white p-2 rounded-full shadow-lg hover:bg-indigo-700 transition-colors"
              title="Scroll to bottom and resume auto-scroll"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 14l-7 7m0 0l-7-7m7 7V3" />
              </svg>
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default RealTimeTail;