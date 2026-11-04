import type { HeritageCategory } from '@/entities/heritage-object';

// Заявка на новый объект. На сервере ей соответствует POST /api/submissions (kind=new_object).
export interface ObjectDraft {
  title: string;
  address: string;
  category: HeritageCategory;
  yearBuilt: string;
  description: string;
  sources: string;
  authorName: string;
  authorEmail: string;
}

export const EMPTY_DRAFT: ObjectDraft = {
  title: '',
  address: '',
  category: 'wooden',
  yearBuilt: '',
  description: '',
  sources: '',
  authorName: '',
  authorEmail: '',
};

export const STEPS = ['Локация и адрес', 'Описание и история', 'Фотографии и автор'] as const;

// Подписи категорий в форме заявки — в единственном числе («какой это объект»).
export const CATEGORY_OPTIONS: { value: HeritageCategory; label: string }[] = [
  { value: 'wooden', label: 'Деревянное зодчество' },
  { value: 'manor', label: 'Купеческий особняк' },
  { value: 'temple', label: 'Культовое сооружение' },
  { value: 'modern', label: 'Модерн' },
  { value: 'lost', label: 'Утраченное здание' },
];
