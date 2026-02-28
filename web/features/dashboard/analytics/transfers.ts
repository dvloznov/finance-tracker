import type { TransactionVM } from '@/features/transactions/types';

const MS_PER_DAY = 24 * 60 * 60 * 1000;
const TRANSFER_WINDOW_DAYS = 3;

/**
 * Detect inter-account transfer pairs in a set of transactions and return
 * the transaction_ids of both legs of each matched pair.
 *
 * Two transactions are considered a transfer when ALL of the following hold:
 *  - They belong to different accounts
 *  - Absolute amounts are equal (compared as 2-decimal strings to avoid float noise)
 *  - Same currency
 *  - Transaction dates are within ±3 days of each other
 *  - One is a debit (amount < 0) and the other is a credit (amount > 0)
 *
 * The algorithm is greedy: each credit can only be matched to one debit.
 * When multiple credits match the same debit, the one closest in date is chosen.
 *
 * Only meaningful in "all accounts" mode — single-account views cannot have
 * cross-account transfers.
 */
export function detectTransferIds(transactions: TransactionVM[]): Set<string> {
  const transferIds = new Set<string>();

  const debits = transactions.filter((t) => parseFloat(t.amount) < 0);
  // Credits pool — entries are removed once matched.
  const availableCredits = transactions
    .filter((t) => parseFloat(t.amount) > 0)
    .map((t) => ({ txn: t, matched: false }));

  for (const debit of debits) {
    const debitDate = new Date(debit.transaction_date).getTime();
    if (isNaN(debitDate)) continue;

    const debitAbs = Math.abs(parseFloat(debit.amount)).toFixed(2);

    let bestMatch: { index: number; diffMs: number } | null = null;

    for (let i = 0; i < availableCredits.length; i++) {
      const entry = availableCredits[i];
      if (entry.matched) continue;

      const credit = entry.txn;

      // Must be on a different account.
      if (credit.account_id === debit.account_id) continue;

      // Same currency.
      if (credit.currency !== debit.currency) continue;

      // Absolute amounts must match.
      const creditAbs = Math.abs(parseFloat(credit.amount)).toFixed(2);
      if (creditAbs !== debitAbs) continue;

      // Within the date window.
      const creditDate = new Date(credit.transaction_date).getTime();
      if (isNaN(creditDate)) continue;

      const diffMs = Math.abs(creditDate - debitDate);
      if (diffMs > TRANSFER_WINDOW_DAYS * MS_PER_DAY) continue;

      // Pick the closest-in-time match.
      if (bestMatch === null || diffMs < bestMatch.diffMs) {
        bestMatch = { index: i, diffMs };
      }
    }

    if (bestMatch !== null) {
      const matched = availableCredits[bestMatch.index];
      matched.matched = true;
      transferIds.add(debit.transaction_id);
      transferIds.add(matched.txn.transaction_id);
    }
  }

  return transferIds;
}
