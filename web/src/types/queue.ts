// Queue message types
export interface QueueMessage {
  id: string;
  size: number;
  age: string;
  sender: string;
  recipients: string[];
  status: 'queued' | 'deferred' | 'frozen';
  retry_count: number;
  last_attempt: string;
  next_retry: string;
}

// Message details for inspection
export interface MessageDetails {
  id: string;
  envelope: {
    sender: string;
    recipients: string[];
    received_at: string;
    size: number;
  };
  headers: Record<string, string>;
  smtp_logs: SMTPLogEntry[];
  content_preview?: string;
  status: 'queued' | 'deferred' | 'frozen';
  retry_count: number;
  last_attempt?: string;
  next_retry?: string;
  delivery_attempts: DeliveryAttempt[];
}

// SMTP transaction log entry
export interface SMTPLogEntry {
  timestamp: string;
  event: string;
  host?: string;
  ip_address?: string;
  message: string;
}

// Delivery attempt information
export interface DeliveryAttempt {
  id: number;
  timestamp: string;
  recipient: string;
  host?: string;
  ip_address?: string;
  status: 'success' | 'defer' | 'bounce';
  smtp_code?: string;
  error_message?: string;
}

// Queue search filters
export interface QueueSearchFilters {
  sender?: string;
  recipient?: string;
  message_id?: string;
  subject?: string;
  status?: string;
  age_min?: number;
  age_max?: number;
  retry_count_min?: number;
  retry_count_max?: number;
}

// Queue operations
export type QueueOperation = 'deliver' | 'freeze' | 'thaw' | 'delete';

// Bulk operation result
export interface BulkOperationResult {
  operation: QueueOperation;
  total_requested: number;
  successful: number;
  failed: number;
  errors?: Array<{
    message_id: string;
    error: string;
  }>;
}

// Bulk operation progress
export interface BulkOperationProgress {
  operation: QueueOperation;
  total: number;
  completed: number;
  failed: number;
  in_progress: boolean;
}

// Queue metrics
export interface QueueMetrics {
  total_messages: number;
  deferred_count: number;
  frozen_count: number;
  oldest_message_age: number;
}