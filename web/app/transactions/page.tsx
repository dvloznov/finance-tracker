'use client';

import type { Transaction, Category } from '@/lib/api-client';
import { cardClass } from '@/lib/ui';
import { AppNav } from '@/components/app-nav';
import { getTransactionColumns } from '@/lib/columns/transactions';
import { useCategories } from '@/lib/hooks/useCategories';
import { useTransactions } from '@/lib/hooks/useTransactions';
import { useState, useMemo } from 'react';
import {
  useReactTable,
  getCoreRowModel,
  getSortedRowModel,
  getFilteredRowModel,
  flexRender,
  SortingState,
  ColumnDef,
} from '@tanstack/react-table';

export default function TransactionsPage() {
  const [sorting, setSorting] = useState<SortingState>([]);
  const [globalFilter, setGlobalFilter] = useState('');

  const { data: transactions, isLoading: transactionsLoading } = useTransactions();
  const { data: categories } = useCategories();

  const columns = useMemo<ColumnDef<Transaction>[]>(
    () => getTransactionColumns(categories),
    [categories]
  );

  const table = useReactTable({
    data: transactions || [],
    columns,
    state: {
      sorting,
      globalFilter,
    },
    onSortingChange: setSorting,
    onGlobalFilterChange: setGlobalFilter,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
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
            ) : (
              <p className="p-6 text-sm text-slate-600">No transactions found</p>
            )}
          </div>
        </div>
      </main>
    </div>
  );
}
