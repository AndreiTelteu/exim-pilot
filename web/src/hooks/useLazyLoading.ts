import { useState, useCallback, useRef, useEffect } from 'react';

interface LazyLoadingConfig {
  initialPageSize: number;
  maxPageSize?: number;
  threshold?: number; // How close to the end before loading more
}

interface LazyLoadingState<T> {
  items: T[];
  loading: boolean;
  error: string | null;
  hasMore: boolean;
  currentPage: number;
  totalItems: number;
}

interface LazyLoadingActions<T> {
  loadMore: () => Promise<void>;
  reset: () => void;
  setItems: (items: T[]) => void;
  appendItems: (items: T[]) => void;
}

type LazyLoadingReturn<T> = LazyLoadingState<T> & LazyLoadingActions<T>;

// Custom hook for lazy loading with infinite scroll
export function useLazyLoading<T>(
  fetchFunction: (page: number, pageSize: number) => Promise<{
    data: T[];
    total: number;
    hasMore: boolean;
  }>,
  config: LazyLoadingConfig
): LazyLoadingReturn<T> {
  const [state, setState] = useState<LazyLoadingState<T>>({
    items: [],
    loading: false,
    error: null,
    hasMore: true,
    currentPage: 1,
    totalItems: 0,
  });

  const loadingRef = useRef(false);
  const configRef = useRef(config);
  configRef.current = config;

  const loadMore = useCallback(async () => {
    if (loadingRef.current || !state.hasMore) {
      return;
    }

    loadingRef.current = true;
    setState(prev => ({ ...prev, loading: true, error: null }));

    try {
      const result = await fetchFunction(state.currentPage, configRef.current.initialPageSize);
      
      setState(prev => ({
        ...prev,
        items: [...prev.items, ...result.data],
        currentPage: prev.currentPage + 1,
        totalItems: result.total,
        hasMore: result.hasMore,
        loading: false,
      }));
    } catch (error) {
      setState(prev => ({
        ...prev,
        error: error instanceof Error ? error.message : 'Failed to load data',
        loading: false,
      }));
    } finally {
      loadingRef.current = false;
    }
  }, [state.currentPage, state.hasMore, fetchFunction]);

  const reset = useCallback(() => {
    setState({
      items: [],
      loading: false,
      error: null,
      hasMore: true,
      currentPage: 1,
      totalItems: 0,
    });
    loadingRef.current = false;
  }, []);

  const setItems = useCallback((items: T[]) => {
    setState(prev => ({
      ...prev,
      items,
      totalItems: items.length,
      hasMore: false,
    }));
  }, []);

  const appendItems = useCallback((items: T[]) => {
    setState(prev => ({
      ...prev,
      items: [...prev.items, ...items],
      totalItems: prev.totalItems + items.length,
    }));
  }, []);

  return {
    ...state,
    loadMore,
    reset,
    setItems,
    appendItems,
  };
}

// Hook for intersection observer-based lazy loading
export function useIntersectionObserver(
  callback: () => void,
  options: IntersectionObserverInit = {}
) {
  const targetRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const target = targetRef.current;
    if (!target) return;

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting) {
          callback();
        }
      },
      {
        threshold: 0.1,
        rootMargin: '100px',
        ...options,
      }
    );

    observer.observe(target);

    return () => {
      observer.unobserve(target);
    };
  }, [callback, options]);

  return targetRef;
}

