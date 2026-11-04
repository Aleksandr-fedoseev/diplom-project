import type { WoodenArtifact } from './types';

export const ARTIFACT_CATEGORIES: { id: WoodenArtifact['category']; label: string }[] = [
  { id: 'platband', label: 'Наличники' },
  { id: 'door', label: 'Двери' },
  { id: 'cornice', label: 'Карнизы' },
  { id: 'bracket', label: 'Кронштейны' },
];
