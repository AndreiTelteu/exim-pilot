import { useState, useEffect, useCallback } from 'react';
import { QueueMessage, QueueSearchFilters, QueueMetrics } from '@/types/queue';
import { APIResponse } from '@/types/api';
import { apiService } from '@/services/api';
import { webSocketService } from '@/services/websocket';

interface UseQueueOptions {
  autoRefresh?: boolean;
  refreshInterval?: number;
}

export function useQueue(options: UseQueueOptions = {}) {
  const { autoRefresh = true, refreshInterval = 30000 } = options;
  
  const [messages, setMessages] = useState<QueueMessage[]>([]);
  const [metrics, setMetrics] = useState<QueueMetrics | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch queue messages
  const fetchMessages = useCallback(async (
    filters?: QueueSearchFilters,
    page: number = 1,
    perPage: number = 25,
    sortField: string = 'age',
    sortDirection: string = 'desc'
  ): Promise<APIResponse<QueueMessage[]>> => {
    try {
      setLoading(true);
      setError(null);

      const params: any = {
        page,
        per_page: perPage,
        sort_field: sortField,
        sort_direction: sortDirection,
        ...filters,
      };

      const response = await apiService.get<QueueMessage[]>('/v1/queue', params);
      
      if (response.success && response.data) {
        setMessages(response.data);
      }
      
      return response;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch queue messages';
      setError(errorMessage);
      throw err;
    } finally {
      setLoading(false);
    }
  }, []);

  // Fetch queue metrics
  const fetchMetrics = useCallback(async () => {
    try {
      const response = await apiService.get<QueueMetrics>('/v1/queue/metrics');
      if (response.success && response.data) {
        setMetrics(response.data);
      }
    } catch (err) {
      console.error('Failed to fetch queue metrics:', err);
    }
  }, []);

  // Queue operations
  const deliverMessage = useCallback(async (messageId: string) => {
    try {
      const response = await apiService.post(`/v1/queue/${messageId}/deliver`);
      if (response.success) {
        // Refresh messages after operation
        await fetchMessages();
      }
      return response;
    } catch (err) {
      throw err;
    }
  }, [fetchMessages]);

  const freezeMessage = useCallback(async (messageId: string) => {
    try {
      const response = await apiService.post(`/v1/queue/${messageId}/freeze`);
      if (response.success) {
        await fetchMessages();
      }
      return response;
    } catch (err) {
      throw err;
    }
  }, [fetchMessages]);

  const thawMessage = useCallback(async (messageId: string) => {
    try {
      const response = await apiService.post(`/v1/queue/${messageId}/thaw`);
      if (response.success) {
        await fetchMessages();
      }
      return response;
    } catch (err) {
      throw err;
    }
  }, [fetchMessages]);

  const deleteMessage = useCallback(async (messageId: string) => {
    try {
      const response = await apiService.delete(`/v1/queue/${messageId}`);
      if (response.success) {
        await fetchMessages();
      }
      return response;
    } catch (err) {
      throw err;
    }
  }, [fetchMessages]);

  // Bulk operations
  const bulkOperation = useCallback(async (
    operation: 'deliver' | 'freeze' | 'thaw' | 'delete',
    messageIds: string[]
  ) => {
    try {
      const response = await apiService.post('/v1/queue/bulk', {
        operation,
        message_ids: messageIds,
      });
      if (response.success) {
        await fetchMessages();
      }
      return response;
    } catch (err) {
      throw err;
    }
  }, [fetchMessages]);

  // WebSocket event handlers
  useEffect(() => {
    if (!autoRefresh) return;

    const handleQueueUpdate = (data: any) => {
      if (data.type === 'queue_update') {
        // Refresh messages when queue updates
        fetchMessages();
        fetchMetrics();
      }
    };

    webSocketService.on('queue_update', handleQueueUpdate);
    
    return () => {
      webSocketService.off('queue_update', handleQueueUpdate);
    };
  }, [autoRefresh, fetchMessages, fetchMetrics]);

  // Auto-refresh interval
  useEffect(() => {
    if (!autoRefresh) return;

    const interval = setInterval(() => {
      fetchMessages();
      fetchMetrics();
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval, fetchMessages, fetchMetrics]);

  return {
    messages,
    metrics,
    loading,
    error,
    fetchMessages,
    fetchMetrics,
    deliverMessage,
    freezeMessage,
    thawMessage,
    deleteMessage,
    bulkOperation,
  };
}