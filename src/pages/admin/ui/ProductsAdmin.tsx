'use client';

import { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { PRODUCT_CATEGORIES, productApi, productKeys, type ProductInput, type ProductItem } from '@/entities/product';
import { formatRub, slugify } from '@/shared/lib';
import { ErrorMessage, Field, FormSelect, Input, QueryState, Textarea } from '@/shared/ui';

const EMPTY: ProductInput = { slug: '', title: '', description: '', category: 'clothes', priceKopecks: 0, stock: 0, imageUrl: '', isActive: true };

const toInput = (p: ProductItem): ProductInput => ({
  slug: p.slug,
  title: p.title,
  description: p.description,
  category: p.category,
  priceKopecks: Math.round(p.price * 100),
  stock: p.stock,
  imageUrl: p.coverImage,
  isActive: p.isActive,
});

function ProductForm({ initial, id, onDone }: { initial: ProductInput; id?: number; onDone: () => void }) {
  const qc = useQueryClient();
  const [f, setF] = useState(initial);
  const [price, setPrice] = useState(initial.priceKopecks ? String(initial.priceKopecks / 100) : '');
  const set = (patch: Partial<ProductInput>) => setF((v) => ({ ...v, ...patch }));
  const save = useMutation({
    mutationFn: () => {
      const input = { ...f, priceKopecks: Math.round(Number(price.replace(',', '.')) * 100) };
      return id ? productApi.update(id, input) : productApi.create(input);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: productKeys.all });
      onDone();
    },
  });
  return (
    <form
      className="p-4 bg-slate-50 border border-slate-200 rounded-xl space-y-3"
      onSubmit={(e) => {
        e.preventDefault();
        save.mutate();
      }}
    >
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <Field label="Название" required>
          <Input required value={f.title} onChange={(e) => set({ title: e.target.value, ...(id ? {} : { slug: slugify(e.target.value) }) })} />
        </Field>
        <Field label="Адрес (slug)" required>
          <Input required value={f.slug} onChange={(e) => set({ slug: e.target.value })} />
        </Field>
        <Field label="Категория">
          <FormSelect value={f.category} onChange={(e) => set({ category: e.target.value })}>
            {PRODUCT_CATEGORIES.map((c) => (
              <option key={c.id} value={c.id}>
                {c.label}
              </option>
            ))}
          </FormSelect>
        </Field>
        <Field label="Цена, ₽" required>
          <Input required inputMode="decimal" pattern="[0-9]+([.,][0-9]{1,2})?" value={price} onChange={(e) => setPrice(e.target.value)} />
        </Field>
        <Field label="Остаток, шт.">
          <Input type="number" min={0} value={f.stock} onChange={(e) => set({ stock: Number(e.target.value) })} />
        </Field>
        <Field label="Картинка (адрес)">
          <Input placeholder="/merch/hoodie.jpg" value={f.imageUrl} onChange={(e) => set({ imageUrl: e.target.value })} />
        </Field>
      </div>
      <Field label="Описание">
        <Textarea rows={2} value={f.description} onChange={(e) => set({ description: e.target.value })} />
      </Field>
      <label className="flex items-center gap-2 text-sm">
        <input type="checkbox" checked={f.isActive ?? true} onChange={(e) => set({ isActive: e.target.checked })} />
        <span>В продаже</span>
      </label>
      {save.error && <ErrorMessage error={save.error} />}
      <div className="flex gap-2">
        <button type="submit" disabled={save.isPending} className="px-4 py-2 bg-slate-900 text-white rounded-lg text-xs font-bold uppercase">
          {id ? 'Сохранить' : 'Добавить товар'}
        </button>
        <button type="button" onClick={onDone} className="px-4 py-2 bg-slate-200 rounded-lg text-xs font-bold uppercase">
          Отмена
        </button>
      </div>
    </form>
  );
}

// Товары магазина, включая снятые с продажи.
export function ProductsAdmin() {
  const products = useQuery({ queryKey: productKeys.list(true), queryFn: () => productApi.list(true) });
  const [open, setOpen] = useState<number | 'new' | null>(null);
  return (
    <div className="space-y-4">
      {open === 'new' ? (
        <ProductForm initial={EMPTY} onDone={() => setOpen(null)} />
      ) : (
        <button type="button" onClick={() => setOpen('new')} className="px-4 py-2.5 bg-slate-900 text-white rounded-xl text-xs font-bold uppercase tracking-wider">
          Новый товар
        </button>
      )}
      <QueryState query={products}>
        {(list) => (
          <ul className="space-y-2">
            {list.map((p) =>
              open === p.id ? (
                <li key={p.id}>
                  <ProductForm initial={toInput(p)} id={p.id} onDone={() => setOpen(null)} />
                </li>
              ) : (
                <li key={p.id}>
                  <button
                    type="button"
                    onClick={() => setOpen(p.id)}
                    className="w-full text-left p-3 bg-white border border-slate-200 rounded-xl hover:border-slate-400 flex justify-between gap-3"
                  >
                    <span className="text-sm font-bold">{p.title}</span>
                    <span className="text-xs text-slate-500 shrink-0">
                      {formatRub(p.price)} · {p.stock} шт.{!p.isActive && ' · снят с продажи'}
                    </span>
                  </button>
                </li>
              ),
            )}
          </ul>
        )}
      </QueryState>
    </div>
  );
}
