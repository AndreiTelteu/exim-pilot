import React, { useState, useEffect } from 'react';
import ReactECharts from 'echarts-for-react';
import { EChartsOption } from 'echarts';
import { reportsService } from '@/services/reports';
import { VolumeReport as VolumeReportType } from '@/types/reports';
import { LoadingSpinner } from '@/components/Common';
import TimeRangeSelector from './TimeRangeSelector';

export default function VolumeReport() {
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [volumeData, setVolumeData] = useState<VolumeReportType | null>(null);
  const [currentStartTime, setCurrentStartTime] = useState<string>('');
  const [currentEndTime, setCurrentEndTime] = useState<string>('');
  const [groupBy, setGroupBy] = useState<string>('day');

  const fetchVolumeData = async (startTime: string, endTime: string, groupBy: string) => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await reportsService.getVolumeReport(startTime, endTime, groupBy);

      if (response.success && response.data) {
        setVolumeData(response.data);
      } else {
        setError(response.error || 'Failed to fetch volume data');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch volume data');
    } finally {
      setLoading(false);
    }
  };

  const handleTimeRangeChange = (startTime: string, endTime: string) => {
    setCurrentStartTime(startTime);
    setCurrentEndTime(endTime);
    fetchVolumeData(startTime, endTime, groupBy);
  };

  const handleGroupByChange = (newGroupBy: string) => {
    setGroupBy(newGroupBy);
    if (currentStartTime && currentEndTime) {
      fetchVolumeData(currentStartTime, currentEndTime, newGroupBy);
    }
  };

  // Volume trend line chart
  const getVolumeTrendChart = (): EChartsOption => {
    if (!volumeData?.time_series) return {};

    const dates = volumeData.time_series.map(point => {
      const date = new Date(point.timestamp);
      return groupBy === 'hour' 
        ? date.toLocaleString('en-US', { month: 'short', day: 'numeric', hour: '2-digit' })
        : groupBy === 'day'
        ? date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
        : date.toLocaleDateString('en-US', { month: 'short', year: 'numeric' });
    });

    const counts = volumeData.time_series.map(point => point.count);

    return {
      title: {
        text: 'Email Volume Trend',
        left: 'center',
      },
      tooltip: {
        trigger: 'axis',
        formatter: (params: any) => {
          const data = params[0];
          return `${data.name}<br/>Messages: ${formatNumber(data.value)}`;
        },
      },
      xAxis: {
        type: 'category',
        data: dates,
        axisLabel: {
          rotate: groupBy === 'hour' ? 45 : 0,
          interval: 'auto',
        },
      },
      yAxis: {
        type: 'value',
        axisLabel: {
          formatter: (value: number) => formatNumber(value),
        },
      },
      series: [
        {
          name: 'Messages',
          type: 'line',
          data: counts,
          smooth: true,
          itemStyle: {
            color: '#3b82f6',
          },
          areaStyle: {
            color: {
              type: 'linear',
              x: 0,
              y: 0,
              x2: 0,
              y2: 1,
              colorStops: [
                { offset: 0, color: 'rgba(59, 130, 246, 0.3)' },
                { offset: 1, color: 'rgba(59, 130, 246, 0.1)' },
              ],
            },
          },
        },
      ],
      grid: {
        left: '3%',
        right: '4%',
        bottom: '3%',
        containLabel: true,
      },
    };
  };

  // Volume distribution bar chart
  const getVolumeDistributionChart = (): EChartsOption => {
    if (!volumeData?.time_series) return {};

    const data = volumeData.time_series.map(point => ({
      timestamp: point.timestamp,
      count: point.count,
    }));

    // Group by hour of day for distribution analysis
    const hourlyDistribution = new Array(24).fill(0);
    data.forEach(point => {
      const hour = new Date(point.timestamp).getHours();
      hourlyDistribution[hour] += point.count;
    });

    return {
      title: {
        text: 'Volume Distribution by Hour',
        left: 'center',
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow',
        },
        formatter: (params: any) => {
          const data = params[0];
          return `Hour ${data.name}:00<br/>Messages: ${formatNumber(data.value)}`;
        },
      },
      xAxis: {
        type: 'category',
        data: Array.from({ length: 24 }, (_, i) => i.toString().padStart(2, '0')),
        axisLabel: {
          formatter: (value: string) => `${value}:00`,
        },
      },
      yAxis: {
        type: 'value',
        axisLabel: {
          formatter: (value: number) => formatNumber(value),
        },
      },
      series: [
        {
          name: 'Messages',
          type: 'bar',
          data: hourlyDistribution,
          itemStyle: {
            color: '#10b981',
          },
        },
      ],
    };
  };

  const formatNumber = (num: number): string => {
    return new Intl.NumberFormat().format(num);
  };

  const formatBytes = (bytes: number): string => {
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    if (bytes === 0) return '0 Bytes';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
  };

  const exportToCSV = () => {
    if (!volumeData) return;

    const csvData = [
      ['Volume Report'],
      ['Period', `${volumeData.period.start} to ${volumeData.period.end}`],
      ['Group By', volumeData.group_by],
      ['Total Volume', volumeData.total_volume.toString()],
      ['Average Volume', volumeData.average_volume.toString()],
      ['Peak Volume', volumeData.peak_volume.toString()],
      [''],
      ['Time Series Data'],
      ['Timestamp', 'Count'],
      ...volumeData.time_series.map(point => [
        point.timestamp,
        point.count.toString()
      ])
    ];

    const csvContent = csvData.map(row => row.join(',')).join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', `volume-report-${new Date().toISOString().split('T')[0]}.csv`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-md p-4">
        <div className="flex">
          <div className="ml-3">
            <h3 className="text-sm font-medium text-red-800">Error loading volume report</h3>
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
          <h1 className="text-2xl font-bold text-gray-900">Volume Analysis</h1>
          {volumeData && (
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
        
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          <TimeRangeSelector onTimeRangeChange={handleTimeRangeChange} />
          
          <div className="bg-white p-4 rounded-lg shadow-sm border">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Group By
            </label>
            <select
              value={groupBy}
              onChange={(e) => handleGroupByChange(e.target.value)}
              className="block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
            >
              <option value="hour">Hour</option>
              <option value="day">Day</option>
              <option value="week">Week</option>
              <option value="month">Month</option>
            </select>
          </div>
        </div>
      </div>
      
      {loading ? (
        <div className="flex justify-center items-center h-64">
          <LoadingSpinner />
        </div>
      ) : (
        <>
          
          {/* Summary Metrics */}
          {volumeData && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="bg-white p-6 rounded-lg shadow-sm border">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-blue-100 rounded-md flex items-center justify-center">
                      <svg className="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-500">Total Volume</p>
                    <p className="text-2xl font-semibold text-blue-600">{formatNumber(volumeData.total_volume)}</p>
                  </div>
                </div>
              </div>

              <div className="bg-white p-6 rounded-lg shadow-sm border">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-green-100 rounded-md flex items-center justify-center">
                      <svg className="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-500">Average Volume</p>
                    <p className="text-2xl font-semibold text-green-600">{formatNumber(Math.round(volumeData.average_volume))}</p>
                    <p className="text-xs text-gray-500">per {groupBy}</p>
                  </div>
                </div>
              </div>

              <div className="bg-white p-6 rounded-lg shadow-sm border">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-orange-100 rounded-md flex items-center justify-center">
                      <svg className="w-5 h-5 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-500">Peak Volume</p>
                    <p className="text-2xl font-semibold text-orange-600">{formatNumber(volumeData.peak_volume)}</p>
                    <p className="text-xs text-gray-500">highest {groupBy}</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Charts */}
          <div className="space-y-6">
            {/* Volume Trend Chart */}
            <div className="bg-white p-6 rounded-lg shadow-sm border">
              <ReactECharts option={getVolumeTrendChart()} style={{ height: '400px' }} />
            </div>

            {/* Volume Distribution Chart */}
            <div className="bg-white p-6 rounded-lg shadow-sm border">
              <ReactECharts option={getVolumeDistributionChart()} style={{ height: '400px' }} />
            </div>
          </div>

          {/* Data Table */}
          {volumeData?.time_series && (
            <div className="bg-white rounded-lg shadow-sm border">
              <div className="px-6 py-4 border-b border-gray-200">
                <h3 className="text-lg font-medium text-gray-900">Volume Data</h3>
              </div>
              <div className="overflow-x-auto max-h-96">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50 sticky top-0">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        {groupBy === 'hour' ? 'Date & Hour' : 
                        groupBy === 'day' ? 'Date' :
                        groupBy === 'week' ? 'Week' : 'Month'}
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Message Count
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Percentage of Total
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {volumeData.time_series.map((point, index) => {
                      const date = new Date(point.timestamp);
                      const percentage = (point.count / volumeData.total_volume) * 100;
                      
                      return (
                        <tr key={index} className={index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {groupBy === 'hour' 
                              ? date.toLocaleString('en-US', { 
                                  month: 'short', 
                                  day: 'numeric', 
                                  hour: '2-digit',
                                  minute: '2-digit'
                                })
                              : groupBy === 'day'
                              ? date.toLocaleDateString('en-US', { 
                                  weekday: 'short',
                                  month: 'short', 
                                  day: 'numeric',
                                  year: 'numeric'
                                })
                              : date.toLocaleDateString('en-US', { 
                                  month: 'short', 
                                  year: 'numeric' 
                                })
                            }
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {formatNumber(point.count)}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            <div className="flex items-center">
                              <div className="flex-1">
                                <div className="flex items-center">
                                  <div className="w-16 bg-gray-200 rounded-full h-2 mr-2">
                                    <div 
                                      className="bg-blue-600 h-2 rounded-full" 
                                      style={{ width: `${Math.min(percentage, 100)}%` }}
                                    ></div>
                                  </div>
                                  <span className="text-sm text-gray-600">
                                    {percentage.toFixed(1)}%
                                  </span>
                                </div>
                              </div>
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          )}
          
        </>
      )}

    </div>
  );
}