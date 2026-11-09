import type { ReactNode } from 'react';
import { AlertTriangle } from 'lucide-react';
import { errorMessage } from '@/shared/api';
import { cn } from '@/shared/lib';

export function Spinner({ className }: { className?: string }) {
  return (
    <div
      role="status"
      aria-label="Загрузка"
      className={cn('w-8 h-8 border-3 border-blue-600 border-t-transparent rounded-full animate-spin', className)}
    />
  );
}

export function Loading({ label = 'Загрузка...', className }: { label?: string; className?: string }) {
  return (
    <div className={cn('py-16 flex flex-col items-center justify-center gap-3 text-slate-500 text-sm', className)}>
      <Spinner />
      <span>{label}</span>
    </div>
  );
}

export function ErrorMessage({ error, onRetry, className }: { error: unknown; onRetry?: () => void; className?: string }) {
  return (
    <div role="alert" className={cn('p-4 rounded-xl border border-rose-200 bg-rose-50 text-sm text-rose-900 flex items-start gap-3', className)}>
      <AlertTriangle className="w-5 h-5 text-rose-600 shrink-0" />
      <div className="space-y-2">
        <p>{errorMessage(error)}</p>
        {onRetry && (
          <button type="button" onClick={onRetry} className="text-xs font-bold uppercase tracking-wider text-rose-700 hover:text-rose-900">
            Повторить
          </button>
        )}
      </div>
    </div>
  );
}

interface QueryStateProps<T> {
  query: { data: T | undefined; error: unknown; isPending: boolean; refetch: () => unknown };
  children: (data: T) => ReactNode;
  loadingLabel?: string;
}

// Загрузка → ошибка с кнопкой «Повторить» → данные. Одинаково на всех страницах.
export function QueryState<T>({ query, children, loadingLabel }: QueryStateProps<T>) {
  if (query.error) return <ErrorMessage error={query.error} onRetry={() => query.refetch()} />;
  if (query.isPending || query.data === undefined) return <Loading label={loadingLabel} />;
  return <>{children(query.data)}</>;
}
