import { useQuery } from '@tanstack/react-query';
import { feedbacksApi, type FeedbackListParams } from '@/api/feedbacks';

export const feedbackKeys = {
  all: ['feedbacks'] as const,
  lists: () => [...feedbackKeys.all, 'list'] as const,
  list: (params: FeedbackListParams) => [...feedbackKeys.lists(), params] as const,
  stats: (params?: { repo?: string }) => [...feedbackKeys.all, 'stats', params] as const,
};

export function useFeedbacks(params: FeedbackListParams = {}) {
  return useQuery({
    queryKey: feedbackKeys.list(params),
    queryFn: () => feedbacksApi.list(params),
    staleTime: 0,
    refetchOnMount: true,
  });
}

export function useFeedbackStats(params: { repo?: string; start_date?: string; end_date?: string } = {}) {
  return useQuery({
    queryKey: feedbackKeys.stats(params),
    queryFn: () => feedbacksApi.getStats(params),
  });
}
