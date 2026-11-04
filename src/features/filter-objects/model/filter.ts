import { hasPanorama, type HeritageCategory, type HeritageObject, type ProtectionStatus } from '@/entities/heritage-object';

export interface ObjectFilter {
  query: string; // ищется в названии, адресе и стиле
  category: HeritageCategory | 'all';
  status: ProtectionStatus | 'all';
  only360: boolean;
}

export const EMPTY_FILTER: ObjectFilter = { query: '', category: 'all', status: 'all', only360: false };

export function filterObjects(objects: HeritageObject[], f: ObjectFilter): HeritageObject[] {
  const q = f.query.trim().toLowerCase();
  return objects.filter(
    (o) =>
      (f.category === 'all' || o.category === f.category) &&
      (f.status === 'all' || o.protectionStatus === f.status) &&
      (!f.only360 || hasPanorama(o)) &&
      (q === '' || [o.title, o.address, o.architecturalStyle].some((s) => s.toLowerCase().includes(q))),
  );
}
