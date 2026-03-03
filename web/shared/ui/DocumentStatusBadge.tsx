/**
 * Shared badge for document parsing status.
 * Handles COMPLETED, FAILED, PROCESSING, RUNNING, PENDING.
 * Worker uses PROCESSING; parsing_runs uses RUNNING - both map to the same "in progress" style.
 */
export function DocumentStatusBadge({ status }: { status?: string | null }) {
  const s = status ?? '';
  const isCompleted = s === 'COMPLETED';
  const isFailed = s === 'FAILED';
  const isInProgress = s === 'PROCESSING' || s === 'RUNNING';
  const isPending = s === 'PENDING';

  const className = `px-3 py-1 rounded-full text-xs font-medium inline-block ${
    isCompleted
      ? 'bg-green-100 text-green-800'
      : isFailed
        ? 'bg-red-100 text-red-800'
        : isInProgress
          ? 'bg-blue-100 text-blue-800'
          : isPending
            ? 'bg-yellow-100 text-yellow-800'
            : 'bg-slate-100 text-slate-700'
  }`;

  const displayLabel = isInProgress ? 'Processing' : s || '—';

  return <span className={className}>{displayLabel}</span>;
}
