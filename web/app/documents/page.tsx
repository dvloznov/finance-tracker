'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, Document } from '@/lib/api-client';
import Link from 'next/link';
import { useState } from 'react';

export default function DocumentsPage() {
  const [uploading, setUploading] = useState(false);
  const [uploadStatus, setUploadStatus] = useState<string>('');
  const queryClient = useQueryClient();

  const { data: documents, isLoading } = useQuery({
    queryKey: ['documents'],
    queryFn: () => apiClient.listDocuments(),
  });

  const uploadMutation = useMutation({
    mutationFn: async (file: File) => {
      setUploading(true);
      setUploadStatus('Creating upload URL...');
      
      const { upload_url, document_id } = await apiClient.createUploadUrl(file.name);
      
      setUploadStatus('Uploading file...');
      const uploadResponse = await fetch(upload_url, {
        method: 'PUT',
        body: file,
        headers: {
          'Content-Type': file.type,
        },
      });

      if (!uploadResponse.ok) {
        throw new Error('Upload failed');
      }

      setUploadStatus('Enqueueing parsing...');
      const gcsUri = `gs://YOUR_BUCKET/${file.name}`;
      await apiClient.enqueueParsing(document_id, gcsUri);
      
      return document_id;
    },
    onSuccess: () => {
      setUploadStatus('Upload successful!');
      queryClient.invalidateQueries({ queryKey: ['documents'] });
      setTimeout(() => {
        setUploading(false);
        setUploadStatus('');
      }, 2000);
    },
    onError: (error: Error) => {
      setUploadStatus(`Error: ${error.message}`);
      setUploading(false);
    },
  });

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      uploadMutation.mutate(file);
    }
  };

  return (
    <div className="min-h-screen bg-slate-50">
      <nav className="border-b bg-white shadow-sm">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <Link href="/" className="text-2xl font-bold text-slate-900">
              Finance Tracker
            </Link>
            <div className="flex gap-4">
              <Link href="/dashboard" className="px-4 py-2 text-slate-600 hover:text-slate-900 font-medium">
                Dashboard
              </Link>
              <Link href="/documents" className="px-4 py-2 text-slate-900 font-medium border-b-2 border-slate-900">
                Documents
              </Link>
              <Link href="/transactions" className="px-4 py-2 text-slate-600 hover:text-slate-900 font-medium">
                Transactions
              </Link>
            </div>
          </div>
        </div>
      </nav>

      <main className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-slate-900 mb-2">Documents</h1>
          <p className="text-slate-600">Upload and manage your bank statements</p>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6 mb-8">
          <h2 className="text-xl font-semibold mb-4">Upload New Document</h2>
          <div className="border-2 border-dashed border-slate-300 rounded-lg p-8 text-center">
            <input
              type="file"
              accept=".pdf"
              onChange={handleFileChange}
              disabled={uploading}
              className="hidden"
              id="file-upload"
            />
            <label
              htmlFor="file-upload"
              className="cursor-pointer inline-block px-6 py-3 bg-slate-900 text-white rounded-lg hover:bg-slate-800 disabled:opacity-50"
            >
              {uploading ? 'Uploading...' : 'Choose PDF File'}
            </label>
            {uploadStatus && (
              <p className="mt-4 text-sm text-slate-600">{uploadStatus}</p>
            )}
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-xl font-semibold mb-4">Uploaded Documents</h2>
          
          {isLoading ? (
            <p className="text-slate-600">Loading documents...</p>
          ) : documents && documents.length > 0 ? (
            <div className="space-y-3">
              {documents.map((doc) => (
                <div
                  key={doc.document_id}
                  className="flex items-center justify-between p-4 border border-slate-200 rounded-lg hover:bg-slate-50"
                >
                  <div>
                    <p className="font-medium text-slate-900">{doc.original_filename || 'Untitled'}</p>
                    <p className="text-sm text-slate-600">
                      Uploaded: {new Date(doc.upload_ts).toLocaleDateString()}
                    </p>
                  </div>
                  <div className="flex items-center gap-4">
                    <span
                      className={`px-3 py-1 rounded-full text-sm font-medium ${
                        doc.parsing_status === 'COMPLETED'
                          ? 'bg-green-100 text-green-800'
                          : doc.parsing_status === 'FAILED'
                          ? 'bg-red-100 text-red-800'
                          : 'bg-yellow-100 text-yellow-800'
                      }`}
                    >
                      {doc.parsing_status}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-slate-600">No documents uploaded yet</p>
          )}
        </div>
      </main>
    </div>
  );
}
