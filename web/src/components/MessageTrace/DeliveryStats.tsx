import React from 'react';
import ReactECharts from 'echarts-for-react';
import { EChartsOption } from 'echarts';

interface MessageDeliveryTrace {
  message_id: string;
  message?: {
    id: string;
    timestamp: string;
    sender: string;
    size?: number;
    status: string;
  };
  recipients: RecipientDeliveryStatus[];
  delivery_timeline: DeliveryTimelineEvent[];
  retry_schedule: RetryScheduleEntry[];
  summary: DeliveryTraceSummary;
  generated_at: string;
}

interface RecipientDeliveryStatus {
  recipient: string;
  status: string;
  delivered_at?: string;
  last_attempt_at?: string;
  next_retry_at?: string;
  attempt_count: number;
  last_smtp_code?: string;
  last_error_text?: string;
  delivery_history: DeliveryAttempt[];
}

interface DeliveryTimelineEvent {
  timestamp: string;
  event_type: string;
  recipient?: string;
  host?: string;
  ip_address?: string;
  smtp_code?: string;
  error_text?: string;
  description: string;
  source: string;
  source_id?: number;
}

interface RetryScheduleEntry {
  recipient: string;
  scheduled_at: string;
  attempt_number: number;
  reason: string;
  is_estimated: boolean;
}

interface DeliveryAttempt {
  id: number;
  message_id: string;
  recipient: string;
  timestamp: string;
  host?: string;
  ip_address?: string;
  status: string;
  smtp_code?: string;
  error_message?: string;
}

interface DeliveryTraceSummary {
  total_recipients: number;
  delivered_count: number;
  deferred_count: number;
  bounced_count: number;
  pending_count: number;
  total_attempts: number;
  first_attempt_at?: string;
  last_attempt_at?: string;
  average_delivery_time_seconds?: number;
}

interface DeliveryStatsProps {
  trace: MessageDeliveryTrace;
}

