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