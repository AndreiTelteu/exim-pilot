// Log entry types
export interface LogEntry {
  id: number;
  timestamp: string;
  message_id?: string;
  log_type: 'main' | 'reject' | 'panic';
  event: string;
  host?: string;
  sender?: string;
  recipients?: string[];
  size?: number;
  status?: string;
  error_code?: string;
  error_text?: string;
  raw_line: string;
}

// Log search filters
export interface LogSearchFilters {
  log_type?: string;
  event?: string;
  message_id?: string;
  sender?: string;
  recipient?: string;
  keyword?: string;
  start_date?: string;
  end_date?: string;
}

// Real-time log message
export interface RealTimeLogMessage {
  type: 'log_entry';
  data: LogEntry;
}

// Enhanced log search parameters
export interface LogSearchParams {
  log_type?: string;
  event?: string;
  message_id?: string;
  sender?: string;
  recipient?: string;
  host?: string;
  start_date?: string;
  end_date?: string;
  keyword?: string;
  page?: number;
  per_page?: number;
}

// Log statistics interface
export interface LogStatistics {
  total_entries: number;
  main_log_count: number;
  reject_log_count: number;
  panic_log_count: number;
  recent_entries: number;
  top_events: Array<{
    event: string;
    count: number;
  }>;
  error_trends: Array<{
    date: string;
    count: number;
  }>;
}

// Log export options
export interface LogExportOptions {
  format: 'csv' | 'txt' | 'json';
  filters?: LogSearchParams;
  include_raw?: boolean;
  date_range?: {
    start: string;
    end: string;
  };
}

// Message history for tracing
export interface MessageHistory {
  message_id: string;
  entries: LogEntry[];
  timeline: Array<{
    timestamp: string;
    event: string;
    status: string;
    details?: string;
  }>;
}

// Log search response structure (matches backend)
export interface LogSearchResponse {
  entries: LogEntry[] | null;
  search_time: string;
  aggregations: any;
}