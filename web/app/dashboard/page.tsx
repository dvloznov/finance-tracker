'use client';

import { useQuery } from '@tanstack/react-query';
import { apiClient, Transaction } from '@/lib/api-client';
import Link from 'next/link';
import { useMemo } from 'react';
import { format, startOfMonth, endOfMonth, subMonths } from 'date-fns';
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D'];

export default function DashboardPage() {
  const { data: transactions, isLoading } = useQuery({
    queryKey: ['transactions'],
    queryFn: () => apiClient.listTransactions(),
  });

  const stats = useMemo(() => {
    if (!transactions) return null;

    const totalIncome = transactions
      .filter((t) => parseFloat(t.amount) > 0)
      .reduce((sum, t) => sum + parseFloat(t.amount), 0);

    const totalExpenses = Math.abs(
      transactions
        .filter((t) => parseFloat(t.amount) < 0)
        .reduce((sum, t) => sum + parseFloat(t.amount), 0)
    );

    const netBalance = totalIncome - totalExpenses;

    return {
      totalIncome,
      totalExpenses,
      netBalance,
      transactionCount: transactions.length,
    };
  }, [transactions]);

  const monthlyData = useMemo(() => {
    if (!transactions) return [];

    const monthlyMap = new Map<string, { income: number; expenses: number }>();

    transactions.forEach((txn) => {
      const date = new Date(txn.transaction_date);
      const monthKey = format(date, 'MMM yyyy');
      const amount = parseFloat(txn.amount);

      if (!monthlyMap.has(monthKey)) {
        monthlyMap.set(monthKey, { income: 0, expenses: 0 });
      }

      const data = monthlyMap.get(monthKey)!;
      if (amount > 0) {
        data.income += amount;
      } else {
        data.expenses += Math.abs(amount);
      }
    });

    return Array.from(monthlyMap.entries())
      .map(([month, data]) => ({ month, ...data }))
      .slice(-6);
  }, [transactions]);

  const categoryData = useMemo(() => {
    if (!transactions) return [];

    const categoryMap = new Map<string, number>();

    transactions
      .filter((t) => parseFloat(t.amount) < 0 && t.category_name)
      .forEach((txn) => {
        const category = txn.category_name!;
        const amount = Math.abs(parseFloat(txn.amount));
        categoryMap.set(category, (categoryMap.get(category) || 0) + amount);
      });

    return Array.from(categoryMap.entries())
      .map(([name, value]) => ({ name, value }))
      .sort((a, b) => b.value - a.value)
      .slice(0, 6);
  }, [transactions]);

  return (
    <div className="min-h-screen bg-slate-50">
      <nav className="border-b bg-white shadow-sm">
        <div className="container mx-auto px-4 py-4">
          <div className="flex items-center justify-between">
            <Link href="/" className="text-2xl font-bold text-slate-900">
              Finance Tracker
            </Link>
            <div className="flex gap-4">
              <Link href="/dashboard" className="px-4 py-2 text-slate-900 font-medium border-b-2 border-slate-900">
                Dashboard
              </Link>
              <Link href="/documents" className="px-4 py-2 text-slate-600 hover:text-slate-900 font-medium">
                Documents
              </Link>
              <Link href="/transactions" className="px-4 py-2 text-slate-600 hover:text-slate-900 font-medium">
                Transactions
              </Link>
            </div>
          </div>
        </div>
      </nav>

      <main className="container mx-auto px-4 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-slate-900 mb-2">Dashboard</h1>
          <p className="text-slate-600">Overview of your financial activity</p>
        </div>

        {isLoading ? (
          <p className="text-slate-600">Loading data...</p>
        ) : stats ? (
          <>
            <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
              <div className="bg-white rounded-lg shadow-md p-6">
                <p className="text-sm text-slate-600 mb-1">Total Income</p>
                <p className="text-2xl font-bold text-green-600">
                  ${stats.totalIncome.toFixed(2)}
                </p>
              </div>
              <div className="bg-white rounded-lg shadow-md p-6">
                <p className="text-sm text-slate-600 mb-1">Total Expenses</p>
                <p className="text-2xl font-bold text-red-600">
                  ${stats.totalExpenses.toFixed(2)}
                </p>
              </div>
              <div className="bg-white rounded-lg shadow-md p-6">
                <p className="text-sm text-slate-600 mb-1">Net Balance</p>
                <p
                  className={`text-2xl font-bold ${
                    stats.netBalance >= 0 ? 'text-green-600' : 'text-red-600'
                  }`}
                >
                  ${stats.netBalance.toFixed(2)}
                </p>
              </div>
              <div className="bg-white rounded-lg shadow-md p-6">
                <p className="text-sm text-slate-600 mb-1">Transactions</p>
                <p className="text-2xl font-bold text-slate-900">{stats.transactionCount}</p>
              </div>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
              <div className="bg-white rounded-lg shadow-md p-6">
                <h2 className="text-xl font-semibold mb-4">Monthly Overview</h2>
                <ResponsiveContainer width="100%" height={300}>
                  <BarChart data={monthlyData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="month" />
                    <YAxis />
                    <Tooltip />
                    <Legend />
                    <Bar dataKey="income" fill="#10b981" name="Income" />
                    <Bar dataKey="expenses" fill="#ef4444" name="Expenses" />
                  </BarChart>
                </ResponsiveContainer>
              </div>

              <div className="bg-white rounded-lg shadow-md p-6">
                <h2 className="text-xl font-semibold mb-4">Spending by Category</h2>
                {categoryData.length > 0 ? (
                  <ResponsiveContainer width="100%" height={300}>
                    <PieChart>
                      <Pie
                        data={categoryData}
                        cx="50%"
                        cy="50%"
                        labelLine={false}
                        label={({ name, percent }) => `${name}: ${((percent || 0) * 100).toFixed(0)}%`}
                        outerRadius={80}
                        fill="#8884d8"
                        dataKey="value"
                      >
                        {categoryData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <Tooltip />
                    </PieChart>
                  </ResponsiveContainer>
                ) : (
                  <p className="text-slate-600 text-center py-12">
                    No categorized expenses yet
                  </p>
                )}
              </div>
            </div>

            <div className="bg-white rounded-lg shadow-md p-6">
              <h2 className="text-xl font-semibold mb-4">Recent Transactions</h2>
              {transactions && transactions.length > 0 ? (
                <div className="space-y-3">
                  {transactions.slice(0, 10).map((txn) => (
                    <div
                      key={txn.transaction_id}
                      className="flex items-center justify-between p-3 border border-slate-200 rounded-lg"
                    >
                      <div>
                        <p className="font-medium text-slate-900">{txn.raw_description}</p>
                        <p className="text-sm text-slate-600">
                          {format(new Date(txn.transaction_date), 'MMM dd, yyyy')}
                        </p>
                      </div>
                      <div className="text-right">
                        <p
                          className={`font-semibold ${
                            parseFloat(txn.amount) < 0 ? 'text-red-600' : 'text-green-600'
                          }`}
                        >
                          {parseFloat(txn.amount).toFixed(2)} {txn.currency}
                        </p>
                        {txn.category_name && (
                          <p className="text-sm text-slate-600">{txn.category_name}</p>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-slate-600">No transactions available</p>
              )}
            </div>
          </>
        ) : (
          <p className="text-slate-600">No data available</p>
        )}
      </main>
    </div>
  );
}
