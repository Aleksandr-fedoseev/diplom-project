import type { SubmissionInput } from '@/entities/submission';
import { parseCoordinates } from '@/shared/lib';

// Заявка на новый объект или дополнение к существующему (POST /api/submissions).
// Категорию и статус охраны выбирает модератор при одобрении.
export interface ObjectDraft {
  title: string;
  address: string;
  coordinates: string; // «46.3538, 48.0285» — необязательно, без них модератор поставит точку сам
  period: string; // время постройки — в свободной форме
  description: string;
  source: string;
  archive: boolean; // снимки архивные, а не современные
  yearTaken: string; // год архивного снимка
}

export const EMPTY_DRAFT: ObjectDraft = {
  title: '',
  address: '',
  coordinates: '',
  period: '',
  description: '',
  source: '',
  archive: false,
  yearTaken: '',
};

export const STEPS = ['Объект и адрес', 'Описание и история', 'Фотографии'] as const;

export const MAX_FILE_MB = 15; // как на сервере для фото (media.Rule)

export function toSubmissionInput(d: ObjectDraft, objectSlug?: string): SubmissionInput {
  return {
    kind: objectSlug ? 'addition' : 'new_object',
    objectSlug,
    title: d.title.trim(),
    description: d.description.trim(),
    address: d.address.trim(),
    coordinates: parseCoordinates(d.coordinates) ?? undefined,
    period: d.period.trim(),
    source: d.source.trim(),
  };
}

export function fileForm(file: File, d: ObjectDraft): FormData {
  const form = new FormData();
  form.set('file', file);
  form.set('kind', d.archive ? 'archive_photo' : 'photo');
  if (d.archive && d.yearTaken.trim()) form.set('yearTaken', d.yearTaken.trim());
  if (d.source.trim()) form.set('sourceNote', d.source.trim());
  return form;
}
