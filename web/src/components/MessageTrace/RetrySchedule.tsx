import React from 'react';

interface RetryScheduleEntry {
  recipient: string;
  scheduled_at: string;
  attempt_number: number;
  reason: string;
  is_estimated: boolean;
}

interface RetryScheduleProps {
  schedule: RetryScheduleEntry[];
  messageId: string;
}

export const RetrySchedule: React.FC<RetryScheduleProps> = ({ schedule, messageId }) => {
  const formatTimestamp = (timestamp: string) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = date.getTime() - now.getTime();
    const diffMinutes = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));

    let relativeTime = '';
    if (diffMs < 0) {
      relativeTime = 'Overdue';
    } else if (diffMinutes < 60) {
      relativeTime = `in ${diffMinutes} minute${diffMinutes !== 1 ? 's' : ''}`;
    } else if (diffHours < 24) {
      relativeTime = `in ${diffHours} hour${diffHours !== 1 ? 's' : ''}`;
    } else {
      relativeTime = `in ${diffDays} day${diffDays !== 1 ? 's' : ''}`;
    }

    return {
      absolute: date.toLocaleString(),
      relative: relativeTime,
      isPast: diffMs < 0,
      isNear: diffMs > 0 && diffMs < 60 * 60 * 1000, // within 1 hour
    };
  };

  const getRetryIcon = (attemptNumber: number, isEstimated: boolean) => {
    if (isEstimated) {
      return (
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-blue-100">
          <svg className="h-4 w-4 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
          </svg>
        </div>
      );
    } else {
      return (
        <div className="flex h-8 w-8 items-center justify-center rounded-full bg-green-100">
          <svg className="h-4 w-4 text-green-600" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M4 2a1 1 0 011 1v2.101a7.002 7.002 0 0111.601 2.566 1 1 0 11-1.885.666A5.002 5.002 0 005.999 7H9a1 1 0 010 2H4a1 1 0 01-1-1V3a1 1 0 011-1zm.008 9.057a1 1 0 011.276.61A5.002 5.002 0 0014.001 13H11a1 1 0 110-2h5a1 1 0 011 1v5a1 1 0 11-2 0v-2.101a7.002 7.002 0 01-11.601-2.566 1 1 0 01.61-1.276z" clipRule="evenodd" />
          </svg>
        </div>
      );
    }
  };

  // Sort schedule by scheduled time
  const sortedSchedule = [...schedule].sort((a, b) => 
    new Date(a.scheduled_at).getTime() - new Date(b.scheduled_at).getTime()
  );

  // Group by recipient
  const groupedByRecipient = sortedSchedule.reduce((acc, entry) => {
    if (!acc[entry.recipient]) {
      acc[entry.recipient] = [];
    }
    acc[entry.recipient].push(entry);
    return acc;
  }, {} as Record<string, RetryScheduleEntry[]>);

  return (
    <div className="space-y-6">
      {schedule.length > 0 ? (
        <>
          {/* Summary */}
          <div className="bg-blue-50 p-4 rounded-lg">
            <div className="flex items-center">
              <svg className="h-5 w-5 text-blue-400 mr-2" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clipRule="evenodd" />
              </svg>
              <h3 className="text-sm font-medium text-blue-900">
                {schedule.length} scheduled retr{schedule.length === 1 ? 'y' : 'ies'} for {Object.keys(groupedByRecipient).length} recipient{Object.keys(groupedByRecipient).length === 1 ? '' : 's'}
              </h3>
            </div>
            <p className="mt-1 text-sm text-blue-700">
              {schedule.filter(s => s.is_estimated).length} estimated, {schedule.filter(s => !s.is_estimated).length} confirmed
            </p>
          </div>

          {/* Timeline View */}
          <div className="space-y-6">
            <h3 className="text-lg font-medium text-gray-900">Retry Timeline</h3>
            
            <div className="flow-root">
              <ul className="-mb-8">
                {sortedSchedule.map((entry, entryIdx) => {
                  const timing = formatTimestamp(entry.scheduled_at);
                  
                  return (
                    <li key={`${entry.recipient}-${entry.attempt_number}`}>
                      <div className="relative pb-8">
                        {entryIdx !== sortedSchedule.length - 1 ? (
                          <span
                            className="absolute top-4 left-4 -ml-px h-full w-0.5 bg-gray-200"
                            aria-hidden="true"
                          />
                        ) : null}
                        <div className="relative flex space-x-3">
                          <div>{getRetryIcon(entry.attempt_number, entry.is_estimated)}</div>
                          <div className="flex min-w-0 flex-1 justify-between space-x-4 pt-1.5">
                            <div className="min-w-0 flex-1">
                              <div className="flex items-center space-x-2">
                                <p className="text-sm font-medium text-gray-900">
                                  Attempt #{entry.attempt_number} for {entry.recipient}
                                </p>
                                {entry.is_estimated && (
                                  <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                                    Estimated
                                  </span>
                                )}
                                {timing.isPast && (
                                  <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-red-100 text-red-800">
                                    Overdue
                                  </span>
                                )}
                                {timing.isNear && !timing.isPast && (
                                  <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-yellow-100 text-yellow-800">
                                    Soon
                                  </span>
                                )}
                              </div>
                              
                              <p className="mt-1 text-sm text-gray-600">
                                {entry.reason}
                              </p>
                            </div>
                            <div className="whitespace-nowrap text-right text-sm">
                              <div className={`font-medium ${timing.isPast ? 'text-red-600' : timing.isNear ? 'text-yellow-600' : 'text-gray-900'}`}>
                                {timing.relative}
                              </div>
                              <div className="text-gray-500 text-xs">
                                {timing.absolute}
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </li>
                  );
                })}
              </ul>
            </div>
          </div>

          {/* Grouped by Recipient */}
          <div className="space-y-4">
            <h3 className="text-lg font-medium text-gray-900">By Recipient</h3>
            
            {Object.entries(groupedByRecipient).map(([recipient, entries]) => (
              <div key={recipient} className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-900 mb-3">{recipient}</h4>
                
                <div className="space-y-2">
                  {entries.map((entry) => {
                    const timing = formatTimestamp(entry.scheduled_at);
                    
                    return (
                      <div key={entry.attempt_number} className="flex items-center justify-between py-2 px-3 bg-gray-50 rounded">
                        <div className="flex items-center space-x-3">
                          <span className="text-sm font-medium text-gray-700">
                            Attempt #{entry.attempt_number}
                          </span>
                          {entry.is_estimated && (
                            <span className="inline-flex px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800">
                              Estimated
                            </span>
                          )}
                          <span className="text-sm text-gray-600">
                            {entry.reason}
                          </span>
                        </div>
                        <div className="text-right">
                          <div className={`text-sm font-medium ${timing.isPast ? 'text-red-600' : timing.isNear ? 'text-yellow-600' : 'text-gray-900'}`}>
                            {timing.relative}
                          </div>
                          <div className="text-xs text-gray-500">
                            {timing.absolute}
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>
            ))}
          </div>
        </>
      ) : (
        <div className="text-center py-8">
          <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900">No retries scheduled</h3>
          <p className="mt-1 text-sm text-gray-500">
            All recipients have been successfully delivered or permanently failed.
          </p>
        </div>
      )}
    </div>
  );
};