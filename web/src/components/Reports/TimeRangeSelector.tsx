import React, { useState } from 'react';
import { TIME_RANGE_OPTIONS, TimeRangeOption } from '@/types/reports';

interface TimeRangeSelectorProps {
  onTimeRangeChange: (startTime: string, endTime: string) => void;
  className?: string;
}

export default function TimeRangeSelector({ onTimeRangeChange, className = '' }: TimeRangeSelectorProps) {
  const [selectedRange, setSelectedRange] = useState<string>('7d');
  const [customStartDate, setCustomStartDate] = useState<string>('');
  const [customEndDate, setCustomEndDate] = useState<string>('');
  const [showCustomInputs, setShowCustomInputs] = useState<boolean>(false);

  const handleRangeChange = (value: string) => {
    setSelectedRange(value);
    setShowCustomInputs(value === 'custom');

    if (value !== 'custom') {
      const option = TIME_RANGE_OPTIONS.find(opt => opt.value === value);
      if (option) {
        const endTime = new Date();
        const startTime = new Date(endTime.getTime() - (option.days * 24 * 60 * 60 * 1000));
        
        onTimeRangeChange(
          startTime.toISOString(),
          endTime.toISOString()
        );
      }
    }
  };

  const handleCustomDateChange = () => {
    if (customStartDate && customEndDate) {
      const startTime = new Date(customStartDate);
      const endTime = new Date(customEndDate);
      
      // Set end time to end of day
      endTime.setHours(23, 59, 59, 999);
      
      onTimeRangeChange(
        startTime.toISOString(),
        endTime.toISOString()
      );
    }
  };

  // Initialize with default range
  React.useEffect(() => {
    handleRangeChange('7d');
  }, []);

  return (
    <div className={`bg-white p-4 rounded-lg shadow-sm border ${className}`}>
      <div className="flex flex-col space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Time Range
          </label>
          <select
            value={selectedRange}
            onChange={(e) => handleRangeChange(e.target.value)}
            className="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
          >
            {TIME_RANGE_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>

        {showCustomInputs && (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Start Date
              </label>
              <input
                type="date"
                value={customStartDate}
                onChange={(e) => setCustomStartDate(e.target.value)}
                className="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                End Date
              </label>
              <input
                type="date"
                value={customEndDate}
                onChange={(e) => setCustomEndDate(e.target.value)}
                className="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              />
            </div>
            {customStartDate && customEndDate && (
              <div className="md:col-span-2">
                <button
                  onClick={handleCustomDateChange}
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
                >
                  Apply Custom Range
                </button>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}