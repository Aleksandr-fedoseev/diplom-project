import { cn } from '@/shared/lib';

const tones = {
  gray: 'bg-slate-100 text-slate-700',
  amber: 'bg-amber-100 text-amber-800',
  green: 'bg-emerald-100 text-emerald-800',
  red: 'bg-rose-100 text-rose-800',
  blue: 'bg-blue-100 text-blue-800',
};

export function StatusBadge({ tone, children }: { tone: keyof typeof tones; children: React.ReactNode }) {
  return <span className={cn('px-2 py-0.5 rounded text-[11px] font-bold uppercase tracking-wider', tones[tone])}>{children}</span>;
}
