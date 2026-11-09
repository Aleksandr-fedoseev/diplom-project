import type { SubmissionKind, SubmissionStatus } from '../model/types';

export const SUBMISSION_STATUS_LABELS: Record<SubmissionStatus, string> = {
  pending: 'На модерации',
  approved: 'Одобрена',
  rejected: 'Отклонена',
};

export const SUBMISSION_KIND_LABELS: Record<SubmissionKind, string> = {
  new_object: 'Новый объект',
  addition: 'Дополнение',
};
