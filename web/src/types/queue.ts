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

// Queue metrics
export interface QueueMetrics {
  total_messages: number;
  deferred_count: number;
  frozen_count: number;
  oldest_message_age: number;
}