import type { Document } from '@/lib/api-client';
import type { DocumentVM } from '@/features/documents/types';

export function toDocumentVM(document: Document): DocumentVM {
  return document;
}
