import { createColumnHelper, type CellContext } from '@tanstack/react-table';
import { Trash2, AlertCircle } from 'lucide-react';
import { formatMonthDay, formatShortDateTime } from '@/shared/formatters/date';
import { DocumentStatusBadge } from '@/shared/ui/DocumentStatusBadge';
import type { DocumentVM } from '@/features/documents/types';
import type { Institution } from '@/shared/types/api';

const columnHelper = createColumnHelper<DocumentVM>();

type SetDeleteConfirm = (value: { show: boolean; documentId: string | null }) => void;

export function getDocumentColumns(
  setDeleteConfirm: SetDeleteConfirm,
  institutions: Institution[] = []
) {
  const institutionLookup = new Map(institutions.map((i) => [i.institution_id, i.name]));

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
      cell: (info) => {
        const id = info.getValue();
        const name = id ? institutionLookup.get(id) : null;
        return (
          <span className="text-slate-700">
            {name ?? id ?? '-'}
          </span>
        );
      },
    }),
    columnHelper.accessor('statement_start_date', {
      header: 'Statement Period',
      cell: (info) => {
        const start = info.getValue();
        const end = info.row.original.statement_end_date;
        if (!start || !end) return <span className="text-slate-500">-</span>;
        try {
          return (
            <span className="text-slate-700">
              {formatMonthDay(start)} - {formatMonthDay(end)}
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
          return (
            <span className="text-slate-700">
              {formatShortDateTime(info.getValue())}
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
        const doc = info.row.original;
        return (
          <div className="flex items-center gap-2">
            <DocumentStatusBadge status={status} />
            {status === 'FAILED' && doc.error_message && (
              <span
                title={doc.error_message}
                className="text-red-500 cursor-help inline-flex"
              >
                <AlertCircle size={16} />
              </span>
            )}
          </div>
        );
      },
      filterFn: (row, columnId, filterValue) => {
        const status = row.getValue(columnId) as string;
        if (filterValue === 'RUNNING') {
          return status === 'RUNNING' || status === 'PROCESSING';
        }
        return status === filterValue;
      },
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
      cell: (info: CellContext<DocumentVM, unknown>) => {
        const document = info.row.original;
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
