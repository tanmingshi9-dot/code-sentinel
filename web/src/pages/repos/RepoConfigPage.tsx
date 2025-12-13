import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { ArrowLeft, Save } from 'lucide-react';
import { useRepo, useUpdateRepo } from '@/hooks';
import { Button, Input, Select, Switch } from '@/components/ui';
import { PageHeader, PageLoading, ErrorMessage } from '@/components/common';
import type { RepoConfig } from '@/types';

const defaultConfig: RepoConfig = {
  llm_provider: 'openai',
  model: 'gpt-4-turbo',
  max_tokens: 4096,
  system_prompt: '',
  review_focus: ['security', 'performance', 'logic'],
  min_severity: 'P1',
  languages: ['go', 'python', 'javascript'],
  ignore_files: ['*.test.go', 'vendor/*'],
  max_diff_lines: 1000,
  auto_review: true,
  // 仓库级配置（可选）
  llm_api_key: '',
  llm_base_url: '',
  github_token: '',
};

const llmProviders = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'qwen', label: '通义千问' },
  { value: 'azure', label: 'Azure OpenAI' },
  { value: 'ollama', label: 'Ollama (本地)' },
];

const defaultBaseURLs: Record<string, string> = {
  openai: 'https://api.openai.com/v1',
  qwen: 'https://dashscope.aliyuncs.com/compatible-mode/v1',
  azure: '',
  ollama: 'http://localhost:11434/v1',
};

const models: Record<string, string[]> = {
  openai: ['gpt-4-turbo', 'gpt-4', 'gpt-3.5-turbo'],
  qwen: ['qwen-max', 'qwen-plus', 'qwen-turbo'],
  azure: ['gpt-4', 'gpt-35-turbo'],
  ollama: ['llama2', 'codellama', 'mistral'],
};

const reviewFocusOptions = [
  { value: 'security', label: '安全问题' },
  { value: 'performance', label: '性能问题' },
  { value: 'logic', label: '逻辑错误' },
  { value: 'style', label: '代码风格' },
];

const languageOptions = ['go', 'python', 'javascript', 'typescript', 'java', 'rust', 'c', 'cpp', 'markdown', 'yaml', 'json', 'shell'];

// 默认系统提示词模板
const DEFAULT_SYSTEM_PROMPT = `你是资深代码审查专家，精通多种编程语言开发。

你的任务是审查代码变更，识别潜在问题，并提供详细的修复建议。

## 审查重点
- 安全问题：SQL 注入、XSS、硬编码密钥、敏感信息泄露、不安全的加密
- 性能问题：循环内查库、N+1 查询、不必要的重复计算、内存泄漏
- 逻辑错误：空指针、边界条件、异常处理不当、死循环、竞态条件
- 代码风格：命名规范、注释质量、代码可读性、过长函数

## 严重程度定义
- P0（严重）：安全漏洞、会导致系统崩溃或数据泄露的问题
- P1（重要）：性能问题、明显的逻辑错误、潜在的 Bug
- P2（建议）：代码风格、注释质量、可读性改进

## 输出格式要求
请严格按照以下 JSON 格式输出，不要添加任何额外内容：

{
  "summary": "本次审查总体评价（1-2句话）",
  "issues": [
    {
      "severity": "P0|P1|P2",
      "category": "security|performance|logic|style",
      "file": "文件路径",
      "line": 行号,
      "title": "问题标题（简短）",
      "description": "问题详细描述",
      "suggestion": "修复建议",
      "code_fix": "修复后的代码片段（可选）"
    }
  ],
  "stats": {
    "p0_count": 0,
    "p1_count": 0,
    "p2_count": 0
  }
}

## 注意事项
- 如果代码没有问题，issues 返回空数组，summary 写 "代码质量良好，未发现明显问题"
- code_fix 字段仅在能提供具体修复代码时填写
- 保持客观和专业，避免主观判断
- 确保输出的是合法的 JSON，不要包含注释或额外文本`;

