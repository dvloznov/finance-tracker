'use client';

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, Document } from '@/lib/api-client';
import { cardClass } from '@/lib/ui';
import { AppNav } from '@/components/app-nav';
import { getDocumentColumns } from '@/lib/columns/documents';
import { useDocuments } from '@/lib/hooks/useDocuments';
import { useState, useMemo } from 'react';
import { Trash2, X } from 'lucide-react';
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  flexRender,
  SortingState,
  ColumnFiltersState,
} from '@tanstack/react-table';

type UploadCardProps = {
  isDragging: boolean;
  uploading: boolean;
  uploadStatus: string;
  handleDragEnter: (e: React.DragEvent) => void;
  handleDragOver: (e: React.DragEvent) => void;
  handleDragLeave: (e: React.DragEvent) => void;
  handleDrop: (e: React.DragEvent) => void;
  handleFileChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
};

function UploadCard({
  isDragging,
  uploading,
  uploadStatus,
  handleDragEnter,
  handleDragOver,
  handleDragLeave,
  handleDrop,
  handleFileChange,
}: UploadCardProps) {
  return (
    <div className={cardClass}>
      <h2 className="text-sm font-semibold text-slate-900 mb-6">Upload New Document</h2>
      <div
        className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
          isDragging
            ? 'border-slate-900 bg-slate-50'
            : 'border-slate-300 hover:border-slate-400'
        }`}
        onDragEnter={handleDragEnter}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
      >
        <input
          type="file"
          accept=".pdf"
          onChange={handleFileChange}
          disabled={uploading}
          className="hidden"
          id="file-upload"
        />
        <div className="space-y-4">
          <div className="text-slate-600">
            <svg
              className="mx-auto h-12 w-12 text-slate-400"
              stroke="currentColor"
              fill="none"
              viewBox="0 0 48 48"
              aria-hidden="true"
            >
              <path
                d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02"
                strokeWidth={2}
                strokeLinecap="round"
                strokeLinejoin="round"
              />
            </svg>
            <p className="mt-2 text-sm font-medium">
              {isDragging ? 'Drop your PDF file here' : 'Drag and drop your PDF file here'}
            </p>
            <p className="mt-1 text-xs text-slate-500">or</p>
          </div>
          <label
            htmlFor="file-upload"
            className="cursor-pointer inline-block px-6 py-3 bg-slate-900 text-white rounded-lg hover:bg-slate-800 disabled:opacity-50 text-sm font-medium"
          >
            {uploading ? 'Uploading...' : 'Choose PDF File'}
          </label>
        </div>
        {uploadStatus && (
          <p className="mt-4 text-sm text-slate-600">{uploadStatus}</p>
        )}
      </div>
    </div>
  );
}

type DocumentsTableCardProps = {
  isLoading: boolean;
  documents: Document[] | undefined;
  table: ReturnType<typeof useReactTable<Document>>;
  globalFilter: string;
  setGlobalFilter: (value: string) => void;
  columnFilters: ColumnFiltersState;
  setColumnFilters: React.Dispatch<React.SetStateAction<ColumnFiltersState>>;
};

function DocumentsTableCard({
  isLoading,
  documents,
  table,
  globalFilter,
  setGlobalFilter,
  columnFilters,
  setColumnFilters,
}: DocumentsTableCardProps) {
  return (
    <div className={cardClass}>
      <div className="mb-6 space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold text-slate-900">Uploaded Documents</h2>
        </div>

        <div className="flex flex-col gap-3 sm:flex-row">
          <input
            type="text"
            placeholder="Search documents..."
            value={globalFilter ?? ''}
            onChange={(e) => setGlobalFilter(e.target.value)}
            className="flex-1 px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-slate-900/10 focus:border-slate-300 text-sm"
          />
          <select
            value={(columnFilters.find((f) => f.id === 'parsing_status')?.value as string) ?? ''}
            onChange={(e) => {
              const value = e.target.value;
              setColumnFilters((prev) =>
                value
                  ? [...prev.filter((f) => f.id !== 'parsing_status'), { id: 'parsing_status', value }]
                  : prev.filter((f) => f.id !== 'parsing_status')
              );
            }}
            className="px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-slate-900/10 focus:border-slate-300 text-sm"
          >
            <option value="">All Statuses</option>
            <option value="PENDING">Pending</option>
            <option value="RUNNING">Running</option>
            <option value="COMPLETED">Completed</option>
            <option value="FAILED">Failed</option>
          </select>
        </div>
      </div>

      {isLoading ? (
        <p className="text-sm text-slate-600">Loading documents...</p>
      ) : documents && documents.length > 0 ? (
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-slate-50 border-b border-slate-100">
              {table.getHeaderGroups().map((headerGroup) => (
                <tr key={headerGroup.id}>
                  {headerGroup.headers.map((header) => (
                    <th
                      key={header.id}
                      className="px-4 py-3 text-left text-[11px] font-medium text-slate-500 uppercase tracking-wider cursor-pointer hover:bg-slate-100"
                      onClick={header.column.getToggleSortingHandler()}
                    >
                      <div className="flex items-center gap-2">
                        {flexRender(header.column.columnDef.header, header.getContext())}
                        {{
                          asc: ' ↑',
                          desc: ' ↓',
                        }[header.column.getIsSorted() as string] ?? null}
                      </div>
                    </th>
                  ))}
                </tr>
              ))}
            </thead>
            <tbody className="bg-white divide-y divide-slate-100">
              {table.getRowModel().rows.map((row) => (
                <tr key={row.id} className="hover:bg-slate-50">
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className="px-4 py-3 text-sm">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>

          {table.getRowModel().rows.length === 0 && (
            <p className="text-sm text-slate-500 text-center py-8">No documents match your search</p>
          )}
        </div>
      ) : (
        <p className="text-sm text-slate-600">No documents uploaded yet</p>
      )}
    </div>
  );
}

type DeleteConfirmModalProps = {
  show: boolean;
  isPending: boolean;
  deleteError: string | null;
  onCancel: () => void;
  onConfirm: () => void;
};

function DeleteConfirmModal({
  show,
  isPending,
  deleteError,
  onCancel,
  onConfirm,
}: DeleteConfirmModalProps) {
  if (!show) return null;

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-2xl ring-1 ring-black/5 shadow-sm max-w-md w-full mx-4">
        <div className="p-6">
          <div className="flex items-start gap-4">
            <div className="flex-shrink-0">
              <div className="w-12 h-12 rounded-full bg-red-100 flex items-center justify-center">
                <Trash2 className="text-red-600" size={24} />
              </div>
            </div>
            <div className="flex-1">
              <h3 className="text-base font-semibold text-slate-900 mb-2">
                Delete Document
              </h3>
              <p className="text-sm text-slate-600 mb-4">
                Are you sure you want to delete this document? This will also delete all associated transactions and cannot be undone.
              </p>

              {deleteError && (
                <div className="mb-4 p-3 bg-red-50 ring-1 ring-red-200 rounded-lg">
                  <p className="text-sm text-red-800 flex items-start gap-2">
                    <X size={16} className="flex-shrink-0 mt-0.5" />
                    <span>{deleteError}</span>
                  </p>
                </div>
              )}
            </div>
          </div>
        </div>
        <div className="bg-slate-50 px-6 py-4 flex gap-3 justify-end rounded-b-2xl">
          <button
            onClick={onCancel}
            disabled={isPending}
            className="px-4 py-2 text-sm font-medium text-slate-700 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            disabled={isPending}
            className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {isPending ? (
              <>
                <span className="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                Deleting...
              </>
            ) : (
              'Delete'
            )}
          </button>
        </div>
      </div>
    </div>
  );
}

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
