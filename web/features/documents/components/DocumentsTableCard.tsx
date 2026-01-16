import { flexRender, type Table } from '@tanstack/react-table';
import type { ColumnFiltersState } from '@tanstack/react-table';
import type { DocumentVM } from '@/features/documents/types';
import { Card } from '@/shared/ui/Card';
import { EmptyState } from '@/shared/ui/EmptyState';

export type DocumentsTableCardProps = {
  isLoading: boolean;
  documents: DocumentVM[] | undefined;
  table: Table<DocumentVM>;
  globalFilter: string;
  setGlobalFilter: (value: string) => void;
  columnFilters: ColumnFiltersState;
  setColumnFilters: React.Dispatch<React.SetStateAction<ColumnFiltersState>>;
};

export function DocumentsTableCard({
  isLoading,
  documents,
  table,
  globalFilter,
  setGlobalFilter,
  columnFilters,
  setColumnFilters,
}: DocumentsTableCardProps) {
  return (
    <Card title="Uploaded Documents">
      <div className="mb-6 space-y-4">
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
        <EmptyState title="Loading documents..." className="py-2" />
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
            <EmptyState title="No documents match your search" className="py-8" />
          )}
        </div>
      ) : (
        <EmptyState title="No documents uploaded yet" className="py-2" />
      )}
    </Card>
  );
}
