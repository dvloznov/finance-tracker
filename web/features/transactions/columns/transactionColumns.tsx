import type { ColumnDef } from '@tanstack/react-table';
import type { Category } from '@/shared/types/api';
import { formatShortDate } from '@/shared/formatters/date';
import type { TransactionVM } from '@/features/transactions/types';
import { AmountCell } from '@/features/transactions/components/cells/AmountCell';
import { CategoryCell } from '@/features/transactions/components/cells/CategoryCell';

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
        return (
          <AmountCell
            amount={getValue<string>()}
            currency={row.original.currency}
          />
        );
      },
    },
    {
      accessorKey: 'category_name',
      header: 'Category',
      cell: ({ getValue, row }) => {
        return (
          <CategoryCell
            categories={categories}
            currentCategory={getValue<string | undefined>()}
            transactionId={row.original.transaction_id}
          />
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
