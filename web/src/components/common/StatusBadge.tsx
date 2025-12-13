import { Badge } from '@/components/ui';
import type { ReviewStatus, Severity } from '@/types';

interface StatusBadgeProps {
  status: ReviewStatus;
}

const statusConfig: Record<ReviewStatus, { label: string; variant: 'success' | 'warning' | 'error' | 'secondary' }> = {
  pending: { label: '等待中', variant: 'secondary' },
  running: { label: '运行中', variant: 'warning' },
  completed: { label: '已完成', variant: 'success' },
  failed: { label: '失败', variant: 'error' },
  skipped: { label: '已跳过', variant: 'secondary' },
};

export function StatusBadge({ status }: StatusBadgeProps) {
  const config = statusConfig[status] || { label: status, variant: 'secondary' as const };
  return <Badge variant={config.variant}>{config.label}</Badge>;
}

interface SeverityBadgeProps {
  severity: Severity;
}

const severityConfig: Record<Severity, { label: string; variant: 'error' | 'warning' | 'success' }> = {
  P0: { label: 'P0', variant: 'error' },
  P1: { label: 'P1', variant: 'warning' },
  P2: { label: 'P2', variant: 'success' },
};

export function SeverityBadge({ severity }: SeverityBadgeProps) {
  const config = severityConfig[severity] || { label: severity, variant: 'secondary' as const };
  return <Badge variant={config.variant}>{config.label}</Badge>;
}
