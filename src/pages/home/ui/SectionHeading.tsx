import type { ReactNode } from 'react';

export function SectionHeading({ kicker, title, action }: { kicker: string; title: string; action?: ReactNode }) {
  return (
    <div className="flex items-end justify-between mb-8">
      <div>
        <span className="text-xs font-bold uppercase tracking-wider text-slate-400">{kicker}</span>
        <h2 className="text-2xl sm:text-3xl font-black text-slate-950 tracking-tight mt-1">{title}</h2>
      </div>
      {action}
    </div>
  );
}
