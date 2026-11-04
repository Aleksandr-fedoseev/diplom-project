import type { ReactNode } from 'react';
import { cn } from '@/shared/lib';

export interface TabItem<T extends string> {
  id: T;
  label: ReactNode;
  icon?: ReactNode;
}

interface TabsProps<T extends string> {
  items: TabItem<T>[];
  value: T;
  onChange: (id: T) => void;
  size?: 'sm' | 'md';
  className?: string;
}

// Вкладки с подчёркиванием активной.
export function Tabs<T extends string>({ items, value, onChange, size = 'md', className }: TabsProps<T>) {
  return (
    <div role="tablist" className={cn('flex items-center border-b border-slate-200 overflow-x-auto', size === 'md' ? 'gap-2' : 'text-xs font-semibold', className)}>
      {items.map((tab) => (
        <button
          key={tab.id}
          type="button"
          role="tab"
          aria-selected={value === tab.id}
          onClick={() => onChange(tab.id)}
          className={cn(
            'border-b-2 whitespace-nowrap transition-colors flex items-center',
            size === 'md' ? 'px-4 py-3 text-sm font-semibold gap-2' : 'py-3 px-3 gap-1.5',
            value === tab.id
              ? cn('border-blue-600 text-blue-600', size === 'sm' && 'font-bold')
              : 'border-transparent text-slate-600 hover:text-slate-900',
          )}
        >
          {tab.icon}
          {tab.label}
        </button>
      ))}
    </div>
  );
}
