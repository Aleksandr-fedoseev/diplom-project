'use client';

import { HERITAGE_OBJECTS, ObjectCard } from '@/entities/heritage-object';
import { FilterSelects, useObjectFilter } from '@/features/filter-objects';
import { PageHeader, SearchInput } from '@/shared/ui';

export default function ObjectsPage() {
  const { filter, update, filtered } = useObjectFilter(HERITAGE_OBJECTS);

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 sm:py-14 w-full">
      <PageHeader
        kicker="Памятники Астрахани"
        title="Исторические здания и архитектурные памятники"
        lead="Собрание старинных купеческих усадеб, деревянных теремов, кремлевских построек и утраченных храмов Астраханской области."
      />

      <div className="bg-white p-4 rounded-xl border border-slate-200 shadow-xs mb-8 flex flex-col md:flex-row gap-3">
        <SearchInput
          className="flex-1"
          placeholder="Поиск по названию, адресу, стилю..."
          value={filter.query}
          onChange={(query) => update({ query })}
        />
        <FilterSelects filter={filter} onChange={update} />
      </div>

      {filtered.length === 0 ? (
        <p className="py-16 text-center text-sm text-slate-500">По заданному фильтру объектов не найдено.</p>
      ) : (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {filtered.map((object) => (
            <ObjectCard key={object.id} object={object} />
          ))}
        </div>
      )}
    </div>
  );
}
