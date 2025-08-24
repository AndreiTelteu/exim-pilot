import React, { useState } from 'react';
import { QueueMessage, QueueSearchFilters, QueueOperation } from '@/types/queue';
import { useQueue } from '@/hooks/useQueue';
import QueueList from './QueueList';
import QueueSearch from './QueueSearch';
import MessageDetails from './MessageDetails';
import BulkActions from './BulkActions';
import { HelpTooltip, HelpSection } from '../Common/HelpTooltip';
import { getHelpContent } from '../../utils/helpContent';

export default function Queue() {
  const [searchFilters, setSearchFilters] = useState<QueueSearchFilters>({});
  const [selectedMessages, setSelectedMessages] = useState<string[]>([]);
  const [selectedMessage, setSelectedMessage] = useState<QueueMessage | null>(null);
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  
  const { bulkOperation } = useQueue({ autoRefresh: false });

  const handleFiltersChange = (filters: QueueSearchFilters) => {
    setSearchFilters(filters);
  };

  const handleMessageSelect = (message: QueueMessage) => {
    setSelectedMessage(message);
  };

  const handleSelectionChange = (selectedIds: string[]) => {
    setSelectedMessages(selectedIds);
  };

  const handleCloseDetails = () => {
    setSelectedMessage(null);
  };

  const handleOperationComplete = () => {
    // Trigger refresh of the queue list
    setRefreshTrigger(prev => prev + 1);
  };

  const handleBulkOperation = async (operation: QueueOperation, messageIds: string[]) => {
    return await bulkOperation(operation, messageIds);
  };

  const handleClearSelection = () => {
    setSelectedMessages([]);
  };

  return (
    <div className="space-y-6">
      {/* Header with Help */}
      <div className="bg-white rounded-lg shadow-sm p-6">
        <div className="flex items-center gap-2 mb-4">
          <h1 className="text-2xl font-semibold text-gray-800">Mail Queue Management</h1>
          <HelpTooltip 
            content={getHelpContent('queue', 'overview')}
            position="right"
          />
        </div>
        <p className="text-gray-600">
          View, search, and manage messages in the mail queue. Use the search filters to find specific messages, 
          and perform individual or bulk operations as needed.
        </p>
      </div>

      {/* Search Interface */}
      <QueueSearch
        onFiltersChange={handleFiltersChange}
        initialFilters={searchFilters}
      />

      {/* Bulk Actions */}
      {selectedMessages.length > 0 && (
        <BulkActions
          selectedMessages={selectedMessages}
          onClearSelection={handleClearSelection}
          onBulkOperation={handleBulkOperation}
          onOperationComplete={handleOperationComplete}
        />
      )}

      {/* Queue List */}
      <QueueList
        searchFilters={searchFilters}
        onMessageSelect={handleMessageSelect}
        selectedMessages={selectedMessages}
        onSelectionChange={handleSelectionChange}
        refreshTrigger={refreshTrigger}
      />

      {/* Message Details Modal */}
      {selectedMessage && (
        <MessageDetails
          messageId={selectedMessage.id}
          onClose={handleCloseDetails}
          onOperationComplete={handleOperationComplete}
        />
      )}

      {/* Help Section */}
      <div className="space-y-4">
        <HelpSection title="Queue Management Help" className="bg-green-50 border-green-200">
          <div className="space-y-4">
            <div>
              <h4 className="font-semibold text-gray-800 mb-2">Message Status Types</h4>
              <ul className="space-y-2 text-sm text-gray-700">
                <li><strong>Queued:</strong> {getHelpContent('queue', 'status').split('.')[0]}</li>
                <li><strong>Deferred:</strong> Messages that failed delivery temporarily and are scheduled for retry</li>
                <li><strong>Frozen:</strong> Messages that have been paused and won't be retried until manually thawed</li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold text-gray-800 mb-2">Available Operations</h4>
              <ul className="space-y-2 text-sm text-gray-700">
                <li><strong>Deliver Now:</strong> {getHelpContent('queue', 'operations').split('deliver:')[1]?.split('freeze:')[0]?.trim()}</li>
                <li><strong>Freeze:</strong> Pause the message to prevent further delivery attempts</li>
                <li><strong>Thaw:</strong> Resume a frozen message and return it to normal retry scheduling</li>
                <li><strong>Delete:</strong> Permanently remove the message from the queue (cannot be undone)</li>
              </ul>
            </div>
            <div>
              <h4 className="font-semibold text-gray-800 mb-2">Search Tips</h4>
              <ul className="space-y-2 text-sm text-gray-700">
                <li>Use wildcards: <code>*@domain.com</code> to match all senders from a domain</li>
                <li>Age filters: <code>&gt;2h</code> (older than 2 hours), <code>&lt;30m</code> (newer than 30 minutes)</li>
                <li>Combine multiple filters for precise results</li>
                <li>Use bulk operations with caution - they affect all selected messages</li>
              </ul>
            </div>
          </div>
        </HelpSection>
      </div>
    </div>
  );
}