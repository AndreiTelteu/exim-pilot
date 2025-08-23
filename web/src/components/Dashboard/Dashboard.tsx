import { useEffect } from 'react';
import { useApp } from '@/context/AppContext';
import { LoadingSpinner } from '@/components/Common';
import { MetricsCard } from './MetricsCard';
import { WeeklyChart } from './WeeklyChart';
import { useDashboard } from '@/hooks/useDashboard';
import { webSocketService } from '@/services/websocket';

function formatAge(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h`;
  return `${Math.floor(seconds / 86400)}d`;
}

export default function Dashboard() {
  const { state, actions } = useApp();
  const { metrics, weeklyData, loading, error, refreshData } = useDashboard();

  useEffect(() => {
    // Connect to WebSocket for real-time updates
    if (!webSocketService.isConnected()) {
      webSocketService.connect()
        .then(() => {
          actions.setConnectionStatus('connected');
        })
        .catch((error) => {
          console.error('Failed to connect to WebSocket:', error);
          actions.setConnectionStatus('disconnected');
          actions.addNotification({
            type: 'warning',
            message: 'Real-time updates unavailable. Data will refresh periodically.'
          });
        });
    }
  }, [actions]);

  if (loading && !metrics) {
    return <LoadingSpinner size="lg" className="py-12" />;
  }

  if (error && !metrics) {
    return (
      <div className="space-y-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-6">
          <h3 className="text-lg font-semibold text-red-800 mb-2">
            Failed to Load Dashboard
          </h3>
          <p className="text-red-700 mb-4">{error}</p>
          <button
            onClick={refreshData}
            className="bg-red-600 text-white px-4 py-2 rounded-md hover:bg-red-700 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Welcome Section */}
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-2xl font-semibold text-gray-800">
              Welcome to Exim-Pilot
            </h2>
            <p className="text-gray-600 mt-1">
              Your comprehensive web-based management interface for Exim mail servers.
            </p>
          </div>
          <div className="flex items-center space-x-2">
            <div className={`w-3 h-3 rounded-full ${
              state.connectionStatus === 'connected' ? 'bg-green-500' : 
              state.connectionStatus === 'connecting' ? 'bg-yellow-500' : 'bg-red-500'
            }`}></div>
            <span className="text-sm text-gray-600">
              {state.connectionStatus === 'connected' ? 'Live' : 
               state.connectionStatus === 'connecting' ? 'Connecting' : 'Offline'}
            </span>
          </div>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-blue-50 p-4 rounded-lg">
            <h3 className="font-semibold text-blue-900">Queue Management</h3>
            <p className="text-blue-700 text-sm">Monitor and manage mail queue</p>
          </div>
          <div className="bg-green-50 p-4 rounded-lg">
            <h3 className="font-semibold text-green-900">Log Monitoring</h3>
            <p className="text-green-700 text-sm">Real-time log analysis</p>
          </div>
          <div className="bg-purple-50 p-4 rounded-lg">
            <h3 className="font-semibold text-purple-900">Reports</h3>
            <p className="text-purple-700 text-sm">Deliverability analytics</p>
          </div>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <MetricsCard
          title="Queue Messages"
          value={metrics?.queue.total ?? 0}
          subtitle="Total messages in queue"
          color="blue"
          trend={metrics?.queue.recent_growth ? {
            value: metrics.queue.recent_growth,
            direction: metrics.queue.recent_growth > 0 ? 'up' : 
                     metrics.queue.recent_growth < 0 ? 'down' : 'stable'
          } : undefined}
          loading={loading}
        />
        
        <MetricsCard
          title="Delivered Today"
          value={metrics?.delivery.delivered_today ?? 0}
          subtitle={`${(metrics?.delivery.success_rate ?? 0).toFixed(1)}% success rate`}
          color="green"
          loading={loading}
        />
        
        <MetricsCard
          title="Deferred"
          value={metrics?.queue.deferred ?? 0}
          subtitle="Temporary delivery failures"
          color="yellow"
          loading={loading}
        />
        
        <MetricsCard
          title="Frozen"
          value={metrics?.queue.frozen ?? 0}
          subtitle={metrics?.queue.oldest_message_age ? 
            `Oldest: ${formatAge(metrics.queue.oldest_message_age)}` : 
            'No frozen messages'
          }
          color="red"
          loading={loading}
        />
      </div>

      {/* Additional Metrics Row */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <MetricsCard
          title="Failed Today"
          value={metrics?.delivery.failed_today ?? 0}
          subtitle="Permanent delivery failures"
          color="red"
          loading={loading}
        />
        
        <MetricsCard
          title="Pending Today"
          value={metrics?.delivery.pending_today ?? 0}
          subtitle="Awaiting delivery"
          color="yellow"
          loading={loading}
        />
        
        <MetricsCard
          title="Log Entries"
          value={metrics?.system.log_entries_today ?? 0}
          subtitle="Processed today"
          color="gray"
          loading={loading}
        />
      </div>

      {/* Weekly Overview Chart */}
      <WeeklyChart data={weeklyData} loading={loading} />

      {/* Last Updated */}
      {metrics?.system.last_updated && (
        <div className="text-center text-sm text-gray-500">
          Last updated: {new Date(metrics.system.last_updated).toLocaleString()}
        </div>
      )}
    </div>
  );
}