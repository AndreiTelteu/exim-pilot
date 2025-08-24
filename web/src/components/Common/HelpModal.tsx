import React, { useState } from 'react';
import { QuestionMarkCircleIcon, XMarkIcon } from '@heroicons/react/24/outline';
import { getHelpContent } from '../../utils/helpContent';

interface HelpModalProps {
  isOpen: boolean;
  onClose: () => void;
  section: string;
}

export const HelpModal: React.FC<HelpModalProps> = ({ isOpen, onClose, section }) => {
  if (!isOpen) return null;

  const getHelpSections = (section: string) => {
    switch (section) {
      case 'dashboard':
        return {
          title: 'Dashboard Help',
          sections: [
            {
              title: 'Overview',
              content: getHelpContent('dashboard', 'overview')
            },
            {
              title: 'Key Metrics',
              content: (
                <div className="space-y-3">
                  <div><strong>Queue Messages:</strong> {getHelpContent('dashboard', 'queueMessages')}</div>
                  <div><strong>Delivered Today:</strong> {getHelpContent('dashboard', 'deliveredToday')}</div>
                  <div><strong>Deferred:</strong> {getHelpContent('dashboard', 'deferred')}</div>
                  <div><strong>Frozen:</strong> {getHelpContent('dashboard', 'frozen')}</div>
                </div>
              )
            },
            {
              title: 'Weekly Chart',
              content: getHelpContent('dashboard', 'weeklyChart')
            }
          ]
        };
      
      case 'queue':
        return {
          title: 'Queue Management Help',
          sections: [
            {
              title: 'Overview',
              content: getHelpContent('queue', 'overview')
            },
            {
              title: 'Message Information',
              content: (
                <div className="space-y-3">
                  <div><strong>Message ID:</strong> {getHelpContent('queue', 'messageId')}</div>
                  <div><strong>Sender:</strong> {getHelpContent('queue', 'sender')}</div>
                  <div><strong>Recipients:</strong> {getHelpContent('queue', 'recipients')}</div>
                  <div><strong>Size:</strong> {getHelpContent('queue', 'size')}</div>
                  <div><strong>Age:</strong> {getHelpContent('queue', 'age')}</div>
                  <div><strong>Retry Count:</strong> {getHelpContent('queue', 'retryCount')}</div>
                </div>
              )
            },
            {
              title: 'Message Operations',
              content: (
                <div className="space-y-3">
                  <div><strong>Deliver Now:</strong> Force immediate delivery attempt, bypassing the normal retry schedule.</div>
                  <div><strong>Freeze:</strong> Pause the message to prevent further delivery attempts.</div>
                  <div><strong>Thaw:</strong> Resume a frozen message and return it to normal retry scheduling.</div>
                  <div><strong>Delete:</strong> Permanently remove the message from the queue. This action cannot be undone.</div>
                </div>
              )
            },
            {
              title: 'Search and Filtering',
              content: (
                <div className="space-y-3">
                  <div><strong>Sender Filter:</strong> Use wildcards like *@domain.com to match all senders from a domain</div>
                  <div><strong>Age Filter:</strong> Examples: >2h (older than 2 hours), <30m (newer than 30 minutes)</div>
                  <div><strong>Status Filter:</strong> Filter by queued, deferred, or frozen status</div>
                  <div><strong>Bulk Operations:</strong> Select multiple messages for batch operations</div>
                </div>
              )
            }
          ]
        };
      
      case 'logs':
        return {
          title: 'Log Viewer Help',
          sections: [
            {
              title: 'Overview',
              content: getHelpContent('logs', 'overview')
            },
            {
              title: 'Log Types',
              content: (
                <div className="space-y-3">
                  <div><strong>Main Log:</strong> Contains message arrivals, deliveries, deferrals, and bounces</div>
                  <div><strong>Reject Log:</strong> Contains rejected messages and connections</div>
                  <div><strong>Panic Log:</strong> Contains daemon errors and critical failures</div>
                </div>
              )
            },
            {
              title: 'Real-time Monitoring',
              content: getHelpContent('logs', 'realTimeTail')
            },
            {
              title: 'Search and Export',
              content: (
                <div className="space-y-3">
                  <div><strong>Date Range:</strong> Specify start and end dates to limit search</div>
                  <div><strong>Text Search:</strong> Search within log entry text for keywords</div>
                  <div><strong>Export:</strong> Download filtered logs in CSV or TXT format</div>
                </div>
              )
            }
          ]
        };
      
      case 'reports':
        return {
          title: 'Reports and Analytics Help',
          sections: [
            {
              title: 'Overview',
              content: getHelpContent('reports', 'overview')
            },
            {
              title: 'Deliverability Reports',
              content: (
                <div className="space-y-3">
                  <div><strong>Success Rate:</strong> Percentage of messages delivered successfully</div>
                  <div><strong>Defer Rate:</strong> Percentage of messages that failed temporarily</div>
                  <div><strong>Bounce Rate:</strong> Percentage of messages that failed permanently</div>
                  <div><strong>Domain Analysis:</strong> Per-domain deliverability statistics</div>
                </div>
              )
            },
            {
              title: 'Volume Analysis',
              content: (
                <div className="space-y-3">
                  <div><strong>Traffic Trends:</strong> Hourly and daily volume patterns</div>
                  <div><strong>Peak Hours:</strong> Busiest times for capacity planning</div>
                  <div><strong>Growth Trends:</strong> Volume changes over time</div>
                </div>
              )
            },
            {
              title: 'Failure Analysis',
              content: (
                <div className="space-y-3">
                  <div><strong>Failure Categories:</strong> Common failure types grouped by cause</div>
                  <div><strong>Bounce Codes:</strong> SMTP response codes and frequencies</div>
                  <div><strong>Problem Domains:</strong> Domains with high failure rates</div>
                </div>
              )
            }
          ]
        };
      
      default:
        return {
          title: 'Help',
          sections: [
            {
              title: 'General Help',
              content: 'Help content not available for this section.'
            }
          ]
        };
    }
  };

  const helpData = getHelpSections(section);

  return (
    <div className="fixed inset-0 z-50 overflow-y-auto">
      <div className="flex items-center justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
        {/* Background overlay */}
        <div 
          className="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75" 
          onClick={onClose}
        />

        {/* Modal positioning */}
        <span className="hidden sm:inline-block sm:align-middle sm:h-screen">&#8203;</span>

        {/* Modal content */}
        <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-4xl sm:w-full">
          {/* Header */}
          <div className="bg-white px-6 pt-6 pb-4 border-b border-gray-200">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <QuestionMarkCircleIcon className="h-8 w-8 text-blue-600" />
                <h3 className="text-xl font-semibold text-gray-900">{helpData.title}</h3>
              </div>
              <button
                type="button"
                className="text-gray-400 hover:text-gray-600 focus:outline-none focus:text-gray-600 transition-colors"
                onClick={onClose}
              >
                <XMarkIcon className="h-6 w-6" />
              </button>
            </div>
          </div>

          {/* Content */}
          <div className="bg-white px-6 py-4 max-h-96 overflow-y-auto">
            <div className="space-y-6">
              {helpData.sections.map((section, index) => (
                <div key={index} className="border-b border-gray-100 last:border-b-0 pb-4 last:pb-0">
                  <h4 className="text-lg font-semibold text-gray-800 mb-3">{section.title}</h4>
                  <div className="text-sm text-gray-700 leading-relaxed">
                    {typeof section.content === 'string' ? (
                      <p>{section.content}</p>
                    ) : (
                      section.content
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Footer */}
          <div className="bg-gray-50 px-6 py-4 sm:flex sm:flex-row-reverse">
            <button
              type="button"
              className="w-full inline-flex justify-center rounded-md border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm transition-colors"
              onClick={onClose}
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

interface HelpButtonProps {
  section: string;
  className?: string;
}

export const HelpButton: React.FC<HelpButtonProps> = ({ section, className = '' }) => {
  const [isModalOpen, setIsModalOpen] = useState(false);

  return (
    <>
      <button
        type="button"
        className={`inline-flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors ${className}`}
        onClick={() => setIsModalOpen(true)}
      >
        <QuestionMarkCircleIcon className="h-4 w-4" />
        Help
      </button>
      
      <HelpModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        section={section}
      />
    </>
  );
};