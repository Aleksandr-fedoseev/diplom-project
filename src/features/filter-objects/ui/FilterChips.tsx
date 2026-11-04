import { Eye } from 'lucide-react';
import { CATEGORY_LABELS, type HeritageCategory } from '@/entities/heritage-object';
import { cn } from '@/shared/lib';
import type { ObjectFilter } from '../model/filter';

const CATEGORIES: { id: HeritageCategory | 'all'; label: string }[] = [
  { id: 'all', label: 'Все объекты' },
  ...(Object.entries(CATEGORY_LABELS) as [HeritageCategory, string][]).map(([id, label]) => ({ id, label })),
];

const chip = 'px-3 py-1.5 rounded-lg text-xs font-medium whitespace-nowrap transition-colors';
const idle = 'bg-slate-100 text-slate-700 hover:bg-slate-200';

interface FilterChipsProps {
  filter: ObjectFilter;
  onChange: (patch: Partial<ObjectFilter>) => void;
}

// Категории кнопками и переключатель «Только с 360°» — для панели над картой.
export function FilterChips({ filter, onChange }: FilterChipsProps) {
  return (
    <div className="flex items-center gap-1.5 overflow-x-auto pb-1 sm:pb-0 scrollbar-none">
      {CATEGORIES.map((c) => (
        <button
          key={c.id}
          type="button"
          aria-pressed={filter.category === c.id}
          onClick={() => onChange({ category: c.id })}
          className={cn(chip, filter.category === c.id ? 'bg-slate-900 text-white' : idle)}
        >
          {c.label}
        </button>
      ))}
      <button
        type="button"
        aria-pressed={filter.only360}
        onClick={() => onChange({ only360: !filter.only360 })}
        className={cn(chip, 'flex items-center gap-1.5', filter.only360 ? 'bg-blue-600 text-white' : idle)}
      >
        <Eye className="w-3.5 h-3.5" />
        <span>Только с 360°</span>
      </button>
    </div>
  );
}
