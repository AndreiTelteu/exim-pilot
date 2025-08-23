import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { apiService } from '../../services/api';
import LoadingSpinner from '../Common/LoadingSpinner';
import { DeliveryTimeline } from './DeliveryTimeline';
import { RecipientStatus } from './RecipientStatus';
import { RetrySchedule } from './RetrySchedule';
import { DeliveryStats } from './DeliveryStats';
import { NotesAndTags } from './NotesAndTags';
import { ThreadedTimeline } from './ThreadedTimeline';

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

export const MessageTrace: React.FC = () => {
  const { messageId } = useParams<{ messageId: string }>();
  const navigate = useNavigate();
  const [trace, setTrace] = useState<MessageDeliveryTrace | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'timeline' | 'recipients' | 'retry' | 'stats' | 'notes' | 'threaded'>('timeline');

  useEffect(() => {
    if (messageId) {
      fetchMessageTrace();
    }
  }, [messageId]);

  const fetchMessageTrace = async () => {
    if (!messageId) return;

    try {
      setLoading(true);
      setError(null);
      const response = await apiService.get(`/messages/${messageId}/delivery-trace`);
      setTrace(response.data as MessageDeliveryTrace);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load message trace');
    } finally {
      setLoading(false);
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const formatDuration = (seconds?: number) => {
    if (!seconds) return 'N/A';
    
    if (seconds < 60) {
      return `${Math.round(seconds)}s`;
    } else if (seconds < 3600) {
      return `${Math.round(seconds / 60)}m`;
    } else {
      return `${Math.round(seconds / 3600)}h`;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'delivered':
        return 'text-green-600 bg-green-100';
      case 'deferred':
        return 'text-yellow-600 bg-yellow-100';
      case 'bounced':
        return 'text-red-600 bg-red-100';
      case 'pending':
        return 'text-blue-600 bg-blue-100';
      case 'frozen':
        return 'text-gray-600 bg-gray-100';
      default:
        return 'text-gray-600 bg-gray-100';
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <LoadingSpinner />
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
            <h3 className="text-sm font-medium text-red-800">Error Loading Message Trace</h3>
            <div className="mt-2 text-sm text-red-700">
              <p>{error}</p>
            </div>
            <div className="mt-4">
              <button
                onClick={() => navigate('/queue')}
                className="bg-red-100 px-3 py-2 rounded-md text-sm font-medium text-red-800 hover:bg-red-200"
              >
                Back to Queue
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!trace) {
    return (
      <div className="text-center py-8">
        <p className="text-gray-500">No trace data available</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Message Delivery Trace</h1>
            <p className="text-sm text-gray-500 mt-1">Message ID: {trace.message_id}</p>
          </div>
          <button
            onClick={() => navigate('/queue')}
            className="bg-gray-100 hover:bg-gray-200 px-4 py-2 rounded-md text-sm font-medium text-gray-700"
          >
            Back to Queue
          </button>
        </div>

        {/* Message Info */}
        {trace.message && (
          <div className="mt-6 grid grid-cols-1 md:grid-cols-4 gap-4">
            <div>
              <dt className="text-sm font-medium text-gray-500">Sender</dt>
              <dd className="mt-1 text-sm text-gray-900">{trace.message.sender}</dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Size</dt>
              <dd className="mt-1 text-sm text-gray-900">
                {trace.message.size ? `${Math.round(trace.message.size / 1024)} KB` : 'N/A'}
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Status</dt>
              <dd className="mt-1">
                <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(trace.message.status)}`}>
                  {trace.message.status}
                </span>
              </dd>
            </div>
            <div>
              <dt className="text-sm font-medium text-gray-500">Received</dt>
              <dd className="mt-1 text-sm text-gray-900">{formatTimestamp(trace.message.timestamp)}</dd>
            </div>
          </div>
        )}

        {/* Summary Stats */}
        <div className="mt-6 grid grid-cols-2 md:grid-cols-5 gap-4">
          <div className="bg-gray-50 p-3 rounded-lg">
            <dt className="text-sm font-medium text-gray-500">Recipients</dt>
            <dd className="mt-1 text-2xl font-semibold text-gray-900">{trace.summary.total_recipients}</dd>
          </div>
          <div className="bg-green-50 p-3 rounded-lg">
            <dt className="text-sm font-medium text-green-600">Delivered</dt>
            <dd className="mt-1 text-2xl font-semibold text-green-900">{trace.summary.delivered_count}</dd>
          </div>
          <div className="bg-yellow-50 p-3 rounded-lg">
            <dt className="text-sm font-medium text-yellow-600">Deferred</dt>
            <dd className="mt-1 text-2xl font-semibold text-yellow-900">{trace.summary.deferred_count}</dd>
          </div>
          <div className="bg-red-50 p-3 rounded-lg">
            <dt className="text-sm font-medium text-red-600">Bounced</dt>
            <dd className="mt-1 text-2xl font-semibold text-red-900">{trace.summary.bounced_count}</dd>
          </div>
          <div className="bg-blue-50 p-3 rounded-lg">
            <dt className="text-sm font-medium text-blue-600">Attempts</dt>
            <dd className="mt-1 text-2xl font-semibold text-blue-900">{trace.summary.total_attempts}</dd>
          </div>
        </div>

        {/* Average Delivery Time */}
        {trace.summary.average_delivery_time_seconds && (
          <div className="mt-4">
            <p className="text-sm text-gray-600">
              Average delivery time: <span className="font-medium">{formatDuration(trace.summary.average_delivery_time_seconds)}</span>
            </p>
          </div>
        )}
      </div>

      {/* Tabs */}
      <div className="bg-white shadow rounded-lg">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8 px-6">
            {[
              { key: 'timeline', label: 'Delivery Timeline', count: trace.delivery_timeline.length },
              { key: 'recipients', label: 'Recipients', count: trace.recipients.length },
              { key: 'retry', label: 'Retry Schedule', count: trace.retry_schedule.length },
              { key: 'threaded', label: 'Threaded View', count: null },
              { key: 'notes', label: 'Notes & Tags', count: null },
              { key: 'stats', label: 'Statistics', count: null },
            ].map((tab) => (
              <button
                key={tab.key}
                onClick={() => setActiveTab(tab.key as any)}
                className={`py-4 px-1 border-b-2 font-medium text-sm ${
                  activeTab === tab.key
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                {tab.label}
                {tab.count !== null && (
                  <span className="ml-2 bg-gray-100 text-gray-900 py-0.5 px-2.5 rounded-full text-xs">
                    {tab.count}
                  </span>
                )}
              </button>
            ))}
          </nav>
        </div>

        <div className="p-6">
          {activeTab === 'timeline' && (
            <DeliveryTimeline 
              timeline={trace.delivery_timeline} 
              messageId={trace.message_id}
            />
          )}
          {activeTab === 'recipients' && (
            <RecipientStatus 
              recipients={trace.recipients} 
              messageId={trace.message_id}
            />
          )}
          {activeTab === 'retry' && (
            <RetrySchedule 
              schedule={trace.retry_schedule} 
              messageId={trace.message_id}
            />
          )}
          {activeTab === 'threaded' && (
            <ThreadedTimeline 
              messageId={trace.message_id}
            />
          )}
          {activeTab === 'notes' && (
            <NotesAndTags 
              messageId={trace.message_id}
            />
          )}
          {activeTab === 'stats' && (
            <DeliveryStats 
              trace={trace}
            />
          )}
        </div>
      </div>
    </div>
  );
};