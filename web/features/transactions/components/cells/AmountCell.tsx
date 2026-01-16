'use client';

import { formatCurrencyWithCode } from '@/shared/formatters/currency';

type AmountCellProps = {
  amount: string;
  currency: string;
};

export function AmountCell({ amount, currency }: AmountCellProps) {
  const parsedAmount = parseFloat(amount);

  return (
    <span className={parsedAmount < 0 ? 'text-red-600 font-semibold' : 'text-green-600 font-semibold'}>
      {formatCurrencyWithCode(parsedAmount, currency)}
    </span>
  );
}