export const DeliveryStats: React.FC<DeliveryStatsProps> = ({ trace }) => {
  const formatDuration = (seconds?: number) => {
    if (!seconds) return 'N/A';
    
    if (seconds < 60) {
      return `${Math.round(seconds)}s`;
    } else if (seconds < 3600) {
      return `${Math.round(seconds / 60)}m ${Math.round(seconds % 60)}s`;
    } else {
      const hours = Math.floor(seconds / 3600);
      const minutes = Math.floor((seconds % 3600) / 60);
      return `${hours}h ${minutes}m`;
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  // Calculate additional statistics
  const calculateStats = () => {
    const stats = {
      eventCounts: {} as Record<string, number>,
      sourceCounts: {} as Record<string, number>,
      attemptsByRecipient: {} as Record<string, number>,
      errorCodes: {} as Record<string, number>,
      hostStats: {} as Record<string, number>,
    };

    // Count events by type and source
    trace.delivery_timeline.forEach(event => {
      stats.eventCounts[event.event_type] = (stats.eventCounts[event.event_type] || 0) + 1;
      stats.sourceCounts[event.source] = (stats.sourceCounts[event.source] || 0) + 1;
    });

    // Count attempts by recipient and collect error codes
    trace.recipients.forEach(recipient => {
      stats.attemptsByRecipient[recipient.recipient] = recipient.attempt_count;
      
      recipient.delivery_history.forEach(attempt => {
        if (attempt.smtp_code) {
          stats.errorCodes[attempt.smtp_code] = (stats.errorCodes[attempt.smtp_code] || 0) + 1;
        }
        if (attempt.host) {
          stats.hostStats[attempt.host] = (stats.hostStats[attempt.host] || 0) + 1;
        }
      });
    });

    return stats;
  };

  const stats = calculateStats();

  // Delivery Status Chart
  const deliveryStatusOption: EChartsOption = {
    title: { text: 'Delivery Status Distribution', left: 'center' },
    tooltip: { trigger: 'item', formatter: '{a} <br/>{b}: {c} ({d}%)' },
    series: [
      {
        name: 'Recipients',
        type: 'pie',
        radius: '50%',
        data: [
          { value: trace.summary.delivered_count, name: 'Delivered', itemStyle: { color: '#10b981' } },
          { value: trace.summary.deferred_count, name: 'Deferred', itemStyle: { color: '#f59e0b' } },
          { value: trace.summary.bounced_count, name: 'Bounced', itemStyle: { color: '#ef4444' } },
          { value: trace.summary.pending_count, name: 'Pending', itemStyle: { color: '#3b82f6' } },
        ].filter(item => item.value > 0),
        emphasis: {
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)'
          }
        }
      }
    ]
  };

  // Event Timeline Chart
  const eventTimelineOption: EChartsOption = {
    title: { text: 'Event Types Distribution', left: 'center' },
    tooltip: { trigger: 'item' },
    xAxis: {
      type: 'category',
      data: Object.keys(stats.eventCounts),
      axisLabel: { rotate: 45 }
    },
    yAxis: { type: 'value' },
    series: [
      {
        name: 'Events',
        type: 'bar',
        data: Object.values(stats.eventCounts),
        itemStyle: { color: '#3b82f6' }
      }
    ]
  };

  // Attempts per Recipient Chart
  const attemptsOption: EChartsOption = {
    title: { text: 'Delivery Attempts per Recipient', left: 'center' },
    tooltip: { trigger: 'item' },
    xAxis: {
      type: 'category',
      data: Object.keys(stats.attemptsByRecipient),
      axisLabel: { 
        rotate: 45,
        formatter: (value: string) => value.length > 20 ? value.substring(0, 20) + '...' : value
      }
    },
    yAxis: { type: 'value' },
    series: [
      {
        name: 'Attempts',
        type: 'bar',
        data: Object.values(stats.attemptsByRecipient),
        itemStyle: { color: '#8b5cf6' }
      }
    ]
  };

  return (
    <div className="space-y-8">
      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-blue-50 p-4 rounded-lg">
          <dt className="text-sm font-medium text-blue-600">Total Events</dt>
          <dd className="mt-1 text-2xl font-semibold text-blue-900">{trace.delivery_timeline.length}</dd>
        </div>
        <div className="bg-green-50 p-4 rounded-lg">
          <dt className="text-sm font-medium text-green-600">Success Rate</dt>
          <dd className="mt-1 text-2xl font-semibold text-green-900">
            {trace.summary.total_recipients > 0 
              ? Math.round((trace.summary.delivered_count / trace.summary.total_recipients) * 100)
              : 0
            }%
          </dd>
        </div>
        <div className="bg-yellow-50 p-4 rounded-lg">
          <dt className="text-sm font-medium text-yellow-600">Avg Attempts</dt>
          <dd className="mt-1 text-2xl font-semibold text-yellow-900">
            {trace.summary.total_recipients > 0 
              ? Math.round((trace.summary.total_attempts / trace.summary.total_recipients) * 10) / 10
              : 0
            }
          </dd>
        </div>
        <div className="bg-purple-50 p-4 rounded-lg">
          <dt className="text-sm font-medium text-purple-600">Avg Delivery Time</dt>
          <dd className="mt-1 text-2xl font-semibold text-purple-900">
            {formatDuration(trace.summary.average_delivery_time_seconds)}
          </dd>
        </div>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white p-4 border border-gray-200 rounded-lg">
          <ReactECharts option={deliveryStatusOption} style={{ height: '300px' }} />
        </div>
        <div className="bg-white p-4 border border-gray-200 rounded-lg">
          <ReactECharts option={eventTimelineOption} style={{ height: '300px' }} />
        </div>
      </div>

      {/* Attempts Chart */}
      {Object.keys(stats.attemptsByRecipient).length > 0 && (
        <div className="bg-white p-4 border border-gray-200 rounded-lg">
          <ReactECharts option={attemptsOption} style={{ height: '400px' }} />
        </div>
      )}

      {/* Detailed Statistics Tables */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Event Sources */}
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Event Sources</h3>
          <div className="space-y-2">
            {Object.entries(stats.sourceCounts).map(([source, count]) => (
              <div key={source} className="flex justify-between items-center py-2 px-3 bg-gray-50 rounded">
                <span className="text-sm font-medium text-gray-700 capitalize">{source}</span>
                <span className="text-sm text-gray-900">{count} events</span>
              </div>
            ))}
          </div>
        </div>

        {/* SMTP Error Codes */}
        {Object.keys(stats.errorCodes).length > 0 && (
          <div className="bg-white border border-gray-200 rounded-lg p-4">
            <h3 className="text-lg font-medium text-gray-900 mb-4">SMTP Response Codes</h3>
            <div className="space-y-2">
              {Object.entries(stats.errorCodes)
                .sort(([,a], [,b]) => b - a)
                .slice(0, 10)
                .map(([code, count]) => (
                <div key={code} className="flex justify-between items-center py-2 px-3 bg-gray-50 rounded">
                  <span className="text-sm font-medium text-gray-700">{code}</span>
                  <span className="text-sm text-gray-900">{count} times</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Host Statistics */}
      {Object.keys(stats.hostStats).length > 0 && (
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Delivery Hosts</h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {Object.entries(stats.hostStats)
              .sort(([,a], [,b]) => b - a)
              .slice(0, 10)
              .map(([host, count]) => (
              <div key={host} className="flex justify-between items-center py-2 px-3 bg-gray-50 rounded">
                <span className="text-sm font-medium text-gray-700 truncate">{host}</span>
                <span className="text-sm text-gray-900">{count} attempts</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Timeline Summary */}
      <div className="bg-white border border-gray-200 rounded-lg p-4">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Timeline Summary</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {trace.summary.first_attempt_at && (
            <div>
              <dt className="text-sm font-medium text-gray-500">First Attempt</dt>
              <dd className="mt-1 text-sm text-gray-900">{formatTimestamp(trace.summary.first_attempt_at)}</dd>
            </div>
          )}
          {trace.summary.last_attempt_at && (
            <div>
              <dt className="text-sm font-medium text-gray-500">Last Attempt</dt>
              <dd className="mt-1 text-sm text-gray-900">{formatTimestamp(trace.summary.last_attempt_at)}</dd>
            </div>
          )}
          <div>
            <dt className="text-sm font-medium text-gray-500">Report Generated</dt>
            <dd className="mt-1 text-sm text-gray-900">{formatTimestamp(trace.generated_at)}</dd>
          </div>
        </div>
      </div>
    </div>
  );
};