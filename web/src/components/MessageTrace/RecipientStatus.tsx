import React, { useState } from 'react';

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

interface RecipientStatusProps {
  recipients: RecipientDeliveryStatus[];
  messageId: string;
}

export const RecipientStatus: React.FC<RecipientStatusProps> = ({ recipients, messageId }) => {
  const [expandedRecipient, setExpandedRecipient] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<string>('');

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'delivered':
        return 'bg-green-100 text-green-800';
      case 'deferred':
        return 'bg-yellow-100 text-yellow-800';
      case 'bounced':
        return 'bg-red-100 text-red-800';
      case 'pending':
        return 'bg-blue-100 text-blue-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getAttemptStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'success':
        return 'text-green-600';
      case 'defer':
        return 'text-yellow-600';
      case 'bounce':
        return 'text-red-600';
      case 'timeout':
        return 'text-gray-600';
      default:
        return 'text-gray-600';
    }
  };

  const getAttemptStatusIcon = (status: string) => {
    switch (status.toLowerCase()) {
      case 'success':
        return (
          <svg className="h-4 w-4 text-green-500" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
          </svg>
        );
      case 'defer':
        return (
          <svg className="h-4 w-4 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
          </svg>
        );
      case 'bounce':
        return (
          <svg className="h-4 w-4 text-red-500" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
          </svg>
        );
      default:
        return (
          <svg className="h-4 w-4 text-gray-500" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
          </svg>
        );
    }
  };

  // Get unique statuses for filter
  const uniqueStatuses = [...new Set(recipients.map(r => r.status))];

  // Apply status filter
  const filteredRecipients = statusFilter 
    ? recipients.filter(r => r.status === statusFilter)
    : recipients;

  const toggleRecipientExpansion = (recipient: string) => {
    setExpandedRecipient(expandedRecipient === recipient ? null : recipient);
  };

  return (
    <div className="space-y-6">
      {/* Status Filter */}
      <div className="bg-gray-50 p-4 rounded-lg">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-medium text-gray-900">Filter by Status</h3>
          <div className="flex space-x-2">
            <button
              onClick={() => setStatusFilter('')}
              className={`px-3 py-1 text-xs font-medium rounded-full ${
                statusFilter === '' 
                  ? 'bg-blue-100 text-blue-800' 
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              All ({recipients.length})
            </button>
            {uniqueStatuses.map(status => {
              const count = recipients.filter(r => r.status === status).length;
              return (
                <button
                  key={status}
                  onClick={() => setStatusFilter(status)}
                  className={`px-3 py-1 text-xs font-medium rounded-full ${
                    statusFilter === status 
                      ? getStatusColor(status)
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                  }`}
                >
                  {status} ({count})
                </button>
              );
            })}
          </div>
        </div>
      </div>

      {/* Recipients List */}
      <div className="space-y-4">
        {filteredRecipients.map((recipient) => (
          <div key={recipient.recipient} className="bg-white border border-gray-200 rounded-lg">
            <div 
              className="p-4 cursor-pointer hover:bg-gray-50"
              onClick={() => toggleRecipientExpansion(recipient.recipient)}
            >
              <div className="flex items-center justify-between">
                <div className="flex-1 min-w-0">
                  <div className="flex items-center space-x-3">
                    <h4 className="text-sm font-medium text-gray-900 truncate">
                      {recipient.recipient}
                    </h4>
                    <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getStatusColor(recipient.status)}`}>
                      {recipient.status}
                    </span>
                    <span className="text-xs text-gray-500">
                      {recipient.attempt_count} attempt{recipient.attempt_count !== 1 ? 's' : ''}
                    </span>
                  </div>
                  
                  <div className="mt-2 grid grid-cols-1 md:grid-cols-3 gap-4 text-xs text-gray-600">
                    {recipient.delivered_at && (
                      <div>
                        <span className="font-medium">Delivered:</span> {formatTimestamp(recipient.delivered_at)}
                      </div>
                    )}
                    {recipient.last_attempt_at && (
                      <div>
                        <span className="font-medium">Last Attempt:</span> {formatTimestamp(recipient.last_attempt_at)}
                      </div>
                    )}
                    {recipient.next_retry_at && (
                      <div>
                        <span className="font-medium">Next Retry:</span> {formatTimestamp(recipient.next_retry_at)}
                      </div>
                    )}
                  </div>

                  {recipient.last_error_text && (
                    <div className="mt-2 text-xs text-red-600 bg-red-50 p-2 rounded">
                      <span className="font-medium">Last Error:</span> {recipient.last_error_text}
                      {recipient.last_smtp_code && ` (${recipient.last_smtp_code})`}
                    </div>
                  )}
                </div>
                
                <div className="flex-shrink-0">
                  <svg 
                    className={`h-5 w-5 text-gray-400 transform transition-transform ${
                      expandedRecipient === recipient.recipient ? 'rotate-180' : ''
                    }`} 
                    fill="currentColor" 
                    viewBox="0 0 20 20"
                  >
                    <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd" />
                  </svg>
                </div>
              </div>
            </div>

            {/* Expanded Delivery History */}
            {expandedRecipient === recipient.recipient && (
              <div className="border-t border-gray-200 p-4 bg-gray-50">
                <h5 className="text-sm font-medium text-gray-900 mb-3">Delivery History</h5>
                
                {recipient.delivery_history.length > 0 ? (
                  <div className="space-y-3">
                    {recipient.delivery_history.map((attempt, index) => (
                      <div key={attempt.id} className="bg-white p-3 rounded border">
                        <div className="flex items-start justify-between">
                          <div className="flex items-center space-x-2">
                            {getAttemptStatusIcon(attempt.status)}
                            <div>
                              <div className="text-sm font-medium text-gray-900">
                                Attempt #{index + 1}
                              </div>
                              <div className="text-xs text-gray-500">
                                {formatTimestamp(attempt.timestamp)}
                              </div>
                            </div>
                          </div>
                          <span className={`text-xs font-medium ${getAttemptStatusColor(attempt.status)}`}>
                            {attempt.status}
                          </span>
                        </div>
                        
                        <div className="mt-2 space-y-1 text-xs text-gray-600">
                          {attempt.host && (
                            <div>
                              <span className="font-medium">Host:</span> {attempt.host}
                              {attempt.ip_address && ` (${attempt.ip_address})`}
                            </div>
                          )}
                          {attempt.smtp_code && (
                            <div>
                              <span className="font-medium">SMTP Code:</span> {attempt.smtp_code}
                            </div>
                          )}
                          {attempt.error_message && (
                            <div className="text-red-600 bg-red-50 p-2 rounded">
                              <span className="font-medium">Error:</span> {attempt.error_message}
                            </div>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-sm text-gray-500">No delivery attempts recorded.</p>
                )}
              </div>
            )}
          </div>
        ))}
      </div>

      {filteredRecipients.length === 0 && (
        <div className="text-center py-8">
          <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900">No recipients found</h3>
          <p className="mt-1 text-sm text-gray-500">
            {recipients.length === 0 
              ? 'No recipients recorded for this message.'
              : 'No recipients match the current filter.'
            }
          </p>
        </div>
      )}
    </div>
  );
};