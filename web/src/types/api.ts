// API response types
export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  meta?: {
    page?: number;
    per_page?: number;
    total?: number;
    total_pages?: number;
  };
}

// Common API error type
export interface APIError {
  code: string;
  message: string;
  details?: string;
}