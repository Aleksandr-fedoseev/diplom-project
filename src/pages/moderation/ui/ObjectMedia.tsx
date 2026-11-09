'use client';

import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { objectApi, objectKeys, type HeritageObject } from '@/entities/heritage-object';
import { ErrorMessage, Field, FileDrop, FormSelect, Input } from '@/shared/ui';

type Kind = 'photo' | 'archive_photo' | 'panorama' | 'model3d';

const KINDS: { id: Kind; label: string; accept: string; hint: string }[] = [
  { id: 'photo', label: 'Современное фото', accept: 'image/jpeg,image/png,image/webp', hint: 'JPG, PNG, WebP до 15 МБ' },
  { id: 'archive_photo', label: 'Архивный снимок', accept: 'image/jpeg,image/png,image/webp', hint: 'JPG, PNG, WebP до 15 МБ' },
  { id: 'panorama', label: 'Панорама 360°', accept: 'image/jpeg,image/webp', hint: 'JPG, WebP до 30 МБ, проекция 2:1' },
  { id: 'model3d', label: '3D-модель', accept: '.glb,model/gltf-binary', hint: 'GLB до 50 МБ' },
];

const KIND_LABEL = Object.fromEntries(KINDS.map((k) => [k.id, k.label])) as Record<string, string>;

// Файлы объекта: список с удалением и загрузка нового (POST /objects/{id}/media).
export function ObjectMedia({ object }: { object: HeritageObject }) {
  const qc = useQueryClient();
  const [kind, setKind] = useState<Kind>('photo');
  const [files, setFiles] = useState<File[]>([]);
  const [caption, setCaption] = useState('');
  const [yearTaken, setYearTaken] = useState('');
  const [ai, setAi] = useState(false);
  const [partial, setPartial] = useState(false);
  const [hAngle, setHAngle] = useState('180');
  const [author, setAuthor] = useState('');
  const [license, setLicense] = useState('');
  const refresh = () => qc.invalidateQueries({ queryKey: objectKeys.all });

  const upload = useMutation({
    mutationFn: () => {
      const form = new FormData();
      form.set('file', files[0]);
      form.set('kind', kind);
      form.set('caption', caption.trim());
      form.set('source', ai ? 'ai_reconstruction' : 'photo');
      if (author.trim()) form.set('author', author.trim());
      if (license.trim()) form.set('license', license.trim());
      if (kind === 'archive_photo' && yearTaken.trim()) form.set('yearTaken', yearTaken.trim());
      if (kind === 'panorama') {
        form.set('projection', partial ? 'partial' : 'full_360');
        if (partial) form.set('hAngle', hAngle);
      }
      return objectApi.uploadMedia(object.id, form);
    },
    onSuccess: () => {
      setFiles([]);
      setCaption('');
      refresh();
    },
  });
  const remove = useMutation({ mutationFn: objectApi.removeMedia, onSuccess: refresh });
  const current = KINDS.find((k) => k.id === kind)!;

  return (
    <div className="space-y-4">
      {(object.files?.length ?? 0) > 0 ? (
        <ul className="grid grid-cols-2 sm:grid-cols-3 gap-2">
          {object.files!.map((f) => (
            <li key={f.id} className="border border-slate-200 rounded-lg overflow-hidden text-[11px]">
              {f.kind === 'model3d' ? (
                <div className="h-24 bg-slate-100 flex items-center justify-center text-slate-500">3D</div>
              ) : (
                // eslint-disable-next-line @next/next/no-img-element
                <img src={f.url} alt={f.caption || f.originalName} className="w-full h-24 object-cover" />
              )}
              <div className="p-2 space-y-1">
                <div className="font-semibold truncate">
                  {KIND_LABEL[f.kind] ?? f.kind}
                  {f.source === 'ai_reconstruction' && ' · ИИ'}
                </div>
                <div className="text-slate-500 truncate">{f.caption || f.originalName}</div>
                <button
                  type="button"
                  onClick={() => confirm(`Удалить файл «${f.originalName}»?`) && remove.mutate(f.id)}
                  className="font-bold text-rose-700 hover:text-rose-900"
                >
                  Удалить
                </button>
              </div>
            </li>
          ))}
        </ul>
      ) : (
        <p className="text-xs text-slate-500">Файлов пока нет.</p>
      )}
      {remove.error && <ErrorMessage error={remove.error} />}

      <form
        className="p-4 bg-slate-50 border border-slate-200 rounded-xl space-y-3"
        onSubmit={(e) => {
          e.preventDefault();
          if (files.length) upload.mutate();
        }}
      >
        <div className="text-xs font-bold uppercase tracking-wider text-slate-700">Добавить файл</div>
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <Field label="Вид">
            <FormSelect value={kind} onChange={(e) => setKind(e.target.value as Kind)}>
              {KINDS.map((k) => (
                <option key={k.id} value={k.id}>
                  {k.label}
                </option>
              ))}
            </FormSelect>
          </Field>
          <Field label="Подпись">
            <Input value={caption} onChange={(e) => setCaption(e.target.value)} />
          </Field>
          {kind === 'archive_photo' && (
            <Field label="Год съёмки">
              <Input value={yearTaken} onChange={(e) => setYearTaken(e.target.value)} />
            </Field>
          )}
          <Field label="Автор снимка">
            <Input value={author} onChange={(e) => setAuthor(e.target.value)} />
          </Field>
          <Field label="Лицензия или разрешение">
            <Input placeholder="Например: с разрешения автора" value={license} onChange={(e) => setLicense(e.target.value)} />
          </Field>
        </div>
        {kind !== 'model3d' && (
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={ai} onChange={(e) => setAi(e.target.checked)} />
            <span>Создано нейросетью (на сайте — пометка «реконструкция ИИ»)</span>
          </label>
        )}
        {kind === 'panorama' && (
          <div className="flex flex-wrap items-center gap-3 text-sm">
            <label className="flex items-center gap-2">
              <input type="checkbox" checked={partial} onChange={(e) => setPartial(e.target.checked)} />
              <span>Неполный круг</span>
            </label>
            {partial && (
              <label className="flex items-center gap-2">
                <span>Угол обзора, °</span>
                <Input type="number" min={30} max={359} className="w-24" value={hAngle} onChange={(e) => setHAngle(e.target.value)} />
              </label>
            )}
          </div>
        )}
        <FileDrop title="Выберите файл" hint={current.hint} accept={current.accept} files={files} onFiles={setFiles} />
        {upload.error && <ErrorMessage error={upload.error} />}
        <button
          type="submit"
          disabled={!files.length || upload.isPending}
          className="px-4 py-2 bg-slate-900 hover:bg-slate-800 disabled:bg-slate-300 text-white rounded-lg text-xs font-bold uppercase tracking-wider"
        >
          {upload.isPending ? 'Загрузка...' : 'Загрузить'}
        </button>
      </form>
    </div>
  );
}
