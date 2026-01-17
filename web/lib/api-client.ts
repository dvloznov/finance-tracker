import type { Account, Category, Document, Institution, Job, Transaction } from '@/shared/types/api';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string = API_BASE_URL) {
    this.baseUrl = baseUrl;
  }

  private async fetch<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...options?.headers,
      },
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: response.statusText }));
      throw new Error(error.error || `API Error: ${response.status}`);
    }

    return response.json();
  }

  // Documents
  async listDocuments(params?: { institution_id?: string; account_id?: string }): Promise<Document[]> {
    const query = new URLSearchParams(params as Record<string, string>);
    const endpoint = `/api/documents${query.toString() ? `?${query}` : ''}`;
    const response = await this.fetch<{ documents: Document[] }>(endpoint);
    return response.documents || [];
  }

  async createUploadUrl(filename: string): Promise<{ upload_url: string; document_id: string; gcs_uri: string; object_name: string }> {
    return this.fetch('/api/documents/upload-url', {
      method: 'POST',
      body: JSON.stringify({ filename }),
    });
  }

  async enqueueParsing(documentId: string, gcsUri: string): Promise<{ job_id: string }> {
    return this.fetch('/api/documents/parse', {
      method: 'POST',
      body: JSON.stringify({ document_id: documentId, gcs_uri: gcsUri }),
    });
  }

  async deleteDocument(documentId: string): Promise<{ document_id: string; status: string }> {
    return this.fetch(`/api/documents/${documentId}`, {
      method: 'DELETE',
    });
  }

  // Transactions
  async listTransactions(params?: {
    start_date?: string;
    end_date?: string;
    institution_id?: string;
    account_id?: string;
  }): Promise<Transaction[]> {
    const query = new URLSearchParams(params as Record<string, string>);
    const endpoint = `/api/transactions${query.toString() ? `?${query}` : ''}`;
    return this.fetch<Transaction[]>(endpoint);
  }

  // Categories
  async listCategories(): Promise<Category[]> {
    return this.fetch<Category[]>('/api/categories');
  }

  // Accounts
  async listAccounts(): Promise<Account[]> {
    const response = await this.fetch<{ accounts: Account[] }>('/api/accounts');
    return response.accounts || [];
  }

  // Institutions
  async listInstitutions(): Promise<Institution[]> {
    const response = await this.fetch<{ institutions: Institution[] }>('/api/institutions');
    return response.institutions || [];
  }

  // Jobs
  async getJob(jobId: string): Promise<Job> {
    return this.fetch<Job>(`/api/jobs/${jobId}`);
  }

  async listJobs(params?: { document_id?: string; status?: string }): Promise<Job[]> {
    const query = new URLSearchParams(params as Record<string, string>);
    const endpoint = `/api/jobs${query.toString() ? `?${query}` : ''}`;
    return this.fetch<Job[]>(endpoint);
  }
}

export const apiClient = new ApiClient();
