import React, { useMemo, useCallback, forwardRef } from 'react';
import { FixedSizeList as List, ListChildComponentProps } from 'react-window';
import InfiniteLoader from 'react-window-infinite-loader';

interface VirtualizedListProps<T> {
  items: T[];
  height: number;
  itemHeight: number;
  renderItem: (item: T, index: number, style: React.CSSProperties) => React.ReactNode;
  hasNextPage?: boolean;
  isNextPageLoading?: boolean;
  loadNextPage?: () => Promise<void>;
  className?: string;
  overscanCount?: number;
}

// Generic virtualized list component for performance optimization
function VirtualizedList<T>({
  items,
  height,
  itemHeight,
  renderItem,
  hasNextPage = false,
  isNextPageLoading = false,
  loadNextPage,
  className = '',
  overscanCount = 5,
}: VirtualizedListProps<T>) {
  // Calculate total item count including loading placeholder
  const itemCount = hasNextPage ? items.length + 1 : items.length;

  // Check if item is loaded
  const isItemLoaded = useCallback(
    (index: number) => !!items[index],
    [items]
  );

  // Item renderer component
  const ItemRenderer = useCallback(
    ({ index, style }: ListChildComponentProps) => {
      const item = items[index];
      
      // Show loading placeholder for unloaded items
      if (!item) {
        return (
          <div style={style} className="flex items-center justify-center py-4">
            <div className="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600"></div>
            <span className="ml-2 text-sm text-gray-500">Loading...</span>
          </div>
        );
      }

      return (
        <div style={style}>
          {renderItem(item, index, style)}
        </div>
      );
    },
    [items, renderItem]
  );

  // Memoize the list component to prevent unnecessary re-renders
  const listComponent = useMemo(() => {
    if (loadNextPage && hasNextPage) {
      return (
        <InfiniteLoader
          isItemLoaded={isItemLoaded}
          itemCount={itemCount}
          loadMoreItems={loadNextPage}
        >
          {({ onItemsRendered, ref }) => (
            <List
              ref={ref}
              height={height}
              itemCount={itemCount}
              itemSize={itemHeight}
              onItemsRendered={onItemsRendered}
              overscanCount={overscanCount}
              className={className}
            >
              {ItemRenderer}
            </List>
          )}
        </InfiniteLoader>
      );
    }

    return (
      <List
        height={height}
        itemCount={itemCount}
        itemSize={itemHeight}
        overscanCount={overscanCount}
        className={className}
      >
        {ItemRenderer}
      </List>
    );
  }, [
    height,
    itemCount,
    itemHeight,
    overscanCount,
    className,
    ItemRenderer,
    loadNextPage,
    hasNextPage,
    isItemLoaded,
  ]);

  return listComponent;
}

export default VirtualizedList;

// Specialized virtualized queue list component
interface VirtualizedQueueListProps {
  messages: any[];
  height: number;
  onMessageSelect?: (message: any) => void;
  selectedMessages?: string[];
  onSelectionChange?: (selectedIds: string[]) => void;
  hasNextPage?: boolean;
  isNextPageLoading?: boolean;
  loadNextPage?: () => Promise<void>;
}

