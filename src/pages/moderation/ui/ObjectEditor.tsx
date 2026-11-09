'use client';

import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { objectApi, objectKeys, PROTECTION_LABELS, useCategories, type ProtectionStatus } from '@/entities/heritage-object';
import { slugify } from '@/shared/lib';
import { ErrorMessage, Field, FormSelect, Input, Textarea } from '@/shared/ui';
import { toObjectInput, type ObjectForm } from '../model/objectForm';

const STATUSES = Object.keys(PROTECTION_LABELS) as ProtectionStatus[];

interface ObjectEditorProps {
  initial: ObjectForm;
  objectId?: number; // нет — создаётся новый объект
  onSaved: () => void;
}

// Паспорт объекта: создание (POST /objects) и правка (PUT /objects/{id}) — роль модератора.
export function ObjectEditor({ initial, objectId, onSaved }: ObjectEditorProps) {
  const qc = useQueryClient();
  const categories = useCategories();
  const [form, setForm] = useState(initial);
  const set = (patch: Partial<ObjectForm>) => setForm((f) => ({ ...f, ...patch }));

  const save = useMutation({
    mutationFn: () => {
      const input = toObjectInput(form);
      return objectId ? objectApi.update(objectId, input) : objectApi.create(input);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: objectKeys.all });
      onSaved();
    },
  });

  return (
    <form
      className="space-y-3"
      onSubmit={(e) => {
        e.preventDefault();
        save.mutate();
      }}
    >
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <Field label="Название" required>
          <Input
            required
            value={form.title}
            onChange={(e) => set({ title: e.target.value, ...(objectId ? {} : { slug: slugify(e.target.value) }) })}
          />
        </Field>
        <Field label="Адрес страницы (slug)" required>
          <Input required pattern="[a-z0-9\-]+" value={form.slug} onChange={(e) => set({ slug: e.target.value })} />
        </Field>
        <Field label="Категория">
          <FormSelect value={form.category} onChange={(e) => set({ category: e.target.value })}>
            {(categories.data ?? []).map((c) => (
              <option key={c.slug} value={c.slug}>
                {c.name}
              </option>
            ))}
          </FormSelect>
        </Field>
        <Field label="Статус охраны">
          <FormSelect value={form.protectionStatus} onChange={(e) => set({ protectionStatus: e.target.value })}>
            {STATUSES.map((s) => (
              <option key={s} value={s}>
                {PROTECTION_LABELS[s].long}
              </option>
            ))}
          </FormSelect>
        </Field>
        <Field label="Номер в реестре ОКН">
          <Input value={form.registryNumber} onChange={(e) => set({ registryNumber: e.target.value })} />
        </Field>
        <Field label="Год постройки">
          <Input value={form.yearBuilt} onChange={(e) => set({ yearBuilt: e.target.value })} />
        </Field>
        <Field label="Архитектор">
          <Input value={form.architect} onChange={(e) => set({ architect: e.target.value })} />
        </Field>
        <Field label="Архитектурный стиль">
          <Input value={form.architecturalStyle} onChange={(e) => set({ architecturalStyle: e.target.value })} />
        </Field>
        <Field label="Адрес" required>
          <Input required value={form.address} onChange={(e) => set({ address: e.target.value })} />
        </Field>
        <Field label="Координаты (широта, долгота)" required>
          <Input required placeholder="46.3538, 48.0285" value={form.coordinates} onChange={(e) => set({ coordinates: e.target.value })} />
        </Field>
      </div>
      <Field label="Краткое описание">
        <Textarea rows={2} value={form.shortDescription} onChange={(e) => set({ shortDescription: e.target.value })} />
      </Field>
      <Field label="Историческая справка">
        <Textarea rows={5} value={form.fullHistory} onChange={(e) => set({ fullHistory: e.target.value })} />
      </Field>
      <Field label="Теги через запятую">
        <Input value={form.tags} onChange={(e) => set({ tags: e.target.value })} />
      </Field>
      <label className="flex items-center gap-2 text-sm">
        <input type="checkbox" checked={form.isPublished} onChange={(e) => set({ isPublished: e.target.checked })} />
        <span>Опубликован (виден всем на сайте)</span>
      </label>
      {save.error && <ErrorMessage error={save.error} />}
      <button
        type="submit"
        disabled={save.isPending}
        className="px-5 py-2.5 bg-slate-900 hover:bg-slate-800 text-white rounded-lg text-xs font-bold uppercase tracking-wider"
      >
        {save.isPending ? 'Сохранение...' : objectId ? 'Сохранить' : 'Создать объект'}
      </button>
    </form>
  );
}
