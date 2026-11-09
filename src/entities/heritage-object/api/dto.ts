// Форматы JSON API бэкенда (internal/transport/http/catalog и media).

export interface ObjectDto {
  id: number;
  slug: string;
  title: string;
  category: string;
  protectionStatus: string;
  registryNumber?: string;
  yearBuilt?: string;
  architect?: string;
  architecturalStyle?: string;
  address: string;
  coordinates: { lat: number; lng: number };
  shortDescription: string;
  fullHistory?: string;
  tags: string[];
  isPublished: boolean;
  coverUrl: string;
  hasPanorama: boolean;
  updatedAt: string;
}

export type MediaKind = 'photo' | 'archive_photo' | 'panorama' | 'model3d' | 'audio';

export interface MediaFileDto {
  id: number;
  objectId?: number;
  submissionId?: number;
  kind: MediaKind;
  originalName: string;
  contentType: string;
  sizeBytes: number;
  caption: string;
  yearTaken: string;
  sourceNote: string;
  source: 'photo' | 'ai_reconstruction';
  projection?: 'full_360' | 'partial';
  hAngle?: number;
  author: string;
  license: string;
  url: string;
  createdAt: string;
}

export interface ObjectCardDto {
  object: ObjectDto;
  media: MediaFileDto[];
}

export interface CategoryDto {
  id: number;
  slug: string;
  name: string;
}

// Тело POST/PUT /objects — то, что правит модератор.
export interface ObjectInput {
  slug: string;
  title: string;
  category: string;
  protectionStatus: string;
  registryNumber?: string | null;
  yearBuilt?: string | null;
  architect?: string | null;
  architecturalStyle?: string | null;
  address: string;
  coordinates: { lat: number; lng: number };
  shortDescription: string;
  fullHistory: string;
  tags: string[];
  isPublished?: boolean;
}
