'use client';

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  SUBMISSION_KIND_LABELS,
  SUBMISSION_STATUS_LABELS,
  submissionApi,
  submissionKeys,
  type SubmissionStatus,
} from '@/entities/submission';
import { formatDateTime } from '@/shared/lib';
import { Chip, QueryState } from '@/shared/ui';
import { SubmissionReview } from './SubmissionReview';

const STATUSES: SubmissionStatus[] = ['pending', 'approved', 'rejected'];

// Очередь заявок: ожидающие решения — старые первыми (так их отдаёт сервер).
export function SubmissionQueue() {
  const [status, setStatus] = useState<SubmissionStatus>('pending');
  const [open, setOpen] = useState<number | null>(null);
  const query = useQuery({ queryKey: submissionKeys.queue(status), queryFn: () => submissionApi.queue(status) });

  return (
    <div className="space-y-4">
      <div className="flex gap-2">
        {STATUSES.map((s) => (
          <Chip key={s} active={status === s} onClick={() => setStatus(s)} className="px-3 py-1.5 text-xs">
            {SUBMISSION_STATUS_LABELS[s]}
          </Chip>
        ))}
      </div>
      <QueryState query={query}>
        {(list) =>
          list.length === 0 ? (
            <p className="py-12 text-center text-sm text-slate-500">
              {status === 'pending' ? 'Новых заявок нет.' : 'Здесь пока пусто.'}
            </p>
          ) : (
            <ul className="space-y-2">
              {list.map((s) =>
                open === s.id ? (
                  <li key={s.id}>
                    <SubmissionReview id={s.id} onClose={() => setOpen(null)} />
                  </li>
                ) : (
                  <li key={s.id}>
                    <button
                      type="button"
                      onClick={() => setOpen(s.id)}
                      className="w-full text-left p-4 bg-white border border-slate-200 rounded-xl hover:border-slate-400"
                    >
                      <div className="text-xs text-slate-500">
                        {SUBMISSION_KIND_LABELS[s.kind]}
                        {s.objectTitle && ` к «${s.objectTitle}»`} · № {s.id} · {formatDateTime(s.createdAt)}
                      </div>
                      <div className="text-sm font-bold text-slate-900">{s.title}</div>
                    </button>
                  </li>
                ),
              )}
            </ul>
          )
        }
      </QueryState>
    </div>
  );
}
