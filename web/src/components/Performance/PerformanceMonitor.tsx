import React, { useState, useEffect } from 'react';
import { apiService } from '@/services/api';
import { LoadingSpinner } from '@/components/Common';

interface DatabaseStats {
  timestamp: string;
  database_size: number;
  table_stats: Record<string, { row_count: number }>;
  index_stats: Record<string, { table_name: string }>;
}

interface RetentionStatus {
  config: {
    log_entries_retention_days: number;
    audit_log_retention_days: number;
    queue_snapshots_retention_days: number;
    delivery_attempts_retention_days: number;
    sessions_retention_days: number;
    enable_auto_cleanup: boolean;
  };
  table_stats: Record<string, {
    table_name: string;
    retention_days: number;
    total_rows: number;
    expired_rows: number;
    oldest_record?: string;
    newest_record?: string;
  }>;
}

interface PerformanceMetrics {
  database: DatabaseStats;
  retention: RetentionStatus;
  system: {
    timestamp: string;
  };
}

export default function PerformanceMonitor() {
  const [metrics, setMetrics] = useState<PerformanceMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [optimizing, setOptimizing] = useState(false);
  const [cleaning, setCleaning] = useState(false);

  const fetchMetrics = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await apiService.get('/v1/performance/metrics');
      if (response.success) {
        setMetrics(response.data as PerformanceMetrics);
      } else {
        setError(response.error || 'Failed to fetch performance metrics');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch performance metrics');
    } finally {
      setLoading(false);
    }
  };

  const optimizeDatabase = async () => {
    try {
      setOptimizing(true);
      const response = await apiService.post('/v1/performance/database/optimize', {});
      
      if (response.success) {
        // Refresh metrics after optimization
        await fetchMetrics();
        alert('Database optimization completed successfully');
      } else {
        alert('Database optimization failed: ' + (response.error || 'Unknown error'));
      }
    } catch (err) {
      alert('Database optimization failed: ' + (err instanceof Error ? err.message : 'Unknown error'));
    } finally {
      setOptimizing(false);
    }
  };

  const cleanupExpiredData = async () => {
    try {
      setCleaning(true);
      const response = await apiService.post('/v1/performance/retention/cleanup', {});
      
      if (response.success) {
        // Refresh metrics after cleanup
        await fetchMetrics();
        const result = response.data as { total_rows_deleted: number; duration: string };
        alert(`Data cleanup completed: ${result.total_rows_deleted} rows deleted in ${result.duration}`);
      } else {
        alert('Data cleanup failed: ' + (response.error || 'Unknown error'));
      }
    } catch (err) {
      alert('Data cleanup failed: ' + (err instanceof Error ? err.message : 'Unknown error'));
    } finally {
      setCleaning(false);
    }
  };

  useEffect(() => {
    fetchMetrics();
    
    // Refresh metrics every 30 seconds
    const interval = setInterval(fetchMetrics, 30000);
    return () => clearInterval(interval);
  }, []);

  const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  };

  const formatNumber = (num: number): string => {
    return new Intl.NumberFormat().format(num);
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center py-12">
        <LoadingSpinner size="lg" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4">
        <div className="flex">
          <div className="flex-shrink-0">
            <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
          </div>
          <div className="ml-3">
            <h3 className="text-sm font-medium text-red-800">Error loading performance metrics</h3>
            <div className="mt-2 text-sm text-red-700">
              <p>{error}</p>
            </div>
            <div className="mt-4">
              <button
                onClick={fetchMetrics}
                className="bg-red-100 px-3 py-2 rounded-md text-sm font-medium text-red-800 hover:bg-red-200"
              >
                Try again
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!metrics) {
    return <div>No performance metrics available</div>;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="sm:flex sm:items-center sm:justify-between">
            <div>
              <h3 className="text-lg leading-6 font-medium text-gray-900">Performance Monitor</h3>
              <p className="mt-1 text-sm text-gray-500">
                Database performance metrics and optimization tools
              </p>
            </div>
            <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none space-x-3">
              <button
                onClick={optimizeDatabase}
                disabled={optimizing}
                className="inline-flex items-center justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {optimizing ? (
                  <>
                    <LoadingSpinner size="sm" className="mr-2" />
                    Optimizing...
                  </>
                ) : (
                  'Optimize Database'
                )}
              </button>
              <button
                onClick={cleanupExpiredData}
                disabled={cleaning}
                className="inline-flex items-center justify-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {cleaning ? (
                  <>
                    <LoadingSpinner size="sm" className="mr-2" />
                    Cleaning...
                  </>
                ) : (
                  'Cleanup Data'
                )}
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Database Statistics */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h4 className="text-lg font-medium text-gray-900 mb-4">Database Statistics</h4>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
            <div className="bg-blue-50 rounded-lg p-4">
              <div className="text-2xl font-bold text-blue-600">
                {formatBytes(metrics.database.database_size)}
              </div>
              <div className="text-sm text-blue-800">Database Size</div>
            </div>
            
            <div className="bg-green-50 rounded-lg p-4">
              <div className="text-2xl font-bold text-green-600">
                {Object.keys(metrics.database.table_stats).length}
              </div>
              <div className="text-sm text-green-800">Tables</div>
            </div>
            
            <div className="bg-purple-50 rounded-lg p-4">
              <div className="text-2xl font-bold text-purple-600">
                {Object.keys(metrics.database.index_stats).length}
              </div>
              <div className="text-sm text-purple-800">Indexes</div>
            </div>
            
            <div className="bg-orange-50 rounded-lg p-4">
              <div className="text-2xl font-bold text-orange-600">
                {formatNumber(
                  Object.values(metrics.database.table_stats).reduce(
                    (sum, table) => sum + table.row_count, 0
                  )
                )}
              </div>
              <div className="text-sm text-orange-800">Total Rows</div>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Table
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Row Count
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {Object.entries(metrics.database.table_stats).map(([tableName, stats]) => (
                  <tr key={tableName}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {tableName}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatNumber(stats.row_count)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Data Retention Status */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h4 className="text-lg font-medium text-gray-900 mb-4">Data Retention Status</h4>
          
          <div className="mb-4">
            <div className="flex items-center">
              <span className="text-sm font-medium text-gray-700">Auto Cleanup:</span>
              <span className={`ml-2 inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                metrics.retention.config.enable_auto_cleanup
                  ? 'bg-green-100 text-green-800'
                  : 'bg-red-100 text-red-800'
              }`}>
                {metrics.retention.config.enable_auto_cleanup ? 'Enabled' : 'Disabled'}
              </span>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Table
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Retention (Days)
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Total Rows
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Expired Rows
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Oldest Record
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {Object.entries(metrics.retention.table_stats).map(([tableName, stats]) => (
                  <tr key={tableName}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      {stats.table_name}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {stats.retention_days > 0 ? stats.retention_days : 'Disabled'}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {formatNumber(stats.total_rows)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {stats.expired_rows > 0 ? (
                        <span className="text-red-600 font-medium">
                          {formatNumber(stats.expired_rows)}
                        </span>
                      ) : (
                        '0'
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {stats.oldest_record 
                        ? new Date(stats.oldest_record).toLocaleDateString()
                        : '-'
                      }
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Last Updated */}
      <div className="text-center text-sm text-gray-500">
        Last updated: {new Date(metrics.system.timestamp).toLocaleString()}
      </div>
    </div>
  );
}