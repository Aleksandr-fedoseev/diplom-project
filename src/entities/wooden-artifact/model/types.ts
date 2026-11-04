export interface WoodenArtifact {
  id: string;
  title: string;
  category: 'platband' | 'door' | 'cornice' | 'bracket';
  categoryLabel: string;
  originalAddress: string;
  yearBuilt: string;
  acquisitionYear: string;
  status: 'restored' | 'in_storage' | 'under_restoration';
  statusLabel: string;
  coverImage: string;
  description: string;
  history: string;
}
