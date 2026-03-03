import type { Account, Category, Document, Institution, Job, Merchant, Transaction } from '@/shared/types/api';

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

    if (response.status === 204) {
      return undefined as T;
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

  async createUploadUrl(
    filename: string,
    accountId?: string
  ): Promise<{ upload_url: string; document_id: string; gcs_uri: string; object_name: string; account_id?: string }> {
    const body: { filename: string; account_id?: string } = { filename };
    if (accountId) body.account_id = accountId;
    return this.fetch('/api/documents/upload-url', {
      method: 'POST',
      body: JSON.stringify(body),
    });
  }

  async updateDocument(
    documentId: string,
    payload: { account_id?: string | null }
  ): Promise<{ document_id: string; account_id: string; institution_id: string; status: string }> {
    return this.fetch(`/api/documents/${documentId}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
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

  // Merchants
  async listMerchants(params?: { start_date?: string; end_date?: string }): Promise<Merchant[]> {
    const query = new URLSearchParams(params as Record<string, string>);
    const endpoint = `/api/merchants${query.toString() ? `?${query}` : ''}`;
    return this.fetch<Merchant[]>(endpoint);
  }

  async updateMerchantCategory(merchantId: string, categoryId: string): Promise<void> {
    await this.fetch(`/api/merchants/${merchantId}/category`, {
      method: 'PUT',
      body: JSON.stringify({ category_id: categoryId }),
    });
  }

  async mergeMerchant(merchantId: string, canonicalMerchantId: string): Promise<void> {
    await this.fetch(`/api/merchants/${merchantId}/merge`, {
      method: 'PUT',
      body: JSON.stringify({ canonical_merchant_id: canonicalMerchantId }),
    });
  }

  async unmergeMerchant(merchantId: string): Promise<void> {
    await this.fetch(`/api/merchants/${merchantId}/merge`, {
      method: 'DELETE',
    });
  }

  // Accounts
  async listAccounts(): Promise<Account[]> {
    const response = await this.fetch<{ accounts: Account[] }>('/api/accounts');
    return response.accounts || [];
  }

  async createAccount(payload: {
    institution_id: string;
    account_name?: string;
    account_number?: string;
    sort_code?: string;
    iban?: string;
    currency?: string;
    account_type?: string;
  }): Promise<{ account_id: string; account: Account }> {
    return this.fetch('/api/accounts', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  }

  async updateAccount(
    accountId: string,
    payload: Partial<{
      institution_id: string;
      account_name: string;
      account_number: string;
      sort_code: string;
      iban: string;
      currency: string;
      account_type: string;
    }>
  ): Promise<{ account_id: string; status: string }> {
    return this.fetch(`/api/accounts/${accountId}`, {
      method: 'PATCH',
      body: JSON.stringify(payload),
    });
  }

  async deleteAccount(accountId: string): Promise<{ account_id: string; status: string }> {
    return this.fetch(`/api/accounts/${accountId}`, {
      method: 'DELETE',
    });
  }

  // Institutions
  async listInstitutions(): Promise<Institution[]> {
    const response = await this.fetch<{ institutions: Institution[] }>('/api/institutions');
    return response.institutions || [];
  }

  async createInstitution(name: string): Promise<{ institution_id: string; name: string }> {
    return this.fetch('/api/institutions', {
      method: 'POST',
      body: JSON.stringify({ name }),
    });
  }

  async updateInstitution(
    institutionId: string,
    name: string
  ): Promise<{ institution_id: string; status: string }> {
    return this.fetch(`/api/institutions/${institutionId}`, {
      method: 'PATCH',
      body: JSON.stringify({ name }),
    });
  }

  async deleteInstitution(institutionId: string): Promise<{ institution_id: string; status: string }> {
    return this.fetch(`/api/institutions/${institutionId}`, {
      method: 'DELETE',
    });
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
