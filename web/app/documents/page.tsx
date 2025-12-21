'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, Document } from '@/lib/api-client';
import Link from 'next/link';
import { useState, useMemo } from 'react';
import { format } from 'date-fns';
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  flexRender,
  createColumnHelper,
  SortingState,
  ColumnFiltersState,
} from '@tanstack/react-table';

const columnHelper = createColumnHelper<Document>();

export default function DocumentsPage() {
  const [uploading, setUploading] = useState(false);
  const [uploadStatus, setUploadStatus] = useState<string>('');
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [globalFilter, setGlobalFilter] = useState('');
  const queryClient = useQueryClient();

  const { data: documents, isLoading } = useQuery({
    queryKey: ['documents'],
    queryFn: () => apiClient.listDocuments(),
  });

  const columns = useMemo(() => [
    columnHelper.accessor('original_filename', {
      header: 'Filename',
      cell: (info) => (
        <span className="font-medium text-slate-900">
          {info.getValue() || 'Untitled'}
        </span>
      ),
    }),
    columnHelper.accessor('document_type', {
      header: 'Type',
      cell: (info) => (
        <span className="text-slate-700">
          {info.getValue() || '-'}
        </span>
      ),
    }),
    columnHelper.accessor('institution_id', {
      header: 'Institution',
      cell: (info) => (
        <span className="text-slate-700">
          {info.getValue() || '-'}
        </span>
      ),
    }),
    columnHelper.accessor('statement_start_date', {
      header: 'Statement Period',
      cell: (info) => {
        const start = info.getValue();
        const end = info.row.original.statement_end_date;
        if (!start || !end) return <span className="text-slate-500">-</span>;
        try {
          const startDate = new Date(start);
          const endDate = new Date(end);
          return (
            <span className="text-slate-700">
              {format(startDate, 'MMM d')} - {format(endDate, 'MMM d, yyyy')}
            </span>
          );
        } catch {
          return <span className="text-slate-500">-</span>;
        }
      },
    }),
    columnHelper.accessor('upload_ts', {
      header: 'Uploaded',
      cell: (info) => {
        try {
          const date = new Date(info.getValue());
          return (
            <span className="text-slate-700">
              {format(date, 'MMM d, yyyy HH:mm')}
            </span>
          );
        } catch {
          return <span className="text-slate-500">-</span>;
        }
      },
    }),
    columnHelper.accessor('parsing_status', {
      header: 'Status',
      cell: (info) => {
        const status = info.getValue();
        return (
          <span
            className={`px-3 py-1 rounded-full text-xs font-medium inline-block ${
              status === 'COMPLETED'
                ? 'bg-green-100 text-green-800'
                : status === 'FAILED'
                ? 'bg-red-100 text-red-800'
                : status === 'RUNNING'
                ? 'bg-blue-100 text-blue-800'
                : 'bg-yellow-100 text-yellow-800'
            }`}
          >
            {status}
          </span>
        );
      },
      filterFn: 'equals',
    }),
    columnHelper.accessor('file_mime_type', {
      header: 'File Type',
      cell: (info) => (
        <span className="text-slate-600 text-sm">
          {info.getValue()?.split('/')[1]?.toUpperCase() || '-'}
        </span>
      ),
    }),
  ], []);

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
          <div className="mb-6">
            <h2 className="text-xl font-semibold mb-4">Uploaded Documents</h2>
            
            <div className="flex gap-4 mb-4">
              <input
                type="text"
                placeholder="Search documents..."
                value={globalFilter ?? ''}
                onChange={(e) => setGlobalFilter(e.target.value)}
                className="flex-1 px-4 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-slate-900"
              />
              <select
                value={(columnFilters.find(f => f.id === 'parsing_status')?.value as string) ?? ''}
                onChange={(e) => {
                  const value = e.target.value;
                  setColumnFilters(prev => 
                    value 
                      ? [...prev.filter(f => f.id !== 'parsing_status'), { id: 'parsing_status', value }]
                      : prev.filter(f => f.id !== 'parsing_status')
                  );
                }}
                className="px-4 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-slate-900"
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
            <p className="text-slate-600">Loading documents...</p>
          ) : documents && documents.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-slate-50 border-b border-slate-200">
                  {table.getHeaderGroups().map((headerGroup) => (
                    <tr key={headerGroup.id}>
                      {headerGroup.headers.map((header) => (
                        <th
                          key={header.id}
                          className="px-4 py-3 text-left text-xs font-medium text-slate-700 uppercase tracking-wider cursor-pointer hover:bg-slate-100"
                          onClick={header.column.getToggleSortingHandler()}
                        >
                          <div className="flex items-center gap-2">
                            {flexRender(
                              header.column.columnDef.header,
                              header.getContext()
                            )}
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
                <tbody className="bg-white divide-y divide-slate-200">
                  {table.getRowModel().rows.map((row) => (
                    <tr key={row.id} className="hover:bg-slate-50">
                      {row.getVisibleCells().map((cell) => (
                        <td key={cell.id} className="px-4 py-3 text-sm">
                          {flexRender(
                            cell.column.columnDef.cell,
                            cell.getContext()
                          )}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
              
              {table.getRowModel().rows.length === 0 && (
                <p className="text-slate-600 text-center py-8">No documents match your search</p>
              )}
            </div>
          ) : (
            <p className="text-slate-600">No documents uploaded yet</p>
          )}
        </div>
      </main>
    </div>
  );
}
