const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface Document {
  document_id: string;
  user_id: string;
  gcs_uri: string;
  document_type?: string;
  source_system?: string;
  institution_id?: string;
  account_id?: string;
  statement_start_date?: string;
  statement_end_date?: string;
  upload_ts: string;
  processed_ts?: string;
  parsing_status: string;
  original_filename: string;
  file_mime_type?: string;
  text_gcs_uri?: string;
  checksum_sha256?: string;
  metadata?: Record<string, any>;
}

export interface Transaction {
  transaction_id: string;
  document_id: string;
  transaction_date: string;
  amount: string;
  currency: string;
  raw_description: string;
  category_name?: string;
  balance_after?: string;
}

export interface Category {
  category_id: string;
  category_name: string;
  subcategory_name?: string;
  slug: string;
}

export interface Job {
  job_id: string;
  document_id: string;
  gcs_uri: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'retrying';
  created_at: string;
  started_at?: string;
  completed_at?: string;
  error?: string;
}

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
  async listDocuments(): Promise<Document[]> {
    const response = await this.fetch<{ documents: Document[] }>('/api/documents');
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
  async listTransactions(params?: { start_date?: string; end_date?: string }): Promise<Transaction[]> {
    const query = new URLSearchParams(params as Record<string, string>);
    const endpoint = `/api/transactions${query.toString() ? `?${query}` : ''}`;
    return this.fetch<Transaction[]>(endpoint);
  }

  // Categories
  async listCategories(): Promise<Category[]> {
    return this.fetch<Category[]>('/api/categories');
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
