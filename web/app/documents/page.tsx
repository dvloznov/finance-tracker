'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/lib/api-client';
import { AppNav } from '@/shared/ui/AppNav';
import { getDocumentColumns } from '@/features/documents/columns/documentColumns';
import { DeleteConfirmModal } from '@/features/documents/components/DeleteConfirmModal';
import { DocumentsTableCard } from '@/features/documents/components/DocumentsTableCard';
import { UploadCard } from '@/features/documents/components/UploadCard';
import { useDocuments } from '@/features/documents/hooks/useDocuments';
import { useState, useMemo } from 'react';
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  SortingState,
  ColumnFiltersState,
} from '@tanstack/react-table';


export default function DocumentsPage() {
  const [uploading, setUploading] = useState(false);
  const [uploadStatus, setUploadStatus] = useState<string>('');
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [globalFilter, setGlobalFilter] = useState('');
  const [isDragging, setIsDragging] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState<{ show: boolean; documentId: string | null }>({ show: false, documentId: null });
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const queryClient = useQueryClient();

  const { data: documents, isLoading } = useDocuments();

  const columns = useMemo(() => getDocumentColumns(setDeleteConfirm), [setDeleteConfirm]);

  const table = useReactTable({
    data: documents || [],
    columns,
    state: {
      sorting,
      columnFilters,
      globalFilter,
    },
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
  });

  const uploadMutation = useMutation({
    mutationFn: async (file: File) => {
      setUploading(true);
      setUploadStatus('Creating upload URL...');
      
      const { upload_url, document_id, gcs_uri } = await apiClient.createUploadUrl(file.name);
      
      setUploadStatus('Uploading file to cloud storage...');
      
      // Build full URL - if upload_url is relative, prepend API base URL
      const apiBaseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      const fullUploadUrl = upload_url.startsWith('http') 
        ? upload_url 
        : `${apiBaseUrl}${upload_url}`;
      
      // Append filename as query parameter for the API
      const uploadUrlWithFilename = `${fullUploadUrl}?filename=${encodeURIComponent(file.name)}`;
      
      const uploadResponse = await fetch(uploadUrlWithFilename, {
        method: 'POST',
        body: file,
        headers: {
          'Content-Type': file.type,
        },
      });

      if (!uploadResponse.ok) {
        const error = await uploadResponse.json().catch(() => ({ error: 'Upload failed' }));
        throw new Error(error.error || 'Upload failed');
      }

      setUploadStatus('Triggering document parsing...');
      await apiClient.enqueueParsing(document_id, gcs_uri);
      
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

  const deleteMutation = useMutation({
    mutationFn: async (documentId: string) => {
      return apiClient.deleteDocument(documentId);
    },
    onSuccess: () => {
      setDeleteConfirm({ show: false, documentId: null });
      setDeleteError(null);
      queryClient.invalidateQueries({ queryKey: ['documents'] });
      queryClient.invalidateQueries({ queryKey: ['transactions'] });
    },
    onError: (error: Error) => {
      console.error('Delete failed:', error);
      setDeleteError(error.message);
    },
  });

  const handleDeleteConfirm = () => {
    if (deleteConfirm.documentId) {
      setDeleteError(null);
      deleteMutation.mutate(deleteConfirm.documentId);
    }
  };

  const handleDeleteCancel = () => {
    setDeleteConfirm({ show: false, documentId: null });
    setDeleteError(null);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      uploadMutation.mutate(file);
    }
  };

  const handleDragEnter = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      const file = files[0];
      if (file.type === 'application/pdf') {
        uploadMutation.mutate(file);
      } else {
        setUploadStatus('Error: Please upload a PDF file');
        setTimeout(() => setUploadStatus(''), 3000);
      }
    }
  };

  return (
    <div className="min-h-screen bg-slate-50">
      <AppNav active="documents" />

      <main className="container mx-auto px-6 py-8">
        <div className="space-y-6">
          <div className="space-y-1">
            <h1 className="text-3xl font-semibold tracking-tight text-slate-900">Documents</h1>
            <p className="text-sm text-slate-600">Upload and manage your bank statements</p>
          </div>

          <UploadCard
            isDragging={isDragging}
            uploading={uploading}
            uploadStatus={uploadStatus}
            handleDragEnter={handleDragEnter}
            handleDragOver={handleDragOver}
            handleDragLeave={handleDragLeave}
            handleDrop={handleDrop}
            handleFileChange={handleFileChange}
          />

          <DocumentsTableCard
            isLoading={isLoading}
            documents={documents}
            table={table}
            globalFilter={globalFilter}
            setGlobalFilter={setGlobalFilter}
            columnFilters={columnFilters}
            setColumnFilters={setColumnFilters}
          />
        </div>
      </main>

      <DeleteConfirmModal
        show={deleteConfirm.show}
        isPending={deleteMutation.isPending}
        deleteError={deleteError}
        onCancel={handleDeleteCancel}
        onConfirm={handleDeleteConfirm}
      />
    </div>
  );
}
