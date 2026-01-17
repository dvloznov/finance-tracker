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
  account_id?: string;
  institution_id?: string;
  transaction_date: string;
  amount: string;
  currency: string;
  raw_description: string;
  category_name?: string;
  subcategory_name?: string;
  balance_after?: string;
}

export interface Account {
  account_id: string;
  institution_id?: string;
  account_name?: string;
  account_number?: string;
  account_type?: string;
  currency?: string;
  created_at?: string;
  updated_at?: string;
}

export interface Institution {
  institution_id: string;
  name: string;
  country?: string;
  created_at?: string;
  updated_at?: string;
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
