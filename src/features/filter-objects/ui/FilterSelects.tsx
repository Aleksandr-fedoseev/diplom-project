import { CATEGORY_LABELS, type HeritageCategory, type ProtectionStatus } from '@/entities/heritage-object';
import { Select } from '@/shared/ui';
import type { ObjectFilter } from '../model/filter';

const CATEGORY_OPTIONS: { value: HeritageCategory | 'all'; label: string }[] = [
  { value: 'all', label: 'Все категории' },
  ...(Object.entries(CATEGORY_LABELS) as [HeritageCategory, string][]).map(([value, label]) => ({ value, label })),
];

const STATUS_OPTIONS: { value: ProtectionStatus | 'all'; label: string }[] = [
  { value: 'all', label: 'Любой статус охраны' },
  { value: 'federal', label: 'Федеральное значение' },
  { value: 'regional', label: 'Региональное значение' },
  { value: 'identified', label: 'Выявленный объект' },
  { value: 'lost', label: 'Утраченный памятник' },
];

interface FilterSelectsProps {
  filter: ObjectFilter;
  onChange: (patch: Partial<ObjectFilter>) => void;
}

// Категория и статус охраны выпадающими списками — для каталога.
export function FilterSelects({ filter, onChange }: FilterSelectsProps) {
  return (
    <div className="flex items-center gap-2">
      <Select label="Категория" value={filter.category} options={CATEGORY_OPTIONS} onChange={(category) => onChange({ category })} />
      <Select label="Статус охраны" value={filter.status} options={STATUS_OPTIONS} onChange={(status) => onChange({ status })} />
    </div>
  );
}
