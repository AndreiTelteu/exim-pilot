import React, { useState, useEffect } from 'react';
import { apiService } from '../../services/api';

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

interface DeliveryThread {
  thread_id: string;
  recipient?: string;
  thread_type: string;
  events: DeliveryTimelineEvent[];
  summary: string;
  status: string;
}

interface MessageNote {
  id: number;
  message_id: string;
  user_id: string;
  note: string;
  is_public: boolean;
  created_at: string;
  updated_at: string;
}

interface MessageTag {
  id: number;
  message_id: string;
  tag: string;
  color?: string;
  user_id: string;
  created_at: string;
}

interface CorrelatedIncident {
  id: string;
  type: string;
  title: string;
  description: string;
  severity: string;
  message_ids: string[];
  start_time: string;
  end_time?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

interface ThreadedTimelineView {
  message_id: string;
  threads: DeliveryThread[];
  notes: MessageNote[];
  tags: MessageTag[];
  correlated_incidents: CorrelatedIncident[];
}

interface ThreadedTimelineProps {
  messageId: string;
}

export const ThreadedTimeline: React.FC<ThreadedTimelineProps> = ({ messageId }) => {
  const [threadedView, setThreadedView] = useState<ThreadedTimelineView | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [expandedThreads, setExpandedThreads] = useState<Set<string>>(new Set());
  const [filterType, setFilterType] = useState<string>('');

  useEffect(() => {
    fetchThreadedTimeline();
  }, [messageId]);

  const fetchThreadedTimeline = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiService.get(`/messages/${messageId}/threaded-timeline`);
      setThreadedView(response.data as ThreadedTimelineView);
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load threaded timeline');
    } finally {
      setLoading(false);
    }
  };

  const toggleThread = (threadId: string) => {
    const newExpanded = new Set(expandedThreads);
    if (newExpanded.has(threadId)) {
      newExpanded.delete(threadId);
    } else {
      newExpanded.add(threadId);
    }
    setExpandedThreads(newExpanded);
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const getThreadStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'success':
        return 'bg-green-100 text-green-800 border-green-200';
      case 'error':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'warning':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'info':
        return 'bg-blue-100 text-blue-800 border-blue-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getThreadIcon = (threadType: string) => {
    switch (threadType) {
      case 'recipient':
        return (
          <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
            <path d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" />
          </svg>
        );
      case 'host':
        return (
          <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zm0 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V8zm0 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1v-2z" clipRule="evenodd" />
          </svg>
        );
      case 'system':
        return (
          <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clipRule="evenodd" />
          </svg>
        );
      default:
        return (
          <svg className="h-5 w-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
          </svg>
        );
    }
  };

  const getEventIcon = (eventType: string) => {
    switch (eventType.toLowerCase()) {
      case 'arrival':
        return '📨';
      case 'delivery':
        return '✅';
      case 'defer':
        return '⏳';
      case 'bounce':
        return '❌';
      case 'attempt':
        return '🔄';
      case 'freeze':
        return '🧊';
      case 'thaw':
        return '🔥';
      case 'delete':
        return '🗑️';
      default:
        return '📋';
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity.toLowerCase()) {
      case 'critical':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'high':
        return 'bg-orange-100 text-orange-800 border-orange-200';
      case 'medium':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'low':
        return 'bg-blue-100 text-blue-800 border-blue-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  // Filter threads based on type
  const filteredThreads = filterType 
    ? threadedView?.threads.filter(thread => thread.thread_type === filterType) || []
    : threadedView?.threads || [];

  if (loading) {
    return (
      <div className="flex justify-center items-center py-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4">
        <div className="text-sm text-red-700">{error}</div>
      </div>
    );
  }

  if (!threadedView) {
    return (
      <div className="text-center py-8">
        <p className="text-gray-500">No threaded timeline data available</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Filter Controls */}
      <div className="bg-gray-50 p-4 rounded-lg">
        <div className="flex items-center space-x-4">
          <label className="text-sm font-medium text-gray-700">Filter by thread type:</label>
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            className="border border-gray-300 rounded-md px-3 py-1 text-sm"
          >
            <option value="">All Threads</option>
            <option value="recipient">Recipients</option>
            <option value="host">Hosts</option>
            <option value="system">System</option>
          </select>
          <span className="text-sm text-gray-500">
            {filteredThreads.length} thread{filteredThreads.length !== 1 ? 's' : ''}
          </span>
        </div>
      </div>

      {/* Correlated Incidents */}
      {threadedView.correlated_incidents.length > 0 && (
        <div className="bg-white border border-gray-200 rounded-lg p-4">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Correlated Incidents</h3>
          <div className="space-y-3">
            {threadedView.correlated_incidents.map(incident => (
              <div key={incident.id} className={`border rounded-lg p-3 ${getSeverityColor(incident.severity)}`}>
                <div className="flex items-center justify-between">
                  <h4 className="font-medium">{incident.title}</h4>
                  <span className="text-xs px-2 py-1 rounded-full bg-white bg-opacity-50">
                    {incident.severity}
                  </span>
                </div>
                <p className="text-sm mt-1">{incident.description}</p>
                <div className="text-xs mt-2 space-x-4">
                  <span>Type: {incident.type}</span>
                  <span>Status: {incident.status}</span>
                  <span>Messages: {incident.message_ids.length}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Delivery Threads */}
      <div className="space-y-4">
        {filteredThreads.map(thread => (
          <div key={thread.thread_id} className={`border rounded-lg ${getThreadStatusColor(thread.status)}`}>
            <div 
              className="p-4 cursor-pointer hover:bg-opacity-50"
              onClick={() => toggleThread(thread.thread_id)}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  {getThreadIcon(thread.thread_type)}
                  <div>
                    <h4 className="font-medium">
                      {thread.thread_type === 'recipient' && thread.recipient ? (
                        `Recipient: ${thread.recipient}`
                      ) : thread.thread_type === 'host' ? (
                        `Host Thread`
                      ) : (
                        'System Events'
                      )}
                    </h4>
                    <p className="text-sm opacity-75">{thread.summary}</p>
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-xs px-2 py-1 rounded-full bg-white bg-opacity-50">
                    {thread.events.length} events
                  </span>
                  <svg 
                    className={`h-5 w-5 transform transition-transform ${
                      expandedThreads.has(thread.thread_id) ? 'rotate-180' : ''
                    }`} 
                    fill="currentColor" 
                    viewBox="0 0 20 20"
                  >
                    <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd" />
                  </svg>
                </div>
              </div>
            </div>

            {/* Thread Events */}
            {expandedThreads.has(thread.thread_id) && (
              <div className="border-t border-current border-opacity-20 p-4 bg-white bg-opacity-50">
                <div className="space-y-3">
                  {thread.events.map((event, index) => (
                    <div key={`${event.timestamp}-${index}`} className="flex items-start space-x-3">
                      <div className="flex-shrink-0 text-lg">
                        {getEventIcon(event.event_type)}
                      </div>
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between">
                          <p className="text-sm font-medium text-gray-900">
                            {event.description}
                          </p>
                          <span className="text-xs text-gray-500">
                            {formatTimestamp(event.timestamp)}
                          </span>
                        </div>
                        
                        {/* Event Details */}
                        <div className="mt-1 space-y-1 text-xs text-gray-600">
                          {event.host && (
                            <div>Host: {event.host}{event.ip_address && ` (${event.ip_address})`}</div>
                          )}
                          {event.smtp_code && (
                            <div>SMTP Code: {event.smtp_code}</div>
                          )}
                          {event.error_text && (
                            <div className="text-red-600 bg-red-50 p-1 rounded">
                              Error: {event.error_text}
                            </div>
                          )}
                          <div className="text-gray-500">
                            Source: {event.source}{event.source_id && ` #${event.source_id}`}
                          </div>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      {filteredThreads.length === 0 && (
        <div className="text-center py-8">
          <p className="text-gray-500">No threads found for the selected filter.</p>
        </div>
      )}
    </div>
  );
};