import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Plus, Settings, Trash2, Search } from 'lucide-react';
import { useRepos, useDeleteRepo, useToggleRepo } from '@/hooks';
import { Button, Input, Switch, Table, TableHeader, TableBody, TableRow, TableHead, TableCell } from '@/components/ui';
import { PageHeader, PageLoading, EmptyState, ErrorMessage, Pagination, ConfirmDialog } from '@/components/common';
import { formatRelativeTime } from '@/lib/utils';

export function RepoListPage() {
  const navigate = useNavigate();
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [deleteId, setDeleteId] = useState<number | null>(null);

  const { data, isLoading, isError, refetch } = useRepos({ page, page_size: 20, search });
  const deleteRepo = useDeleteRepo();
  const toggleRepo = useToggleRepo();

  const handleDelete = async () => {
    if (deleteId) {
      await deleteRepo.mutateAsync(deleteId);
      setDeleteId(null);
    }
  };

  const handleToggle = (id: number, enabled: boolean) => {
    toggleRepo.mutate({ id, enabled });
  };

  if (isLoading) return <PageLoading />;
  if (isError) return <ErrorMessage onRetry={refetch} />;

  return (
    <div>
      <PageHeader
        title="仓库管理"
        description="管理接入的 GitHub 仓库及其审查配置"
        action={
          <Button onClick={() => navigate('/repos/new')}>
            <Plus className="w-4 h-4 mr-2" />
            新建仓库
          </Button>
        }
      />

      {/* Search */}
      <div className="mb-4">
        <div className="relative w-64">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <Input
            placeholder="搜索仓库..."
            value={search}
            onChange={(e) => {
              setSearch(e.target.value);
              setPage(1);
            }}
            className="pl-9"
          />
        </div>
      </div>

      {/* Table */}
      {!data?.items.length ? (
        <EmptyState
          message="暂无仓库"
          description="点击上方按钮添加第一个仓库"
          action={
            <Button onClick={() => navigate('/repos/new')}>
              <Plus className="w-4 h-4 mr-2" />
              新建仓库
            </Button>
          }
        />
      ) : (
        <div className="bg-white rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>仓库名称</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>审查次数</TableHead>
                <TableHead>最后审查</TableHead>
                <TableHead className="w-[120px]">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {data.items.map((repo) => (
                <TableRow key={repo.id}>
                  <TableCell className="font-medium">{repo.full_name}</TableCell>
                  <TableCell>
                    <Switch
                      checked={repo.enabled}
                      onCheckedChange={(enabled) => handleToggle(repo.id, enabled)}
                    />
                  </TableCell>
                  <TableCell>{repo.review_count}</TableCell>
                  <TableCell className="text-gray-500">
                    {formatRelativeTime(repo.last_review_at)}
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-2">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => navigate(`/repos/${repo.id}`)}
                      >
                        <Settings className="w-4 h-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setDeleteId(repo.id)}
                      >
                        <Trash2 className="w-4 h-4 text-red-500" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          <Pagination
            page={page}
            pageSize={20}
            total={data.total}
            onChange={setPage}
          />
        </div>
      )}

      {/* Delete Confirm */}
      <ConfirmDialog
        open={!!deleteId}
        onOpenChange={() => setDeleteId(null)}
        title="删除仓库"
        description="确定要删除该仓库吗？删除后将无法恢复，相关的审查记录也会被删除。"
        variant="destructive"
        confirmText="删除"
        onConfirm={handleDelete}
        loading={deleteRepo.isPending}
      />
    </div>
  );
}
