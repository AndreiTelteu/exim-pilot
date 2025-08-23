import React, { useState } from 'react';
import DeliverabilityReport from './DeliverabilityReport';
import FailureAnalysis from './FailureAnalysis';
import VolumeReport from './VolumeReport';

type ReportType = 'deliverability' | 'volume' | 'failures';

export default function Reports() {
  const [activeReport, setActiveReport] = useState<ReportType>('deliverability');

  const reportTabs = [
    { id: 'deliverability', label: 'Deliverability', icon: '📊' },
    { id: 'volume', label: 'Volume Analysis', icon: '📈' },
    { id: 'failures', label: 'Failure Analysis', icon: '⚠️' },
  ];

  const renderActiveReport = () => {
    switch (activeReport) {
      case 'deliverability':
        return <DeliverabilityReport />;
      case 'volume':
        return <VolumeReport />;
      case 'failures':
        return <FailureAnalysis />;
      default:
        return <DeliverabilityReport />;
    }
  };

  return (
    <div className="container mx-auto px-4 py-6">
      {/* Tab Navigation */}
      <div className="bg-white shadow-sm rounded-lg mb-6">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8 px-6" aria-label="Tabs">
            {reportTabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveReport(tab.id as ReportType)}
                className={`py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${
                  activeReport === tab.id
                    ? 'border-blue-500 text-blue-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                }`}
              >
                <span className="mr-2">{tab.icon}</span>
                {tab.label}
              </button>
            ))}
          </nav>
        </div>
      </div>

      {/* Report Content */}
      {renderActiveReport()}
    </div>
  );
}