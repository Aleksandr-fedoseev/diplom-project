import type { Panorama } from '@/entities/panorama/@x/heritage-object';

export type ProtectionStatus = 'federal' | 'regional' | 'identified' | 'lost';

// Стартовые категории. Администратор может добавить новые — тогда приходит другой slug,
// и подпись берётся из справочника категорий API.
export type HeritageCategory =
  | 'wooden' // Деревянное зодчество
  | 'manor' // Купеческие особняки и усадьбы
  | 'temple' // Храмы и кремль
  | 'modern' // Архитектура модерна
  | 'lost'; // Утраченное наследие

// Файл объекта — для управления медиа в кабинете модератора.
export interface ObjectFile {
  id: number;
  kind: 'photo' | 'archive_photo' | 'panorama' | 'model3d' | 'audio';
  url: string;
  caption: string;
  source: 'photo' | 'ai_reconstruction';
  originalName: string;
}

export interface HeritageObject {
  id: number; // id в базе: по нему загружаются файлы и правится объект
  slug: string; // адрес страницы: /objects/<slug>
  title: string;
  category: HeritageCategory;
  protectionStatus: ProtectionStatus;
  registryNumber?: string;
  yearBuilt: string; // пусто, если неизвестен
  architect?: string;
  architecturalStyle: string;
  address: string;
  coordinates: { lat: number; lng: number };
  coverImage: string; // пусто — фотографий нет
  hasPanorama: boolean;
  archiveImage?: string;
  archiveYear?: string;
  panoramas?: Panorama[]; // только в карточке объекта
  files?: ObjectFile[]; // только в карточке объекта
  shortDescription: string;
  fullHistory: string; // в списке пусто, есть в карточке
  tags: string[];
  isPublished: boolean;
}
