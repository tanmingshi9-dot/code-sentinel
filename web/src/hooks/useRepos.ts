import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { reposApi, type RepoListParams, type CreateRepoRequest, type UpdateRepoRequest } from '@/api/repos';

export const repoKeys = {
  all: ['repos'] as const,
  lists: () => [...repoKeys.all, 'list'] as const,
  list: (params: RepoListParams) => [...repoKeys.lists(), params] as const,
  details: () => [...repoKeys.all, 'detail'] as const,
  detail: (id: number) => [...repoKeys.details(), id] as const,
  templates: () => [...repoKeys.all, 'templates'] as const,
};

export function useRepos(params: RepoListParams = {}) {
  return useQuery({
    queryKey: repoKeys.list(params),
    queryFn: () => reposApi.list(params),
  });
}

export function useRepo(id: number) {
  return useQuery({
    queryKey: repoKeys.detail(id),
    queryFn: () => reposApi.get(id),
    enabled: !!id,
  });
}

export function useConfigTemplates() {
  return useQuery({
    queryKey: repoKeys.templates(),
    queryFn: () => reposApi.getTemplates(),
  });
}

export function useCreateRepo() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateRepoRequest) => reposApi.create(data),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: repoKeys.lists() });
      if (data.restored) {
        toast.success('仓库已恢复（之前被删除过）');
      } else {
        toast.success('仓库创建成功');
      }
    },
  });
}

export function useUpdateRepo() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateRepoRequest }) =>
      reposApi.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: repoKeys.lists() });
      queryClient.invalidateQueries({ queryKey: repoKeys.detail(id) });
      toast.success('仓库更新成功');
    },
  });
}


export function useDeleteRepo() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: number) => reposApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: repoKeys.lists() });
      toast.success('仓库删除成功');
    },
  });
}

export function useToggleRepo() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, enabled }: { id: number; enabled: boolean }) =>
      reposApi.toggle(id, enabled),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: repoKeys.lists() });
    },
  });
}
