import { describe, expect, it } from 'vitest';
import type { MediaFileDto, ObjectDto } from '../api/dto';
import { fromDto } from './fromDto';

const dto: ObjectDto = {
  id: 12,
  slug: 'dom-kononova',
  title: 'Дом Кононова',
  category: 'wooden',
  protectionStatus: 'regional',
  address: 'ул. Мусы Джалиля, 18',
  coordinates: { lat: 46.35, lng: 48.04 },
  shortDescription: 'Кратко',
  tags: ['Резьба'],
  isPublished: true,
  coverUrl: '/api/media/5/file',
  hasPanorama: true,
  updatedAt: '2026-11-01T00:00:00Z',
};

const file = (patch: Partial<MediaFileDto>): MediaFileDto => ({
  id: 1,
  kind: 'photo',
  originalName: 'a.jpg',
  contentType: 'image/jpeg',
  sizeBytes: 1,
  caption: '',
  yearTaken: '',
  sourceNote: '',
  source: 'photo',
  author: '',
  license: '',
  url: '/api/media/1/file',
  createdAt: '',
  ...patch,
});

describe('fromDto', () => {
  it('адреса файлов дополняются адресом сервера, пустые поля — пустыми строками', () => {
    const o = fromDto(dto);
    expect(o.coverImage).toBe('http://localhost:8080/api/media/5/file');
    expect(o.yearBuilt).toBe('');
    expect(o.architecturalStyle).toBe('');
    expect(o.fullHistory).toBe('');
    expect(o.panoramas).toBeUndefined();
  });

  it('нет обложки — пустая строка, а не адрес сервера', () => {
    expect(fromDto({ ...dto, coverUrl: '' }).coverImage).toBe('');
  });

  it('карточка: панорамы с пометкой ИИ и архивный снимок с годом', () => {
    const o = fromDto(dto, [
      file({ id: 7, kind: 'panorama', source: 'ai_reconstruction', caption: 'Фасад', url: '/api/media/7/file' }),
      file({ id: 8, kind: 'archive_photo', yearTaken: '1910', url: '/api/media/8/file' }),
      file({ id: 9, kind: 'model3d', url: '/api/media/9/file' }),
    ]);
    expect(o.panoramas).toEqual([{ url: 'http://localhost:8080/api/media/7/file', source: 'ai_reconstruction', caption: 'Фасад' }]);
    expect(o.archiveImage).toBe('http://localhost:8080/api/media/8/file');
    expect(o.archiveYear).toBe('1910');
  });
});
