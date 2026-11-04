import React from 'react';
import { cn } from '@/shared/lib';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  // primary — тёмная основная кнопка во всю ширину формы; secondary — компактная тёмная.
  variant?: 'primary' | 'secondary';
}

const variants = {
  primary:
    'w-full py-4 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-black uppercase tracking-wider transition-colors shadow-sm',
  secondary:
    'px-6 py-3 bg-slate-950 hover:bg-slate-800 text-white rounded-xl text-xs font-bold uppercase tracking-wider transition-colors',
};

export function Button({ variant = 'primary', type = 'button', className, ...props }: ButtonProps) {
  return <button type={type} {...props} className={cn(variants[variant], className)} />;
}

interface ChipProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  active: boolean;
  // filled — серый фон (вкладки, фильтры); outline — белый с рамкой (варианты на выбор).
  variant?: 'filled' | 'outline';
}

const chipStyles = {
  filled: { on: 'bg-slate-950 text-white shadow-xs', off: 'bg-slate-100 text-slate-700 hover:bg-slate-200' },
  outline: {
    on: 'border bg-slate-950 text-white border-slate-950 shadow-xs',
    off: 'border bg-white text-slate-800 border-slate-200 hover:border-slate-400',
  },
};

// Кнопка-переключатель: вкладки, выбор суммы, фильтры. Выбранная — тёмная.
export function Chip({ active, variant = 'filled', className, ...props }: ChipProps) {
  const style = chipStyles[variant];
  return (
    <button
      type="button"
      aria-pressed={active}
      {...props}
      className={cn('rounded-xl font-bold transition-all', active ? style.on : style.off, className)}
    />
  );
}
