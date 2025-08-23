import { apiService } from './api';
import { APIResponse } from '@/types/api';
import {
  DeliverabilityReport,
  VolumeReport,
  FailureReport,
  TopSendersReport,
  TopRecipientsReport,
  DomainAnalysis,
} from '@/types/reports';

export class ReportsService {
  async getDeliverabilityReport(
    startTime?: string,
    endTime?: string
  ): Promise<APIResponse<DeliverabilityReport>> {
    const params: Record<string, string> = {};
    if (startTime) params.start_time = startTime;
    if (endTime) params.end_time = endTime;

    return apiService.get<DeliverabilityReport>('/v1/reports/deliverability', params);
  }

  async getVolumeReport(
    startTime?: string,
    endTime?: string,
    groupBy: string = 'day'
  ): Promise<APIResponse<VolumeReport>> {
    const params: Record<string, string> = { group_by: groupBy };
    if (startTime) params.start_time = startTime;
    if (endTime) params.end_time = endTime;

    return apiService.get<VolumeReport>('/v1/reports/volume', params);
  }

  async getFailureReport(
    startTime?: string,
    endTime?: string,
    limit: number = 20
  ): Promise<APIResponse<FailureReport>> {
    const params: Record<string, string> = { limit: limit.toString() };
    if (startTime) params.start_time = startTime;
    if (endTime) params.end_time = endTime;

    return apiService.get<FailureReport>('/v1/reports/failures', params);
  }

  async getTopSendersReport(
    startTime?: string,
    endTime?: string,
    limit: number = 50
  ): Promise<APIResponse<TopSendersReport>> {
    const params: Record<string, string> = { limit: limit.toString() };
    if (startTime) params.start_time = startTime;
    if (endTime) params.end_time = endTime;

    return apiService.get<TopSendersReport>('/v1/reports/top-senders', params);
  }

  async getTopRecipientsReport(
    startTime?: string,
    endTime?: string,
    limit: number = 50
  ): Promise<APIResponse<TopRecipientsReport>> {
    const params: Record<string, string> = { limit: limit.toString() };
    if (startTime) params.start_time = startTime;
    if (endTime) params.end_time = endTime;

    return apiService.get<TopRecipientsReport>('/v1/reports/top-recipients', params);
  }

  async getDomainAnalysis(
    startTime?: string,
    endTime?: string,
    analysisType: string = 'both',
    limit: number = 50
  ): Promise<APIResponse<DomainAnalysis>> {
    const params: Record<string, string> = {
      type: analysisType,
      limit: limit.toString(),
    };
    if (startTime) params.start_time = startTime;
    if (endTime) params.end_time = endTime;

    return apiService.get<DomainAnalysis>('/v1/reports/domains', params);
  }
}

export const reportsService = new ReportsService();