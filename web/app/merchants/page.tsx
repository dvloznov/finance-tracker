'use client';

import { useState, useMemo } from 'react';
import { AppNav } from '@/shared/ui/AppNav';
import { cardClass } from '@/lib/ui';
import { useMerchants } from '@/features/merchants/hooks/useMerchants';
import { useUpdateMerchantCategory } from '@/features/merchants/hooks/useUpdateMerchantCategory';
import { useMergeMerchant } from '@/features/merchants/hooks/useMergeMerchant';
import { useUnmergeMerchant } from '@/features/merchants/hooks/useUnmergeMerchant';
import { useCategories } from '@/features/categories/hooks/useCategories';
import type { Category, Merchant } from '@/shared/types/api';

// The categories table is denormalized: one row per (category, subcategory) pair.
// A merchant's category_id points to one of these leaf rows.
// We present two selects: one for the parent category name, one for the subcategory.

function CategorySelects({
  merchant,
  categories,
}: {
  merchant: Merchant;
  categories: Category[];
}) {
  const { mutate, isPending } = useUpdateMerchantCategory();

  // Find the row that matches the merchant's current category_id.
  const currentRow = categories.find((c) => c.category_id === merchant.category_id);
  const [selectedCategoryName, setSelectedCategoryName] = useState(
    currentRow?.category_name ?? ''
  );
  const [selectedCategoryId, setSelectedCategoryId] = useState(merchant.category_id);

  // Unique parent category names, sorted alphabetically.
  // Include merchant's current category if not in the list (handles stale/missing data).
  const parentNames = useMemo(() => {
    const seen = new Set<string>();
    const names: string[] = [];
    for (const cat of categories) {
      const name = cat.category_name;
      if (name && !seen.has(name)) {
        seen.add(name);
        names.push(name);
      }
    }
    if (merchant.category_name && !seen.has(merchant.category_name)) {
      names.push(merchant.category_name);
    }
    return names.sort((a, b) => a.localeCompare(b));
  }, [categories, merchant.category_name]);

  // Subcategory rows that belong to the selected parent category.
  const subcategoryRows = useMemo(
    () => categories.filter((c) => c.category_name === selectedCategoryName),
    [categories, selectedCategoryName]
  );

  const selectClass = `w-full text-sm rounded-lg border px-2 py-1 focus:outline-none focus:ring-2 focus:ring-slate-900/10 focus:border-slate-300 ${
    isPending
      ? 'border-slate-200 bg-slate-50 text-slate-400 cursor-wait'
      : 'border-slate-200 bg-white text-slate-900 hover:border-slate-300'
  }`;

  const handleCategoryChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newName = e.target.value;
    setSelectedCategoryName(newName);
    // Auto-select the first subcategory row for the new parent and save immediately.
    const firstRow = categories.find((c) => c.category_name === newName);
    if (firstRow) {
      setSelectedCategoryId(firstRow.category_id);
      mutate({ merchantId: merchant.merchant_id, categoryId: firstRow.category_id });
    }
  };

  const handleSubcategoryChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const newId = e.target.value;
    setSelectedCategoryId(newId);
    mutate({ merchantId: merchant.merchant_id, categoryId: newId });
  };

  return (
    <div className="flex flex-col gap-1.5 min-w-[160px]">
      {/* Parent category select */}
      <select
        value={selectedCategoryName}
        onChange={handleCategoryChange}
        disabled={isPending}
        className={selectClass}
      >
        {selectedCategoryName === '' && (
          <option value="" disabled>— pick category —</option>
        )}
        {parentNames.map((name) => (
          <option key={name} value={name}>
            {name}
          </option>
        ))}
      </select>

      {/* Subcategory select — only shown when the selected category has subcategories */}
      {subcategoryRows.length > 1 || (subcategoryRows.length === 1 && subcategoryRows[0].subcategory_name) ? (
        <select
          value={selectedCategoryId}
          onChange={handleSubcategoryChange}
          disabled={isPending}
          className={selectClass}
        >
          {subcategoryRows.map((row) => (
            <option key={row.category_id} value={row.category_id}>
              {row.subcategory_name ?? '(general)'}
            </option>
          ))}
        </select>
      ) : null}
    </div>
  );
}

