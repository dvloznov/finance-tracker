import type { ReactNode } from 'react';
import { cardClass } from '@/lib/ui';

export type CardProps = {
  title?: string;
  description?: string;
  action?: ReactNode;
  children: ReactNode;
  className?: string;
};

export function Card({ title, description, action, children, className }: CardProps) {
  return (
    <div className={className ? `${cardClass} ${className}` : cardClass}>
      {(title || description || action) && (
        <div className="mb-6 space-y-2">
          <div className="flex items-center justify-between gap-4">
            {title && <h2 className="text-sm font-semibold text-slate-900">{title}</h2>}
            {action}
          </div>
          {description && <p className="text-sm text-slate-600">{description}</p>}
        </div>
      )}
      {children}
    </div>
  );
}
