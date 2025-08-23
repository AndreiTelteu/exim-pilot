import React, { useState } from 'react';

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

interface DeliveryTimelineProps {
  timeline: DeliveryTimelineEvent[];
  messageId: string;
}

export const DeliveryTimeline: React.FC<DeliveryTimelineProps> = ({ timeline, messageId }) => {
  const [filter, setFilter] = useState({
    eventType: '',
    recipient: '',
    source: '',
  });

  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    return {
      date: date.toLocaleDateString(),
      time: date.toLocaleTimeString(),
    };
  };

  const getEventIcon = (eventType: string) => {
    switch (eventType.toLowerCase()) {
      case 'arrival':
        return (
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-100">
            <svg className="h-4 w-4 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-8.293l-3-3a1 1 0 00-1.414 0l-3 3a1 1 0 001.414 1.414L9 9.414V13a1 1 0 102 0V9.414l1.293 1.293a1 1 0 001.414-1.414z" clipRule="evenodd" />
            </svg>
          </div>
        );
      case 'delivery':
        return (
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-green-100">
            <svg className="h-4 w-4 text-green-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
            </svg>
          </div>
        );
      case 'defer':
        return (
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-yellow-100">
            <svg className="h-4 w-4 text-yellow-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
            </svg>
          </div>
        );
      case 'bounce':
        return (
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-red-100">
            <svg className="h-4 w-4 text-red-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
          </div>
        );
      case 'attempt':
        return (
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-purple-100">
            <svg className="h-4 w-4 text-purple-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clipRule="evenodd" />
            </svg>
          </div>
        );
      case 'freeze':
        return (
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-100">
            <svg className="h-4 w-4 text-gray-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
            </svg>
          </div>
        );
      case 'thaw':
        return (
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-100">
            <svg className="h-4 w-4 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
              <path d="M10 2L3 7v11a1 1 0 001 1h12a1 1 0 001-1V7l-7-5zM8 15v-3h4v3H8z" />
            </svg>
          </div>
        );
      case 'delete':
        return (
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-red-100">
            <svg className="h-4 w-4 text-red-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M9 2a1 1 0 000 2h2a1 1 0 100-2H9z" clipRule="evenodd" />
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8 7a1 1 0 012 0v6a1 1 0 11-2 0V7zm4 0a1 1 0 012 0v6a1 1 0 11-2 0V7z" clipRule="evenodd" />
            </svg>
          </div>
        );
      default:
        return (
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-gray-100">
            <svg className="h-4 w-4 text-gray-600" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
            </svg>
          </div>
        );
    }
  };

  const getSourceBadgeColor = (source: string) => {
    switch (source.toLowerCase()) {
      case 'log':
        return 'bg-blue-100 text-blue-800';
      case 'queue':
        return 'bg-green-100 text-green-800';
      case 'audit':
        return 'bg-purple-100 text-purple-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  // Get unique values for filters
  const uniqueEventTypes = [...new Set(timeline.map(event => event.event_type))];
  const uniqueRecipients = [...new Set(timeline.map(event => event.recipient).filter(Boolean))];
  const uniqueSources = [...new Set(timeline.map(event => event.source))];

  // Apply filters
  const filteredTimeline = timeline.filter(event => {
    if (filter.eventType && event.event_type !== filter.eventType) return false;
    if (filter.recipient && event.recipient !== filter.recipient) return false;
    if (filter.source && event.source !== filter.source) return false;
    return true;
  });

  return (
    <div className="space-y-6">
      {/* Filters */}
      <div className="bg-gray-50 p-4 rounded-lg">
        <h3 className="text-sm font-medium text-gray-900 mb-3">Filters</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1">Event Type</label>
            <select
              value={filter.eventType}
              onChange={(e) => setFilter({ ...filter, eventType: e.target.value })}
              className="w-full text-sm border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">All Events</option>
              {uniqueEventTypes.map(type => (
                <option key={type} value={type}>{type}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1">Recipient</label>
            <select
              value={filter.recipient}
              onChange={(e) => setFilter({ ...filter, recipient: e.target.value })}
              className="w-full text-sm border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">All Recipients</option>
              {uniqueRecipients.map(recipient => (
                <option key={recipient} value={recipient}>{recipient}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="block text-xs font-medium text-gray-700 mb-1">Source</label>
            <select
              value={filter.source}
              onChange={(e) => setFilter({ ...filter, source: e.target.value })}
              className="w-full text-sm border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="">All Sources</option>
              {uniqueSources.map(source => (
                <option key={source} value={source}>{source}</option>
              ))}
            </select>
          </div>
        </div>
        {(filter.eventType || filter.recipient || filter.source) && (
          <div className="mt-3">
            <button
              onClick={() => setFilter({ eventType: '', recipient: '', source: '' })}
              className="text-sm text-blue-600 hover:text-blue-800"
            >
              Clear all filters
            </button>
          </div>
        )}
      </div>

      {/* Timeline */}
      <div className="flow-root">
        <ul className="-mb-8">
          {filteredTimeline.map((event, eventIdx) => {
            const { date, time } = formatTimestamp(event.timestamp);
            
            return (
              <li key={`${event.timestamp}-${eventIdx}`}>
                <div className="relative pb-8">
                  {eventIdx !== filteredTimeline.length - 1 ? (
                    <span
                      className="absolute top-4 left-4 -ml-px h-full w-0.5 bg-gray-200"
                      aria-hidden="true"
                    />
                  ) : null}
                  <div className="relative flex space-x-3">
                    <div>{getEventIcon(event.event_type)}</div>
                    <div className="flex min-w-0 flex-1 justify-between space-x-4 pt-1.5">
                      <div className="min-w-0 flex-1">
                        <p className="text-sm font-medium text-gray-900">
                          {event.description}
                        </p>
                        
                        {/* Event Details */}
                        <div className="mt-2 space-y-1">
                          {event.recipient && (
                            <p className="text-xs text-gray-600">
                              <span className="font-medium">Recipient:</span> {event.recipient}
                            </p>
                          )}
                          {event.host && (
                            <p className="text-xs text-gray-600">
                              <span className="font-medium">Host:</span> {event.host}
                              {event.ip_address && ` (${event.ip_address})`}
                            </p>
                          )}
                          {event.smtp_code && (
                            <p className="text-xs text-gray-600">
                              <span className="font-medium">SMTP Code:</span> {event.smtp_code}
                            </p>
                          )}
                          {event.error_text && (
                            <p className="text-xs text-red-600 bg-red-50 p-2 rounded">
                              <span className="font-medium">Error:</span> {event.error_text}
                            </p>
                          )}
                        </div>

                        {/* Source Badge */}
                        <div className="mt-2">
                          <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${getSourceBadgeColor(event.source)}`}>
                            {event.source}
                            {event.source_id && ` #${event.source_id}`}
                          </span>
                        </div>
                      </div>
                      <div className="whitespace-nowrap text-right text-sm text-gray-500">
                        <div>{date}</div>
                        <div className="font-medium">{time}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </li>
            );
          })}
        </ul>
      </div>

      {filteredTimeline.length === 0 && (
        <div className="text-center py-8">
          <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900">No events found</h3>
          <p className="mt-1 text-sm text-gray-500">
            {timeline.length === 0 
              ? 'No delivery events recorded for this message.'
              : 'No events match the current filters.'
            }
          </p>
        </div>
      )}
    </div>
  );
};