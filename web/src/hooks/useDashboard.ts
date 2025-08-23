import { useState, useEffect, useCallback } from 'react';
import { apiService } from '@/services/api';
import { webSocketService } from '@/services/websocket';
import { DashboardMetrics, WeeklyOverviewData } from '@/types/dashboard';
import { useApp } from '@/context/AppContext';

export function useDashboard() {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [weeklyData, setWeeklyData] = useState<WeeklyOverviewData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const { actions } = useApp();

  const fetchDashboardData = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      // Fetch dashboard metrics
      const metricsResponse = await apiService.get<DashboardMetrics>('/v1/dashboard');
      if (metricsResponse.success && metricsResponse.data) {
        setMetrics(metricsResponse.data);
      }

      // Fetch weekly overview data
      const weeklyResponse = await apiService.get<WeeklyOverviewData>('/v1/reports/weekly-overview');
      if (weeklyResponse.success && weeklyResponse.data) {
        setWeeklyData(weeklyResponse.data);
      }

    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch dashboard data';
      setError(errorMessage);
      actions.addNotification({
        type: 'error',
        message: errorMessage
      });
    } finally {
      setLoading(false);
    }
  }, [actions]);

  const handleMetricsUpdate = useCallback((data: DashboardMetrics) => {
    setMetrics(data);
  }, []);

  const handleWeeklyDataUpdate = useCallback((data: WeeklyOverviewData) => {
    setWeeklyData(data);
  }, []);

  useEffect(() => {
    // Initial data fetch
    fetchDashboardData();

    // Set up WebSocket listeners for real-time updates
    webSocketService.on('dashboard_metrics', handleMetricsUpdate);
    webSocketService.on('weekly_overview', handleWeeklyDataUpdate);

    // Set up periodic refresh as fallback
    const refreshInterval = setInterval(fetchDashboardData, 30000); // 30 seconds

    return () => {
      webSocketService.off('dashboard_metrics', handleMetricsUpdate);
      webSocketService.off('weekly_overview', handleWeeklyDataUpdate);
      clearInterval(refreshInterval);
    };
  }, [fetchDashboardData, handleMetricsUpdate, handleWeeklyDataUpdate]);

  const refreshData = useCallback(() => {
    fetchDashboardData();
  }, [fetchDashboardData]);

  return {
    metrics,
    weeklyData,
    loading,
    error,
    refreshData
  };
}