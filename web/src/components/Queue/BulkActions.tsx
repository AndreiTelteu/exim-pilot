import React, { useState } from 'react';
import { QueueOperation, BulkOperationResult, BulkOperationProgress } from '@/types/queue';

interface BulkActionsProps {
  selectedMessages: string[];
  onClearSelection: () => void;
  onBulkOperation: (operation: QueueOperation, messageIds: string[]) => Promise<any>;
  onOperationComplete?: () => void;
}

interface ConfirmationDialogProps {
  isOpen: boolean;
  operation: QueueOperation;
  messageCount: number;
  onConfirm: () => void;
  onCancel: () => void;
}

interface ProgressIndicatorProps {
  progress: BulkOperationProgress;
}

interface ResultFeedbackProps {
  result: BulkOperationResult | null;
  onClose: () => void;
}

const ConfirmationDialog: React.FC<ConfirmationDialogProps> = ({
  isOpen,
  operation,
  messageCount,
  onConfirm,
  onCancel,
}) => {
  if (!isOpen) return null;

  const getOperationText = (op: QueueOperation) => {
    switch (op) {
      case 'deliver':
        return 'force delivery of';
      case 'freeze':
        return 'freeze';
      case 'thaw':
        return 'thaw';
      case 'delete':
        return 'permanently delete';
      default:
        return 'perform operation on';
    }
  };

  const getOperationColor = (op: QueueOperation) => {
    switch (op) {
      case 'delete':
        return 'text-red-600';
      case 'freeze':
        return 'text-yellow-600';
      default:
        return 'text-blue-600';
    }
  };

  const isDestructive = operation === 'delete';

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div className="mt-3 text-center">
          <div className={`mx-auto flex items-center justify-center h-12 w-12 rounded-full ${
            isDestructive ? 'bg-red-100' : 'bg-blue-100'
          }`}>
            {isDestructive ? (
              <svg className="h-6 w-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
              </svg>
            ) : (
              <svg className="h-6 w-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            )}
          </div>
          <h3 className="text-lg leading-6 font-medium text-gray-900 mt-4">
            Confirm Bulk Operation
          </h3>
          <div className="mt-2 px-7 py-3">
            <p className="text-sm text-gray-500">
              Are you sure you want to{' '}
              <span className={`font-medium ${getOperationColor(operation)}`}>
                {getOperationText(operation)}
              </span>{' '}
              <span className="font-medium">{messageCount}</span> message{messageCount !== 1 ? 's' : ''}?
            </p>
            {isDestructive && (
              <p className="text-sm text-red-600 mt-2 font-medium">
                This action cannot be undone.
              </p>
            )}
          </div>
          <div className="items-center px-4 py-3">
            <div className="flex space-x-3">
              <button
                onClick={onCancel}
                className="px-4 py-2 bg-gray-500 text-white text-base font-medium rounded-md w-full shadow-sm hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-300"
              >
                Cancel
              </button>
              <button
                onClick={onConfirm}
                className={`px-4 py-2 text-white text-base font-medium rounded-md w-full shadow-sm focus:outline-none focus:ring-2 ${
                  isDestructive
                    ? 'bg-red-600 hover:bg-red-700 focus:ring-red-300'
                    : 'bg-blue-600 hover:bg-blue-700 focus:ring-blue-300'
                }`}
              >
                {operation === 'delete' ? 'Delete' : 'Confirm'}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

const ProgressIndicator: React.FC<ProgressIndicatorProps> = ({ progress }) => {
  const percentage = progress.total > 0 ? Math.round((progress.completed / progress.total) * 100) : 0;
  
  return (
    <div className="bg-white border border-gray-200 rounded-md p-4 shadow-sm">
      <div className="flex items-center justify-between mb-2">
        <h4 className="text-sm font-medium text-gray-900">
          {progress.operation.charAt(0).toUpperCase() + progress.operation.slice(1)} Operation in Progress
        </h4>
        <span className="text-sm text-gray-500">
          {progress.completed} of {progress.total}
        </span>
      </div>
      
      <div className="w-full bg-gray-200 rounded-full h-2 mb-2">
        <div
          className="bg-blue-600 h-2 rounded-full transition-all duration-300"
          style={{ width: `${percentage}%` }}
        />
      </div>
      
      <div className="flex justify-between text-xs text-gray-500">
        <span>{percentage}% complete</span>
        {progress.failed > 0 && (
          <span className="text-red-600">{progress.failed} failed</span>
        )}
      </div>
    </div>
  );
};

const ResultFeedback: React.FC<ResultFeedbackProps> = ({ result, onClose }) => {
  if (!result) return null;

  const isSuccess = result.failed === 0;
  const hasPartialFailure = result.successful > 0 && result.failed > 0;

  return (
    <div className={`border rounded-md p-4 ${
      isSuccess ? 'bg-green-50 border-green-200' : 
      hasPartialFailure ? 'bg-yellow-50 border-yellow-200' : 
      'bg-red-50 border-red-200'
    }`}>
      <div className="flex">
        <div className="flex-shrink-0">
          {isSuccess ? (
            <svg className="h-5 w-5 text-green-400" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
            </svg>
          ) : hasPartialFailure ? (
            <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
            </svg>
          ) : (
            <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
              <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
            </svg>
          )}
        </div>
        <div className="ml-3 flex-1">
          <h3 className={`text-sm font-medium ${
            isSuccess ? 'text-green-800' : 
            hasPartialFailure ? 'text-yellow-800' : 
            'text-red-800'
          }`}>
            {isSuccess ? 'Operation Completed Successfully' :
             hasPartialFailure ? 'Operation Partially Completed' :
             'Operation Failed'}
          </h3>
          <div className={`mt-2 text-sm ${
            isSuccess ? 'text-green-700' : 
            hasPartialFailure ? 'text-yellow-700' : 
            'text-red-700'
          }`}>
            <p>
              {result.successful} of {result.total_requested} messages processed successfully.
              {result.failed > 0 && ` ${result.failed} failed.`}
            </p>
            
            {result.errors && result.errors.length > 0 && (
              <div className="mt-3">
                <p className="font-medium">Errors:</p>
                <ul className="mt-1 list-disc list-inside space-y-1">
                  {result.errors.slice(0, 5).map((error, index) => (
                    <li key={index} className="text-xs">
                      <span className="font-mono">{error.message_id}</span>: {error.error}
                    </li>
                  ))}
                  {result.errors.length > 5 && (
                    <li className="text-xs">
                      ... and {result.errors.length - 5} more errors
                    </li>
                  )}
                </ul>
              </div>
            )}
          </div>
        </div>
        <div className="ml-auto pl-3">
          <div className="-mx-1.5 -my-1.5">
            <button
              onClick={onClose}
              className={`inline-flex rounded-md p-1.5 focus:outline-none focus:ring-2 focus:ring-offset-2 ${
                isSuccess ? 'text-green-500 hover:bg-green-100 focus:ring-green-600' :
                hasPartialFailure ? 'text-yellow-500 hover:bg-yellow-100 focus:ring-yellow-600' :
                'text-red-500 hover:bg-red-100 focus:ring-red-600'
              }`}
            >
              <span className="sr-only">Dismiss</span>
              <svg className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default function BulkActions({
  selectedMessages,
  onClearSelection,
  onBulkOperation,
  onOperationComplete,
}: BulkActionsProps) {
  const [showConfirmation, setShowConfirmation] = useState(false);
  const [pendingOperation, setPendingOperation] = useState<QueueOperation | null>(null);
  const [progress, setProgress] = useState<BulkOperationProgress | null>(null);
  const [result, setResult] = useState<BulkOperationResult | null>(null);

  const handleOperationClick = (operation: QueueOperation) => {
    setPendingOperation(operation);
    setShowConfirmation(true);
  };

  const handleConfirmOperation = async () => {
    if (!pendingOperation) return;

    setShowConfirmation(false);
    
    // Initialize progress
    setProgress({
      operation: pendingOperation,
      total: selectedMessages.length,
      completed: 0,
      failed: 0,
      in_progress: true,
    });

    try {
      const response = await onBulkOperation(pendingOperation, selectedMessages);
      
      // Simulate progress updates (in a real implementation, this would come from WebSocket or polling)
      const updateProgress = (completed: number, failed: number = 0) => {
        setProgress(prev => prev ? {
          ...prev,
          completed,
          failed,
          in_progress: completed + failed < selectedMessages.length,
        } : null);
      };

      // Simulate incremental progress
      for (let i = 1; i <= selectedMessages.length; i++) {
        await new Promise(resolve => setTimeout(resolve, 100));
        updateProgress(i);
      }

      // Set final result
      const operationResult: BulkOperationResult = {
        operation: pendingOperation,
        total_requested: selectedMessages.length,
        successful: response.data?.successful || selectedMessages.length,
        failed: response.data?.failed || 0,
        errors: response.data?.errors || [],
      };

      setResult(operationResult);
      setProgress(null);
      
      // Clear selection after successful operation
      onClearSelection();
      
      // Notify parent component
      onOperationComplete?.();
      
    } catch (error) {
      // Handle operation error
      const errorResult: BulkOperationResult = {
        operation: pendingOperation,
        total_requested: selectedMessages.length,
        successful: 0,
        failed: selectedMessages.length,
        errors: [{ message_id: 'all', error: error instanceof Error ? error.message : 'Unknown error' }],
      };
      
      setResult(errorResult);
      setProgress(null);
    }

    setPendingOperation(null);
  };

  const handleCancelOperation = () => {
    setShowConfirmation(false);
    setPendingOperation(null);
  };

  const handleCloseResult = () => {
    setResult(null);
  };

  const getOperationButtonClass = (operation: QueueOperation, disabled: boolean = false) => {
    const baseClass = "inline-flex items-center px-3 py-2 border text-sm leading-4 font-medium rounded-md focus:outline-none focus:ring-2 focus:ring-offset-2";
    
    if (disabled) {
      return `${baseClass} border-gray-300 text-gray-400 bg-gray-100 cursor-not-allowed`;
    }

    switch (operation) {
      case 'delete':
        return `${baseClass} border-red-300 text-red-700 bg-white hover:bg-red-50 focus:ring-red-500`;
      case 'freeze':
        return `${baseClass} border-yellow-300 text-yellow-700 bg-white hover:bg-yellow-50 focus:ring-yellow-500`;
      default:
        return `${baseClass} border-transparent text-blue-700 bg-blue-100 hover:bg-blue-200 focus:ring-blue-500`;
    }
  };

  const isOperationInProgress = progress?.in_progress || false;

  return (
    <div className="space-y-4">
      {/* Progress Indicator */}
      {progress && <ProgressIndicator progress={progress} />}
      
      {/* Result Feedback */}
      {result && <ResultFeedback result={result} onClose={handleCloseResult} />}
      
      {/* Bulk Actions Bar */}
      <div className="bg-blue-50 border border-blue-200 rounded-md p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center">
            <svg className="h-5 w-5 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <span className="ml-2 text-sm font-medium text-blue-800">
              {selectedMessages.length} message{selectedMessages.length !== 1 ? 's' : ''} selected
            </span>
          </div>
          <div className="flex items-center space-x-2">
            <button
              type="button"
              onClick={() => handleOperationClick('deliver')}
              disabled={isOperationInProgress}
              className={getOperationButtonClass('deliver', isOperationInProgress)}
            >
              <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8" />
              </svg>
              Deliver Now
            </button>
            <button
              type="button"
              onClick={() => handleOperationClick('freeze')}
              disabled={isOperationInProgress}
              className={getOperationButtonClass('freeze', isOperationInProgress)}
            >
              <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 2L3 14h7v8l7-12h-7V2z" />
              </svg>
              Freeze
            </button>
            <button
              type="button"
              onClick={() => handleOperationClick('thaw')}
              disabled={isOperationInProgress}
              className={getOperationButtonClass('thaw', isOperationInProgress)}
            >
              <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707" />
              </svg>
              Thaw
            </button>
            <button
              type="button"
              onClick={() => handleOperationClick('delete')}
              disabled={isOperationInProgress}
              className={getOperationButtonClass('delete', isOperationInProgress)}
            >
              <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              Delete
            </button>
            <button
              type="button"
              onClick={onClearSelection}
              disabled={isOperationInProgress}
              className={getOperationButtonClass('thaw', isOperationInProgress)}
            >
              Clear Selection
            </button>
          </div>
        </div>
      </div>

      {/* Confirmation Dialog */}
      <ConfirmationDialog
        isOpen={showConfirmation}
        operation={pendingOperation || 'deliver'}
        messageCount={selectedMessages.length}
        onConfirm={handleConfirmOperation}
        onCancel={handleCancelOperation}
      />
    </div>
  );
}