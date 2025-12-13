import { useQuery } from '@tanstack/react-query';
import { reviewsApi, type ReviewListParams } from '@/api/reviews';

export const reviewKeys = {
  all: ['reviews'] as const,
  lists: () => [...reviewKeys.all, 'list'] as const,
  list: (params: ReviewListParams) => [...reviewKeys.lists(), params] as const,
  details: () => [...reviewKeys.all, 'detail'] as const,
  detail: (id: number) => [...reviewKeys.details(), id] as const,
};

export function useReviews(params: ReviewListParams = {}) {
  return useQuery({
    queryKey: reviewKeys.list(params),
    queryFn: () => reviewsApi.list(params),
    staleTime: 0,
    refetchOnMount: true,
  });
}

export function useReview(id: number) {
  return useQuery({
    queryKey: reviewKeys.detail(id),
    queryFn: () => reviewsApi.get(id),
    enabled: !!id,
  });
}
