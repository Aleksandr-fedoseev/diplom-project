// Заявка пользователя: новый объект или дополнение к существующему (internal/transport/http/submission).
export type SubmissionKind = 'new_object' | 'addition';
export type SubmissionStatus = 'pending' | 'approved' | 'rejected';

export interface SubmissionFile {
  id: number;
  kind: string;
  originalName: string;
  contentType: string;
  sizeBytes: number;
  caption: string;
  yearTaken: string;
  sourceNote: string;
  url: string;
  createdAt: string;
}

export interface Submission {
  id: number;
  authorId: number;
  author?: { id: number; name: string; email: string; isBlocked: boolean }; // только модератору
  kind: SubmissionKind;
  objectId?: number;
  objectSlug?: string;
  objectTitle?: string;
  title: string;
  description: string;
  address: string;
  coordinates?: { lat: number; lng: number };
  period: string;
  source: string;
  status: SubmissionStatus;
  decision?: { decision: 'approved' | 'rejected'; reason?: string; moderatorId?: number; createdAt: string };
  media?: SubmissionFile[]; // только в карточке заявки
  createdAt: string;
}

export interface SubmissionInput {
  kind: SubmissionKind;
  objectSlug?: string; // для дополнения
  title: string;
  description: string;
  address?: string;
  coordinates?: { lat: number; lng: number };
  period?: string;
  source?: string;
}

// Поля объекта каталога при одобрении новой заявки; незаполненные сервер берёт из заявки.
export interface ApproveInput {
  slug?: string;
  title?: string;
  category?: string;
  protectionStatus?: string;
  address?: string;
  coordinates?: { lat: number; lng: number };
  shortDescription?: string;
  fullHistory?: string;
  isPublished?: boolean;
}
