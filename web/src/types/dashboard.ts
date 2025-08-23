// Dashboard-related types
export interface DashboardMetrics {
  queue: QueueMetrics;
  delivery: DeliveryMetrics;
  system: SystemMetrics;
}

export interface QueueMetrics {
  total: number;
  deferred: number;
  frozen: number;
  oldest_message_age: number; // in seconds
  recent_growth: number; // percentage change
}

export interface DeliveryMetrics {
  delivered_today: number;
  failed_today: number;
  pending_today: number;
  success_rate: number; // percentage
}

export interface SystemMetrics {
  uptime: number; // in seconds
  log_entries_today: number;
  last_updated: string; // ISO timestamp
}

export interface WeeklyOverviewData {
  dates: string[];
  delivered: number[];
  failed: number[];
  pending: number[];
  deferred: number[];
}

export interface MetricsCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  color?: 'blue' | 'green' | 'yellow' | 'red' | 'purple' | 'gray';
  trend?: {
    value: number;
    direction: 'up' | 'down' | 'stable';
  };
  loading?: boolean;
}