import type { InputHTMLAttributes } from 'react';
import type { LucideIcon } from 'lucide-react';

// Поле со значком слева.
export function IconInput({ icon: Icon, ...props }: InputHTMLAttributes<HTMLInputElement> & { icon: LucideIcon }) {
  return (
    <div className="relative">
      <Icon className="w-4 h-4 absolute left-3.5 top-1/2 -translate-y-1/2 text-slate-400" />
      <input
        {...props}
        className="w-full bg-slate-50 border border-slate-200 pl-10 pr-3 py-2.5 rounded-lg text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-blue-600 focus:bg-white"
      />
    </div>
  );
}
