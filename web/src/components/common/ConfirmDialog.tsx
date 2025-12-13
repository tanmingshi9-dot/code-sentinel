import { AlertTriangle } from 'lucide-react';
import { Dialog, DialogHeader, DialogContent, DialogFooter, Button } from '@/components/ui';

interface ConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  confirmText?: string;
  cancelText?: string;
  variant?: 'default' | 'destructive';
  onConfirm: () => void;
  loading?: boolean;
}

export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmText = '确认',
  cancelText = '取消',
  variant = 'default',
  onConfirm,
  loading,
}: ConfirmDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <div className="w-[400px]">
        <DialogHeader onClose={() => onOpenChange(false)}>{title}</DialogHeader>
        <DialogContent>
          <div className="flex gap-4">
            {variant === 'destructive' && (
              <div className="flex-shrink-0">
                <AlertTriangle className="w-6 h-6 text-red-500" />
              </div>
            )}
            <p className="text-sm text-gray-600">{description}</p>
          </div>
        </DialogContent>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {cancelText}
          </Button>
          <Button
            variant={variant === 'destructive' ? 'destructive' : 'default'}
            onClick={onConfirm}
            disabled={loading}
          >
            {loading ? '处理中...' : confirmText}
          </Button>
        </DialogFooter>
      </div>
    </Dialog>
  );
}