function MergeModal({
  variant,
  merchants,
  onSelect,
  onClose,
}: {
  variant: Merchant;
  merchants: Merchant[];
  onSelect: (canonical: Merchant) => void;
  onClose: () => void;
}) {
  const [search, setSearch] = useState('');
  const candidates = useMemo(() => {
    return merchants.filter(
      (m) =>
        m.merchant_id !== variant.merchant_id &&
        !m.merged_into_merchant_id &&
        (search.trim() === '' ||
          m.merchant_name.toLowerCase().includes(search.toLowerCase()) ||
          m.normalized_name.toLowerCase().includes(search.toLowerCase()))
    );
  }, [merchants, variant.merchant_id, search]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={onClose}>
      <div
        className="bg-white rounded-xl shadow-xl max-w-md w-full mx-4 max-h-[80vh] flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-4 border-b border-slate-100">
          <h3 className="font-semibold text-slate-900">Merge &quot;{variant.merchant_name}&quot; into</h3>
          <p className="text-xs text-slate-500 mt-1">Select the canonical merchant to merge into</p>
        </div>
        <div className="p-4 border-b border-slate-100">
          <input
            type="text"
            placeholder="Search merchants..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full px-3 py-2 text-sm border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-slate-900/10"
          />
        </div>
        <div className="overflow-y-auto flex-1 p-2">
          {candidates.length === 0 ? (
            <p className="p-4 text-sm text-slate-500 text-center">No merchants match</p>
          ) : (
            <ul className="space-y-0.5">
              {candidates.map((m) => (
                <li key={m.merchant_id}>
                  <button
                    type="button"
                    onClick={() => onSelect(m)}
                    className="w-full text-left px-3 py-2 rounded-lg hover:bg-slate-100 text-sm"
                  >
                    <span className="font-medium text-slate-900">{m.merchant_name}</span>
                    <span className="text-slate-400 ml-2 text-xs">{m.normalized_name}</span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </div>
        <div className="p-4 border-t border-slate-100">
          <button
            type="button"
            onClick={onClose}
            className="text-sm text-slate-600 hover:text-slate-900"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}

export default function MerchantsPage() {
  const [globalFilter, setGlobalFilter] = useState('');
  const [mergeVariant, setMergeVariant] = useState<Merchant | null>(null);
  const { data: merchants, isLoading } = useMerchants();
  const { data: categories } = useCategories();
  const { mutate: mergeMerchant, isPending: isMergePending } = useMergeMerchant();
  const { mutate: unmergeMerchant, isPending: isUnmergePending } = useUnmergeMerchant();

  const filtered = useMemo(() => {
    if (!merchants) return [];
    const q = globalFilter.trim().toLowerCase();
    const list = q
      ? merchants.filter(
          (m) =>
            m.merchant_name.toLowerCase().includes(q) ||
            m.normalized_name.toLowerCase().includes(q) ||
            (m.category_name ?? '').toLowerCase().includes(q)
        )
      : merchants;
    // Merged merchants at the bottom, canonical first (preserve transaction_count order within each group)
    return [...list].sort((a, b) => {
      const aMerged = !!a.merged_into_merchant_id ? 1 : 0;
      const bMerged = !!b.merged_into_merchant_id ? 1 : 0;
      return aMerged - bMerged;
    });
  }, [merchants, globalFilter]);

  return (
    <div className="min-h-screen bg-slate-50">
      <AppNav active="merchants" />

      <main className="container mx-auto px-6 py-8">
        <div className="space-y-6">
          <div className="space-y-1">
            <h1 className="text-3xl font-semibold tracking-tight text-slate-900">Merchants</h1>
            <p className="text-sm text-slate-600">
              All unique merchants found in your statements. Change a category to update it across
              all transactions and dashboard analytics.
            </p>
          </div>

          <div className={cardClass}>
            <input
              type="text"
              placeholder="Search merchants..."
              value={globalFilter}
              onChange={(e) => setGlobalFilter(e.target.value)}
              className="w-full px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-slate-900/10 focus:border-slate-300 text-sm"
            />
          </div>

          <div className="bg-white rounded-2xl ring-1 ring-black/5 shadow-sm overflow-hidden">
            {isLoading ? (
              <p className="p-6 text-sm text-slate-600">Loading merchants...</p>
            ) : filtered.length === 0 ? (
              <p className="p-6 text-sm text-slate-600">
                {globalFilter ? 'No merchants match your search.' : 'No merchants found.'}
              </p>
            ) : (
              <>
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead className="bg-slate-50 border-b border-slate-100">
                      <tr>
                        <th className="px-6 py-3 text-left text-[11px] font-medium text-slate-500 uppercase tracking-wider">
                          Merchant
                        </th>
                        <th className="px-6 py-3 text-left text-[11px] font-medium text-slate-500 uppercase tracking-wider">
                          Category
                        </th>
                        <th className="px-6 py-3 text-right text-[11px] font-medium text-slate-500 uppercase tracking-wider">
                          Transactions
                        </th>
                        <th className="px-6 py-3 text-right text-[11px] font-medium text-slate-500 uppercase tracking-wider">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-slate-100">
                      {filtered.map((merchant) => (
                        <tr
                          key={merchant.merchant_id}
                          className={
                            merchant.merged_into_merchant_id
                              ? 'bg-slate-50/60 hover:bg-slate-100/60 text-slate-500'
                              : 'hover:bg-slate-50'
                          }
                        >
                          <td className="px-6 py-4">
                            <p
                              className={
                                merchant.merged_into_merchant_id
                                  ? 'text-sm font-medium text-slate-500'
                                  : 'text-sm font-medium text-slate-900'
                              }
                            >
                              {merchant.merchant_name}
                              {merchant.merged_into_merchant_id && (
                                <span className="inline-flex items-center ml-2 px-1.5 py-0.5 rounded text-[10px] font-medium bg-slate-100 text-slate-600">
                                  Merged into {merchant.canonical_merchant_name ?? merchant.merged_into_merchant_id}
                                </span>
                              )}
                            </p>
                            <p className="text-xs text-slate-400 mt-0.5">{merchant.normalized_name}</p>
                          </td>
                          <td className="px-6 py-4">
                            {categories ? (
                              <CategorySelects merchant={merchant} categories={categories} />
                            ) : (
                              <div>
                                <span className="text-sm text-slate-400">{merchant.category_name ?? '—'}</span>
                                {merchant.subcategory_name && (
                                  <p className="text-xs text-slate-400 mt-0.5">{merchant.subcategory_name}</p>
                                )}
                              </div>
                            )}
                          </td>
                          <td className="px-6 py-4 text-right">
                            <span className="text-sm tabular-nums text-slate-600">{merchant.transaction_count}</span>
                          </td>
                          <td className="px-6 py-4 text-right">
                            {merchant.merged_into_merchant_id ? (
                              <button
                                type="button"
                                onClick={() => unmergeMerchant(merchant.merchant_id)}
                                disabled={isUnmergePending}
                                className="text-xs text-slate-600 hover:text-slate-900 hover:underline"
                              >
                                Unmerge
                              </button>
                            ) : (
                              <button
                                type="button"
                                onClick={() => setMergeVariant(merchant)}
                                disabled={isMergePending}
                                className="text-xs text-slate-600 hover:text-slate-900 hover:underline"
                              >
                                Merge into…
                              </button>
                            )}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                <div className="px-6 py-3 border-t border-slate-100 text-xs text-slate-400">
                  {filtered.length} merchant{filtered.length !== 1 ? 's' : ''}
                  {globalFilter ? ` matching "${globalFilter}"` : ''}
                </div>
              </>
            )}
          </div>
        </div>
      </main>

      {mergeVariant && merchants && (
        <MergeModal
          variant={mergeVariant}
          merchants={merchants}
          onSelect={(canonical) => {
            mergeMerchant(
              { merchantId: mergeVariant.merchant_id, canonicalMerchantId: canonical.merchant_id },
              { onSuccess: () => setMergeVariant(null) }
            );
          }}
          onClose={() => setMergeVariant(null)}
        />
      )}
    </div>
  );
}
