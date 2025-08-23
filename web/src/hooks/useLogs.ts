import { useState, useEffect, useCallback } from 'react';
import { LogEntry, LogSearchFilters } from '@/types/logs';
import { APIResponse } from '@/types/api';
import { apiService } from '@/services/api';

interface UseLogsOptions {
  autoFetch?: boolean;
  initialPage?: number;
  itemsPerPage?: number;
}

interface UseLogsReturn {
  logs: LogEntry[];
  loading: boolean;
  error: string | null;
  currentPage: number;
  totalPages: number;
  totalItems: number;
  filters: LogSearchFilters;
  fetchLogs: (page?: number, searchFilters?: LogSearchFilters) => Promise<void>;
  setPage: (page: number) => void;
  setFilters: (filters: LogSearchFilters) => void;
  refresh: () => Promise<void>;
}

export function useLogs(options: UseLogsOptions = {}): UseLogsReturn {
  const {
    autoFetch = true,
    initialPage = 1,
    itemsPerPage = 50,
  } = options;

  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPage, setCurrentPage] = useState(initialPage);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [filters, setFilters] = useState<LogSearchFilters>({});

  const fetchLogs = useCallback(async (page: number = currentPage, searchFilters: LogSearchFilters = filters) => {
    try {
      setLoading(true);
      setError(null);

      const params: Record<string, any> = {
        page,
        per_page: itemsPerPage,
        ...searchFilters,
      };

      // Remove empty filters
      Object.keys(params).forEach(key => {
        if (params[key] === '' || params[key] === undefined || params[key] === null) {
          delete params[key];
        }
      });

      const response: APIResponse<LogEntry[]> = await apiService.get('/v1/logs', params);
      
      if (response.success && response.data) {
        setLogs(response.data);
        setTotalPages(response.meta?.total_pages || 1);
        setTotalItems(response.meta?.total || 0);
        setCurrentPage(page);
      } else {
        throw new Error(response.error || 'Failed to fetch logs');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch logs');
      setLogs([]);
    } finally {
      setLoading(false);
    }
  }, [currentPage, filters, itemsPerPage]);

  const setPage = useCallback((page: number) => {
    setCurrentPage(page);
    fetchLogs(page, filters);
  }, [fetchLogs, filters]);

  const updateFilters = useCallback((newFilters: LogSearchFilters) => {
    setFilters(newFilters);
    setCurrentPage(1);
    fetchLogs(1, newFilters);
  }, [fetchLogs]);

  const refresh = useCallback(() => {
    return fetchLogs(currentPage, filters);
  }, [fetchLogs, currentPage, filters]);

  useEffect(() => {
    if (autoFetch) {
      fetchLogs();
    }
  }, []);

  return {
    logs,
    loading,
    error,
    currentPage,
    totalPages,
    totalItems,
    filters,
    fetchLogs,
    setPage,
    setFilters: updateFilters,
    refresh,
  };
}