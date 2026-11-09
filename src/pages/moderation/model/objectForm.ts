import type { HeritageObject, ObjectInput } from '@/entities/heritage-object';
import { parseCoordinates } from '@/shared/lib';

// Поля формы объекта — строки, как в input; в API уходят через toObjectInput.
export interface ObjectForm {
  slug: string;
  title: string;
  category: string;
  protectionStatus: string;
  registryNumber: string;
  yearBuilt: string;
  architect: string;
  architecturalStyle: string;
  address: string;
  coordinates: string;
  shortDescription: string;
  fullHistory: string;
  tags: string; // через запятую
  isPublished: boolean;
}

export const EMPTY_OBJECT_FORM: ObjectForm = {
  slug: '',
  title: '',
  category: 'manor',
  protectionStatus: 'identified',
  registryNumber: '',
  yearBuilt: '',
  architect: '',
  architecturalStyle: '',
  address: '',
  coordinates: '',
  shortDescription: '',
  fullHistory: '',
  tags: '',
  isPublished: true,
};

export function toObjectForm(o: HeritageObject): ObjectForm {
  return {
    slug: o.slug,
    title: o.title,
    category: o.category,
    protectionStatus: o.protectionStatus,
    registryNumber: o.registryNumber ?? '',
    yearBuilt: o.yearBuilt,
    architect: o.architect ?? '',
    architecturalStyle: o.architecturalStyle,
    address: o.address,
    coordinates: `${o.coordinates.lat}, ${o.coordinates.lng}`,
    shortDescription: o.shortDescription,
    fullHistory: o.fullHistory,
    tags: o.tags.join(', '),
    isPublished: o.isPublished,
  };
}

// Пустое необязательное поле уходит как null — сервер очистит значение.
const optional = (s: string) => s.trim() || null;

export function toObjectInput(f: ObjectForm): ObjectInput {
  const coordinates = parseCoordinates(f.coordinates);
  if (!coordinates) throw new Error('Укажите координаты: широта и долгота, например 46.3538, 48.0285');
  return {
    slug: f.slug.trim(),
    title: f.title.trim(),
    category: f.category,
    protectionStatus: f.protectionStatus,
    registryNumber: optional(f.registryNumber),
    yearBuilt: optional(f.yearBuilt),
    architect: optional(f.architect),
    architecturalStyle: optional(f.architecturalStyle),
    address: f.address.trim(),
    coordinates,
    shortDescription: f.shortDescription.trim(),
    fullHistory: f.fullHistory.trim(),
    tags: f.tags
      .split(',')
      .map((t) => t.trim())
      .filter(Boolean),
    isPublished: f.isPublished,
  };
}
