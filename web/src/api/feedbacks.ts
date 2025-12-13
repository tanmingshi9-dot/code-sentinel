import { client } from './client';
import type { Feedback, PaginatedData, FeedbackStats } from '@/types';

export interface FeedbackListParams {
  page?: number;
  page_size?: number;
  repo?: string;
  category?: string;
  severity?: string;
  start_date?: string;
  end_date?: string;
}

export const feedbacksApi = {
  list: (params: FeedbackListParams = {}) =>
    client.get<unknown, PaginatedData<Feedback>>('/feedbacks', { params }),

  getStats: (params: { repo?: string; start_date?: string; end_date?: string } = {}) =>
    client.get<unknown, FeedbackStats>('/feedbacks/stats', { params }),
};