export function RepoConfigPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: repo, isLoading, isError, refetch } = useRepo(Number(id));
  const updateRepo = useUpdateRepo();

  const [config, setConfig] = useState<RepoConfig>(defaultConfig);
  const [enabled, setEnabled] = useState(true);
  const [webhookSecret, setWebhookSecret] = useState('');

  useEffect(() => {
    if (repo) {
      setEnabled(repo.enabled);
      setWebhookSecret(''); // webhook_secret 不回显，只能重新设置
      if (repo.config) {
        try {
          const parsed = JSON.parse(repo.config);
          // 如果 system_prompt 为空，预填默认模板
          if (!parsed.system_prompt) {
            parsed.system_prompt = DEFAULT_SYSTEM_PROMPT;
          }
          setConfig({ ...defaultConfig, ...parsed });
        } catch {
          setConfig({ ...defaultConfig, system_prompt: DEFAULT_SYSTEM_PROMPT });
        }
      } else {
        setConfig({ ...defaultConfig, system_prompt: DEFAULT_SYSTEM_PROMPT });
      }
    }
  }, [repo]);

  const handleSave = async () => {
    await updateRepo.mutateAsync({
      id: Number(id),
      data: {
        enabled,
        config,
        webhook_secret: webhookSecret || undefined,
      },
    });
  };

  const toggleFocus = (focus: string) => {
    setConfig((prev) => ({
      ...prev,
      review_focus: prev.review_focus.includes(focus as never)
        ? prev.review_focus.filter((f) => f !== focus)
        : [...prev.review_focus, focus as never],
    }));
  };

  const toggleLanguage = (lang: string) => {
    setConfig((prev) => ({
      ...prev,
      languages: prev.languages.includes(lang)
        ? prev.languages.filter((l) => l !== lang)
        : [...prev.languages, lang],
    }));
  };

  if (isLoading) return <PageLoading />;
  if (isError) return <ErrorMessage onRetry={refetch} />;

  return (
    <div>
      <PageHeader
        title={repo?.full_name || '仓库配置'}
        action={
          <div className="flex gap-2">
            <Button variant="ghost" onClick={() => navigate('/repos')}>
              <ArrowLeft className="w-4 h-4 mr-2" />
              返回
            </Button>
            <Button onClick={handleSave} disabled={updateRepo.isPending}>
              <Save className="w-4 h-4 mr-2" />
              {updateRepo.isPending ? '保存中...' : '保存配置'}
            </Button>
          </div>
        }
      />

      <div className="space-y-6">
        {/* Basic Config */}
        <section className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold mb-4">基础配置</h2>
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <label className="block text-sm font-medium text-gray-700">启用审查</label>
                <p className="text-sm text-gray-500">关闭后将不再自动审查该仓库的 PR</p>
              </div>
              <Switch checked={enabled} onCheckedChange={setEnabled} />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">LLM 提供商</label>
                <Select
                  value={config.llm_provider}
                  onChange={(e) => setConfig((prev) => ({ ...prev, llm_provider: e.target.value as never, model: models[e.target.value][0] }))}
                >
                  {llmProviders.map((p) => (
                    <option key={p.value} value={p.value}>{p.label}</option>
                  ))}
                </Select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">模型</label>
                <Select
                  value={config.model}
                  onChange={(e) => setConfig((prev) => ({ ...prev, model: e.target.value }))}
                >
                  {models[config.llm_provider]?.map((m) => (
                    <option key={m} value={m}>{m}</option>
                  ))}
                </Select>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">最大 Token</label>
              <Input
                type="number"
                value={config.max_tokens}
                onChange={(e) => setConfig((prev) => ({ ...prev, max_tokens: Number(e.target.value) }))}
                className="w-32"
              />
            </div>
          </div>
        </section>

        {/* Repo-level Credentials */}
        <section className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold mb-4">仓库级凭证配置</h2>
          <p className="text-sm text-gray-500 mb-4">
            可选配置。如果留空，将使用全局配置。配置后该仓库将使用独立的 API Key 和 Token。
          </p>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">LLM API Key</label>
              <Input
                type="password"
                value={config.llm_api_key || ''}
                onChange={(e) => setConfig((prev) => ({ ...prev, llm_api_key: e.target.value }))}
                placeholder="留空使用全局配置"
              />
              <p className="mt-1 text-sm text-gray-500">用于调用大模型 API 的密钥</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">LLM Base URL</label>
              <Input
                value={config.llm_base_url || ''}
                onChange={(e) => setConfig((prev) => ({ ...prev, llm_base_url: e.target.value }))}
                placeholder={defaultBaseURLs[config.llm_provider] || '留空使用默认地址'}
              />
              <p className="mt-1 text-sm text-gray-500">API 地址，留空使用提供商默认地址</p>
            </div>

            <div className="pt-4 border-t">
              <label className="block text-sm font-medium text-gray-700 mb-1">GitHub Token</label>
              <Input
                type="password"
                value={config.github_token || ''}
                onChange={(e) => setConfig((prev) => ({ ...prev, github_token: e.target.value }))}
                placeholder="留空使用全局配置"
              />
              <p className="mt-1 text-sm text-gray-500">
                用于访问该仓库的 GitHub Personal Access Token，适用于跨组织/账号场景
              </p>
            </div>

            <div className="pt-4 border-t">
              <label className="block text-sm font-medium text-gray-700 mb-1">Webhook Secret</label>
              <Input
                type="password"
                value={webhookSecret}
                onChange={(e) => setWebhookSecret(e.target.value)}
                placeholder="留空保持不变"
              />
              <p className="mt-1 text-sm text-gray-500">
                用于验证 GitHub Webhook 请求，需与 GitHub 仓库 Webhook 配置一致
              </p>
            </div>
          </div>
        </section>

        {/* Review Config */}
        <section className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold mb-4">审查配置</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">审查重点</label>
              <div className="flex flex-wrap gap-2">
                {reviewFocusOptions.map((opt) => (
                  <button
                    key={opt.value}
                    type="button"
                    onClick={() => toggleFocus(opt.value)}
                    className={`px-3 py-1.5 rounded-full text-sm border transition-colors ${
                      config.review_focus.includes(opt.value as never)
                        ? 'bg-blue-50 border-blue-500 text-blue-700'
                        : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
                    }`}
                  >
                    {opt.label}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">最小报告级别</label>
              <Select
                value={config.min_severity}
                onChange={(e) => setConfig((prev) => ({ ...prev, min_severity: e.target.value as never }))}
                className="w-48"
              >
                <option value="P0">仅 P0（严重问题）</option>
                <option value="P1">P0 + P1（严重 + 重要）</option>
                <option value="P2">全部（P0 + P1 + P2）</option>
              </Select>
            </div>

            <div>
              <div className="flex items-center justify-between mb-1">
                <label className="block text-sm font-medium text-gray-700">系统提示词</label>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => setConfig((prev) => ({ ...prev, system_prompt: DEFAULT_SYSTEM_PROMPT }))}
                >
                  重置为默认模板
                </Button>
              </div>
              <textarea
                value={config.system_prompt}
                onChange={(e) => setConfig((prev) => ({ ...prev, system_prompt: e.target.value }))}
                className="w-full h-80 px-3 py-2 border border-gray-300 rounded-md text-sm font-mono focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="mt-1 text-sm text-gray-500">
                定义 AI 审查代码时使用的系统提示词，支持自定义审查规则和输出格式
              </p>
            </div>
          </div>
        </section>

        {/* Filter Config */}
        <section className="bg-white rounded-lg border p-6">
          <h2 className="text-lg font-semibold mb-4">过滤规则</h2>
          <div className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">支持语言</label>
              <div className="flex flex-wrap gap-2">
                {languageOptions.map((lang) => (
                  <button
                    key={lang}
                    type="button"
                    onClick={() => toggleLanguage(lang)}
                    className={`px-3 py-1.5 rounded-full text-sm border transition-colors ${
                      config.languages.includes(lang)
                        ? 'bg-blue-50 border-blue-500 text-blue-700'
                        : 'bg-white border-gray-300 text-gray-700 hover:bg-gray-50'
                    }`}
                  >
                    {lang}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">忽略文件</label>
              <Input
                value={config.ignore_files.join(', ')}
                onChange={(e) => setConfig((prev) => ({ ...prev, ignore_files: e.target.value.split(',').map((s) => s.trim()).filter(Boolean) }))}
                placeholder="*.test.go, vendor/*, node_modules/*"
              />
              <p className="mt-1 text-sm text-gray-500">多个规则用逗号分隔，支持 glob 模式</p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">最大 Diff 行数</label>
              <Input
                type="number"
                value={config.max_diff_lines}
                onChange={(e) => setConfig((prev) => ({ ...prev, max_diff_lines: Number(e.target.value) }))}
                className="w-32"
              />
              <p className="mt-1 text-sm text-gray-500">超过此行数的 PR 将跳过审查</p>
            </div>

            <div className="flex items-center justify-between pt-4 border-t">
              <div>
                <label className="block text-sm font-medium text-gray-700">自动审查</label>
                <p className="text-sm text-gray-500">关闭后需手动触发审查</p>
              </div>
              <Switch
                checked={config.auto_review}
                onCheckedChange={(checked) => setConfig((prev) => ({ ...prev, auto_review: checked }))}
              />
            </div>
          </div>
        </section>
      </div>
    </div>
  );
}
