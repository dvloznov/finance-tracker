import { apiClient } from '@/lib/api-client';
import type { Document } from '@/shared/types/api';

type UploadUrlResponse = {
  upload_url: string;
  document_id: string;
  gcs_uri: string;
  object_name: string;
};

type ListDocumentsParams = {
  institution_id?: string;
  account_id?: string;
};

export async function listDocuments(params?: ListDocumentsParams): Promise<Document[]> {
  return apiClient.listDocuments(params);
}

export async function createUploadUrl(filename: string): Promise<UploadUrlResponse> {
  return apiClient.createUploadUrl(filename);
}

export async function enqueueParsing(documentId: string, gcsUri: string): Promise<{ job_id: string }> {
  return apiClient.enqueueParsing(documentId, gcsUri);
}

export async function deleteDocument(documentId: string): Promise<{ document_id: string; status: string }> {
  return apiClient.deleteDocument(documentId);
}
