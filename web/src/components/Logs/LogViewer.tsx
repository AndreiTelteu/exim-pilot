import { useState, useMemo } from 'react';
import { LogEntry } from '../../types/logs';
import { Pagination, VirtualizedList } from '../Common';
import { formatDistanceToNow } from 'date-fns';

interface LogViewerProps {
  logs: LogEntry[];
  currentPage: number;
  totalPages: number;
  totalEntries: number;
  onPageChange: (page: number) => void;
  onRefresh: () => void;
}

const LogViewer: React.FC<LogViewerProps> = ({
  logs,
  currentPage,
  totalPages,
  totalEntries,
  onPageChange,
  onRefresh
}) => {
  const [selectedLogId, setSelectedLogId] = useState<number | null>(null);
  const [expandedLogs, setExpandedLogs] = useState<Set<number>>(new Set());

  const getLogTypeColor = (logType: string) => {
    switch (logType) {
      case 'main':
        return 'bg-blue-100 text-blue-800';
      case 'reject':
        return 'bg-red-100 text-red-800';
      case 'panic':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
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

  const toggleExpanded = (logId: number) => {
    const newExpanded = new Set(expandedLogs);
    if (newExpanded.has(logId)) {
      newExpanded.delete(logId);
    } else {
      newExpanded.add(logId);
    }
    setExpandedLogs(newExpanded);
  };

  const formatTimestamp = (timestamp: string) => {
    try {
      const date = new Date(timestamp);
      return {
        relative: formatDistanceToNow(date, { addSuffix: true }),
        absolute: date.toLocaleString(),
      };
    } catch {
      return {
        relative: 'Invalid date',
        absolute: timestamp,
      };
    }
  };

  const renderLogEntry = (log: LogEntry, index: number) => {
    const isExpanded = expandedLogs.has(log.id);
    const isSelected = selectedLogId === log.id;
    const timeInfo = formatTimestamp(log.timestamp);

    return (
      <div
        key={log.id}
        className={`border border-gray-200 rounded-lg p-4 transition-colors ${
          isSelected ? 'ring-2 ring-indigo-500 bg-indigo-50' : 'bg-white hover:bg-gray-50'
        }`}
      >
        <div className="flex items-start justify-between">
          <div className="flex-1 min-w-0">
            <div className="flex items-center space-x-3 mb-2">
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getLogTypeColor(log.log_type)}`}>
                {log.log_type}
              </span>
              <span className={`font-medium ${getEventColor(log.event)}`}>
                {log.event}
              </span>
              {log.message_id && (
                <span className="text-sm text-gray-500 font-mono">
                  ID: {log.message_id.length > 16 ? `${log.message_id.substring(0, 16)}...` : log.message_id}
                </span>
              )}
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm">
              <div>
                <span className="font-medium text-gray-700">Time:</span>
                <div className="text-gray-600">
                  <div title={timeInfo.absolute}>{timeInfo.relative}</div>
                  <div className="text-xs text-gray-500">{timeInfo.absolute}</div>
                </div>
              </div>

              {log.sender && (
                <div>
                  <span className="font-medium text-gray-700">Sender:</span>
                  <div className="text-gray-600 font-mono text-xs break-all">
                    {log.sender}
                  </div>
                </div>
              )}

              {log.recipients && log.recipients.length > 0 && (
                <div>
                  <span className="font-medium text-gray-700">Recipients:</span>
                  <div className="text-gray-600 font-mono text-xs">
                    {log.recipients.length === 1 
                      ? log.recipients[0] 
                      : `${log.recipients[0]} (+${log.recipients.length - 1} more)`
                    }
                  </div>
                </div>
              )}

              {log.host && (
                <div>
                  <span className="font-medium text-gray-700">Host:</span>
                  <div className="text-gray-600 font-mono text-xs">
                    {log.host}
                  </div>
                </div>
              )}

              {log.size && (
                <div>
                  <span className="font-medium text-gray-700">Size:</span>
                  <div className="text-gray-600">
                    {(log.size / 1024).toFixed(1)} KB
                  </div>
                </div>
              )}

              {log.status && (
                <div>
                  <span className="font-medium text-gray-700">Status:</span>
                  <div className="text-gray-600">
                    {log.status}
                  </div>
                </div>
              )}
            </div>

            {(log.error_code || log.error_text) && (
              <div className="mt-3 p-3 bg-red-50 border border-red-200 rounded-md">
                <div className="text-sm">
                  {log.error_code && (
                    <div className="font-medium text-red-800">
                      Error Code: {log.error_code}
                    </div>
                  )}
                  {log.error_text && (
                    <div className="text-red-700 mt-1">
                      {log.error_text}
                    </div>
                  )}
                </div>
              </div>
            )}

            {isExpanded && (
              <div className="mt-4 p-3 bg-gray-50 border border-gray-200 rounded-md">
                <div className="text-sm font-medium text-gray-700 mb-2">Raw Log Entry:</div>
                <pre className="text-xs text-gray-600 whitespace-pre-wrap font-mono bg-white p-2 rounded border">
                  {log.raw_line}
                </pre>
              </div>
            )}
          </div>

          <div className="flex items-center space-x-2 ml-4">
            <button
              onClick={() => toggleExpanded(log.id)}
              className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
              title={isExpanded ? 'Collapse details' : 'Expand details'}
            >
              <svg
                className={`w-5 h-5 transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            
            <button
              onClick={() => setSelectedLogId(isSelected ? null : log.id)}
              className={`p-1 transition-colors ${
                isSelected ? 'text-indigo-600' : 'text-gray-400 hover:text-gray-600'
              }`}
              title={isSelected ? 'Deselect' : 'Select'}
            >
              <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
            </button>
          </div>
        </div>
      </div>
    );
  };

  const memoizedLogEntries = useMemo(() => {
    return logs.map((log, index) => renderLogEntry(log, index));
  }, [logs, expandedLogs, selectedLogId]);

  if (logs.length === 0) {
    return (
      <div className="text-center py-12">
        <svg
          className="mx-auto h-12 w-12 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
        <h3 className="mt-2 text-sm font-medium text-gray-900">No log entries found</h3>
        <p className="mt-1 text-sm text-gray-500">
          Try adjusting your search filters or check back later.
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="text-sm text-gray-700">
          Showing {logs.length} of {totalEntries.toLocaleString()} log entries
        </div>
        <button
          onClick={onRefresh}
          className="inline-flex items-center px-3 py-1 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
        >
          <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          Refresh
        </button>
      </div>

      <div className="space-y-3">
        {memoizedLogEntries}
      </div>

      {totalPages > 1 && (
        <div className="flex justify-center">
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={onPageChange}
          />
        </div>
      )}
    </div>
  );
};

export default LogViewer;