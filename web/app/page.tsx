import Link from 'next/link';

export default function Home() {
  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-50 to-slate-100">
      <nav className="border-b bg-white shadow-sm">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <h1 className="text-2xl font-bold text-slate-900">Finance Tracker</h1>
            <div className="flex gap-4">
              <Link
                href="/dashboard"
                className="px-4 py-2 text-slate-600 hover:text-slate-900 font-medium"
              >
                Dashboard
              </Link>
              <Link
                href="/transactions"
                className="px-4 py-2 text-slate-600 hover:text-slate-900 font-medium"
              >
                Transactions
              </Link>
              <Link
                href="/accounts"
                className="px-4 py-2 text-slate-600 hover:text-slate-900 font-medium"
              >
                Accounts
              </Link>
            </div>
          </div>
        </div>
      </nav>

      <main className="container mx-auto px-4 py-12">
        <div className="max-w-4xl mx-auto text-center">
          <h2 className="text-4xl font-bold text-slate-900 mb-4">
            Welcome to Finance Tracker
          </h2>
          <p className="text-xl text-slate-600 mb-8">
            AI-powered document parsing for your bank statements
          </p>
          
          <div className="grid md:grid-cols-3 gap-6 mt-12">
            <Link
              href="/dashboard"
              className="p-6 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow"
            >
              <div className="text-3xl mb-3">📊</div>
              <h3 className="text-lg font-semibold mb-2">Dashboard</h3>
              <p className="text-slate-600 text-sm">
                View spending charts and financial overview
              </p>
            </Link>

            <Link
              href="/transactions"
              className="p-6 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow"
            >
              <div className="text-3xl mb-3">💰</div>
              <h3 className="text-lg font-semibold mb-2">Transactions</h3>
              <p className="text-slate-600 text-sm">
                View and categorize your transactions
              </p>
            </Link>

            <Link
              href="/accounts"
              className="p-6 bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow"
            >
              <div className="text-3xl mb-3">🏦</div>
              <h3 className="text-lg font-semibold mb-2">Accounts</h3>
              <p className="text-slate-600 text-sm">
                Manage institutions, accounts, and upload statements
              </p>
            </Link>
          </div>
        </div>
      </main>
    </div>
  );
}
