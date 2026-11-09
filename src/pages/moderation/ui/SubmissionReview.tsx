'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { objectKeys, PROTECTION_LABELS, useCategories, type ProtectionStatus } from '@/entities/heritage-object';
import { SUBMISSION_KIND_LABELS, submissionApi, submissionKeys, type Submission } from '@/entities/submission';
import { userKeys } from '@/entities/user';
import { formatDateTime, parseCoordinates, slugify } from '@/shared/lib';
import { AuthImage, ErrorMessage, Field, FormSelect, Input, QueryState, Textarea } from '@/shared/ui';

const STATUSES = Object.keys(PROTECTION_LABELS) as ProtectionStatus[];
const button = 'px-4 py-2 rounded-lg text-xs font-bold uppercase tracking-wider disabled:opacity-50';

function ApproveForm({ sub, onDone }: { sub: Submission; onDone: () => void }) {
  const qc = useQueryClient();
  const categories = useCategories();
  const isNew = sub.kind === 'new_object';
  const [slug, setSlug] = useState(slugify(sub.title));
  const [category, setCategory] = useState('manor');
  const [status, setStatus] = useState<ProtectionStatus>('identified');
  const [coords, setCoords] = useState(sub.coordinates ? `${sub.coordinates.lat}, ${sub.coordinates.lng}` : '');
  const [publish, setPublish] = useState(true);

  const approve = useMutation({
    mutationFn: () => {
      if (!isNew) return submissionApi.approve(sub.id);
      const coordinates = parseCoordinates(coords);
      if (!coordinates) throw new Error('Без координат объект не поставить на карту: укажите широту и долготу');
      return submissionApi.approve(sub.id, { slug, category, protectionStatus: status, coordinates, isPublished: publish });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['submissions'] });
      qc.invalidateQueries({ queryKey: objectKeys.all }); // новый объект или новые фото в каталоге
      onDone();
    },
  });

  return (
    <form
      className="space-y-3 p-4 bg-emerald-50 border border-emerald-200 rounded-xl"
      onSubmit={(e) => {
        e.preventDefault();
        approve.mutate();
      }}
    >
      {isNew ? (
        <>
          <p className="text-xs text-emerald-900">
            Будет создан объект каталога. Название, адрес, история и год берутся из заявки — потом их можно поправить во
            вкладке «Объекты».
          </p>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <Field label="Адрес страницы (slug)" required>
              <Input required pattern="[a-z0-9\-]+" value={slug} onChange={(e) => setSlug(e.target.value)} />
            </Field>
            <Field label="Координаты" required>
              <Input required placeholder="46.3538, 48.0285" value={coords} onChange={(e) => setCoords(e.target.value)} />
            </Field>
            <Field label="Категория">
              <FormSelect value={category} onChange={(e) => setCategory(e.target.value)}>
                {(categories.data ?? []).map((c) => (
                  <option key={c.slug} value={c.slug}>
                    {c.name}
                  </option>
                ))}
              </FormSelect>
            </Field>
            <Field label="Статус охраны">
              <FormSelect value={status} onChange={(e) => setStatus(e.target.value as ProtectionStatus)}>
                {STATUSES.map((s) => (
                  <option key={s} value={s}>
                    {PROTECTION_LABELS[s].long}
                  </option>
                ))}
              </FormSelect>
            </Field>
          </div>
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={publish} onChange={(e) => setPublish(e.target.checked)} />
            <span>Сразу опубликовать на сайте</span>
          </label>
        </>
      ) : (
        <p className="text-xs text-emerald-900">Файлы заявки будут добавлены к объекту «{sub.objectTitle}».</p>
      )}
      {approve.error && <ErrorMessage error={approve.error} />}
      <button type="submit" disabled={approve.isPending} className={`${button} bg-emerald-600 hover:bg-emerald-700 text-white`}>
        {approve.isPending ? 'Одобряем...' : 'Одобрить'}
      </button>
    </form>
  );
}

