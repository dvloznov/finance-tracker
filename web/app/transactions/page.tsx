'use client';

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, Transaction, Category } from '@/lib/api-client';
import Link from 'next/link';
import { useState, useMemo } from 'react';
import { format } from 'date-fns';
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
  const queryClient = useQueryClient();

  const { data: transactions, isLoading: transactionsLoading } = useQuery({
    queryKey: ['transactions'],
    queryFn: () => apiClient.listTransactions(),
  });

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: () => apiClient.listCategories(),
  });

  const columns = useMemo<ColumnDef<Transaction>[]>(
    () => [
      {
        accessorKey: 'transaction_date',
        header: 'Date',
        cell: ({ getValue }) => {
          const dateStr = getValue<string>();
          if (!dateStr) return '—';
          const date = new Date(dateStr);
          if (isNaN(date.getTime())) return dateStr;
          return format(date, 'MMM dd, yyyy');
        },
      },
      {
        accessorKey: 'raw_description',
        header: 'Description',
      },
      {
        accessorKey: 'amount',
        header: 'Amount',
        cell: ({ getValue, row }) => {
          const amount = parseFloat(getValue<string>());
          const currency = row.original.currency;
          return (
            <span className={amount < 0 ? 'text-red-600' : 'text-green-600'}>
              {amount.toFixed(2)} {currency}
            </span>
          );
        },
      },
      {
        accessorKey: 'category_name',
        header: 'Category',
        cell: ({ getValue, row }) => {
          const [isEditing, setIsEditing] = useState(false);
          const currentCategory = getValue<string | undefined>();

          if (isEditing) {
            return (
              <select
                className="border border-slate-300 rounded px-2 py-1 text-sm"
                defaultValue={currentCategory || ''}
                onChange={(e) => {
                  // TODO: Implement category update mutation
                  console.log('Update category for', row.original.transaction_id, 'to', e.target.value);
                  setIsEditing(false);
                }}
                onBlur={() => setIsEditing(false)}
                autoFocus
              >
                <option value="">Uncategorized</option>
                {categories?.map((cat) => (
                  <option key={cat.category_id} value={cat.category_name}>
                    {cat.category_name}
                  </option>
                ))}
              </select>
            );
          }

          return (
            <button
              onClick={() => setIsEditing(true)}
              className="text-left px-2 py-1 rounded hover:bg-slate-100 text-sm"
            >
              {currentCategory || (
                <span className="text-slate-400 italic">Click to categorize</span>
              )}
            </button>
          );
        },
      },
      {
        accessorKey: 'balance_after',
        header: 'Balance',
        cell: ({ getValue, row }) => {
          const balance = getValue<string | undefined>();
          if (!balance) return <span className="text-slate-400">—</span>;
          return `${parseFloat(balance).toFixed(2)} ${row.original.currency}`;
        },
      },
    ],
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
              <Link href="/documents" className="px-4 py-2 text-slate-600 hover:text-slate-900 font-medium">
                Documents
              </Link>
              <Link href="/transactions" className="px-4 py-2 text-slate-900 font-medium border-b-2 border-slate-900">
                Transactions
              </Link>
            </div>
          </div>
        </div>
      </nav>

      <main className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-slate-900 mb-2">Transactions</h1>
          <p className="text-slate-600">View and categorize your transactions</p>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6 mb-6">
          <input
            type="text"
            placeholder="Search transactions..."
            value={globalFilter}
            onChange={(e) => setGlobalFilter(e.target.value)}
            className="w-full px-4 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-slate-900"
          />
        </div>

        <div className="bg-white rounded-lg shadow-md overflow-hidden">
          {transactionsLoading ? (
            <p className="p-6 text-slate-600">Loading transactions...</p>
          ) : transactions && transactions.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-slate-50 border-b border-slate-200">
                  {table.getHeaderGroups().map((headerGroup) => (
                    <tr key={headerGroup.id}>
                      {headerGroup.headers.map((header) => (
                        <th
                          key={header.id}
                          className="px-6 py-3 text-left text-xs font-medium text-slate-700 uppercase tracking-wider cursor-pointer hover:bg-slate-100"
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
                <tbody className="bg-white divide-y divide-slate-200">
                  {table.getRowModel().rows.map((row) => (
                    <tr key={row.id} className="hover:bg-slate-50">
                      {row.getVisibleCells().map((cell) => (
                        <td key={cell.id} className="px-6 py-4 whitespace-nowrap text-sm">
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="p-6 text-slate-600">No transactions found</p>
          )}
        </div>
      </main>
    </div>
  );
}
