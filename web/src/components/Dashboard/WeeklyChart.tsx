import ReactECharts from 'echarts-for-react';
import { EChartsOption } from 'echarts';
import { WeeklyOverviewData } from '@/types/dashboard';
import { HelpTooltip } from '../Common/HelpTooltip';
import { getHelpContent } from '../../utils/helpContent';

interface WeeklyChartProps {
  data: WeeklyOverviewData | null;
  loading?: boolean;
}

export function WeeklyChart({ data, loading = false }: WeeklyChartProps) {
  if (loading || !data) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6">
        <h3 className="text-lg font-semibold text-gray-800 mb-4">Weekly Email Overview</h3>
        <div className="h-96 flex items-center justify-center">
          <div className="animate-pulse">
            <div className="h-4 bg-gray-300 rounded w-48 mb-4"></div>
            <div className="space-y-3">
              <div className="h-3 bg-gray-300 rounded w-full"></div>
              <div className="h-3 bg-gray-300 rounded w-5/6"></div>
              <div className="h-3 bg-gray-300 rounded w-4/6"></div>
              <div className="h-3 bg-gray-300 rounded w-3/6"></div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  const option: EChartsOption = {
    title: {
      text: 'Weekly Email Overview',
      left: 'center',
      textStyle: {
        fontSize: 18,
        fontWeight: 'bold',
        color: '#374151'
      }
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      },
      backgroundColor: 'rgba(255, 255, 255, 0.95)',
      borderColor: '#E5E7EB',
      borderWidth: 1,
      textStyle: {
        color: '#374151'
      },
      formatter: function (params: any) {
        const total = params.reduce((sum: number, param: any) => sum + param.value, 0);
        let result = `<div style="font-weight: bold; margin-bottom: 8px; border-bottom: 1px solid #E5E7EB; padding-bottom: 4px;">${params[0].axisValue}</div>`;
        
        params.forEach((param: any) => {
          const color = param.color;
          const name = param.seriesName;
          const value = param.value.toLocaleString();
          const percentage = total > 0 ? ((param.value / total) * 100).toFixed(1) : '0.0';
          result += `<div style="display: flex; align-items: center; margin: 4px 0;">
            <span style="display:inline-block;margin-right:8px;border-radius:50%;width:10px;height:10px;background-color:${color}"></span>
            <span style="color: #374151;">${name}: <strong>${value}</strong> (${percentage}%)</span>
          </div>`;
        });
        
        result += `<div style="margin-top: 8px; padding-top: 4px; border-top: 1px solid #E5E7EB; font-weight: bold;">Total: ${total.toLocaleString()}</div>`;
        return result;
      }
    },
    legend: {
      data: ['Delivered', 'Failed', 'Pending', 'Deferred'],
      bottom: 10,
      textStyle: {
        color: '#6B7280'
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      top: '15%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: data.dates,
      axisLabel: {
        color: '#6B7280',
        fontSize: 12
      },
      axisLine: {
        lineStyle: {
          color: '#E5E7EB'
        }
      }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        color: '#6B7280',
        fontSize: 12,
        formatter: function (value: number) {
          if (value >= 1000000) {
            return (value / 1000000).toFixed(1) + 'M';
          } else if (value >= 1000) {
            return (value / 1000).toFixed(1) + 'K';
          }
          return value.toString();
        }
      },
      axisLine: {
        lineStyle: {
          color: '#E5E7EB'
        }
      },
      splitLine: {
        lineStyle: {
          color: '#F3F4F6'
        }
      }
    },
    series: [
      {
        name: 'Delivered',
        type: 'bar',
        data: data.delivered,
        color: '#10B981',
        emphasis: {
          focus: 'series'
        },
        animationDelay: function (idx: number) {
          return idx * 100;
        }
      },
      {
        name: 'Failed',
        type: 'bar',
        data: data.failed,
        color: '#EF4444',
        emphasis: {
          focus: 'series'
        },
        animationDelay: function (idx: number) {
          return idx * 100 + 100;
        }
      },
      {
        name: 'Pending',
        type: 'bar',
        data: data.pending,
        color: '#F59E0B',
        emphasis: {
          focus: 'series'
        },
        animationDelay: function (idx: number) {
          return idx * 100 + 200;
        }
      },
      {
        name: 'Deferred',
        type: 'bar',
        data: data.deferred,
        color: '#6B7280',
        emphasis: {
          focus: 'series'
        },
        animationDelay: function (idx: number) {
          return idx * 100 + 300;
        }
      }
    ],
    toolbox: {
      show: true,
      feature: {
        dataZoom: {
          yAxisIndex: 'none'
        },
        restore: {},
        saveAsImage: {
          name: 'weekly-email-overview'
        }
      },
      right: 20,
      top: 20
    },
    dataZoom: [
      {
        type: 'inside',
        start: 0,
        end: 100
      },
      {
        start: 0,
        end: 100,
        height: 30,
        bottom: 50
      }
    ],
    animationEasing: 'elasticOut',
    animationDelayUpdate: function (idx: number) {
      return idx * 50;
    }
  };

  const onChartClick = (params: any) => {
    if (params.componentType === 'series') {
      console.log('Chart clicked:', {
        date: params.name,
        series: params.seriesName,
        value: params.value
      });
      // Could trigger navigation to detailed view or show more info
    }
  };

  // Calculate totals for the period
  const totals = {
    delivered: data.delivered.reduce((sum, val) => sum + val, 0),
    failed: data.failed.reduce((sum, val) => sum + val, 0),
    pending: data.pending.reduce((sum, val) => sum + val, 0),
    deferred: data.deferred.reduce((sum, val) => sum + val, 0)
  };
  
  const grandTotal = totals.delivered + totals.failed + totals.pending + totals.deferred;

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      <div className="flex items-center gap-2 mb-4">
        <h3 className="text-lg font-semibold text-gray-800">Weekly Email Overview</h3>
        <HelpTooltip 
          content={getHelpContent('dashboard', 'weeklyChart')}
          position="right"
        />
      </div>
      <ReactECharts 
        option={option} 
        style={{ height: '400px', width: '100%' }} 
        opts={{ renderer: 'canvas' }}
        onEvents={{
          'click': onChartClick
        }}
      />
      
      {/* Summary Statistics */}
      <div className="mt-6 pt-4 border-t border-gray-200">
        <h4 className="text-sm font-semibold text-gray-700 mb-3">Period Summary</h4>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
          <div className="text-center">
            <div className="text-lg font-bold text-green-600">{totals.delivered.toLocaleString()}</div>
            <div className="text-gray-600">Delivered</div>
            <div className="text-xs text-gray-500">
              {grandTotal > 0 ? ((totals.delivered / grandTotal) * 100).toFixed(1) : '0'}%
            </div>
          </div>
          <div className="text-center">
            <div className="text-lg font-bold text-red-600">{totals.failed.toLocaleString()}</div>
            <div className="text-gray-600">Failed</div>
            <div className="text-xs text-gray-500">
              {grandTotal > 0 ? ((totals.failed / grandTotal) * 100).toFixed(1) : '0'}%
            </div>
          </div>
          <div className="text-center">
            <div className="text-lg font-bold text-yellow-600">{totals.pending.toLocaleString()}</div>
            <div className="text-gray-600">Pending</div>
            <div className="text-xs text-gray-500">
              {grandTotal > 0 ? ((totals.pending / grandTotal) * 100).toFixed(1) : '0'}%
            </div>
          </div>
          <div className="text-center">
            <div className="text-lg font-bold text-gray-600">{totals.deferred.toLocaleString()}</div>
            <div className="text-gray-600">Deferred</div>
            <div className="text-xs text-gray-500">
              {grandTotal > 0 ? ((totals.deferred / grandTotal) * 100).toFixed(1) : '0'}%
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}