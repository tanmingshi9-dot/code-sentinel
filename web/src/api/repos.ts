import { client } from './client';
import type { Repo, PaginatedData, RepoConfig, ConfigTemplate } from '@/types';

export interface RepoListParams {
  page?: number;
  page_size?: number;
  search?: string;
}

export interface CreateRepoRequest {
  full_name: string;
  webhook_secret?: string;
  enabled?: boolean;
  config?: Partial<RepoConfig>;
}

export interface UpdateRepoRequest {
  webhook_secret?: string;
  enabled?: boolean;
  config?: RepoConfig;
}

export const reposApi = {
  list: (params: RepoListParams = {}) =>
    client.get<unknown, PaginatedData<Repo>>('/repos', { params }),

  get: (id: number) =>
    client.get<unknown, Repo>(`/repos/${id}`),

  create: (data: CreateRepoRequest) =>
    client.post<unknown, Repo & { restored?: boolean }>('/repos', data),

  update: (id: number, data: UpdateRepoRequest) =>
    client.put<unknown, Repo>(`/repos/${id}`, data),

  delete: (id: number) =>
    client.delete(`/repos/${id}`),

  toggle: (id: number, enabled: boolean) =>
    client.put<unknown, Repo>(`/repos/${id}/toggle`, { enabled }),

  getTemplates: () =>
    client.get<unknown, { templates: ConfigTemplate[] }>('/config-templates'),
};
