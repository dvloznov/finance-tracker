import type { Document } from '@/shared/types/api';
import type { DocumentVM } from '@/features/documents/types';

export function toDocumentVM(document: Document): DocumentVM {
  return document;
}
