import React from 'react';
import { cn } from '@/shared/lib';

const control =
  'w-full bg-white border border-slate-200 p-3 rounded-xl text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-950';

interface FieldProps {
  label: string;
  required?: boolean;
  children: React.ReactElement<{ id?: string }>;
  className?: string;
}

// Подпись над полем. id связывает подпись с полем, чтобы клик по ней ставил курсор в поле.
export function Field({ label, required, children, className }: FieldProps) {
  const id = React.useId();
  return (
    <div className={className}>
      <label htmlFor={id} className="block text-xs font-bold uppercase tracking-wider text-slate-700 mb-1">
        {label}
        {required && '*'}
      </label>
      {React.cloneElement(children, { id })}
    </div>
  );
}

export function Input({ className, ...props }: React.InputHTMLAttributes<HTMLInputElement>) {
  return <input type="text" {...props} className={cn(control, className)} />;
}

export function Textarea({ className, ...props }: React.TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return <textarea rows={4} {...props} className={cn(control, className)} />;
}

// Выпадающий список в стиле полей формы (для фильтров — Select).
export function FormSelect({ className, ...props }: React.SelectHTMLAttributes<HTMLSelectElement>) {
  return <select {...props} className={cn(control, className)} />;
}
