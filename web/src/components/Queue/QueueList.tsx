import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { QueueMessage, QueueSearchFilters } from '@/types/queue';
import { APIResponse } from '@/types/api';
import { apiService } from '@/services/api';
import { webSocketService } from '@/services/websocket';
import { LoadingSpinner, Pagination } from '@/components/Common';
import { VirtualizedQueueList } from '@/components/Common/VirtualizedList';
import { useLazyLoading, useOptimizedDataFetching } from '@/hooks/useLazyLoading';

interface QueueListProps {
  searchFilters?: QueueSearchFilters;
  onMessageSelect?: (message: QueueMessage) => void;
  selectedMessages?: string[];
  onSelectionChange?: (selectedIds: string[]) => void;
  refreshTrigger?: number;
  useVirtualization?: boolean;
  height?: number;
}

type SortField = 'id' | 'sender' | 'recipients' | 'size' | 'age' | 'status' | 'retry_count';
type SortDirection = 'asc' | 'desc';

interface SortConfig {
  field: SortField;
  direction: SortDirection;
}

export default function QueueList({
  searchFilters,
  onMessageSelect,
  selectedMessages = [],
  onSelectionChange,
  refreshTrigger = 0,
  useVirtualization = false,
  height = 600,
}: QueueListProps) {
  const [messages, setMessages] = useState<QueueMessage[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [sortConfig, setSortConfig] = useState<SortConfig>({ field: 'age', direction: 'desc' });
  const itemsPerPage = 25;

  // Optimized fetch function for lazy loading
  const fetchMessagesPage = useCallback(async (page: number, pageSize: number) => {
    const params: any = {
      page,
      per_page: pageSize,
      sort_field: sortConfig.field,
      sort_direction: sortConfig.direction,
      ...searchFilters,
    };

    const response: APIResponse<any> = await apiService.get('/v1/queue', params);
    
    if (response.success && response.data) {
      return {
        data: response.data.messages || [],
        total: response.meta?.total || 0,
        hasMore: (response.meta?.page || 1) < (response.meta?.total_pages || 1),
      };
    } else {
      throw new Error(response.error || 'Failed to fetch queue messages');
    }
  }, [sortConfig, searchFilters]);

  // Use lazy loading for virtualized lists
  const lazyLoading = useLazyLoading(fetchMessagesPage, {
    initialPageSize: useVirtualization ? 100 : itemsPerPage,
    threshold: 0.8,
  });

  const fetchMessages = useCallback(async () => {
    if (useVirtualization) {
      // Reset lazy loading when filters change
      lazyLoading.reset();
      await lazyLoading.loadMore();
      return;
    }

    try {
      setLoading(true);
      setError(null);

      const result = await fetchMessagesPage(currentPage, itemsPerPage);
      setMessages(result.data);
      setTotalPages(Math.ceil(result.total / itemsPerPage));
      setTotalItems(result.total);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch queue messages');
    } finally {
      setLoading(false);
    }
  }, [currentPage, fetchMessagesPage, useVirtualization, itemsPerPage]);

  // Handle real-time updates via WebSocket
  useEffect(() => {
    const handleQueueUpdate = (data: any) => {
      if (data.type === 'queue_update') {
        // Refresh the current page when queue updates
        fetchMessages();
      }
    };

    webSocketService.on('queue_update', handleQueueUpdate);
    
    return () => {
      webSocketService.off('queue_update', handleQueueUpdate);
    };
  }, [fetchMessages]);

  useEffect(() => {
    fetchMessages();
  }, [fetchMessages]);

  const handleSort = (field: SortField) => {
    setSortConfig(prev => ({
      field,
      direction: prev.field === field && prev.direction === 'asc' ? 'desc' : 'asc',
    }));
    setCurrentPage(1); // Reset to first page when sorting
  };

  const handleSelectAll = (checked: boolean) => {
    if (onSelectionChange) {
      onSelectionChange(checked ? messages.map(m => m.id) : []);
    }
  };

  const handleSelectMessage = (messageId: string, checked: boolean) => {
    if (onSelectionChange) {
      const newSelection = checked
        ? [...selectedMessages, messageId]
        : selectedMessages.filter(id => id !== messageId);
      onSelectionChange(newSelection);
    }
  };

  const formatSize = (bytes: number): string => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  const formatAge = (age: string): string => {
    // Age comes as a string like "2h 30m" or "1d 5h"
    return age;
  };

  const getStatusBadgeClass = (status: string): string => {
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
  };

  const getSortIcon = (field: SortField) => {
    if (sortConfig.field !== field) {
      return (
        <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4" />
        </svg>
      );
    }
    
    return sortConfig.direction === 'asc' ? (
      <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
      </svg>
    ) : (
      <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
      </svg>
    );
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center py-12">
        <LoadingSpinner size="lg" />
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
            <h3 className="text-sm font-medium text-red-800">Error loading queue</h3>
            <div className="mt-2 text-sm text-red-700">
              <p>{error}</p>
            </div>
            <div className="mt-4">
              <button
                onClick={fetchMessages}
                className="bg-red-100 px-3 py-2 rounded-md text-sm font-medium text-red-800 hover:bg-red-200"
              >
                Try again
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Determine which messages to display
  const displayMessages = useVirtualization ? lazyLoading.items : messages;
  const displayTotalItems = useVirtualization ? lazyLoading.totalItems : totalItems;
  const displayLoading = useVirtualization ? lazyLoading.loading : loading;
  const displayError = useVirtualization ? lazyLoading.error : error;

  // Render virtualized list if enabled
  if (useVirtualization) {
    if (displayError) {
      return (
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">Error loading queue</h3>
              <div className="mt-2 text-sm text-red-700">
                <p>{displayError}</p>
              </div>
              <div className="mt-4">
                <button
                  onClick={() => {
                    lazyLoading.reset();
                    lazyLoading.loadMore();
                  }}
                  className="bg-red-100 px-3 py-2 rounded-md text-sm font-medium text-red-800 hover:bg-red-200"
                >
                  Try again
                </button>
              </div>
            </div>
          </div>
        </div>
      );
    }

    return (
      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <div className="sm:flex sm:items-center">
            <div className="sm:flex-auto">
              <h1 className="text-lg font-semibold text-gray-900">Mail Queue</h1>
              <p className="mt-2 text-sm text-gray-700">
                {displayTotalItems} message{displayTotalItems !== 1 ? 's' : ''} in queue
                {displayLoading && ' (Loading...)'}
              </p>
            </div>
          </div>

          <div className="mt-8">
            <VirtualizedQueueList
              messages={displayMessages}
              height={height}
              onMessageSelect={onMessageSelect}
              selectedMessages={selectedMessages}
              onSelectionChange={onSelectionChange}
              hasNextPage={lazyLoading.hasMore}
              isNextPageLoading={displayLoading}
              loadNextPage={lazyLoading.loadMore}
            />
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white shadow rounded-lg">
      <div className="px-4 py-5 sm:p-6">
        <div className="sm:flex sm:items-center">
          <div className="sm:flex-auto">
            <h1 className="text-lg font-semibold text-gray-900">Mail Queue</h1>
            <p className="mt-2 text-sm text-gray-700">
              {displayTotalItems} message{displayTotalItems !== 1 ? 's' : ''} in queue
            </p>
          </div>
        </div>

        <div className="mt-8 flow-root">
          <div className="-mx-4 -my-2 overflow-x-auto sm:-mx-6 lg:-mx-8">
            <div className="inline-block min-w-full py-2 align-middle sm:px-6 lg:px-8">
              <table className="min-w-full divide-y divide-gray-300">
                <thead>
                  <tr>
                    {onSelectionChange && (
                      <th scope="col" className="relative px-7 sm:w-12 sm:px-6">
                        <input
                          type="checkbox"
                          className="absolute left-4 top-1/2 -mt-2 h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-600"
                          checked={selectedMessages.length === messages.length && messages.length > 0}
                          onChange={(e) => handleSelectAll(e.target.checked)}
                        />
                      </th>
                    )}
                    <th
                      scope="col"
                      className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide cursor-pointer hover:bg-gray-50"
                      onClick={() => handleSort('id')}
                    >
                      <div className="flex items-center space-x-1">
                        <span>Message ID</span>
                        {getSortIcon('id')}
                      </div>
                    </th>
                    <th
                      scope="col"
                      className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide cursor-pointer hover:bg-gray-50"
                      onClick={() => handleSort('sender')}
                    >
                      <div className="flex items-center space-x-1">
                        <span>Sender</span>
                        {getSortIcon('sender')}
                      </div>
                    </th>
                    <th
                      scope="col"
                      className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide cursor-pointer hover:bg-gray-50"
                      onClick={() => handleSort('recipients')}
                    >
                      <div className="flex items-center space-x-1">
                        <span>Recipients</span>
                        {getSortIcon('recipients')}
                      </div>
                    </th>
                    <th
                      scope="col"
                      className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide cursor-pointer hover:bg-gray-50"
                      onClick={() => handleSort('size')}
                    >
                      <div className="flex items-center space-x-1">
                        <span>Size</span>
                        {getSortIcon('size')}
                      </div>
                    </th>
                    <th
                      scope="col"
                      className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide cursor-pointer hover:bg-gray-50"
                      onClick={() => handleSort('age')}
                    >
                      <div className="flex items-center space-x-1">
                        <span>Age</span>
                        {getSortIcon('age')}
                      </div>
                    </th>
                    <th
                      scope="col"
                      className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide cursor-pointer hover:bg-gray-50"
                      onClick={() => handleSort('status')}
                    >
                      <div className="flex items-center space-x-1">
                        <span>Status</span>
                        {getSortIcon('status')}
                      </div>
                    </th>
                    <th
                      scope="col"
                      className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide cursor-pointer hover:bg-gray-50"
                      onClick={() => handleSort('retry_count')}
                    >
                      <div className="flex items-center space-x-1">
                        <span>Retries</span>
                        {getSortIcon('retry_count')}
                      </div>
                    </th>
                    <th scope="col" className="relative py-3.5 pl-3 pr-4 sm:pr-6">
                      <span className="sr-only">Actions</span>
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {displayMessages.map((message) => (
                    <tr
                      key={message.id}
                      className={`hover:bg-gray-50 ${selectedMessages.includes(message.id) ? 'bg-blue-50' : ''}`}
                    >
                      {onSelectionChange && (
                        <td className="relative px-7 sm:w-12 sm:px-6">
                          <input
                            type="checkbox"
                            className="absolute left-4 top-1/2 -mt-2 h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-600"
                            checked={selectedMessages.includes(message.id)}
                            onChange={(e) => handleSelectMessage(message.id, e.target.checked)}
                          />
                        </td>
                      )}
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        <div className="font-mono text-xs">{message.id}</div>
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        <div className="max-w-xs truncate" title={message.sender}>
                          {message.sender}
                        </div>
                      </td>
                      <td className="px-3 py-4 text-sm text-gray-900">
                        <div className="max-w-xs">
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
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        {formatSize(message.size)}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        {formatAge(message.age)}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        <span className={`inline-flex rounded-full px-2 text-xs font-semibold leading-5 ${getStatusBadgeClass(message.status)}`}>
                          {message.status}
                        </span>
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        {message.retry_count}
                      </td>
                      <td className="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
                        <button
                          onClick={() => onMessageSelect?.(message)}
                          className="text-blue-600 hover:text-blue-900"
                        >
                          View
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>

              {displayMessages.length === 0 && (
                <div className="text-center py-12">
                  <svg
                    className="mx-auto h-12 w-12 text-gray-400"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                    aria-hidden="true"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2M4 13h2m13-8l-4 4m0 0l-4-4m4 4V3"
                    />
                  </svg>
                  <h3 className="mt-2 text-sm font-medium text-gray-900">No messages</h3>
                  <p className="mt-1 text-sm text-gray-500">
                    The mail queue is empty.
                  </p>
                </div>
              )}
            </div>
          </div>
        </div>

        {totalPages > 1 && (
          <Pagination
            currentPage={currentPage}
            totalPages={totalPages}
            onPageChange={setCurrentPage}
            totalItems={totalItems}
            itemsPerPage={itemsPerPage}
          />
        )}
      </div>
    </div>
  );
}