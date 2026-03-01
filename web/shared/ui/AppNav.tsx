import Link from 'next/link';
import { AccountScopeSelect } from '@/shared/ui/AccountScopeSelect';

const linkBase = 'px-4 py-2 text-sm font-medium border-b-2';
const activeLink = `${linkBase} text-slate-900 border-slate-900`;
const inactiveLink = `${linkBase} text-slate-600 hover:text-slate-900 border-transparent`;

type AppNavProps = {
  active: 'dashboard' | 'documents' | 'transactions' | 'merchants';
};

export function AppNav({ active }: AppNavProps) {
  return (
    <nav className="border-b border-slate-100 bg-white">
      <div className="container mx-auto px-6 py-4">
        <div className="flex items-center justify-between">
          <Link href="/" className="text-xl font-semibold tracking-tight text-slate-900">
            Finance Tracker
          </Link>
          <div className="flex items-center gap-4">
            <Link href="/dashboard" className={active === 'dashboard' ? activeLink : inactiveLink}>
              Dashboard
            </Link>
            <Link href="/documents" className={active === 'documents' ? activeLink : inactiveLink}>
              Documents
            </Link>
            <Link href="/transactions" className={active === 'transactions' ? activeLink : inactiveLink}>
              Transactions
            </Link>
            <Link href="/merchants" className={active === 'merchants' ? activeLink : inactiveLink}>
              Merchants
            </Link>
            <AccountScopeSelect />
          </div>
        </div>
      </div>
    </nav>
  );
}
