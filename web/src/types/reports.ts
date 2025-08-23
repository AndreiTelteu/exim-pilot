// Report types based on the backend API structures

export interface Period {
  start: string;
  end: string;
}

export interface FailureReason {
  reason: string;
  count: number;
}

export interface DeliverabilityReport {
  period: Period;
  total_messages: number;
  delivered_count: number;
  deferred_count: number;
  bounced_count: number;
  rejected_count: number;
  delivery_rate: number;
  deferral_rate: number;
  bounce_rate: number;
  rejection_rate: number;
  event_counts: Record<string, number>;
  log_type_counts: Record<string, number>;
  top_failure_reasons: FailureReason[];
}

export interface TimeSeriesPoint {
  timestamp: string;
  count: number;
}

export interface VolumeReport {
  period: Period;
  group_by: string;
  total_volume: number;
  average_volume: number;
  peak_volume: number;
  time_series: TimeSeriesPoint[];
}

export interface FailureCategory {
  category: string;
  count: number;
  percentage: number;
  description: string;
}

export interface ErrorCodeStat {
  code: string;
  description: string;
  count: number;
}

export interface FailureReport {
  period: Period;
  total_failures: number;
  failure_categories: FailureCategory[];
  top_error_codes: ErrorCodeStat[];
}

export interface SenderStat {
  sender: string;
  message_count: number;
  volume_bytes: number;
  delivery_rate: number;
}

export interface TopSendersReport {
  period: Period;
  top_senders: SenderStat[];
}

export interface RecipientStat {
  recipient: string;
  message_count: number;
  volume_bytes: number;
  delivery_rate: number;
}

export interface TopRecipientsReport {
  period: Period;
  top_recipients: RecipientStat[];
}

export interface DomainStat {
  domain: string;
  message_count: number;
  delivery_rate: number;
  bounce_rate: number;
  defer_rate: number;
}

export interface DomainAnalysis {
  period: Period;
  analysis_type: string;
  sender_domains?: DomainStat[];
  recipient_domains?: DomainStat[];
}

// Time range options for reports
export interface TimeRangeOption {
  label: string;
  value: string;
  days: number;
}

export const TIME_RANGE_OPTIONS: TimeRangeOption[] = [
  { label: 'Last 24 hours', value: '1d', days: 1 },
  { label: 'Last 3 days', value: '3d', days: 3 },
  { label: 'Last 7 days', value: '7d', days: 7 },
  { label: 'Last 30 days', value: '30d', days: 30 },
  { label: 'Last 90 days', value: '90d', days: 90 },
  { label: 'Custom', value: 'custom', days: 0 },
];