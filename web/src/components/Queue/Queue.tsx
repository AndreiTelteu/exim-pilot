import React, { useState } from 'react';
import { QueueMessage, QueueSearchFilters, QueueOperation } from '@/types/queue';
import { useQueue } from '@/hooks/useQueue';
import QueueList from './QueueList';
import QueueSearch from './QueueSearch';
import MessageDetails from './MessageDetails';
import BulkActions from './BulkActions';

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
    </div>
  );
}