import { useState } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import type { Category } from '@/lib/api-client';
import { formatCurrencyWithCode } from '@/shared/formatters/currency';
import { formatShortDate } from '@/shared/formatters/date';
import type { TransactionVM } from '@/features/transactions/types';

export function getTransactionColumns(categories: Category[] | undefined): ColumnDef<TransactionVM>[] {
  return [
    {
      accessorKey: 'transaction_date',
      header: 'Date',
      cell: ({ getValue }) => {
        const dateStr = getValue<string>();
        if (!dateStr) return <span className="text-slate-500">—</span>;
        const formatted = formatShortDate(dateStr);
        if (formatted === dateStr) return <span className="text-slate-700">{dateStr}</span>;
        return <span className="text-slate-900 font-medium">{formatted}</span>;
      },
    },
    {
      accessorKey: 'raw_description',
      header: 'Description',
      cell: ({ getValue }) => (
        <span className="text-slate-900">{getValue<string>()}</span>
      ),
    },
    {
      accessorKey: 'amount',
      header: 'Amount',
      cell: ({ getValue, row }) => {
        const amount = parseFloat(getValue<string>());
        const currency = row.original.currency;
        return (
          <span className={amount < 0 ? 'text-red-600 font-semibold' : 'text-green-600 font-semibold'}>
            {formatCurrencyWithCode(amount, currency)}
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
            className="text-left px-2 py-1 rounded hover:bg-slate-100 text-sm text-slate-900"
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
        return (
          <span className="text-slate-900">
            {formatCurrencyWithCode(parseFloat(balance), row.original.currency)}
          </span>
        );
      },
    },
  ];
}
