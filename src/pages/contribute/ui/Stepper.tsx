import { Check } from 'lucide-react';
import { cn } from '@/shared/lib';
import { STEPS } from '../model/draft';

// Шаги заявки: текущий выделен, пройденные отмечены галочкой.
export function Stepper({ step }: { step: number }) {
  return (
    <ol className="grid grid-cols-3 gap-2 mb-8 text-xs font-semibold">
      {STEPS.map((label, i) => {
        const n = i + 1;
        return (
          <li
            key={label}
            aria-current={n === step ? 'step' : undefined}
            className={cn(
              'p-3 rounded-lg border text-center transition-colors flex items-center justify-center gap-1.5',
              n === step && 'bg-blue-50 border-blue-600 text-blue-700 font-bold',
              n < step && 'bg-white border-emerald-200 text-emerald-700',
              n > step && 'bg-white border-slate-200 text-slate-500',
            )}
          >
            {n < step && <Check className="w-3.5 h-3.5" />}
            <span>
              {n}. {label}
            </span>
          </li>
        );
      })}
    </ol>
  );
}
