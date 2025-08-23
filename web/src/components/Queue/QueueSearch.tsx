import React, { useState, useEffect } from 'react';
import { QueueSearchFilters } from '@/types/queue';

interface QueueSearchProps {
  onFiltersChange: (filters: QueueSearchFilters) => void;
  initialFilters?: QueueSearchFilters;
}

export default function QueueSearch({ onFiltersChange, initialFilters = {} }: QueueSearchProps) {
  const [filters, setFilters] = useState<QueueSearchFilters>(initialFilters);
  const [isExpanded, setIsExpanded] = useState(false);

  // Check if any advanced filters are set
  const hasAdvancedFilters = Boolean(
    filters.age_min || 
    filters.age_max || 
    filters.retry_count_min || 
    filters.retry_count_max ||
    filters.status
  );

  useEffect(() => {
    if (hasAdvancedFilters) {
      setIsExpanded(true);
    }
  }, [hasAdvancedFilters]);

  const handleInputChange = (field: keyof QueueSearchFilters, value: string | number | undefined) => {
    const newFilters = { ...filters };
    
    if (value === '' || value === undefined) {
      delete newFilters[field];
    } else {
      (newFilters as any)[field] = value;
    }
    
    setFilters(newFilters);
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onFiltersChange(filters);
  };

  const handleClear = () => {
    setFilters({});
    onFiltersChange({});
  };

  const hasAnyFilters = Object.keys(filters).length > 0;

  return (
    <div className="bg-white shadow rounded-lg">
      <div className="px-4 py-5 sm:p-6">
        <div className="sm:flex sm:items-center sm:justify-between">
          <div>
            <h3 className="text-lg font-medium text-gray-900">Search Queue</h3>
            <p className="mt-1 text-sm text-gray-500">
              Filter messages by sender, recipient, status, and other criteria
            </p>
          </div>
          <div className="mt-4 sm:mt-0">
            <button
              type="button"
              onClick={() => setIsExpanded(!isExpanded)}
              className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              {isExpanded ? 'Hide' : 'Show'} Advanced Filters
              <svg
                className={`ml-2 -mr-0.5 h-4 w-4 transform transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
              </svg>
            </button>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="mt-6 space-y-6">
          {/* Basic Search Fields */}
          <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
            <div>
              <label htmlFor="sender" className="block text-sm font-medium text-gray-700">
                Sender
              </label>
              <input
                type="text"
                id="sender"
                value={filters.sender || ''}
                onChange={(e) => handleInputChange('sender', e.target.value)}
                placeholder="sender@example.com"
                className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              />
            </div>

            <div>
              <label htmlFor="recipient" className="block text-sm font-medium text-gray-700">
                Recipient
              </label>
              <input
                type="text"
                id="recipient"
                value={filters.recipient || ''}
                onChange={(e) => handleInputChange('recipient', e.target.value)}
                placeholder="recipient@example.com"
                className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
              />
            </div>

            <div>
              <label htmlFor="message_id" className="block text-sm font-medium text-gray-700">
                Message ID
              </label>
              <input
                type="text"
                id="message_id"
                value={filters.message_id || ''}
                onChange={(e) => handleInputChange('message_id', e.target.value)}
                placeholder="1hKj2L-0001Ac-2B"
                className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm font-mono text-xs"
              />
            </div>
          </div>

          {/* Advanced Filters */}
          {isExpanded && (
            <div className="border-t border-gray-200 pt-6">
              <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
                <div>
                  <label htmlFor="subject" className="block text-sm font-medium text-gray-700">
                    Subject
                  </label>
                  <input
                    type="text"
                    id="subject"
                    value={filters.subject || ''}
                    onChange={(e) => handleInputChange('subject', e.target.value)}
                    placeholder="Email subject"
                    className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  />
                </div>

                <div>
                  <label htmlFor="status" className="block text-sm font-medium text-gray-700">
                    Status
                  </label>
                  <select
                    id="status"
                    value={filters.status || ''}
                    onChange={(e) => handleInputChange('status', e.target.value || undefined)}
                    className="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                  >
                    <option value="">All statuses</option>
                    <option value="queued">Queued</option>
                    <option value="deferred">Deferred</option>
                    <option value="frozen">Frozen</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Age (hours)
                  </label>
                  <div className="mt-1 flex space-x-2">
                    <input
                      type="number"
                      value={filters.age_min || ''}
                      onChange={(e) => handleInputChange('age_min', e.target.value ? Number(e.target.value) : undefined)}
                      placeholder="Min"
                      min="0"
                      className="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                    />
                    <span className="flex items-center text-gray-500">to</span>
                    <input
                      type="number"
                      value={filters.age_max || ''}
                      onChange={(e) => handleInputChange('age_max', e.target.value ? Number(e.target.value) : undefined)}
                      placeholder="Max"
                      min="0"
                      className="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Retry Count
                  </label>
                  <div className="mt-1 flex space-x-2">
                    <input
                      type="number"
                      value={filters.retry_count_min || ''}
                      onChange={(e) => handleInputChange('retry_count_min', e.target.value ? Number(e.target.value) : undefined)}
                      placeholder="Min"
                      min="0"
                      className="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                    />
                    <span className="flex items-center text-gray-500">to</span>
                    <input
                      type="number"
                      value={filters.retry_count_max || ''}
                      onChange={(e) => handleInputChange('retry_count_max', e.target.value ? Number(e.target.value) : undefined)}
                      placeholder="Max"
                      min="0"
                      className="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
                    />
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex items-center justify-between pt-4 border-t border-gray-200">
            <div className="flex items-center space-x-4">
              <button
                type="submit"
                className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                <svg className="-ml-1 mr-2 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
                Search
              </button>

              {hasAnyFilters && (
                <button
                  type="button"
                  onClick={handleClear}
                  className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  <svg className="-ml-1 mr-2 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                  Clear Filters
                </button>
              )}
            </div>

            {hasAnyFilters && (
              <div className="text-sm text-gray-500">
                {Object.keys(filters).length} filter{Object.keys(filters).length !== 1 ? 's' : ''} applied
              </div>
            )}
          </div>
        </form>

        {/* Active Filters Display */}
        {hasAnyFilters && (
          <div className="mt-4 pt-4 border-t border-gray-200">
            <div className="flex flex-wrap gap-2">
              <span className="text-sm font-medium text-gray-700">Active filters:</span>
              {Object.entries(filters).map(([key, value]) => (
                <span
                  key={key}
                  className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
                >
                  {key.replace('_', ' ')}: {String(value)}
                  <button
                    type="button"
                    onClick={() => handleInputChange(key as keyof QueueSearchFilters, undefined)}
                    className="flex-shrink-0 ml-1.5 h-4 w-4 rounded-full inline-flex items-center justify-center text-blue-400 hover:bg-blue-200 hover:text-blue-500 focus:outline-none focus:bg-blue-500 focus:text-white"
                  >
                    <span className="sr-only">Remove filter</span>
                    <svg className="h-2 w-2" stroke="currentColor" fill="none" viewBox="0 0 8 8">
                      <path strokeLinecap="round" strokeWidth="1.5" d="m1 1 6 6m0-6L1 7" />
                    </svg>
                  </button>
                </span>
              ))}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}