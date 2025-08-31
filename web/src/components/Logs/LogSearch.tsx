import { useState, useEffect } from 'react';
import { LogSearchParams } from '../../types/logs';

interface LogSearchProps {
  onSearch: (params: LogSearchParams) => void;
  onExport: (format: 'csv' | 'txt' | 'json') => void;
  totalEntries: number;
}

const LogSearch: React.FC<LogSearchProps> = ({ onSearch, onExport, totalEntries }) => {
  const [searchForm, setSearchForm] = useState<LogSearchParams>({
    log_type: '',
    event: '',
    message_id: '',
    sender: '',
    recipient: '',
    host: '',
    keyword: '',
    start_date: '',
    end_date: '',
    per_page: 50,
  });

  const [isAdvanced, setIsAdvanced] = useState(false);
  const [showExportOptions, setShowExportOptions] = useState(false);

  // Set default date range to last 24 hours
  useEffect(() => {
    const now = new Date();
    const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000);
    
    setSearchForm(prev => ({
      ...prev,
      start_date: yesterday.toISOString().slice(0, 16),
      end_date: now.toISOString().slice(0, 16),
    }));
  }, []);

  const handleInputChange = (name: string, value: string | number) => {
    setSearchForm(prev => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Filter out empty values
    const filteredParams = Object.fromEntries(
      Object.entries(searchForm).filter(([_, value]) => value !== '' && value !== null && value !== undefined)
    );
    
    onSearch(filteredParams);
  };

  const handleReset = () => {
    const now = new Date();
    const yesterday = new Date(now.getTime() - 24 * 60 * 60 * 1000);
    
    setSearchForm({
      log_type: '',
      event: '',
      message_id: '',
      sender: '',
      recipient: '',
      host: '',
      keyword: '',
      start_date: yesterday.toISOString().slice(0, 16),
      end_date: now.toISOString().slice(0, 16),
      per_page: 50,
    });
    
    onSearch({
      per_page: 50,
      start_date: yesterday.toISOString().slice(0, 16),
      end_date: now.toISOString().slice(0, 16),
    });
  };

  const handleQuickFilter = (filterType: string, value: string) => {
    const newParams = { ...searchForm };
    if (filterType === 'log_type') {
      newParams.log_type = value;
    } else if (filterType === 'timeRange') {
      const now = new Date();
      let startDate: Date;
      
      switch (value) {
        case '1h':
          startDate = new Date(now.getTime() - 60 * 60 * 1000);
          break;
        case '6h':
          startDate = new Date(now.getTime() - 6 * 60 * 60 * 1000);
          break;
        case '24h':
          startDate = new Date(now.getTime() - 24 * 60 * 60 * 1000);
          break;
        case '7d':
          startDate = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
          break;
        default:
          startDate = new Date(now.getTime() - 24 * 60 * 60 * 1000);
      }
      
      newParams.start_date = startDate.toISOString().slice(0, 16);
      newParams.end_date = now.toISOString().slice(0, 16);
    }
    
    setSearchForm(newParams);
    onSearch(Object.fromEntries(
      Object.entries(newParams).filter(([_, v]) => v !== '' && v !== null && v !== undefined)
    ));
  };

  return (
    <div className="bg-white shadow rounded-lg">
      <div className="px-4 py-5 sm:p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-medium text-gray-900">Log Search</h3>
          <div className="flex items-center space-x-2">
            <span className="text-sm text-gray-500">
              {totalEntries.toLocaleString()} total entries
            </span>
            <div className="relative">
              <button
                onClick={() => setShowExportOptions(!showExportOptions)}
                className="inline-flex items-center px-3 py-1 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
              >
                <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
                Export
              </button>
              
              {showExportOptions && (
                <div className="absolute right-0 mt-2 w-36 bg-white border border-gray-200 rounded-md shadow-lg z-10">
                  <div className="py-1">
                    <button
                      onClick={() => { onExport('csv'); setShowExportOptions(false); }}
                      className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                    >
                      Export as CSV
                    </button>
                    <button
                      onClick={() => { onExport('txt'); setShowExportOptions(false); }}
                      className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                    >
                      Export as TXT
                    </button>
                    <button
                      onClick={() => { onExport('json'); setShowExportOptions(false); }}
                      className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                    >
                      Export as JSON
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Quick Filters */}
        <div className="mb-4">
          <div className="flex flex-wrap gap-2">
            <div className="flex items-center space-x-1">
              <span className="text-xs text-gray-500">Log Type:</span>
              <button
                onClick={() => handleQuickFilter('log_type', '')}
                className={`px-2 py-1 text-xs rounded ${
                  searchForm.log_type === '' ? 'bg-indigo-100 text-indigo-800' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                All
              </button>
              <button
                onClick={() => handleQuickFilter('log_type', 'main')}
                className={`px-2 py-1 text-xs rounded ${
                  searchForm.log_type === 'main' ? 'bg-blue-100 text-blue-800' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                Main
              </button>
              <button
                onClick={() => handleQuickFilter('log_type', 'reject')}
                className={`px-2 py-1 text-xs rounded ${
                  searchForm.log_type === 'reject' ? 'bg-red-100 text-red-800' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                Reject
              </button>
              <button
                onClick={() => handleQuickFilter('log_type', 'panic')}
                className={`px-2 py-1 text-xs rounded ${
                  searchForm.log_type === 'panic' ? 'bg-yellow-100 text-yellow-800' : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                Panic
              </button>
            </div>

            <div className="flex items-center space-x-1">
              <span className="text-xs text-gray-500">Time:</span>
              <button
                onClick={() => handleQuickFilter('timeRange', '1h')}
                className="px-2 py-1 text-xs rounded bg-gray-100 text-gray-700 hover:bg-gray-200"
              >
                1h
              </button>
              <button
                onClick={() => handleQuickFilter('timeRange', '6h')}
                className="px-2 py-1 text-xs rounded bg-gray-100 text-gray-700 hover:bg-gray-200"
              >
                6h
              </button>
              <button
                onClick={() => handleQuickFilter('timeRange', '24h')}
                className="px-2 py-1 text-xs rounded bg-gray-100 text-gray-700 hover:bg-gray-200"
              >
                24h
              </button>
              <button
                onClick={() => handleQuickFilter('timeRange', '7d')}
                className="px-2 py-1 text-xs rounded bg-gray-100 text-gray-700 hover:bg-gray-200"
              >
                7d
              </button>
            </div>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Basic Search */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label htmlFor="keyword" className="block text-sm font-medium text-gray-700 mb-1">
                Keyword Search
              </label>
              <input
                type="text"
                id="keyword"
                value={searchForm.keyword || ''}
                onChange={(e) => handleInputChange('keyword', e.target.value)}
                placeholder="Search in log entries..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              />
            </div>

            <div>
              <label htmlFor="message_id" className="block text-sm font-medium text-gray-700 mb-1">
                Message ID
              </label>
              <input
                type="text"
                id="message_id"
                value={searchForm.message_id || ''}
                onChange={(e) => handleInputChange('message_id', e.target.value)}
                placeholder="e.g., 1A2B3C-4D5E6F-GH"
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 font-mono text-sm"
              />
            </div>

            <div>
              <label htmlFor="per_page" className="block text-sm font-medium text-gray-700 mb-1">
                Results per page
              </label>
              <select
                id="per_page"
                value={searchForm.per_page || 50}
                onChange={(e) => handleInputChange('per_page', parseInt(e.target.value))}
                className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              >
                <option value={25}>25</option>
                <option value={50}>50</option>
                <option value={100}>100</option>
                <option value={200}>200</option>
              </select>
            </div>
          </div>

          {/* Advanced Search Toggle */}
          <div>
            <button
              type="button"
              onClick={() => setIsAdvanced(!isAdvanced)}
              className="inline-flex items-center text-sm text-indigo-600 hover:text-indigo-500"
            >
              <svg
                className={`w-4 h-4 mr-1 transition-transform ${isAdvanced ? 'rotate-90' : ''}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
              </svg>
              Advanced Filters
            </button>
          </div>

          {/* Advanced Search Fields */}
          {isAdvanced && (
            <div className="space-y-4 pt-4 border-t border-gray-200">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label htmlFor="sender" className="block text-sm font-medium text-gray-700 mb-1">
                    Sender Email
                  </label>
                  <input
                    type="email"
                    id="sender"
                    value={searchForm.sender || ''}
                    onChange={(e) => handleInputChange('sender', e.target.value)}
                    placeholder="user@example.com"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                  />
                </div>

                <div>
                  <label htmlFor="recipient" className="block text-sm font-medium text-gray-700 mb-1">
                    Recipient Email
                  </label>
                  <input
                    type="email"
                    id="recipient"
                    value={searchForm.recipient || ''}
                    onChange={(e) => handleInputChange('recipient', e.target.value)}
                    placeholder="recipient@example.com"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                  />
                </div>

                <div>
                  <label htmlFor="host" className="block text-sm font-medium text-gray-700 mb-1">
                    Host
                  </label>
                  <input
                    type="text"
                    id="host"
                    value={searchForm.host || ''}
                    onChange={(e) => handleInputChange('host', e.target.value)}
                    placeholder="mail.example.com"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                  />
                </div>

                <div>
                  <label htmlFor="event" className="block text-sm font-medium text-gray-700 mb-1">
                    Event Type
                  </label>
                  <input
                    type="text"
                    id="event"
                    value={searchForm.event || ''}
                    onChange={(e) => handleInputChange('event', e.target.value)}
                    placeholder="delivery, defer, bounce..."
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label htmlFor="start_date" className="block text-sm font-medium text-gray-700 mb-1">
                    Start Date & Time
                  </label>
                  <input
                    type="datetime-local"
                    id="start_date"
                    value={searchForm.start_date || ''}
                    onChange={(e) => handleInputChange('start_date', e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                  />
                </div>

                <div>
                  <label htmlFor="end_date" className="block text-sm font-medium text-gray-700 mb-1">
                    End Date & Time
                  </label>
                  <input
                    type="datetime-local"
                    id="end_date"
                    value={searchForm.end_date || ''}
                    onChange={(e) => handleInputChange('end_date', e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                  />
                </div>
              </div>
            </div>
          )}

          {/* Submit Buttons */}
          <div className="flex items-center space-x-3">
            <button
              type="submit"
              className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
              Search
            </button>
            
            <button
              type="button"
              onClick={handleReset}
              className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
            >
              <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
              </svg>
              Reset
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default LogSearch;