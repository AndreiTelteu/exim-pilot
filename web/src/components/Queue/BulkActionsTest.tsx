import React, { useState } from "react";
import { QueueMessage, QueueOperation } from "@/types/queue";
import BulkActions from "./BulkActions";

// Mock queue messages for testing
const mockMessages: QueueMessage[] = [
  {
    id: "1ABCD-123456-EF",
    size: 2048,
    age: "2h 30m",
    sender: "user@example.com",
    recipients: ["recipient1@domain.com", "recipient2@domain.com"],
    status: "queued",
    retry_count: 0,
    last_attempt: "2024-01-15T10:30:00Z",
    next_retry: "2024-01-15T11:00:00Z",
  },
  {
    id: "2BCDE-234567-FG",
    size: 4096,
    age: "1h 15m",
    sender: "sender@company.com",
    recipients: ["user@example.org"],
    status: "deferred",
    retry_count: 2,
    last_attempt: "2024-01-15T09:45:00Z",
    next_retry: "2024-01-15T12:00:00Z",
  },
  {
    id: "3CDEF-345678-GH",
    size: 1024,
    age: "45m",
    sender: "noreply@service.com",
    recipients: ["admin@domain.net", "support@domain.net"],
    status: "frozen",
    retry_count: 5,
    last_attempt: "2024-01-15T08:30:00Z",
    next_retry: "2024-01-15T16:00:00Z",
  },
];

export default function BulkActionsTest() {
  const [selectedMessages, setSelectedMessages] = useState<string[]>([]);
  const [operationResults, setOperationResults] = useState<string[]>([]);

  const handleSelectionChange = (messageId: string, checked: boolean) => {
    if (checked) {
      setSelectedMessages((prev) => [...prev, messageId]);
    } else {
      setSelectedMessages((prev) => prev.filter((id) => id !== messageId));
    }
  };

  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedMessages(mockMessages.map((m) => m.id));
    } else {
      setSelectedMessages([]);
    }
  };

  const handleClearSelection = () => {
    setSelectedMessages([]);
  };

  const handleBulkOperation = async (
    operation: QueueOperation,
    messageIds: string[]
  ) => {
    // Mock API call with simulated delay
    await new Promise((resolve) => setTimeout(resolve, 1000));

    // Simulate some failures for testing
    const successful = Math.floor(messageIds.length * 0.8);
    const failed = messageIds.length - successful;

    const result = `${operation} operation: ${successful} successful, ${failed} failed`;
    setOperationResults((prev) => [...prev, result]);

    // Mock API response
    return {
      success: true,
      data: {
        successful,
        failed,
        errors:
          failed > 0
            ? [
                {
                  message_id: messageIds[0],
                  error: "Temporary failure - will retry",
                },
              ]
            : [],
      },
    };
  };

  const handleOperationComplete = () => {
    console.log("Operation completed, refreshing queue...");
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

  return (
    <div className="max-w-6xl mx-auto p-6 space-y-6">
      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h1 className="text-2xl font-bold text-gray-900 mb-6">
            Bulk Actions Test
          </h1>

          {/* Mock Queue Table */}
          <div className="mb-6">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Mock Queue Messages
            </h2>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-300">
                <thead>
                  <tr>
                    <th scope="col" className="relative px-7 sm:w-12 sm:px-6">
                      <input
                        type="checkbox"
                        className="absolute left-4 top-1/2 -mt-2 h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-600"
                        checked={
                          selectedMessages.length === mockMessages.length &&
                          mockMessages.length > 0
                        }
                        onChange={(e) => handleSelectAll(e.target.checked)}
                      />
                    </th>
                    <th className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide">
                      Message ID
                    </th>
                    <th className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide">
                      Sender
                    </th>
                    <th className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide">
                      Recipients
                    </th>
                    <th className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide">
                      Size
                    </th>
                    <th className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide">
                      Age
                    </th>
                    <th className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide">
                      Status
                    </th>
                    <th className="px-3 py-3.5 text-left text-xs font-medium text-gray-500 uppercase tracking-wide">
                      Retries
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  {mockMessages.map((message) => (
                    <tr
                      key={message.id}
                      className={`hover:bg-gray-50 ${
                        selectedMessages.includes(message.id)
                          ? "bg-blue-50"
                          : ""
                      }`}
                    >
                      <td className="relative px-7 sm:w-12 sm:px-6">
                        <input
                          type="checkbox"
                          className="absolute left-4 top-1/2 -mt-2 h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-600"
                          checked={selectedMessages.includes(message.id)}
                          onChange={(e) =>
                            handleSelectionChange(message.id, e.target.checked)
                          }
                        />
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        <div className="font-mono text-xs">{message.id}</div>
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        <div
                          className="max-w-xs truncate"
                          title={message.sender}
                        >
                          {message.sender}
                        </div>
                      </td>
                      <td className="px-3 py-4 text-sm text-gray-900">
                        <div className="max-w-xs">
                          {message.recipients.length === 1 ? (
                            <div
                              className="truncate"
                              title={message.recipients[0]}
                            >
                              {message.recipients[0]}
                            </div>
                          ) : (
                            <div>
                              <div
                                className="truncate"
                                title={message.recipients[0]}
                              >
                                {message.recipients[0]}
                              </div>
                              {message.recipients.length > 1 && (
                                <div className="text-xs text-gray-500">
                                  +{message.recipients.length - 1} more
                                </div>
                              )}
                            </div>
                          )}
                        </div>
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        {formatSize(message.size)}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        {message.age}
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        <span
                          className={`inline-flex rounded-full px-2 text-xs font-semibold leading-5 ${getStatusBadgeClass(
                            message.status
                          )}`}
                        >
                          {message.status}
                        </span>
                      </td>
                      <td className="whitespace-nowrap px-3 py-4 text-sm text-gray-900">
                        {message.retry_count}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Bulk Actions Component */}
          {selectedMessages.length > 0 && (
            <BulkActions
              selectedMessages={selectedMessages}
              onClearSelection={handleClearSelection}
              onBulkOperation={handleBulkOperation}
              onOperationComplete={handleOperationComplete}
            />
          )}

          {/* Operation Results */}
          {operationResults.length > 0 && (
            <div className="mt-6">
              <h3 className="text-lg font-medium text-gray-900 mb-3">
                Operation Results
              </h3>
              <div className="bg-gray-50 rounded-md p-4">
                <ul className="space-y-2">
                  {operationResults.map((result, index) => (
                    <li key={index} className="text-sm text-gray-700">
                      {result}
                    </li>
                  ))}
                </ul>
                <button
                  onClick={() => setOperationResults([])}
                  className="mt-3 text-sm text-blue-600 hover:text-blue-800"
                >
                  Clear Results
                </button>
              </div>
            </div>
          )}

          {/* Instructions */}
          <div className="mt-8 bg-blue-50 border border-blue-200 rounded-md p-4">
            <h3 className="text-sm font-medium text-blue-800 mb-2">
              Test Instructions
            </h3>
            <ul className="text-sm text-blue-700 space-y-1">
              <li>1. Select one or more messages using the checkboxes</li>
              <li>
                2. The bulk actions bar will appear with operation buttons
              </li>
              <li>
                3. Click any operation button to see the confirmation dialog
              </li>
              <li>4. Confirm the operation to see the progress indicator</li>
              <li>5. View the operation results and feedback</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
}
