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
  hasPanorama360: boolean;
  panoramaImage?: string;
  hasAudioGuide: boolean;
  audioDuration?: string;
  shortDescription: string;
  fullHistory: string;
  tags: string[];
}

export interface RouteItem {
  id: string;
  slug: string;
  title: string;
  description: string;
  coverImage: string;
  duration: string;
  distance: string;
  stopsCount: number;
  objectIds: string[];
}

export interface MerchItem {
  id: string;
  title: string;
  price: number;
  image: string;
  category: string;
  description: string;
}
