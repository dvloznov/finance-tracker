import type { TransactionVM } from '@/features/transactions/types';

export type CategoryDatum = {
  id: string;
  label: string;
  value: number;
};

export function getCategoryTotals(transactions?: TransactionVM[] | null): CategoryDatum[] {
  if (!transactions || !Array.isArray(transactions)) return [];

  const categoryMap = new Map<string, number>();

  transactions
    .filter((t) => parseFloat(t.amount) < 0 && t.category_name)
    .forEach((txn) => {
      const category = txn.category_name!;
      const amount = Math.abs(parseFloat(txn.amount));
      categoryMap.set(category, (categoryMap.get(category) || 0) + amount);
    });

  return Array.from(categoryMap.entries())
    .map(([id, value]) => ({ id, label: id, value }))
    .sort((a, b) => b.value - a.value)
    .slice(0, 6);
}

export function getSubcategoryTotals(
  transactions?: TransactionVM[] | null,
  selectedCategory?: string | null
): CategoryDatum[] {
  if (!transactions || !Array.isArray(transactions) || !selectedCategory) return [];

  const subcategoryMap = new Map<string, number>();

  transactions
    .filter((t) => parseFloat(t.amount) < 0 && t.category_name === selectedCategory)
    .forEach((txn) => {
      const subcategory = txn.subcategory_name || 'Uncategorized';
      const amount = Math.abs(parseFloat(txn.amount));
      subcategoryMap.set(subcategory, (subcategoryMap.get(subcategory) || 0) + amount);
    });

  return Array.from(subcategoryMap.entries())
    .map(([id, value]) => ({ id, label: id, value }))
    .sort((a, b) => b.value - a.value)
    .slice(0, 6);
}
