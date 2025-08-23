export class AppError extends Error {
  public code: string;
  public statusCode?: number;

  constructor(message: string, code: string = 'UNKNOWN_ERROR', statusCode?: number) {
    super(message);
    this.name = 'AppError';
    this.code = code;
    this.statusCode = statusCode;
  }
}

export function handleAPIError(error: any): AppError {
  if (error instanceof AppError) {
    return error;
  }

  if (error.response) {
    // HTTP error response
    const { status, data } = error.response;
    const message = data?.error || data?.message || `HTTP Error ${status}`;
    const code = data?.code || `HTTP_${status}`;
    return new AppError(message, code, status);
  }

  if (error.request) {
    // Network error
    return new AppError(
      'Network error. Please check your connection.',
      'NETWORK_ERROR'
    );
  }

  if (error instanceof Error) {
    return new AppError(error.message, 'GENERIC_ERROR');
  }

  return new AppError('An unknown error occurred', 'UNKNOWN_ERROR');
}

export function getErrorMessage(error: any): string {
  if (error instanceof AppError) {
    return error.message;
  }

  if (error?.message) {
    return error.message;
  }

  if (typeof error === 'string') {
    return error;
  }

  return 'An unexpected error occurred';
}

export function logError(error: any, context?: string) {
  const errorInfo = {
    message: getErrorMessage(error),
    code: error?.code,
    statusCode: error?.statusCode,
    context,
    timestamp: new Date().toISOString(),
    stack: error?.stack,
  };

  console.error('Application Error:', errorInfo);

  // In production, you might want to send this to an error reporting service
  // Example: Sentry.captureException(error, { extra: errorInfo });
}

export function isNetworkError(error: any): boolean {
  return error?.code === 'NETWORK_ERROR' || 
         error?.message?.includes('Network Error') ||
         error?.message?.includes('fetch');
}

export function isAuthError(error: any): boolean {
  return error?.statusCode === 401 || 
         error?.statusCode === 403 ||
         error?.code === 'UNAUTHORIZED' ||
         error?.code === 'FORBIDDEN';
}