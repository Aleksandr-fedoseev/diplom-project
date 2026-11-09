import { api } from '@/shared/api';
import type { ApproveInput, Submission, SubmissionFile, SubmissionInput, SubmissionStatus } from '../model/types';

export const submissionKeys = {
  mine: ['submissions', 'mine'] as const,
  queue: (status: SubmissionStatus) => ['submissions', 'queue', status] as const,
  one: (id: number) => ['submissions', 'one', id] as const,
};

export const submissionApi = {
  create: (input: SubmissionInput) => api<Submission>('/submissions', { method: 'POST', json: input }),
  // Форма: file, kind (photo | archive_photo), caption, yearTaken, sourceNote.
  uploadMedia: (id: number, form: FormData) =>
    api<SubmissionFile>(`/submissions/${id}/media`, { method: 'POST', form }),
  get: (id: number) => api<Submission>(`/submissions/${id}`),
  mine: () => api<Submission[]>('/me/submissions'),
  queue: (status: SubmissionStatus) =>
    api<Submission[]>('/moderation/submissions', { query: { status, limit: 100 } }),
  approve: (id: number, input?: ApproveInput) =>
    api<Submission>(`/moderation/submissions/${id}/approve`, { method: 'POST', json: input ?? {} }),
  reject: (id: number, reason: string, blockAuthor = false) =>
    api<Submission>(`/moderation/submissions/${id}/reject`, { method: 'POST', json: { reason, blockAuthor } }),
};
