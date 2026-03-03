'use client';

import { Trash2, X } from 'lucide-react';

export type ConfirmModalProps = {
  show: boolean;
  title: string;
  message: string;
  isPending: boolean;
  error: string | null;
  confirmLabel?: string;
  onCancel: () => void;
  onConfirm: () => void;
};

export function ConfirmModal({
  show,
  title,
  message,
  isPending,
  error,
  confirmLabel = 'Delete',
  onCancel,
  onConfirm,
}: ConfirmModalProps) {
  if (!show) return null;

  return (
    <div className="fixed inset-0 bg-black/40 flex items-center justify-center z-50">
      <div className="bg-white rounded-2xl ring-1 ring-black/5 shadow-sm max-w-md w-full mx-4">
        <div className="p-6">
          <div className="flex items-start gap-4">
            <div className="flex-shrink-0">
              <div className="w-12 h-12 rounded-full bg-red-100 flex items-center justify-center">
                <Trash2 className="text-red-600" size={24} />
              </div>
            </div>
            <div className="flex-1">
              <h3 className="text-base font-semibold text-slate-900 mb-2">{title}</h3>
              <p className="text-sm text-slate-600 mb-4">{message}</p>

              {error && (
                <div className="mb-4 p-3 bg-red-50 ring-1 ring-red-200 rounded-lg">
                  <p className="text-sm text-red-800 flex items-start gap-2">
                    <X size={16} className="flex-shrink-0 mt-0.5" />
                    <span>{error}</span>
                  </p>
                </div>
              )}
            </div>
          </div>
        </div>
        <div className="bg-slate-50 px-6 py-4 flex gap-3 justify-end rounded-b-2xl">
          <button
            onClick={onCancel}
            disabled={isPending}
            className="px-4 py-2 text-sm font-medium text-slate-700 bg-white border border-slate-200 rounded-lg hover:bg-slate-50 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            disabled={isPending}
            className="px-4 py-2 text-sm font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            {isPending ? (
              <>
                <span className="inline-block w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                Deleting...
              </>
            ) : (
              confirmLabel
            )}
          </button>
        </div>
      </div>
    </div>
  );
}
