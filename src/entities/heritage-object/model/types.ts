import type { Panorama } from '@/entities/panorama/@x/heritage-object';

export type ProtectionStatus = 'federal' | 'regional' | 'identified' | 'lost';

export type HeritageCategory = 
  | 'wooden'      // Деревянное зодчество
  | 'manor'       // Купеческие особняки и усадьбы
  | 'temple'      // Храмы и кремль
  | 'modern'      // Архитектура модерна
  | 'lost';       // Утраченное наследие

export interface HeritageObject {
  id: string;
  slug: string;
  title: string;
  category: HeritageCategory;
  protectionStatus: ProtectionStatus;
  registryNumber?: string;
  yearBuilt: string;
  architect?: string;
  architecturalStyle: string;
  address: string;
  coordinates: {
    lat: number;
    lng: number;
  };
  coverImage: string;
  archiveImage?: string;
  archiveYear?: string;
  panoramas?: Panorama[];
  shortDescription: string;
  fullHistory: string;
  tags: string[];
}
