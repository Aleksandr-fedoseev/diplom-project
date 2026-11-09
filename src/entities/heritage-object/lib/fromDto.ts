import { mediaUrl } from '@/shared/api';
import type { MediaFileDto, ObjectDto } from '../api/dto';
import type { HeritageCategory, HeritageObject, ProtectionStatus } from '../model/types';

// Ответ API → модель фронтенда. Медиафайлы есть только у карточки объекта.
export function fromDto(dto: ObjectDto, media?: MediaFileDto[]): HeritageObject {
  const object: HeritageObject = {
    id: dto.id,
    slug: dto.slug,
    title: dto.title,
    category: dto.category as HeritageCategory,
    protectionStatus: dto.protectionStatus as ProtectionStatus,
    registryNumber: dto.registryNumber || undefined,
    yearBuilt: dto.yearBuilt ?? '',
    architect: dto.architect || undefined,
    architecturalStyle: dto.architecturalStyle ?? '',
    address: dto.address,
    coordinates: dto.coordinates,
    coverImage: mediaUrl(dto.coverUrl),
    hasPanorama: dto.hasPanorama,
    shortDescription: dto.shortDescription,
    fullHistory: dto.fullHistory ?? '',
    tags: dto.tags,
    isPublished: dto.isPublished,
  };
  if (!media) return object;

  const archive = media.find((m) => m.kind === 'archive_photo');
  return {
    ...object,
    archiveImage: archive ? mediaUrl(archive.url) : undefined,
    archiveYear: archive?.yearTaken || undefined,
    panoramas: media
      .filter((m) => m.kind === 'panorama')
      .map((m) => ({ url: mediaUrl(m.url), source: m.source, caption: m.caption })),
    files: media.map((m) => ({
      id: m.id,
      kind: m.kind,
      url: mediaUrl(m.url),
      caption: m.caption,
      source: m.source,
      originalName: m.originalName,
    })),
  };
}
