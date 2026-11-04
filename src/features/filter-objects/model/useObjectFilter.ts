'use client';

import { useMemo, useState } from 'react';
import type { HeritageObject } from '@/entities/heritage-object';
import { EMPTY_FILTER, filterObjects, type ObjectFilter } from './filter';

// Состояние фильтра и отфильтрованный список. update меняет одно или несколько условий.
export function useObjectFilter(objects: HeritageObject[]) {
  const [filter, setFilter] = useState<ObjectFilter>(EMPTY_FILTER);
  const filtered = useMemo(() => filterObjects(objects, filter), [objects, filter]);
  const update = (patch: Partial<ObjectFilter>) => setFilter((f) => ({ ...f, ...patch }));
  return { filter, update, filtered };
}
