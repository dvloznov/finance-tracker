'use client';

import { useState } from 'react';
import type { Category } from '@/lib/api-client';

type CategoryCellProps = {
  categories: Category[] | undefined;
  currentCategory: string | undefined;
  transactionId: string;
};

export function CategoryCell({ categories, currentCategory, transactionId }: CategoryCellProps) {
  const [isEditing, setIsEditing] = useState(false);

  if (isEditing) {
    return (
      <select
        className="border border-slate-300 rounded px-2 py-1 text-sm"
        defaultValue={currentCategory || ''}
        onChange={(e) => {
          // TODO: Implement category update mutation
          console.log('Update category for', transactionId, 'to', e.target.value);
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
}
