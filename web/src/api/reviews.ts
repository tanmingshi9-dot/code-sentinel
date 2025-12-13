import { client } from './client';
import type { Review, PaginatedData } from '@/types';

export interface ReviewListParams {
  page?: number;
  page_size?: number;
  repo?: string;
  status?: string;
  pr_number?: number;
  start_date?: string;
  end_date?: string;
}

export const reviewsApi = {
  list: (params: ReviewListParams = {}) =>
    client.get<unknown, PaginatedData<Review>>('/reviews', { params }),

  get: (id: number) =>
    client.get<unknown, Review>(`/reviews/${id}`),
};
