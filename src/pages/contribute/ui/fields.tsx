import React from 'react';
import { cn } from '@/shared/lib';

// Поля формы заявки: светлый фон и синяя рамка фокуса, как в остальном каталоге.
const control =
  'w-full bg-slate-50 border border-slate-200 p-3 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white';

export function Label({ htmlFor, children }: { htmlFor: string; children: React.ReactNode }) {
  return (
    <label htmlFor={htmlFor} className="block text-sm font-semibold text-slate-900 mb-1.5">
      {children}
    </label>
  );
}

export const DraftInput = (props: React.InputHTMLAttributes<HTMLInputElement>) => (
  <input type="text" {...props} className={cn(control, props.className)} />
);
export const DraftTextarea = (props: React.TextareaHTMLAttributes<HTMLTextAreaElement>) => (
  <textarea {...props} className={cn(control, props.className)} />
);
export const DraftSelect = (props: React.SelectHTMLAttributes<HTMLSelectElement>) => (
  <select {...props} className={cn(control, props.className)} />
);

const navButton = 'px-6 py-2.5 rounded-lg text-sm font-semibold flex items-center gap-2';
export const nextClass = cn(navButton, 'bg-blue-600 hover:bg-blue-700 text-white');
export const backClass = cn(navButton, 'px-5 bg-slate-100 hover:bg-slate-200 text-slate-700');
export const submitClass = cn(navButton, 'bg-emerald-600 hover:bg-emerald-700 text-white');
