'use client';

import { useQuery } from '@tanstack/react-query';
import { apiClient, Transaction } from '@/lib/api-client';
import Link from 'next/link';
import { useMemo } from 'react';
import { format, startOfMonth, endOfMonth, subMonths } from 'date-fns';
import { ResponsiveLine, SliceTooltipProps } from '@nivo/line';
import { ResponsiveBar, BarTooltipProps } from '@nivo/bar';
import { ResponsivePie } from '@nivo/pie';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D'];

export default function DashboardPage() {
  const { data: transactions, isLoading, error } = useQuery({
    queryKey: ['transactions'],
    queryFn: () => apiClient.listTransactions(),
  });

  const stats = useMemo(() => {
    if (!transactions || !Array.isArray(transactions)) return null;

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
    if (!transactions || !Array.isArray(transactions)) return [];

    const monthlyMap = new Map<string, { income: number; expenses: number }>();

    transactions.forEach((txn) => {
      // Handle civil.Date format from BigQuery
      const dateStr = typeof txn.transaction_date === 'string' 
        ? txn.transaction_date 
        : String(txn.transaction_date || '');
      
      if (!dateStr) return;
      
      const date = new Date(dateStr);
      if (isNaN(date.getTime())) return; // Skip invalid dates
      
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
  }, [transactions]);

  const balanceData = useMemo(() => {
    if (!transactions || !Array.isArray(transactions)) return [];

    // Sort transactions by date
    const sorted = [...transactions]
      .filter(txn => txn.transaction_date) // Filter out transactions without dates
      .sort((a, b) => {
        const dateA = new Date(a.transaction_date);
        const dateB = new Date(b.transaction_date);
        return dateA.getTime() - dateB.getTime();
      });

    // Use balance_after when available, otherwise calculate running balance
    const balanceHistory: Array<{ x: string; y: number }> = [];
    let calculatedBalance: number | null = null;

    // First pass: find if we have any balance_after values to work backwards from
    const txnsWithBalance = sorted.filter(txn => txn.balance_after);
    
    if (txnsWithBalance.length > 0) {
      // Work backwards from the last known balance
      const lastKnownBalance = parseFloat(txnsWithBalance[txnsWithBalance.length - 1].balance_after!);
      let workingBalance = lastKnownBalance;
      
      // Go through transactions in reverse to calculate earlier balances
      for (let i = sorted.length - 1; i >= 0; i--) {
        const txn = sorted[i];
        const date = new Date(txn.transaction_date);
        if (isNaN(date.getTime())) continue;
        
        // If this transaction has balance_after, use it
        if (txn.balance_after) {
          workingBalance = parseFloat(txn.balance_after);
        } else {
          // Calculate balance before this transaction
          workingBalance -= parseFloat(txn.amount);
        }
        
        balanceHistory.unshift({
          x: format(date, 'MMM dd'),
          y: workingBalance,
        });
      }
    } else {
      // Fallback: no balance_after values, calculate running balance from 0
      let runningBalance = 0;
      for (const txn of sorted) {
        const date = new Date(txn.transaction_date);
        if (isNaN(date.getTime())) continue;
        
        runningBalance += parseFloat(txn.amount);
        balanceHistory.push({
          x: format(date, 'MMM dd'),
          y: runningBalance,
        });
      }
    }

    // Sample every nth transaction if too many data points
    if (balanceHistory.length > 30) {
      const step = Math.ceil(balanceHistory.length / 30);
      return [{
        id: 'balance',
        data: balanceHistory.filter((_, i) => i % step === 0)
      }];
    }

    return [{
      id: 'balance',
      data: balanceHistory
    }];
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

        {error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-8">
            <p className="text-red-800">Error loading data: {error instanceof Error ? error.message : 'Unknown error'}</p>
          </div>
        )}

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

            <div className="bg-white rounded-lg shadow-md p-6 mb-8">
              <h2 className="text-xl font-semibold mb-4">Account Balance Over Time</h2>
              {balanceData.length > 0 && balanceData[0].data.length > 0 ? (
                <div style={{ height: 300 }}>
                  <ResponsiveLine
                    data={balanceData}
                    margin={{ top: 20, right: 20, bottom: 50, left: 60 }}
                    xScale={{ type: 'point' }}
                    yScale={{ type: 'linear', min: 'auto', max: 'auto' }}
                    curve="monotoneX"
                    axisTop={null}
                    axisRight={null}
                    axisBottom={{
                      tickSize: 5,
                      tickPadding: 5,
                      tickRotation: 0,
                      legend: 'Date',
                      legendOffset: 36,
                      legendPosition: 'middle'
                    }}
                    axisLeft={{
                      tickSize: 5,
                      tickPadding: 5,
                      tickRotation: 0,
                      legend: 'Balance',
                      legendOffset: -50,
                      legendPosition: 'middle'
                    }}
                    colors={['#3b82f6']}
                    lineWidth={2}
                    pointSize={0}
                    enableGridX={true}
                    enableGridY={true}
                    gridXValues={undefined}
                    gridYValues={undefined}
                    useMesh={true}
                    legends={[
                      {
                        anchor: 'top-right',
                        direction: 'row',
                        justify: false,
                        translateX: 0,
                        translateY: -20,
                        itemsSpacing: 0,
                        itemDirection: 'left-to-right',
                        itemWidth: 80,
                        itemHeight: 20,
                        itemOpacity: 0.75,
                        symbolSize: 12,
                        symbolShape: 'circle',
                      }
                    ]}
                    tooltip={(point) => (
                      <div className="bg-white px-3 py-2 shadow-lg rounded border border-slate-200">
                        <div className="font-medium text-slate-900">{String(point.point.data.x)}</div>
                        <div className="text-sm text-slate-600">
                          Balance: <span className="font-semibold">${Number(point.point.data.y).toFixed(2)}</span>
                        </div>
                      </div>
                    )}
                  />
                </div>
              ) : (
                <p className="text-slate-600 text-center py-12">
                  No transaction history available
                </p>
              )}
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
              <div className="bg-white rounded-lg shadow-md p-6">
                <h2 className="text-xl font-semibold mb-4">Monthly Overview</h2>
                {monthlyData.length > 0 ? (
                  <div style={{ height: 300 }}>
                    <ResponsiveBar
                      data={monthlyData}
                      keys={['income', 'expenses']}
                      indexBy="month"
                      margin={{ top: 20, right: 130, bottom: 50, left: 60 }}
                      padding={0.3}
                      valueScale={{ type: 'linear' }}
                      indexScale={{ type: 'band', round: true }}
                      colors={({ id }) => id === 'income' ? '#10b981' : '#ef4444'}
                      borderColor={{
                        from: 'color',
                        modifiers: [['darker', 1.6]]
                      }}
                      axisTop={null}
                      axisRight={null}
                      axisBottom={{
                        tickSize: 5,
                        tickPadding: 5,
                        tickRotation: 0,
                        legend: 'Month',
                        legendPosition: 'middle',
                        legendOffset: 40
                      }}
                      axisLeft={{
                        tickSize: 5,
                        tickPadding: 5,
                        tickRotation: 0,
                        legend: 'Amount',
                        legendPosition: 'middle',
                        legendOffset: -50
                      }}
                      enableGridY={true}
                      labelSkipWidth={12}
                      labelSkipHeight={12}
                      legends={[
                        {
                          dataFrom: 'keys',
                          anchor: 'bottom-right',
                          direction: 'column',
                          justify: false,
                          translateX: 120,
                          translateY: 0,
                          itemsSpacing: 2,
                          itemWidth: 100,
                          itemHeight: 20,
                          itemDirection: 'left-to-right',
                          itemOpacity: 0.85,
                          symbolSize: 20,
                        }
                      ]}
                      tooltip={({ id, value, indexValue, color }) => (
                        <div className="bg-white px-3 py-2 shadow-lg rounded border border-slate-200">
                          <div className="font-medium text-slate-900">{indexValue}</div>
                          <div className="flex items-center gap-2 text-sm text-slate-600">
                            <div className="w-3 h-3 rounded" style={{ backgroundColor: color }} />
                            <span className="capitalize">{id}:</span>
                            <span className="font-semibold">${Number(value).toFixed(2)}</span>
                          </div>
                        </div>
                      )}
                    />
                  </div>
                ) : (
                  <p className="text-slate-600 text-center py-12">No monthly data available</p>
                )}
              </div>

              <div className="bg-white rounded-lg shadow-md p-6">
                <h2 className="text-xl font-semibold mb-4">Spending by Category</h2>
                {categoryData.length > 0 ? (
                  <div style={{ height: 300 }}>
                    <ResponsivePie
                      data={categoryData}
                      margin={{ top: 20, right: 20, bottom: 20, left: 20 }}
                      innerRadius={0}
                      padAngle={0.7}
                      cornerRadius={3}
                      activeOuterRadiusOffset={8}
                      colors={COLORS}
                      borderWidth={1}
                      borderColor={{
                        from: 'color',
                        modifiers: [['darker', 0.2]]
                      }}
                      arcLinkLabelsSkipAngle={10}
                      arcLinkLabelsTextColor="#333333"
                      arcLinkLabelsThickness={2}
                      arcLinkLabelsColor={{ from: 'color' }}
                      arcLabelsSkipAngle={10}
                      arcLabelsTextColor={{
                        from: 'color',
                        modifiers: [['darker', 2]]
                      }}
                      arcLabel={(d) => `${((d.value / categoryData.reduce((sum, c) => sum + c.value, 0)) * 100).toFixed(0)}%`}
                      tooltip={({ datum }) => (
                        <div className="bg-white px-3 py-2 shadow-lg rounded border border-slate-200">
                          <div className="flex items-center gap-2">
                            <div className="w-3 h-3 rounded" style={{ backgroundColor: datum.color }} />
                            <span className="font-medium text-slate-900">{datum.label}</span>
                          </div>
                          <div className="text-sm text-slate-600 mt-1">
                            Amount: <span className="font-semibold">${Number(datum.value).toFixed(2)}</span>
                          </div>
                          <div className="text-sm text-slate-600">
                            Percentage: <span className="font-semibold">{((datum.value / categoryData.reduce((sum, c) => sum + c.value, 0)) * 100).toFixed(1)}%</span>
                          </div>
                        </div>
                      )}
                    />
                  </div>
                ) : (
                  <p className="text-slate-600 text-center py-12">
                    No categorized expenses yet
                  </p>
                )}
              </div>
            </div>

            <div className="bg-white rounded-lg shadow-md p-6">
              <h2 className="text-xl font-semibold mb-4">Recent Transactions</h2>
              {transactions && Array.isArray(transactions) && transactions.length > 0 ? (
                <div className="space-y-3">
                  {transactions.slice(0, 10).map((txn, idx) => (
                    <div
                      key={txn.transaction_id || `txn-${idx}`}
                      className="flex items-center justify-between p-3 border border-slate-200 rounded-lg"
                    >
                      <div>
                        <p className="font-medium text-slate-900">{txn.raw_description}</p>
                        <p className="text-sm text-slate-600">
                          {txn.transaction_date && (() => {
                            const date = new Date(txn.transaction_date);
                            return !isNaN(date.getTime()) ? format(date, 'MMM dd, yyyy') : txn.transaction_date;
                          })()}
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