export const VirtualizedQueueList: React.FC<VirtualizedQueueListProps> = ({
  messages,
  height,
  onMessageSelect,
  selectedMessages = [],
  onSelectionChange,
  hasNextPage,
  isNextPageLoading,
  loadNextPage,
}) => {
  const formatSize = useCallback((bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }, []);

  const getStatusBadgeClass = useCallback((status: string): string => {
    switch (status) {
      case 'queued':
        return 'bg-blue-100 text-blue-800';
      case 'deferred':
        return 'bg-yellow-100 text-yellow-800';
      case 'frozen':
        return 'bg-red-100 text-red-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  }, []);

  const handleSelectMessage = useCallback((messageId: string, checked: boolean) => {
    if (onSelectionChange) {
      const newSelection = checked
        ? [...selectedMessages, messageId]
        : selectedMessages.filter(id => id !== messageId);
      onSelectionChange(newSelection);
    }
  }, [selectedMessages, onSelectionChange]);

  const renderQueueItem = useCallback((message: any, index: number, style: React.CSSProperties) => {
    const isSelected = selectedMessages.includes(message.id);
    
    return (
      <div
        style={style}
        className={`flex items-center px-4 py-3 border-b border-gray-200 hover:bg-gray-50 ${
          isSelected ? 'bg-blue-50' : ''
        }`}
      >
        {onSelectionChange && (
          <div className="flex-shrink-0 mr-4">
            <input
              type="checkbox"
              className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-600"
              checked={isSelected}
              onChange={(e) => handleSelectMessage(message.id, e.target.checked)}
            />
          </div>
        )}
        
        <div className="flex-1 min-w-0 grid grid-cols-7 gap-4 items-center">
          {/* Message ID */}
          <div className="font-mono text-xs text-gray-900 truncate">
            {message.id}
          </div>
          
          {/* Sender */}
          <div className="text-sm text-gray-900 truncate" title={message.sender}>
            {message.sender}
          </div>
          
          {/* Recipients */}
          <div className="text-sm text-gray-900">
            {message.recipients.length === 1 ? (
              <div className="truncate" title={message.recipients[0]}>
                {message.recipients[0]}
              </div>
            ) : (
              <div>
                <div className="truncate" title={message.recipients[0]}>
                  {message.recipients[0]}
                </div>
                {message.recipients.length > 1 && (
                  <div className="text-xs text-gray-500">
                    +{message.recipients.length - 1} more
                  </div>
                )}
              </div>
            )}
          </div>
          
          {/* Size */}
          <div className="text-sm text-gray-900">
            {formatSize(message.size)}
          </div>
          
          {/* Age */}
          <div className="text-sm text-gray-900">
            {message.age}
          </div>
          
          {/* Status */}
          <div>
            <span className={`inline-flex rounded-full px-2 text-xs font-semibold leading-5 ${getStatusBadgeClass(message.status)}`}>
              {message.status}
            </span>
          </div>
          
          {/* Retries */}
          <div className="text-sm text-gray-900">
            {message.retry_count}
          </div>
        </div>
        
        <div className="flex-shrink-0 ml-4">
          <button
            onClick={() => onMessageSelect?.(message)}
            className="text-blue-600 hover:text-blue-900 text-sm font-medium"
          >
            View
          </button>
        </div>
      </div>
    );
  }, [
    selectedMessages,
    onSelectionChange,
    onMessageSelect,
    formatSize,
    getStatusBadgeClass,
    handleSelectMessage,
  ]);

  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      {/* Header */}
      <div className="bg-gray-50 px-4 py-3 border-b border-gray-200">
        <div className="grid grid-cols-7 gap-4 text-xs font-medium text-gray-500 uppercase tracking-wide">
          <div>Message ID</div>
          <div>Sender</div>
          <div>Recipients</div>
          <div>Size</div>
          <div>Age</div>
          <div>Status</div>
          <div>Retries</div>
        </div>
      </div>
      
      {/* Virtualized List */}
      <VirtualizedList
        items={messages}
        height={height}
        itemHeight={80}
        renderItem={renderQueueItem}
        hasNextPage={hasNextPage}
        isNextPageLoading={isNextPageLoading}
        loadNextPage={loadNextPage}
        overscanCount={10}
      />
    </div>
  );
};

// Specialized virtualized log list component
interface VirtualizedLogListProps {
  logs: any[];
  height: number;
  selectedLogs?: Set<number>;
  onSelectLog?: (logId: number) => void;
  hasNextPage?: boolean;
  isNextPageLoading?: boolean;
  loadNextPage?: () => Promise<void>;
}

export const VirtualizedLogList: React.FC<VirtualizedLogListProps> = ({
  logs,
  height,
  selectedLogs = new Set(),
  onSelectLog,
  hasNextPage,
  isNextPageLoading,
  loadNextPage,
}) => {
  const formatTimestamp = useCallback((timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  }, []);

  const getLogTypeColor = useCallback((logType: string) => {
    switch (logType) {
      case 'main':
        return 'bg-blue-100 text-blue-800';
      case 'reject':
        return 'bg-red-100 text-red-800';
      case 'panic':
        return 'bg-yellow-100 text-yellow-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  }, []);

  const getEventColor = useCallback((event: string) => {
    switch (event) {
      case 'arrival':
        return 'bg-green-100 text-green-800';
      case 'delivery':
        return 'bg-blue-100 text-blue-800';
      case 'defer':
        return 'bg-yellow-100 text-yellow-800';
      case 'bounce':
        return 'bg-red-100 text-red-800';
      case 'reject':
        return 'bg-red-100 text-red-800';
      case 'panic':
        return 'bg-orange-100 text-orange-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  }, []);

  const renderLogItem = useCallback((log: any, index: number, style: React.CSSProperties) => {
    const isSelected = selectedLogs.has(log.id);
    
    return (
      <div
        style={style}
        className={`flex items-center px-4 py-3 border-b border-gray-200 hover:bg-gray-50 ${
          isSelected ? 'bg-gray-50' : ''
        }`}
      >
        {onSelectLog && (
          <div className="flex-shrink-0 mr-4">
            <input
              type="checkbox"
              className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              checked={isSelected}
              onChange={() => onSelectLog(log.id)}
            />
          </div>
        )}
        
        <div className="flex-1 min-w-0 grid grid-cols-8 gap-4 items-center">
          {/* Timestamp */}
          <div className="text-sm text-gray-900">
            {formatTimestamp(log.timestamp)}
          </div>
          
          {/* Message ID */}
          <div className="font-mono text-sm text-gray-900 truncate">
            {log.message_id || '-'}
          </div>
          
          {/* Type */}
          <div>
            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getLogTypeColor(log.log_type)}`}>
              {log.log_type}
            </span>
          </div>
          
          {/* Event */}
          <div>
            <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${getEventColor(log.event)}`}>
              {log.event}
            </span>
          </div>
          
          {/* Sender */}
          <div className="text-sm text-gray-900 truncate">
            {log.sender || '-'}
          </div>
          
          {/* Recipients */}
          <div className="text-sm text-gray-900 truncate">
            {log.recipients ? log.recipients.join(', ') : '-'}
          </div>
          
          {/* Status */}
          <div className="text-sm text-gray-900">
            {log.status || '-'}
            {log.error_code && (
              <span className="ml-2 text-xs text-red-600">
                ({log.error_code})
              </span>
            )}
          </div>
          
          {/* Raw Line Preview */}
          <div className="text-xs text-gray-500 truncate" title={log.raw_line}>
            {log.raw_line}
          </div>
        </div>
      </div>
    );
  }, [
    selectedLogs,
    onSelectLog,
    formatTimestamp,
    getLogTypeColor,
    getEventColor,
  ]);

  return (
    <div className="bg-white shadow rounded-lg overflow-hidden">
      {/* Header */}
      <div className="bg-gray-50 px-4 py-3 border-b border-gray-200">
        <div className="grid grid-cols-8 gap-4 text-xs font-medium text-gray-500 uppercase tracking-wide">
          <div>Timestamp</div>
          <div>Message ID</div>
          <div>Type</div>
          <div>Event</div>
          <div>Sender</div>
          <div>Recipients</div>
          <div>Status</div>
          <div>Raw Line</div>
        </div>
      </div>
      
      {/* Virtualized List */}
      <VirtualizedList
        items={logs}
        height={height}
        itemHeight={60}
        renderItem={renderLogItem}
        hasNextPage={hasNextPage}
        isNextPageLoading={isNextPageLoading}
        loadNextPage={loadNextPage}
        overscanCount={15}
      />
    </div>
  );
};