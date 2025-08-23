import React, { useState, useEffect } from "react";
import {
  MessageDetails as MessageDetailsType,
  QueueOperation,
} from "@/types/queue";
import { APIResponse } from "@/types/api";
import { apiService } from "@/services/api";
import { LoadingSpinner } from "@/components/Common";

interface MessageDetailsProps {
  messageId?: string;
  message?: MessageDetailsType;
  onClose: () => void;
  onOperationComplete?: () => void;
}

interface ConfirmationDialogProps {
  isOpen: boolean;
  operation: QueueOperation;
  messageId: string;
  onConfirm: () => void;
  onCancel: () => void;
}

function ConfirmationDialog({
  isOpen,
  operation,
  messageId,
  onConfirm,
  onCancel,
}: ConfirmationDialogProps) {
  if (!isOpen) return null;

  const getOperationText = () => {
    switch (operation) {
      case "deliver":
        return {
          title: "Force Delivery",
          message:
            "This will attempt to deliver the message immediately, bypassing normal retry scheduling.",
          confirmText: "Deliver Now",
          confirmClass: "bg-blue-600 hover:bg-blue-700",
        };
      case "freeze":
        return {
          title: "Freeze Message",
          message:
            "This will prevent the message from being delivered until it is thawed.",
          confirmText: "Freeze",
          confirmClass: "bg-yellow-600 hover:bg-yellow-700",
        };
      case "thaw":
        return {
          title: "Thaw Message",
          message: "This will resume normal delivery attempts for the message.",
          confirmText: "Thaw",
          confirmClass: "bg-green-600 hover:bg-green-700",
        };
      case "delete":
        return {
          title: "Delete Message",
          message:
            "This will permanently remove the message from the queue. This action cannot be undone.",
          confirmText: "Delete",
          confirmClass: "bg-red-600 hover:bg-red-700",
        };
      default:
        return {
          title: "Confirm Action",
          message: "Are you sure you want to perform this action?",
          confirmText: "Confirm",
          confirmClass: "bg-gray-600 hover:bg-gray-700",
        };
    }
  };

  const { title, message, confirmText, confirmClass } = getOperationText();

  return (
    <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
      <div className="relative top-20 mx-auto p-5 border w-96 shadow-lg rounded-md bg-white">
        <div className="mt-3 text-center">
          <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-yellow-100">
            <svg
              className="h-6 w-6 text-yellow-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
              />
            </svg>
          </div>
          <h3 className="text-lg leading-6 font-medium text-gray-900 mt-4">
            {title}
          </h3>
          <div className="mt-2 px-7 py-3">
            <p className="text-sm text-gray-500">{message}</p>
            <p className="text-xs text-gray-400 mt-2 font-mono">
              Message ID: {messageId}
            </p>
          </div>
          <div className="items-center px-4 py-3">
            <button
              onClick={onConfirm}
              className={`px-4 py-2 ${confirmClass} text-white text-base font-medium rounded-md w-full shadow-sm focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 mr-2`}
            >
              {confirmText}
            </button>
            <button
              onClick={onCancel}
              className="mt-3 px-4 py-2 bg-gray-300 text-gray-800 text-base font-medium rounded-md w-full shadow-sm hover:bg-gray-400 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-500"
            >
              Cancel
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default function MessageDetails({
  messageId,
  message: initialMessage,
  onClose,
  onOperationComplete,
}: MessageDetailsProps) {
  const [message, setMessage] = useState<MessageDetailsType | null>(
    initialMessage || null
  );
  const [loading, setLoading] = useState(!initialMessage);
  const [error, setError] = useState<string | null>(null);
  const [operationLoading, setOperationLoading] =
    useState<QueueOperation | null>(null);
  const [confirmDialog, setConfirmDialog] = useState<{
    isOpen: boolean;
    operation: QueueOperation;
  }>({ isOpen: false, operation: "deliver" });

  const currentMessageId = messageId || initialMessage?.id;

  useEffect(() => {
    if (!initialMessage && currentMessageId) {
      fetchMessageDetails();
    }
  }, [currentMessageId, initialMessage]);

  const fetchMessageDetails = async () => {
    if (!currentMessageId) return;

    try {
      setLoading(true);
      setError(null);

      const response: APIResponse<MessageDetailsType> = await apiService.get(
        `/v1/queue/${currentMessageId}`
      );

      if (response.success && response.data) {
        setMessage(response.data);
      } else {
        setError(response.error || "Failed to fetch message details");
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to fetch message details"
      );
    } finally {
      setLoading(false);
    }
  };

  const handleOperation = async (operation: QueueOperation) => {
    try {
      setOperationLoading(operation);
      setError(null);

      let response: APIResponse<any>;

      switch (operation) {
        case "deliver":
          response = await apiService.post(
            `/v1/queue/${currentMessageId}/deliver`
          );
          break;
        case "freeze":
          response = await apiService.post(
            `/v1/queue/${currentMessageId}/freeze`
          );
          break;
        case "thaw":
          response = await apiService.post(
            `/v1/queue/${currentMessageId}/thaw`
          );
          break;
        case "delete":
          response = await apiService.delete(`/v1/queue/${currentMessageId}`);
          break;
        default:
          throw new Error("Unknown operation");
      }

      if (response.success) {
        // Refresh message details after operation
        if (operation !== "delete") {
          await fetchMessageDetails();
        }
        onOperationComplete?.();

        // Close the modal if message was deleted
        if (operation === "delete") {
          onClose();
        }
      } else {
        setError(response.error || `Failed to ${operation} message`);
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : `Failed to ${operation} message`
      );
    } finally {
      setOperationLoading(null);
      setConfirmDialog({ isOpen: false, operation: "deliver" });
    }
  };

  const openConfirmDialog = (operation: QueueOperation) => {
    setConfirmDialog({ isOpen: true, operation });
  };

  const closeConfirmDialog = () => {
    setConfirmDialog({ isOpen: false, operation: "deliver" });
  };

  const formatTimestamp = (timestamp: string): string => {
    return new Date(timestamp).toLocaleString();
  };

  const formatSize = (bytes: number): string => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + " " + sizes[i];
  };

  const getStatusBadgeClass = (status: string): string => {
    switch (status) {
      case "queued":
        return "bg-blue-100 text-blue-800";
      case "deferred":
        return "bg-yellow-100 text-yellow-800";
      case "frozen":
        return "bg-red-100 text-red-800";
      default:
        return "bg-gray-100 text-gray-800";
    }
  };

  if (loading) {
    return (
      <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-40">
        <div className="relative top-20 mx-auto p-5 border w-4/5 max-w-4xl shadow-lg rounded-md bg-white">
          <div className="flex justify-center items-center py-12">
            <LoadingSpinner size="lg" />
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-40">
        <div className="relative top-20 mx-auto p-5 border w-4/5 max-w-4xl shadow-lg rounded-md bg-white">
          <div className="bg-red-50 border border-red-200 rounded-md p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg
                  className="h-5 w-5 text-red-400"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">
                  Error loading message details
                </h3>
                <div className="mt-2 text-sm text-red-700">
                  <p>{error}</p>
                </div>
                <div className="mt-4 flex space-x-2">
                  <button
                    onClick={fetchMessageDetails}
                    className="bg-red-100 px-3 py-2 rounded-md text-sm font-medium text-red-800 hover:bg-red-200"
                  >
                    Try again
                  </button>
                  <button
                    onClick={onClose}
                    className="bg-gray-100 px-3 py-2 rounded-md text-sm font-medium text-gray-800 hover:bg-gray-200"
                  >
                    Close
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (!message) {
    return null;
  }

  return (
    <>
      <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-40">
        <div className="relative top-10 mx-auto p-5 border w-4/5 max-w-6xl shadow-lg rounded-md bg-white mb-10">
          {/* Header */}
          <div className="flex justify-between items-start mb-6">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">
                Message Details
              </h2>
              <p className="text-sm text-gray-600 font-mono mt-1">
                {message.id}
              </p>
            </div>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600"
            >
              <svg
                className="w-6 h-6"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          {/* Status and Actions */}
          <div className="flex justify-between items-center mb-6 p-4 bg-gray-50 rounded-lg">
            <div className="flex items-center space-x-4">
              <span
                className={`inline-flex rounded-full px-3 py-1 text-sm font-semibold ${getStatusBadgeClass(
                  message.status
                )}`}
              >
                {message.status}
              </span>
              <span className="text-sm text-gray-600">
                Retry Count: {message.retry_count}
              </span>
              {message.last_attempt && (
                <span className="text-sm text-gray-600">
                  Last Attempt: {formatTimestamp(message.last_attempt)}
                </span>
              )}
            </div>

            <div className="flex space-x-2">
              <button
                onClick={() => openConfirmDialog("deliver")}
                disabled={operationLoading !== null}
                className="px-3 py-2 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {operationLoading === "deliver" ? (
                  <LoadingSpinner size="sm" />
                ) : (
                  "Deliver Now"
                )}
              </button>

              {message.status === "frozen" ? (
                <button
                  onClick={() => openConfirmDialog("thaw")}
                  disabled={operationLoading !== null}
                  className="px-3 py-2 bg-green-600 text-white text-sm font-medium rounded-md hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {operationLoading === "thaw" ? (
                    <LoadingSpinner size="sm" />
                  ) : (
                    "Thaw"
                  )}
                </button>
              ) : (
                <button
                  onClick={() => openConfirmDialog("freeze")}
                  disabled={operationLoading !== null}
                  className="px-3 py-2 bg-yellow-600 text-white text-sm font-medium rounded-md hover:bg-yellow-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {operationLoading === "freeze" ? (
                    <LoadingSpinner size="sm" />
                  ) : (
                    "Freeze"
                  )}
                </button>
              )}

              <button
                onClick={() => openConfirmDialog("delete")}
                disabled={operationLoading !== null}
                className="px-3 py-2 bg-red-600 text-white text-sm font-medium rounded-md hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {operationLoading === "delete" ? (
                  <LoadingSpinner size="sm" />
                ) : (
                  "Delete"
                )}
              </button>
            </div>
          </div>

          {/* Content Tabs */}
          <div className="border-b border-gray-200">
            <nav className="-mb-px flex space-x-8">
              <button className="border-blue-500 text-blue-600 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm">
                Envelope
              </button>
            </nav>
          </div>

          <div className="mt-6 space-y-6">
            {/* Envelope Information */}
            <div className="bg-white border border-gray-200 rounded-lg p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">
                Envelope Information
              </h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Sender
                  </label>
                  <p className="mt-1 text-sm text-gray-900 font-mono break-all">
                    {message.envelope.sender}
                  </p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Received At
                  </label>
                  <p className="mt-1 text-sm text-gray-900">
                    {formatTimestamp(message.envelope.received_at)}
                  </p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Size
                  </label>
                  <p className="mt-1 text-sm text-gray-900">
                    {formatSize(message.envelope.size)}
                  </p>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Recipients ({message.envelope.recipients.length})
                  </label>
                  <div className="mt-1 max-h-32 overflow-y-auto">
                    {message.envelope.recipients.map((recipient, index) => (
                      <p
                        key={index}
                        className="text-sm text-gray-900 font-mono break-all"
                      >
                        {recipient}
                      </p>
                    ))}
                  </div>
                </div>
              </div>
            </div>
            {/* Headers */}
            <div className="bg-white border border-gray-200 rounded-lg p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">
                Message Headers
              </h3>
              <div className="bg-gray-50 rounded-md p-4 max-h-64 overflow-y-auto">
                <pre className="text-xs font-mono text-gray-800 whitespace-pre-wrap">
                  {Object.entries(message.headers).map(([key, value]) => (
                    <div key={key} className="mb-1">
                      <span className="font-semibold text-blue-600">
                        {key}:
                      </span>{" "}
                      {value}
                    </div>
                  ))}
                </pre>
              </div>
            </div>

            {/* SMTP Transaction Logs */}
            {message.smtp_logs && message.smtp_logs.length > 0 && (
              <div className="bg-white border border-gray-200 rounded-lg p-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">
                  SMTP Transaction Logs
                </h3>
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {message.smtp_logs.map((log, index) => (
                    <div
                      key={index}
                      className="border-l-4 border-blue-200 pl-4 py-2 bg-gray-50 rounded-r"
                    >
                      <div className="flex justify-between items-start">
                        <div className="flex-1">
                          <p className="text-sm font-medium text-gray-900">
                            {log.event}
                          </p>
                          <p className="text-sm text-gray-600 mt-1">
                            {log.message}
                          </p>
                          {log.host && (
                            <p className="text-xs text-gray-500 mt-1">
                              Host: {log.host}{" "}
                              {log.ip_address && `(${log.ip_address})`}
                            </p>
                          )}
                        </div>
                        <span className="text-xs text-gray-500 ml-4">
                          {formatTimestamp(log.timestamp)}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Delivery Attempts */}
            {message.delivery_attempts &&
              message.delivery_attempts.length > 0 && (
                <div className="bg-white border border-gray-200 rounded-lg p-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">
                    Delivery Attempts
                  </h3>
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Timestamp
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Recipient
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Host
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Status
                          </th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Response
                          </th>
                        </tr>
                      </thead>
                      <tbody className="bg-white divide-y divide-gray-200">
                        {message.delivery_attempts.map((attempt) => (
                          <tr key={attempt.id}>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                              {formatTimestamp(attempt.timestamp)}
                            </td>
                            <td className="px-6 py-4 text-sm text-gray-900 font-mono break-all">
                              {attempt.recipient}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                              {attempt.host && (
                                <div>
                                  <div>{attempt.host}</div>
                                  {attempt.ip_address && (
                                    <div className="text-xs text-gray-500">
                                      {attempt.ip_address}
                                    </div>
                                  )}
                                </div>
                              )}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <span
                                className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                                  attempt.status === "success"
                                    ? "bg-green-100 text-green-800"
                                    : attempt.status === "defer"
                                    ? "bg-yellow-100 text-yellow-800"
                                    : "bg-red-100 text-red-800"
                                }`}
                              >
                                {attempt.status}
                              </span>
                            </td>
                            <td className="px-6 py-4 text-sm text-gray-900">
                              {attempt.smtp_code && (
                                <div className="font-mono text-xs mb-1">
                                  {attempt.smtp_code}
                                </div>
                              )}
                              {attempt.error_message && (
                                <div className="text-xs text-gray-600">
                                  {attempt.error_message}
                                </div>
                              )}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}

            {/* Message Content Preview */}
            {message.content_preview && (
              <div className="bg-white border border-gray-200 rounded-lg p-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">
                  Message Content Preview
                </h3>
                <div className="bg-gray-50 rounded-md p-4 max-h-64 overflow-y-auto">
                  <pre className="text-sm font-mono text-gray-800 whitespace-pre-wrap">
                    {message.content_preview}
                  </pre>
                </div>
                <p className="text-xs text-gray-500 mt-2">
                  * Content is truncated and sanitized for security. Attachments
                  are not displayed.
                </p>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Confirmation Dialog */}
      <ConfirmationDialog
        isOpen={confirmDialog.isOpen}
        operation={confirmDialog.operation}
        messageId={currentMessageId || ""}
        onConfirm={() => handleOperation(confirmDialog.operation)}
        onCancel={closeConfirmDialog}
      />
    </>
  );
}
