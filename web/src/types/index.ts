// 仓库相关类型
export interface Repo {
  id: number;
  full_name: string;
  owner: string;
  name: string;
  enabled: boolean;
  config: string | null;
  last_review_at: string | null;
  review_count: number;
  created_at: string;
  updated_at: string;
}

export interface RepoConfig {
  llm_provider: LLMProvider;
  model: string;
  max_tokens: number;
  system_prompt: string;
  review_focus: ReviewFocus[];
  min_severity: Severity;
  languages: string[];
  ignore_files: string[];
  max_diff_lines: number;
  auto_review: boolean;

  // 仓库级 LLM 配置（可选，覆盖全局配置）
  llm_api_key?: string;
  llm_base_url?: string;

  // 仓库级 GitHub 配置（可选，覆盖全局配置）
  github_token?: string;
}

export type LLMProvider = 'openai' | 'qwen' | 'azure' | 'ollama';
export type Severity = 'P0' | 'P1' | 'P2';
export type ReviewFocus = 'security' | 'performance' | 'logic' | 'style';
export type ReviewStatus = 'pending' | 'running' | 'completed' | 'failed' | 'skipped';

// 审查相关类型
export interface Review {
  id: number;
  repo_id: number;
  repo_full_name: string;
  pr_number: number;
  pr_title: string;
  pr_author: string;
  commit_sha: string;
  status: ReviewStatus;
  result: string;
  token_used: number;
  duration_ms: number;
  error_msg?: string;
  created_at: string;
}

export interface ReviewResult {
  summary: string;
  issues: ReviewIssue[];
  stats: ReviewStats;
  score?: number;
  model?: string;
  duration_ms?: number;
}

export interface ReviewIssue {
  severity: Severity;
  category: string;
  file: string;
  line: number;
  title: string;
  description: string;
  suggestion?: string;
  code_fix?: string;
}

export interface ReviewStats {
  p0_count: number;
  p1_count: number;
  p2_count: number;
}


// 反馈相关类型
export interface Feedback {
  id: number;
  review_id: number;
  repo_full_name: string;
  pr_number: number;
  file: string;
  line: number;
  issue_index: number;
  severity: Severity;
  category: string;
  title: string;
  ai_content: string;
  is_false_positive: boolean;
  reason: string;
  reporter: string;
  created_at: string;
}

export interface FeedbackStats {
  total: number;
  by_category: Record<string, number>;
  by_severity: Record<string, number>;
}

// API 响应类型
export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface PaginatedData<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

// 配置模板
export interface ConfigTemplate {
  name: string;
  description: string;
  config: RepoConfig;
}
