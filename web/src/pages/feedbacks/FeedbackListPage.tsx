import { useState } from 'react';
import { Search } from 'lucide-react';
import { useFeedbacks } from '@/hooks';
import { Input, Select, Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui';
import { PageHeader, PageLoading, EmptyState, ErrorMessage, Pagination, SeverityBadge } from '@/components/common';
import { formatRelativeTime } from '@/lib/utils';

export function FeedbackListPage() {
  const [page, setPage] = useState(1);
  const [repo, setRepo] = useState('');
  const [severity, setSeverity] = useState('');
  const [category, setCategory] = useState('');

  const { data, isLoading, isError, refetch } = useFeedbacks({
    page,
    page_size: 20,
    repo: repo || undefined,
    severity: severity || undefined,
    category: category || undefined,
  });

  if (isLoading) return <PageLoading />;
  if (isError) return <ErrorMessage onRetry={refetch} />;

  return (
    <div>
      <PageHeader
        title="误报反馈"
        description="查看用户标记的误报记录，用于持续优化审查质量"
      />

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
        <Select value={severity} onChange={(e) => { setSeverity(e.target.value); setPage(1); }} className="w-32">
          <option value="">全部级别</option>
          <option value="P0">P0</option>
          <option value="P1">P1</option>
          <option value="P2">P2</option>
        </Select>
        <Select value={category} onChange={(e) => { setCategory(e.target.value); setPage(1); }} className="w-32">
          <option value="">全部类型</option>
          <option value="security">安全</option>
          <option value="performance">性能</option>
          <option value="logic">逻辑</option>
          <option value="style">风格</option>
        </Select>
      </div>

      {/* Table */}
      {!data?.items.length ? (
        <EmptyState message="暂无误报记录" description="当用户在 PR 评论中回复 /false 时，会记录在这里" />
      ) : (
        <div className="bg-white rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>仓库</TableHead>
                <TableHead>PR</TableHead>
                <TableHead>位置</TableHead>
                <TableHead>级别</TableHead>
                <TableHead>问题标题</TableHead>
                <TableHead>误报原因</TableHead>
                <TableHead>反馈人</TableHead>
                <TableHead>时间</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.items.map((feedback) => (
                <TableRow key={feedback.id}>
                  <TableCell className="font-medium">{feedback.repo_full_name}</TableCell>
                  <TableCell>
                    <span className="text-blue-600">#{feedback.pr_number}</span>
                  </TableCell>
                  <TableCell className="text-sm text-gray-500">
                    {feedback.file}:{feedback.line}
                  </TableCell>
                  <TableCell>
                    <SeverityBadge severity={feedback.severity} />
                  </TableCell>
                  <TableCell className="max-w-[200px] truncate">{feedback.title}</TableCell>
                  <TableCell className="max-w-[200px] truncate text-gray-500">
                    {feedback.reason || '-'}
                  </TableCell>
                  <TableCell>{feedback.reporter}</TableCell>
                  <TableCell className="text-gray-500">
                    {formatRelativeTime(feedback.created_at)}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <Pagination page={page} pageSize={20} total={data.total} onChange={setPage} />
        </div>
      )}
    </div>
  );
}
