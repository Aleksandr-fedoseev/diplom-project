import { cn } from '@/shared/lib';
import { PROTECTION_LABELS } from '../model/labels';
import type { ProtectionStatus } from '../model/types';

const tones: Record<ProtectionStatus, { text: string; pill: string; outlined: string }> = {
  federal: { text: 'text-emerald-700', pill: 'bg-emerald-50 text-emerald-800', outlined: 'bg-emerald-50 text-emerald-800 border border-emerald-200' },
  lost: { text: 'text-rose-700', pill: 'bg-rose-50 text-rose-800', outlined: 'bg-rose-50 text-rose-800 border border-rose-200' },
  regional: { text: 'text-slate-600', pill: 'bg-slate-100 text-slate-700', outlined: 'bg-slate-100 text-slate-800 border border-slate-200' },
  identified: { text: 'text-slate-600', pill: 'bg-slate-100 text-slate-700', outlined: 'bg-slate-100 text-slate-800 border border-slate-200' },
};

interface ProtectionBadgeProps {
  status: ProtectionStatus;
  length?: 'short' | 'medium' | 'long';
  // text — цветной текст; pill — плашка; outlined — плашка с рамкой.
  variant?: 'text' | 'pill' | 'outlined';
  className?: string;
}

export function ProtectionBadge({ status, length = 'medium', variant = 'pill', className }: ProtectionBadgeProps) {
  return (
    <span className={cn('font-semibold', variant !== 'text' && 'rounded', tones[status][variant], className)}>
      {PROTECTION_LABELS[status][length]}
    </span>
  );
}
