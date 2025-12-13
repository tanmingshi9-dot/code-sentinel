import axios from 'axios';
import { toast } from 'sonner';
import type { ApiResponse } from '@/types';

export const client = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

client.interceptors.response.use(
  (response) => {
    const data = response.data as ApiResponse<unknown>;
    if (data.code !== 0) {
      toast.error(data.message || '请求失败');
      return Promise.reject(new Error(data.message));
    }
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return data.data as any;
  },
  (error) => {
    if (!error.response) {
      toast.error('网络错误，请检查网络连接');
      return Promise.reject(error);
    }

    const status = error.response.status;
    const messages: Record<number, string> = {
      400: '请求参数错误',
      401: '未授权',
      403: '没有权限',
      404: '资源不存在',
      500: '服务器错误',
    };

    toast.error(messages[status] || `请求失败 (${status})`);
    return Promise.reject(error);
  }
);
