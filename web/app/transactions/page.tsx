'use client';

import type { TransactionVM } from '@/features/transactions/types';
import { cardClass } from '@/lib/ui';
import { AppNav } from '@/shared/ui/AppNav';
import { getTransactionColumns } from '@/features/transactions/columns/transactionColumns';
import { useCategories } from '@/features/categories/hooks/useCategories';
import { useTransactions } from '@/features/transactions/hooks/useTransactions';
import { useAccountScope } from '@/shared/account-scope/context';
import { detectTransferIds } from '@/features/dashboard/analytics/transfers';
import { useState, useMemo } from 'react';
import { Inbox } from 'lucide-react';
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  flexRender,
  SortingState,
  ColumnDef,
} from '@tanstack/react-table';

export default function TransactionsPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [globalFilter, setGlobalFilter] = useState('');
  const [pagination, setPagination] = useState({ pageIndex: 0, pageSize: 20 });
  const { scope } = useAccountScope();

  const { data: transactions, isLoading: transactionsLoading } = useTransactions({ scope });
  const { data: categories } = useCategories();

  const transferIds = useMemo(() => {
    if (scope.mode !== 'all') return new Set<string>();
    return detectTransferIds(transactions ?? []);
  }, [transactions, scope.mode]);

  const columns = useMemo<ColumnDef<TransactionVM>[]>(
    () => getTransactionColumns(categories, transferIds),
    [categories, transferIds]
  );

  const table = useReactTable({
    data: transactions || [],
    columns,
    state: {
      sorting,
      globalFilter,
      pagination,
    },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    onPaginationChange: setPagination,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
  });

  return (
    <div className="min-h-screen bg-slate-50">
      <AppNav active="transactions" />

      <main className="container mx-auto px-6 py-8">
        <div className="space-y-6">
          <div className="space-y-1">
            <h1 className="text-3xl font-semibold tracking-tight text-slate-900">Transactions</h1>
            <p className="text-sm text-slate-600">View and categorize your transactions</p>
          </div>

          <div className={cardClass}>
            <input
              type="text"
              placeholder="Search transactions..."
              value={globalFilter}
              onChange={(e) => setGlobalFilter(e.target.value)}
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-slate-900/10 focus:border-slate-300 text-sm"
            />
          </div>

          <div className="bg-white rounded-2xl ring-1 ring-black/5 shadow-sm overflow-hidden">
            {transactionsLoading ? (
              <p className="p-6 text-sm text-slate-600">Loading transactions...</p>
            ) : transactions && transactions.length > 0 ? (
              <>
                <div className="overflow-x-auto">
                  <table className="w-full">
                  <thead className="bg-slate-50 border-b border-slate-100">
                    {table.getHeaderGroups().map((headerGroup) => (
                      <tr key={headerGroup.id}>
                        {headerGroup.headers.map((header) => (
                          <th
                            key={header.id}
                            className="px-6 py-3 text-left text-[11px] font-medium text-slate-500 uppercase tracking-wider cursor-pointer hover:bg-slate-100"
                            onClick={header.column.getToggleSortingHandler()}
                          >
                            {flexRender(
                              header.column.columnDef.header,
                              header.getContext()
                            )}
                          </th>
                        ))}
                      </tr>
                    ))}
                  </thead>
                  <tbody className="bg-white divide-y divide-slate-100">
                    {table.getRowModel().rows.map((row) => (
                      <tr key={row.id} className="hover:bg-slate-50">
                        {row.getVisibleCells().map((cell) => (
                          <td key={cell.id} className="px-6 py-4 text-sm">
                            {flexRender(cell.column.columnDef.cell, cell.getContext())}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                  </table>
                </div>
                <div className="flex items-center justify-between border-t border-slate-100 px-6 py-3 text-sm text-slate-600">
                <div>
                  Showing {table.getRowModel().rows.length} of {table.getFilteredRowModel().rows.length}
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-xs text-slate-500">
                    Page {pagination.pageIndex + 1} of {table.getPageCount()}
                  </span>
                  <button
                    type="button"
                    onClick={() => table.previousPage()}
                    disabled={!table.getCanPreviousPage()}
                    className="rounded-md border border-slate-200 px-3 py-1 text-xs text-slate-700 disabled:opacity-50"
                  >
                    Previous
                  </button>
                  <button
                    type="button"
                    onClick={() => table.nextPage()}
                    disabled={!table.getCanNextPage()}
                    className="rounded-md border border-slate-200 px-3 py-1 text-xs text-slate-700 disabled:opacity-50"
                  >
                    Next
                  </button>
                </div>
                </div>
              </>
            ) : (
              <div className="flex flex-col items-center justify-center p-6 text-center">
              <Inbox className="w-10 h-10 text-slate-300 mb-3" />
              <p className="text-base text-slate-600">No transactions found</p>
            </div>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
