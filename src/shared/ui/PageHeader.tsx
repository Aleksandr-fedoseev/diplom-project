import type { ReactNode } from 'react';
import { cn } from '@/shared/lib';

interface PageHeaderProps {
  kicker: string; // надзаголовок: раздел сайта
  title: ReactNode;
  lead?: ReactNode;
  className?: string;
}

export function PageHeader({ kicker, title, lead, className }: PageHeaderProps) {
  return (
    <div className={cn('mb-8', className)}>
      <div className="text-xs font-bold uppercase tracking-wider text-blue-600 mb-1">{kicker}</div>
      <h1 className="text-3xl sm:text-4xl font-bold text-slate-900 tracking-tight">{title}</h1>
      {lead && <p className="text-base text-slate-600 mt-2 max-w-2xl">{lead}</p>}
    </div>
  );
}
