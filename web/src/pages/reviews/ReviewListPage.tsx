import { useState } from 'react';
import { Eye, Search } from 'lucide-react';
import { useReviews } from '@/hooks';
import { Button, Input, Select, Table, TableHeader, TableBody, TableRow, TableHead, TableCell, Dialog, DialogHeader, DialogContent } from '@/components/ui';
import { PageHeader, PageLoading, EmptyState, ErrorMessage, Pagination, StatusBadge, SeverityBadge } from '@/components/common';
import { formatRelativeTime } from '@/lib/utils';
import type { Review, ReviewResult, ReviewIssue } from '@/types';

export function ReviewListPage() {
  const [page, setPage] = useState(1);
  const [repo, setRepo] = useState('');
  const [status, setStatus] = useState('');
  const [selectedReview, setSelectedReview] = useState<Review | null>(null);

  const { data, isLoading, isError, refetch } = useReviews({
    page,
    page_size: 20,
    repo: repo || undefined,
    status: status || undefined,
  });

  const parseResult = (result: string): ReviewResult | null => {
    try {
      return JSON.parse(result);
    } catch {
      return null;
    }
  };

  const getIssueCounts = (result: string) => {
    const parsed = parseResult(result);
    if (!parsed?.stats) return null;
    return parsed.stats;
  };

  if (isLoading) return <PageLoading />;
  if (isError) return <ErrorMessage onRetry={refetch} />;

  return (
    <div>
      <PageHeader title="审查历史" description="查看所有 PR 的审查记录" />

      {/* Filters */}
      <div className="flex gap-4 mb-4">
        <div className="relative w-64">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <Input
            placeholder="按仓库筛选..."
            value={repo}
            onChange={(e) => { setRepo(e.target.value); setPage(1); }}
            className="pl-9"
          />
        </div>
        <Select value={status} onChange={(e) => { setStatus(e.target.value); setPage(1); }} className="w-40">
          <option value="">全部状态</option>
          <option value="completed">已完成</option>
          <option value="failed">失败</option>
          <option value="skipped">已跳过</option>
          <option value="running">运行中</option>
        </Select>
      </div>

      {/* Table */}
      {!data?.items.length ? (
        <EmptyState message="暂无审查记录" />
      ) : (
        <div className="bg-white rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>仓库</TableHead>
                <TableHead>PR</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>问题</TableHead>
                <TableHead>Token</TableHead>
                <TableHead>耗时</TableHead>
                <TableHead>时间</TableHead>
                <TableHead className="w-[80px]">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.items.map((review) => {
                const stats = getIssueCounts(review.result);
                return (
                  <TableRow key={review.id}>
                    <TableCell className="font-medium">{review.repo_full_name}</TableCell>
                    <TableCell>
                      <div>
                        <span className="text-blue-600">#{review.pr_number}</span>
                        <p className="text-sm text-gray-500 truncate max-w-[200px]">{review.pr_title}</p>
                      </div>
                    </TableCell>
                    <TableCell>
                      <StatusBadge status={review.status} />
                    </TableCell>
                    <TableCell>
                      {stats ? (
                        <div className="flex gap-1">
                          {stats.p0_count > 0 && <SeverityBadge severity="P0" />}
                          {stats.p1_count > 0 && <SeverityBadge severity="P1" />}
                          {stats.p2_count > 0 && <SeverityBadge severity="P2" />}
                          {stats.p0_count === 0 && stats.p1_count === 0 && stats.p2_count === 0 && (
                            <span className="text-gray-400">-</span>
                          )}
                        </div>
                      ) : (
                        <span className="text-gray-400">-</span>
                      )}
                    </TableCell>
                    <TableCell>{review.token_used || '-'}</TableCell>
                    <TableCell>{review.duration_ms ? `${(review.duration_ms / 1000).toFixed(1)}s` : '-'}</TableCell>
                    <TableCell className="text-gray-500">{formatRelativeTime(review.created_at)}</TableCell>
                    <TableCell>
                      <Button variant="ghost" size="icon" onClick={() => setSelectedReview(review)}>
                        <Eye className="w-4 h-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
          <Pagination page={page} pageSize={20} total={data.total} onChange={setPage} />
        </div>
      )}

      {/* Detail Dialog */}
      <ReviewDetailDialog review={selectedReview} onClose={() => setSelectedReview(null)} />
    </div>
  );
}


interface ReviewDetailDialogProps {
  review: Review | null;
  onClose: () => void;
}

function ReviewDetailDialog({ review, onClose }: ReviewDetailDialogProps) {
  if (!review) return null;

  let result: ReviewResult | null = null;
  try {
    result = JSON.parse(review.result);
  } catch {
    // ignore
  }

  return (
    <Dialog open={!!review} onOpenChange={onClose}>
      <div className="w-[700px] max-h-[80vh] overflow-auto">
        <DialogHeader onClose={onClose}>
          审查详情 - #{review.pr_number}
        </DialogHeader>
        <DialogContent>
          {/* PR Info */}
          <div className="mb-6 p-4 bg-gray-50 rounded-lg">
            <h3 className="font-medium mb-2">{review.pr_title}</h3>
            <div className="text-sm text-gray-600 space-y-1">
              <p>仓库：{review.repo_full_name}</p>
              <p>作者：{review.pr_author}</p>
              <p>Commit：{review.commit_sha?.slice(0, 7)}</p>
              <p>状态：<StatusBadge status={review.status} /></p>
            </div>
          </div>

          {/* Result */}
          {result ? (
            <div className="space-y-4">
              {/* Summary */}
              <div>
                <h4 className="font-medium mb-2">总结</h4>
                <p className="text-sm text-gray-700">{result.summary || '无'}</p>
              </div>

              {/* Issues */}
              {result.issues && result.issues.length > 0 ? (
                <div>
                  <h4 className="font-medium mb-2">问题列表 ({result.issues.length})</h4>
                  <div className="space-y-3">
                    {result.issues.map((issue: ReviewIssue, idx: number) => (
                      <div key={idx} className="p-3 border rounded-lg">
                        <div className="flex items-center gap-2 mb-2">
                          <SeverityBadge severity={issue.severity} />
                          <span className="font-medium">{issue.title}</span>
                        </div>
                        <p className="text-sm text-gray-500 mb-1">
                          {issue.file}:{issue.line}
                        </p>
                        <p className="text-sm text-gray-700 mb-2">{issue.description}</p>
                        {issue.suggestion && (
                          <p className="text-sm text-blue-600">💡 {issue.suggestion}</p>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              ) : (
                <p className="text-sm text-gray-500">✅ 未发现问题</p>
              )}

              {/* Meta */}
              <div className="pt-4 border-t text-sm text-gray-500">
                <p>模型：{result.model || '-'}</p>
                <p>Token：{review.token_used}</p>
                <p>耗时：{review.duration_ms ? `${(review.duration_ms / 1000).toFixed(2)}s` : '-'}</p>
              </div>
            </div>
          ) : (
            <div className="text-sm text-gray-500">
              {review.error_msg ? (
                <p className="text-red-500">错误：{review.error_msg}</p>
              ) : (
                <p>{review.result || '无结果'}</p>
              )}
            </div>
          )}
        </DialogContent>
      </div>
    </Dialog>
  );
}
