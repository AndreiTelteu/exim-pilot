import React, { useState, useEffect } from 'react';
import ReactECharts from 'echarts-for-react';
import { EChartsOption } from 'echarts';
import { reportsService } from '@/services/reports';
import { FailureReport } from '@/types/reports';
import { LoadingSpinner } from '@/components/Common';
import TimeRangeSelector from './TimeRangeSelector';

export default function FailureAnalysis() {
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [failureData, setFailureData] = useState<FailureReport | null>(null);
  const [currentStartTime, setCurrentStartTime] = useState<string>('');
  const [currentEndTime, setCurrentEndTime] = useState<string>('');

  const fetchFailureData = async (startTime: string, endTime: string) => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await reportsService.getFailureReport(startTime, endTime, 50);

      if (response.success && response.data) {
        setFailureData(response.data);
      } else {
        setError(response.error || 'Failed to fetch failure data');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch failure data');
    } finally {
      setLoading(false);
    }
  };

  const handleTimeRangeChange = (startTime: string, endTime: string) => {
    setCurrentStartTime(startTime);
    setCurrentEndTime(endTime);
    fetchFailureData(startTime, endTime);
  };

  // Failure categories pie chart
  const getFailureCategoriesChart = (): EChartsOption => {
    if (!failureData?.failure_categories) return {};

    return {
      title: {
        text: 'Failure Categories',
        left: 'center',
      },
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b}: {c} ({d}%)',
      },
      legend: {
        orient: 'vertical',
        left: 'left',
      },
      series: [
        {
          name: 'Failures',
          type: 'pie',
          radius: '50%',
          data: failureData.failure_categories.map(category => ({
            value: category.count,
            name: category.category,
            itemStyle: {
              color: category.category.toLowerCase().includes('bounce') ? '#ef4444' :
                     category.category.toLowerCase().includes('defer') ? '#f59e0b' :
                     category.category.toLowerCase().includes('reject') ? '#6b7280' :
                     category.category.toLowerCase().includes('timeout') ? '#8b5cf6' :
                     '#06b6d4'
            }
          })),
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

  // Top error codes bar chart
  const getErrorCodesChart = (): EChartsOption => {
    if (!failureData?.top_error_codes) return {};

    return {
      title: {
        text: 'Top Error Codes',
        left: 'center',
      },
      tooltip: {
        trigger: 'axis',
        axisPointer: {
          type: 'shadow',
        },
        formatter: (params: any) => {
          const data = params[0];
          const errorCode = failureData.top_error_codes[data.dataIndex];
          return `${data.name}<br/>Count: ${data.value}<br/>Description: ${errorCode.description}`;
        },
      },
      xAxis: {
        type: 'category',
        data: failureData.top_error_codes.map(code => code.code),
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
          data: failureData.top_error_codes.map(code => code.count),
          itemStyle: {
            color: '#ef4444',
          },
        },
      ],
    };
  };

  const formatNumber = (num: number): string => {
    return new Intl.NumberFormat().format(num);
  };

  const exportToCSV = () => {
    if (!failureData) return;

    const csvData = [
      ['Failure Analysis Report'],
      ['Period', `${failureData.period.start} to ${failureData.period.end}`],
      ['Total Failures', failureData.total_failures.toString()],
      [''],
      ['Failure Categories'],
      ['Category', 'Count', 'Percentage', 'Description'],
      ...failureData.failure_categories.map(cat => [
        cat.category,
        cat.count.toString(),
        `${cat.percentage.toFixed(2)}%`,
        cat.description
      ]),
      [''],
      ['Top Error Codes'],
      ['Code', 'Count', 'Description'],
      ...failureData.top_error_codes.map(code => [
        code.code,
        code.count.toString(),
        code.description
      ])
    ];

    const csvContent = csvData.map(row => row.join(',')).join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    const url = URL.createObjectURL(blob);
    link.setAttribute('href', url);
    link.setAttribute('download', `failure-analysis-${new Date().toISOString().split('T')[0]}.csv`);
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
            <h3 className="text-sm font-medium text-red-800">Error loading failure analysis</h3>
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
          <h1 className="text-2xl font-bold text-gray-900">Failure Analysis</h1>
          {failureData && (
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
      
      {loading ? (
        <div className="flex justify-center items-center h-64">
          <LoadingSpinner />
        </div>
      ) : (
        <>
          
          {/* Summary Metrics */}
          {failureData && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="bg-white p-6 rounded-lg shadow-sm border">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-red-100 rounded-md flex items-center justify-center">
                      <svg className="w-5 h-5 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-500">Total Failures</p>
                    <p className="text-2xl font-semibold text-red-600">{formatNumber(failureData.total_failures)}</p>
                  </div>
                </div>
              </div>

              <div className="bg-white p-6 rounded-lg shadow-sm border">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-orange-100 rounded-md flex items-center justify-center">
                      <svg className="w-5 h-5 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-500">Failure Categories</p>
                    <p className="text-2xl font-semibold text-orange-600">{failureData.failure_categories.length}</p>
                  </div>
                </div>
              </div>

              <div className="bg-white p-6 rounded-lg shadow-sm border">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-purple-100 rounded-md flex items-center justify-center">
                      <svg className="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                      </svg>
                    </div>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-500">Error Codes</p>
                    <p className="text-2xl font-semibold text-purple-600">{failureData.top_error_codes.length}</p>
                  </div>
                </div>
              </div>
            </div>
          )}
          
          {/* Charts */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Failure Categories Pie Chart */}
            <div className="bg-white p-6 rounded-lg shadow-sm border">
              <ReactECharts option={getFailureCategoriesChart()} style={{ height: '400px' }} />
            </div>

            {/* Top Error Codes Bar Chart */}
            <div className="bg-white p-6 rounded-lg shadow-sm border">
              <ReactECharts option={getErrorCodesChart()} style={{ height: '400px' }} />
            </div>
          </div>
          
          {/* Detailed Tables */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Failure Categories Table */}
            {failureData?.failure_categories && (
              <div className="bg-white rounded-lg shadow-sm border">
                <div className="px-6 py-4 border-b border-gray-200">
                  <h3 className="text-lg font-medium text-gray-900">Failure Categories</h3>
                </div>
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Category
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Count
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Percentage
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {failureData.failure_categories.map((category, index) => (
                        <tr key={index}>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="flex items-center">
                              <div className={`w-3 h-3 rounded-full mr-3 ${
                                category.category.toLowerCase().includes('bounce') ? 'bg-red-500' :
                                category.category.toLowerCase().includes('defer') ? 'bg-yellow-500' :
                                category.category.toLowerCase().includes('reject') ? 'bg-gray-500' :
                                category.category.toLowerCase().includes('timeout') ? 'bg-purple-500' :
                                'bg-blue-500'
                              }`}></div>
                              <div>
                                <div className="text-sm font-medium text-gray-900">{category.category}</div>
                                <div className="text-sm text-gray-500">{category.description}</div>
                              </div>
                            </div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {formatNumber(category.count)}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {category.percentage.toFixed(1)}%
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            {/* Top Error Codes Table */}
            {failureData?.top_error_codes && (
              <div className="bg-white rounded-lg shadow-sm border">
                <div className="px-6 py-4 border-b border-gray-200">
                  <h3 className="text-lg font-medium text-gray-900">Top Error Codes</h3>
                </div>
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Code
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Count
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Description
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {failureData.top_error_codes.map((code, index) => (
                        <tr key={index}>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span className="inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800">
                              {code.code}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                            {formatNumber(code.count)}
                          </td>
                          <td className="px-6 py-4 text-sm text-gray-900">
                            {code.description}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}
          </div>
          
        </>
      )}
      
    </div>
  );
}