function RejectForm({ sub, onDone }: { sub: Submission; onDone: () => void }) {
  const qc = useQueryClient();
  const [reason, setReason] = useState('');
  const [block, setBlock] = useState(false);
  const reject = useMutation({
    mutationFn: () => submissionApi.reject(sub.id, reason.trim(), block),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['submissions'] });
      if (block) qc.invalidateQueries({ queryKey: userKeys.list });
      onDone();
    },
  });
  return (
    <form
      className="space-y-3 p-4 bg-rose-50 border border-rose-200 rounded-xl"
      onSubmit={(e) => {
        e.preventDefault();
        reject.mutate();
      }}
    >
      <Field label="Причина отказа (придёт автору письмом)" required>
        <Textarea required rows={3} value={reason} onChange={(e) => setReason(e.target.value)} />
      </Field>
      <label className="flex items-center gap-2 text-sm text-rose-900">
        <input type="checkbox" checked={block} onChange={(e) => setBlock(e.target.checked)} />
        <span>Заблокировать автора (спам, нарушение правил)</span>
      </label>
      {reject.error && <ErrorMessage error={reject.error} />}
      <button type="submit" disabled={reject.isPending} className={`${button} bg-rose-600 hover:bg-rose-700 text-white`}>
        {reject.isPending ? 'Отклоняем...' : 'Отклонить'}
      </button>
    </form>
  );
}

// Карточка заявки для модератора: автор, текст, файлы и решение.
export function SubmissionReview({ id, onClose }: { id: number; onClose: () => void }) {
  const query = useQuery({ queryKey: submissionKeys.one(id), queryFn: () => submissionApi.get(id) });
  const [action, setAction] = useState<'approve' | 'reject' | null>(null);

  return (
    <div className="p-5 bg-white border border-blue-200 rounded-xl shadow-sm space-y-4">
      <QueryState query={query}>
        {(sub) => (
          <>
            <div className="flex items-start justify-between gap-3">
              <div>
                <div className="text-xs text-slate-500">
                  {SUBMISSION_KIND_LABELS[sub.kind]} № {sub.id} · {formatDateTime(sub.createdAt)}
                </div>
                <h3 className="text-lg font-bold text-slate-900">{sub.title}</h3>
                {sub.author && (
                  <div className="text-xs text-slate-600">
                    Автор: {sub.author.name} ({sub.author.email}){sub.author.isBlocked && <b className="text-rose-700"> — заблокирован</b>}
                  </div>
                )}
              </div>
              <button type="button" onClick={onClose} className="text-xs font-bold text-slate-500 hover:text-slate-900">
                Свернуть
              </button>
            </div>

            <dl className="grid grid-cols-1 sm:grid-cols-2 gap-2 text-xs">
              {sub.objectTitle && <Fact label="Объект" value={sub.objectTitle} />}
              {sub.address && <Fact label="Адрес" value={sub.address} />}
              {sub.coordinates && <Fact label="Координаты" value={`${sub.coordinates.lat}, ${sub.coordinates.lng}`} />}
              {sub.period && <Fact label="Время постройки" value={sub.period} />}
              {sub.source && <Fact label="Источник" value={sub.source} />}
            </dl>
            <p className="text-sm text-slate-700 whitespace-pre-line">{sub.description}</p>

            {(sub.media?.length ?? 0) > 0 && (
              <div className="grid grid-cols-2 sm:grid-cols-3 gap-2">
                {sub.media!.map((m) => (
                  <figure key={m.id} className="space-y-1">
                    <AuthImage src={m.url} alt={m.originalName} className="w-full h-32 rounded-lg" />
                    <figcaption className="text-[11px] text-slate-500 truncate">
                      {m.kind === 'archive_photo' ? `Архив${m.yearTaken ? `, ${m.yearTaken}` : ''}` : 'Фото'} · {m.originalName}
                    </figcaption>
                  </figure>
                ))}
              </div>
            )}

            {sub.decision && (
              <p className="text-xs text-slate-600">
                Решение: {sub.decision.decision === 'approved' ? 'одобрена' : `отклонена — ${sub.decision.reason}`} (
                {formatDateTime(sub.decision.createdAt)})
              </p>
            )}

            {sub.status === 'pending' && (
              <div className="space-y-3">
                <div className="flex gap-2">
                  <button type="button" onClick={() => setAction('approve')} className={`${button} bg-emerald-100 text-emerald-900`}>
                    Одобрить…
                  </button>
                  <button type="button" onClick={() => setAction('reject')} className={`${button} bg-rose-100 text-rose-900`}>
                    Отклонить…
                  </button>
                </div>
                {action === 'approve' && <ApproveForm sub={sub} onDone={onClose} />}
                {action === 'reject' && <RejectForm sub={sub} onDone={onClose} />}
              </div>
            )}
          </>
        )}
      </QueryState>
    </div>
  );
}

function Fact({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-slate-400">{label}</dt>
      <dd className="font-semibold text-slate-800">{value}</dd>
    </div>
  );
}
