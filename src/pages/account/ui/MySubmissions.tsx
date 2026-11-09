'use client';

import Link from 'next/link';
import { useQuery } from '@tanstack/react-query';
import { SUBMISSION_KIND_LABELS, SUBMISSION_STATUS_LABELS, submissionApi, submissionKeys } from '@/entities/submission';
import { formatDateTime } from '@/shared/lib';
import { QueryState } from '@/shared/ui';
import { StatusBadge } from './StatusBadge';

const TONE = { pending: 'amber', approved: 'green', rejected: 'red' } as const;

export function MySubmissions() {
  const query = useQuery({ queryKey: submissionKeys.mine, queryFn: submissionApi.mine });
  return (
    <QueryState query={query}>
      {(list) =>
        list.length === 0 ? (
          <div className="py-12 text-center text-sm text-slate-500 space-y-3">
            <p>Вы ещё не отправляли заявок.</p>
            <Link href="/contribute" className="text-blue-600 font-semibold">
              Предложить объект →
            </Link>
          </div>
        ) : (
          <ul className="space-y-3">
            {list.map((s) => (
              <li key={s.id} className="p-4 bg-white border border-slate-200 rounded-xl space-y-1.5">
                <div className="flex flex-wrap items-center gap-2">
                  <StatusBadge tone={TONE[s.status]}>{SUBMISSION_STATUS_LABELS[s.status]}</StatusBadge>
                  <span className="text-xs text-slate-500">
                    {SUBMISSION_KIND_LABELS[s.kind]}
                    {s.objectTitle && ` к «${s.objectTitle}»`} · № {s.id} · {formatDateTime(s.createdAt)}
                  </span>
                </div>
                <h3 className="text-sm font-bold text-slate-900">{s.title}</h3>
                {s.decision?.reason && <p className="text-xs text-rose-700">Причина отказа: {s.decision.reason}</p>}
              </li>
            ))}
          </ul>
        )
      }
    </QueryState>
  );
}
