import { createColumnHelper } from '@tanstack/react-table';
import { format } from 'date-fns';
import { Trash2 } from 'lucide-react';
import type { Document } from '@/lib/api-client';

const columnHelper = createColumnHelper<Document>();

type SetDeleteConfirm = (value: { show: boolean; documentId: string | null }) => void;

export function getDocumentColumns(setDeleteConfirm: SetDeleteConfirm) {
  return [
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
    {
      id: 'actions',
      header: 'Actions',
      cell: (info: any) => {
        const document = info.row.original as Document;
        return (
          <button
            onClick={() => setDeleteConfirm({ show: true, documentId: document.document_id })}
            className="text-red-600 hover:text-red-800 p-1 rounded hover:bg-red-50 transition-colors"
            title="Delete document"
          >
            <Trash2 size={18} />
          </button>
        );
      },
    },
  ];
}
