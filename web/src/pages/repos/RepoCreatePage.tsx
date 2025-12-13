import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { ArrowLeft } from 'lucide-react';
import { useCreateRepo, useConfigTemplates } from '@/hooks';
import { Button, Input, Select, Switch } from '@/components/ui';
import { PageHeader } from '@/components/common';

export function RepoCreatePage() {
  const navigate = useNavigate();
  const createRepo = useCreateRepo();
  const { data: templatesData } = useConfigTemplates();

  const [fullName, setFullName] = useState('');
  const [webhookSecret, setWebhookSecret] = useState('');
  const [enabled, setEnabled] = useState(true);
  const [template, setTemplate] = useState('default');
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    // Validate
    if (!fullName) {
      setError('请输入仓库名称');
      return;
    }

    const repoNameRegex = /^[a-zA-Z0-9_.-]+\/[a-zA-Z0-9_.-]+$/;
    if (!repoNameRegex.test(fullName)) {
      setError('仓库名称格式错误，应为 owner/repo');
      return;
    }

    // Get template config
    const selectedTemplate = templatesData?.templates.find((t) => t.name === template);

    try {
      await createRepo.mutateAsync({
        full_name: fullName,
        webhook_secret: webhookSecret || undefined,
        enabled,
        config: selectedTemplate?.config,
      });
      navigate('/repos');
    } catch {
      // Error handled by mutation
    }
  };

  return (
    <div>
      <PageHeader
        title="新建仓库"
        action={
          <Button variant="ghost" onClick={() => navigate('/repos')}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            返回
          </Button>
        }
      />

      <div className="max-w-xl">
        <form onSubmit={handleSubmit} className="bg-white rounded-lg border p-6 space-y-6">
          {/* Repo Name */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              仓库名称 <span className="text-red-500">*</span>
            </label>
            <Input
              placeholder="owner/repo"
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
            />
            <p className="mt-1 text-sm text-gray-500">
              格式：owner/repo，例如 facebook/react
            </p>
            {error && <p className="mt-1 text-sm text-red-500">{error}</p>}
          </div>

          {/* Template */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              配置模板
            </label>
            <Select value={template} onChange={(e) => setTemplate(e.target.value)}>
              {templatesData?.templates.map((t) => (
                <option key={t.name} value={t.name}>
                  {t.name} - {t.description}
                </option>
              ))}
            </Select>
            <p className="mt-1 text-sm text-gray-500">
              选择预设模板，创建后可在配置页面修改
            </p>
          </div>

          {/* Webhook Secret */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Webhook Secret
            </label>
            <Input
              type="password"
              placeholder="可选，用于验证 GitHub Webhook 请求"
              value={webhookSecret}
              onChange={(e) => setWebhookSecret(e.target.value)}
            />
            <p className="mt-1 text-sm text-gray-500">
              在 GitHub 仓库 Settings → Webhooks 中配置相同的 secret
            </p>
          </div>

          {/* Enabled */}
          <div className="flex items-center justify-between">
            <div>
              <label className="block text-sm font-medium text-gray-700">
                启用审查
              </label>
              <p className="text-sm text-gray-500">
                启用后将自动审查该仓库的 PR
              </p>
            </div>
            <Switch checked={enabled} onCheckedChange={setEnabled} />
          </div>

          {/* Submit */}
          <div className="flex justify-end gap-3 pt-4 border-t">
            <Button type="button" variant="outline" onClick={() => navigate('/repos')}>
              取消
            </Button>
            <Button type="submit" disabled={createRepo.isPending}>
              {createRepo.isPending ? '创建中...' : '创建仓库'}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
