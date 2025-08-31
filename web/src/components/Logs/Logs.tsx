import { useState, useEffect, useCallback } from 'react';
import { LogEntry, LogSearchParams, LogStatistics, LogSearchResponse } from '../../types/logs';
import { APIResponse } from '../../types/api';
import LogViewer from './LogViewer';
import LogSearch from './LogSearch';
import RealTimeTail from './RealTimeTail';
import LogStatisticsPanel from './LogStatisticsPanel';
import { LoadingSpinner } from '../Common';
import { useApp } from '../../context/AppContext';

interface LogsProps {}

const Logs: React.FC<LogsProps> = () => {
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(0);
  const [totalEntries, setTotalEntries] = useState(0);
  const [searchParams, setSearchParams] = useState<LogSearchParams>({
    per_page: 50
  });
  const [statistics, setStatistics] = useState<LogStatistics | null>(null);
  const [activeView, setActiveView] = useState<'search' | 'realtime' | 'statistics'>('search');
  const { actions } = useApp();

  const fetchLogs = useCallback(async (params: LogSearchParams = {}) => {
    try {
      setLoading(true);
      setError(null);
      
      const queryParams = new URLSearchParams();
      const mergedParams = { ...searchParams, ...params };
      
      Object.entries(mergedParams).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          // Map frontend parameter names to backend parameter names
          let backendKey = key;
          if (key === 'start_date') {
            backendKey = 'start_time';
            // Convert datetime-local format to RFC3339
            try {
              const date = new Date(value.toString());
              if (!isNaN(date.getTime())) {
                queryParams.append(backendKey, date.toISOString());
                return;
              }
            } catch (e) {
              console.warn('Invalid start_date format:', value);
            }
          } else if (key === 'end_date') {
            backendKey = 'end_time';
            // Convert datetime-local format to RFC3339
            try {
              const date = new Date(value.toString());
              if (!isNaN(date.getTime())) {
                queryParams.append(backendKey, date.toISOString());
                return;
              }
            } catch (e) {
              console.warn('Invalid end_date format:', value);
            }
          } else if (key === 'keyword') {
            backendKey = 'keywords';
          }
          
          queryParams.append(backendKey, value.toString());
        }
      });

      const response = await fetch(`/api/v1/logs?${queryParams.toString()}`, {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch logs: ${response.statusText}`);
      }

      const data: APIResponse<LogSearchResponse> = await response.json();
      
      if (!data.success) {
        throw new Error(data.error || 'Failed to fetch logs');
      }

      // Handle the nested response structure: data.data.entries
      const entries = data.data?.entries || [];
      setLogs(Array.isArray(entries) ? entries : []);
      setTotalPages(data.meta?.total_pages || 0);
      setTotalEntries(data.meta?.total || 0);
      setCurrentPage(data.meta?.page || 1);

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
  }, [searchParams, actions]);

  const fetchStatistics = useCallback(async () => {
    try {
      const response = await fetch('/api/v1/logs/statistics', {
        method: 'GET',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error(`Failed to fetch statistics: ${response.statusText}`);
      }

      const data: APIResponse<LogStatistics> = await response.json();
      
      if (data.success && data.data) {
        setStatistics(data.data);
      }
    } catch (err) {
      console.error('Failed to fetch log statistics:', err);
    }
  }, []);

  const handleSearch = useCallback((params: LogSearchParams) => {
    setSearchParams(prev => ({ ...prev, ...params, page: 1 }));
    setCurrentPage(1);
  }, []);

  const handlePageChange = useCallback((page: number) => {
    setSearchParams(prev => ({ ...prev, page }));
  }, []);

  const handleExport = useCallback(async (format: 'csv' | 'txt' | 'json') => {
    try {
      const queryParams = new URLSearchParams();
      Object.entries(searchParams).forEach(([key, value]) => {
        if (value !== undefined && value !== null && value !== '') {
          queryParams.append(key, value.toString());
        }
      });
      queryParams.append('format', format);

      const response = await fetch(`/api/v1/logs/export?${queryParams.toString()}`, {
        method: 'GET',
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error(`Export failed: ${response.statusText}`);
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.style.display = 'none';
      a.href = url;
      a.download = `logs_export_${Date.now()}.${format}`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      actions.addNotification({
        type: 'success',
        message: `Logs exported successfully as ${format.toUpperCase()}`
      });
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Export failed';
      actions.addNotification({
        type: 'error',
        message: errorMessage
      });
    }
  }, [searchParams, actions]);

  useEffect(() => {
    fetchLogs(searchParams);
  }, [searchParams, fetchLogs]);

  useEffect(() => {
    if (activeView === 'statistics') {
      fetchStatistics();
    }
  }, [activeView, fetchStatistics]);

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Log Monitoring</h1>
        <div className="flex space-x-2">
          <button
            onClick={() => setActiveView('search')}
            className={`px-4 py-2 rounded-md text-sm font-medium ${
              activeView === 'search'
                ? 'bg-indigo-600 text-white'
                : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
            }`}
          >
            Log Search
          </button>
          <button
            onClick={() => setActiveView('realtime')}
            className={`px-4 py-2 rounded-md text-sm font-medium ${
              activeView === 'realtime'
                ? 'bg-indigo-600 text-white'
                : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
            }`}
          >
            Real-time Tail
          </button>
          <button
            onClick={() => setActiveView('statistics')}
            className={`px-4 py-2 rounded-md text-sm font-medium ${
              activeView === 'statistics'
                ? 'bg-indigo-600 text-white'
                : 'bg-white text-gray-700 border border-gray-300 hover:bg-gray-50'
            }`}
          >
            Statistics
          </button>
        </div>
      </div>

      {activeView === 'search' && (
        <div className="space-y-6">
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
                  <h3 className="text-sm font-medium text-red-800">Error loading logs</h3>
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
      )}

      {activeView === 'realtime' && (
        <RealTimeTail />
      )}

      {activeView === 'statistics' && (
        <LogStatisticsPanel
          statistics={statistics}
          onRefresh={fetchStatistics}
        />
      )}
    </div>
  );
};

export default Logs;