// Hook for debounced search with lazy loading
export function useDebouncedLazySearch<T>(
  searchFunction: (query: string, page: number, pageSize: number) => Promise<{
    data: T[];
    total: number;
    hasMore: boolean;
  }>,
  debounceMs: number = 300,
  pageSize: number = 50
) {
  const [query, setQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [searchResults, setSearchResults] = useState<T[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasMore, setHasMore] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);

  const debounceTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const searchAbortControllerRef = useRef<AbortController | null>(null);

  // Debounce the search query
  useEffect(() => {
    if (debounceTimeoutRef.current) {
      clearTimeout(debounceTimeoutRef.current);
    }

    debounceTimeoutRef.current = setTimeout(() => {
      setDebouncedQuery(query);
      setCurrentPage(1);
      setSearchResults([]);
    }, debounceMs);

    return () => {
      if (debounceTimeoutRef.current) {
        clearTimeout(debounceTimeoutRef.current);
      }
    };
  }, [query, debounceMs]);

  // Perform search when debounced query changes
  useEffect(() => {
    if (!debouncedQuery.trim()) {
      setSearchResults([]);
      setHasMore(false);
      return;
    }

    performSearch(debouncedQuery, 1, true);
  }, [debouncedQuery]);

  const performSearch = useCallback(async (
    searchQuery: string,
    page: number,
    isNewSearch: boolean = false
  ) => {
    // Cancel previous search
    if (searchAbortControllerRef.current) {
      searchAbortControllerRef.current.abort();
    }

    searchAbortControllerRef.current = new AbortController();
    setLoading(true);
    setError(null);

    try {
      const result = await searchFunction(searchQuery, page, pageSize);
      
      if (searchAbortControllerRef.current.signal.aborted) {
        return;
      }

      setSearchResults(prev => 
        isNewSearch ? result.data : [...prev, ...result.data]
      );
      setHasMore(result.hasMore);
      setCurrentPage(page);
    } catch (err) {
      if (searchAbortControllerRef.current.signal.aborted) {
        return;
      }
      
      setError(err instanceof Error ? err.message : 'Search failed');
    } finally {
      setLoading(false);
    }
  }, [searchFunction, pageSize]);

  const loadMore = useCallback(() => {
    if (!loading && hasMore && debouncedQuery.trim()) {
      performSearch(debouncedQuery, currentPage + 1, false);
    }
  }, [loading, hasMore, debouncedQuery, currentPage, performSearch]);

  const clearSearch = useCallback(() => {
    setQuery('');
    setDebouncedQuery('');
    setSearchResults([]);
    setHasMore(false);
    setCurrentPage(1);
    setError(null);
    
    if (searchAbortControllerRef.current) {
      searchAbortControllerRef.current.abort();
    }
  }, []);

  return {
    query,
    setQuery,
    searchResults,
    loading,
    error,
    hasMore,
    loadMore,
    clearSearch,
  };
}

// Hook for optimized data fetching with caching
export function useOptimizedDataFetching<T>(
  key: string,
  fetchFunction: () => Promise<T>,
  options: {
    cacheTime?: number; // milliseconds
    staleTime?: number; // milliseconds
    refetchOnWindowFocus?: boolean;
  } = {}
) {
  const {
    cacheTime = 5 * 60 * 1000, // 5 minutes
    staleTime = 1 * 60 * 1000,  // 1 minute
    refetchOnWindowFocus = true,
  } = options;

  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [lastFetched, setLastFetched] = useState<number>(0);

  const cacheRef = useRef<Map<string, { data: T; timestamp: number }>>(new Map());

  const isStale = useCallback(() => {
    return Date.now() - lastFetched > staleTime;
  }, [lastFetched, staleTime]);

  const fetchData = useCallback(async (force: boolean = false) => {
    // Check cache first
    const cached = cacheRef.current.get(key);
    if (cached && !force && Date.now() - cached.timestamp < cacheTime) {
      setData(cached.data);
      setLastFetched(cached.timestamp);
      return;
    }

    if (loading) return;

    setLoading(true);
    setError(null);

    try {
      const result = await fetchFunction();
      const timestamp = Date.now();
      
      setData(result);
      setLastFetched(timestamp);
      
      // Update cache
      cacheRef.current.set(key, { data: result, timestamp });
      
      // Clean up old cache entries
      for (const [cacheKey, cacheValue] of cacheRef.current.entries()) {
        if (Date.now() - cacheValue.timestamp > cacheTime) {
          cacheRef.current.delete(cacheKey);
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch data');
    } finally {
      setLoading(false);
    }
  }, [key, fetchFunction, cacheTime, loading]);

  // Initial fetch
  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Refetch on window focus if data is stale
  useEffect(() => {
    if (!refetchOnWindowFocus) return;

    const handleFocus = () => {
      if (isStale()) {
        fetchData();
      }
    };

    window.addEventListener('focus', handleFocus);
    return () => window.removeEventListener('focus', handleFocus);
  }, [refetchOnWindowFocus, isStale, fetchData]);

  const refetch = useCallback(() => {
    return fetchData(true);
  }, [fetchData]);

  const invalidateCache = useCallback(() => {
    cacheRef.current.delete(key);
  }, [key]);

  return {
    data,
    loading,
    error,
    refetch,
    invalidateCache,
    isStale: isStale(),
  };
}