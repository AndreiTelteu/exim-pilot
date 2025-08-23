import React, { useState, useEffect } from 'react';
import ReactECharts from 'echarts-for-react';
import { EChartsOption } from 'echarts';
import { reportsService } from '@/services/reports';
import { DeliverabilityReport as DeliverabilityReportType, TopSendersReport, TopRecipientsReport, DomainAnalysis } from '@/types/reports';
import { LoadingSpinner } from '@/components/Common';
import TimeRangeSelector from './TimeRangeSelector';

export default function DeliverabilityReport() {
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [deliverabilityData, setDeliverabilityData] = useState<DeliverabilityReportType | null>(null);
  const [topSendersData, setTopSendersData] = useState<TopSendersReport | null>(null);
  const [topRecipientsData, setTopRecipientsData] = useState<TopRecipientsReport | null>(null);
  const [domainAnalysisData, setDomainAnalysisData] = useState<DomainAnalysis | null>(null);
  const [currentStartTime, setCurrentStartTime] = useState<string>('');
  const [currentEndTime, setCurrentEndTime] = useState<string>('');

  const fetchReportData = async (startTime: string, endTime: string) => {
    setLoading(true);
    setError(null);
    
    try {
      const [deliverabilityResponse, topSendersResponse, topRecipientsResponse, domainAnalysisResponse] = await Promise.all([
        reportsService.getDeliverabilityReport(startTime, endTime),
        reportsService.getTopSendersReport(startTime, endTime, 10),
        reportsService.getTopRecipientsReport(startTime, endTime, 10),
        reportsService.getDomainAnalysis(startTime, endTime, 'both', 10),
      ]);

      if (deliverabilityResponse.success && deliverabilityResponse.data) {
        setDeliverabilityData(deliverabilityResponse.data);
      }

      if (topSendersResponse.success && topSendersResponse.data) {
        setTopSendersData(topSendersResponse.data);
      }

      if (topRecipientsResponse.success && topRecipientsResponse.data) {
        setTopRecipientsData(topRecipientsResponse.data);
      }

      if (domainAnalysisResponse.success && domainAnalysisResponse.data) {
        setDomainAnalysisData(domainAnalysisResponse.data);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch report data');
    } finally {
      setLoading(false);
    }
  };

  const handleTimeRangeChange = (startTime: string, endTime: string) => {
    setCurrentStartTime(startTime);
    setCurrentEndTime(endTime);
    fetchReportData(startTime, endTime);
  };

  // Deliverability rates chart
  const getDeliverabilityRatesChart = (): EChartsOption => {
    if (!deliverabilityData) return {};

    return {
      title: {
        text: 'Deliverability Rates',
        left: 'center',
      },
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b}: {c}% ({d}%)',
      },
      legend: {
        orient: 'vertical',
        left: 'left',
      },
      series: [
        {
          name: 'Deliverability',
          type: 'pie',
          radius: '50%',
          data: [
            { value: deliverabilityData.delivery_rate, name: 'Delivered', itemStyle: { color: '#10b981' } },
            { value: deliverabilityData.deferral_rate, name: 'Deferred', itemStyle: { color: '#f59e0b' } },
            { value: deliverabilityData.bounce_rate, name: 'Bounced', itemStyle: { color: '#ef4444' } },
            { value: deliverabilityData.rejection_rate, name: 'Rejected', itemStyle: { color: '#6b7280' } },
          ],
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: 'rgba(0, 0, 0, 0.5)',
            },
          },
        },
      ],
    };
  };

  // Top failure reasons chart
  const getFailureReasonsChart = (): EChartsOption => {
    if (!deliverabilityData?.top_failure_reasons) return {};

    return {
      title: {
        text: 'Top Failure Reasons',
        left: 'center',
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow',
        },
      },
      xAxis: {
        type: 'category',
        data: deliverabilityData.top_failure_reasons.map(reason => reason.reason),
        axisLabel: {
          rotate: 45,
          interval: 0,
        },
      },
      yAxis: {
        type: 'value',
      },
      series: [
        {
          name: 'Count',
          type: 'bar',
          data: deliverabilityData.top_failure_reasons.map(reason => reason.count),
          itemStyle: {
            color: '#ef4444',
          },
        },
      ],
    };
  };

  // Domain analysis chart
  const getDomainAnalysisChart = (): EChartsOption => {
    if (!domainAnalysisData?.recipient_domains) return {};

    return {
      title: {
        text: 'Top Recipient Domains - Deliverability',
        left: 'center',
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow',
        },
      },
      legend: {
        data: ['Delivery Rate', 'Bounce Rate', 'Defer Rate'],
        top: 30,
      },
      xAxis: {
        type: 'category',
        data: domainAnalysisData.recipient_domains.slice(0, 10).map(domain => domain.domain),
        axisLabel: {
          rotate: 45,
          interval: 0,
        },
      },
      yAxis: {
        type: 'value',
        max: 100,
        axisLabel: {
          formatter: '{value}%',
        },
      },
      series: [
        {
          name: 'Delivery Rate',
          type: 'bar',
          data: domainAnalysisData.recipient_domains.slice(0, 10).map(domain => domain.delivery_rate),
          itemStyle: { color: '#10b981' },
        },
        {
          name: 'Bounce Rate',
          type: 'bar',
          data: domainAnalysisData.recipient_domains.slice(0, 10).map(domain => domain.bounce_rate),
          itemStyle: { color: '#ef4444' },
        },
        {
          name: 'Defer Rate',
          type: 'bar',
          data: domainAnalysisData.recipient_domains.slice(0, 10).map(domain => domain.defer_rate),
          itemStyle: { color: '#f59e0b' },
        },
      ],
    };
  };

  const formatNumber = (num: number): string => {
    return new Intl.NumberFormat().format(num);
  };

  const formatBytes = (bytes: number): string => {
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    if (bytes === 0) return '0 Bytes';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
  };

  const exportToCSV = () => {
    if (!deliverabilityData) return;

    const csvData = [
      ['Deliverability Report'],
      ['Period', `${deliverabilityData.period.start} to ${deliverabilityData.period.end}`],
      ['Total Messages', deliverabilityData.total_messages.toString()],
      ['Delivery Rate', `${deliverabilityData.delivery_rate.toFixed(2)}%`],
      ['Bounce Rate', `${deliverabilityData.bounce_rate.toFixed(2)}%`],
      ['Deferral Rate', `${deliverabilityData.deferral_rate.toFixed(2)}%`],
      ['Rejection Rate', `${deliverabilityData.rejection_rate.toFixed(2)}%`],
      [''],
      ['Top Failure Reasons'],
      ['Reason', 'Count'],
      ...deliverabilityData.top_failure_reasons.map(reason => [
        reason.reason,
        reason.count.toString()
      ]),
      [''],
      ['Top Senders'],
      ['Sender', 'Messages', 'Volume (Bytes)', 'Delivery Rate'],
      ...(topSendersData?.top_senders.map(sender => [
        sender.sender,
        sender.message_count.toString(),
        sender.volume_bytes.toString(),
        `${sender.delivery_rate.toFixed(2)}%`
      ]) || []),
      [''],
      ['Top Recipients'],
      ['Recipient', 'Messages', 'Volume (Bytes)', 'Delivery Rate'],
      ...(topRecipientsData?.top_recipients.map(recipient => [
        recipient.recipient,
        recipient.message_count.toString(),
        recipient.volume_bytes.toString(),
        `${recipient.delivery_rate.toFixed(2)}%`
      ]) || [])
    ];

    const csvContent = csvData.map(row => row.join(',')).join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', `deliverability-report-${new Date().toISOString().split('T')[0]}.csv`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <LoadingSpinner />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4">
        <div className="flex">
          <div className="ml-3">
            <h3 className="text-sm font-medium text-red-800">Error loading deliverability report</h3>
            <div className="mt-2 text-sm text-red-700">
              <p>{error}</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-white shadow-sm rounded-lg p-6">
        <div className="flex justify-between items-start mb-4">
          <h1 className="text-2xl font-bold text-gray-900">Deliverability Report</h1>
          {deliverabilityData && (
            <button
              onClick={exportToCSV}
              className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Export CSV
            </button>
          )}
        </div>
        <TimeRangeSelector onTimeRangeChange={handleTimeRangeChange} />
      </div>

      {/* Summary Metrics */}
      {deliverabilityData && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="w-8 h-8 bg-blue-100 rounded-md flex items-center justify-center">
                  <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 4.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                  </svg>
                </div>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-500">Total Messages</p>
                <p className="text-2xl font-semibold text-gray-900">{formatNumber(deliverabilityData.total_messages)}</p>
              </div>
            </div>
          </div>

          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="w-8 h-8 bg-green-100 rounded-md flex items-center justify-center">
                  <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                  </svg>
                </div>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-500">Delivery Rate</p>
                <p className="text-2xl font-semibold text-green-600">{deliverabilityData.delivery_rate.toFixed(1)}%</p>
              </div>
            </div>
          </div>

          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="w-8 h-8 bg-red-100 rounded-md flex items-center justify-center">
                  <svg className="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </div>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-500">Bounce Rate</p>
                <p className="text-2xl font-semibold text-red-600">{deliverabilityData.bounce_rate.toFixed(1)}%</p>
              </div>
            </div>
          </div>

          <div className="bg-white p-6 rounded-lg shadow-sm border">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <div className="w-8 h-8 bg-yellow-100 rounded-md flex items-center justify-center">
                  <svg className="w-5 h-5 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-500">Deferral Rate</p>
                <p className="text-2xl font-semibold text-yellow-600">{deliverabilityData.deferral_rate.toFixed(1)}%</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Deliverability Rates Pie Chart */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <ReactECharts option={getDeliverabilityRatesChart()} style={{ height: '400px' }} />
        </div>

        {/* Top Failure Reasons */}
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <ReactECharts option={getFailureReasonsChart()} style={{ height: '400px' }} />
        </div>
      </div>

      {/* Domain Analysis Chart */}
      {domainAnalysisData?.recipient_domains && (
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <ReactECharts option={getDomainAnalysisChart()} style={{ height: '400px' }} />
        </div>
      )}

      {/* Top Senders and Recipients Tables */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top Senders */}
        {topSendersData && (
          <div className="bg-white rounded-lg shadow-sm border">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Top Senders</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Sender
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Messages
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Volume
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Delivery Rate
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {topSendersData.top_senders.map((sender, index) => (
                    <tr key={index}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {sender.sender}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {formatNumber(sender.message_count)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {formatBytes(sender.volume_bytes)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                          sender.delivery_rate >= 95 
                            ? 'bg-green-100 text-green-800'
                            : sender.delivery_rate >= 90
                            ? 'bg-yellow-100 text-yellow-800'
                            : 'bg-red-100 text-red-800'
                        }`}>
                          {sender.delivery_rate.toFixed(1)}%
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* Top Recipients */}
        {topRecipientsData && (
          <div className="bg-white rounded-lg shadow-sm border">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Top Recipients</h3>
            </div>
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Recipient
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Messages
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Volume
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Delivery Rate
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {topRecipientsData.top_recipients.map((recipient, index) => (
                    <tr key={index}>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {recipient.recipient}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {formatNumber(recipient.message_count)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        {formatBytes(recipient.volume_bytes)}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                        <span className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                          recipient.delivery_rate >= 95 
                            ? 'bg-green-100 text-green-800'
                            : recipient.delivery_rate >= 90
                            ? 'bg-yellow-100 text-yellow-800'
                            : 'bg-red-100 text-red-800'
                        }`}>
                          {recipient.delivery_rate.toFixed(1)}%
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}