'use client';

import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { objectApi, objectKeys, useCategories, type CategoryDto } from '@/entities/heritage-object';
import { slugify } from '@/shared/lib';
import { ErrorMessage, Input, QueryState } from '@/shared/ui';

const button = 'px-3 py-1.5 rounded-lg text-xs font-bold disabled:opacity-50';

function CategoryRow({ category }: { category: CategoryDto }) {
  const qc = useQueryClient();
  const [name, setName] = useState(category.name);
  const [slug, setSlug] = useState(category.slug);
  const refresh = () => qc.invalidateQueries({ queryKey: objectKeys.all });
  const save = useMutation({ mutationFn: () => objectApi.updateCategory(category.id, slug.trim(), name.trim()), onSuccess: refresh });
  const remove = useMutation({ mutationFn: () => objectApi.removeCategory(category.id), onSuccess: refresh });
  const changed = name.trim() !== category.name || slug.trim() !== category.slug;

  return (
    <li className="p-3 bg-white border border-slate-200 rounded-xl space-y-2">
      <div className="flex flex-col sm:flex-row gap-2">
        <Input aria-label="Название" value={name} onChange={(e) => setName(e.target.value)} />
        <Input aria-label="Код (slug)" className="sm:w-56 font-mono" value={slug} onChange={(e) => setSlug(e.target.value)} />
        <div className="flex gap-2 shrink-0">
          <button type="button" disabled={!changed || save.isPending} onClick={() => save.mutate()} className={`${button} bg-slate-900 text-white`}>
            Сохранить
          </button>
          <button
            type="button"
            disabled={remove.isPending}
            onClick={() => confirm(`Удалить категорию «${category.name}»?`) && remove.mutate()}
            className={`${button} bg-rose-100 text-rose-800`}
          >
            Удалить
          </button>
        </div>
      </div>
      {(save.error || remove.error) && <ErrorMessage error={save.error ?? remove.error} />}
    </li>
  );
}

// Управление категориями объектов: создать, переименовать, удалить. Категорию, в которой
// есть объекты, сервер удалить не даст — сначала объекты нужно перенести.
export function CategoriesAdmin() {
  const qc = useQueryClient();
  const categories = useCategories();
  const [name, setName] = useState('');
  const create = useMutation({
    mutationFn: () => objectApi.createCategory(slugify(name), name.trim()),
    onSuccess: () => {
      setName('');
      qc.invalidateQueries({ queryKey: objectKeys.all });
    },
  });

  return (
    <div className="space-y-4">
      <form
        className="flex flex-col sm:flex-row gap-2"
        onSubmit={(e) => {
          e.preventDefault();
          create.mutate();
        }}
      >
        <Input required placeholder="Новая категория, например «Промышленная архитектура»" value={name} onChange={(e) => setName(e.target.value)} />
        <button type="submit" disabled={create.isPending} className="px-4 py-2.5 bg-slate-900 text-white rounded-xl text-xs font-bold uppercase tracking-wider shrink-0">
          Добавить
        </button>
      </form>
      {create.error && <ErrorMessage error={create.error} />}
      <QueryState query={categories}>
        {(list) => (
          <ul className="space-y-2">
            {list.map((c) => (
              <CategoryRow key={`${c.id}-${c.slug}-${c.name}`} category={c} />
            ))}
          </ul>
        )}
      </QueryState>
    </div>
  );
}
