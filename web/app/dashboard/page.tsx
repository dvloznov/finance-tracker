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
const cardClass = 'bg-white rounded-2xl ring-1 ring-black/5 shadow-sm p-6';
const statLabelClass = 'text-[11px] font-medium uppercase tracking-wider text-slate-500';
const statValueClass = 'text-4xl font-semibold tabular-nums tracking-tight';
const currencyClass = 'text-slate-400 text-2xl font-medium';

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
      <nav className="border-b border-slate-100 bg-white">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <Link href="/" className="text-xl font-semibold tracking-tight text-slate-900">
              Finance Tracker
            </Link>
            <div className="flex gap-1">
              <Link href="/dashboard" className="px-4 py-2 text-sm font-medium text-slate-900 border-b-2 border-slate-900">
                Dashboard
              </Link>
              <Link href="/documents" className="px-4 py-2 text-sm font-medium text-slate-600 hover:text-slate-900 border-b-2 border-transparent">
                Documents
              </Link>
              <Link href="/transactions" className="px-4 py-2 text-sm font-medium text-slate-600 hover:text-slate-900 border-b-2 border-transparent">
                Transactions
              </Link>
            </div>
          </div>
        </div>
      </nav>

      <main className="container mx-auto px-6 py-8">
        <div className="space-y-6">
          <div className="space-y-1">
            <h1 className="text-3xl font-semibold tracking-tight text-slate-900">Dashboard</h1>
            <p className="text-sm text-slate-600">Overview of your financial activity</p>
          </div>

          {error && (
            <div className="bg-red-50 ring-1 ring-red-200 rounded-2xl p-4">
              <p className="text-sm text-red-800">Error loading data: {error instanceof Error ? error.message : 'Unknown error'}</p>
            </div>
          )}

          {isLoading ? (
            <p className="text-sm text-slate-600">Loading data...</p>
          ) : stats ? (
            <div className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className={`${cardClass} flex flex-col gap-3`}>
                  <p className={statLabelClass}>Total Income</p>
                  <p className={`${statValueClass} text-emerald-600 flex items-baseline gap-1`}>
                    <span className={currencyClass}>£</span>
                    <span>{Math.round(stats.totalIncome).toLocaleString()}</span>
                  </p>
                </div>
                <div className={`${cardClass} flex flex-col gap-3`}>
                  <p className={statLabelClass}>Total Expenses</p>
                  <p className={`${statValueClass} text-rose-600 flex items-baseline gap-1`}>
                    <span className={currencyClass}>£</span>
                    <span>{Math.round(stats.totalExpenses).toLocaleString()}</span>
                  </p>
                </div>
                <div className={`${cardClass} flex flex-col gap-3`}>
                  <p className={statLabelClass}>Net Balance</p>
                  <p className={`${statValueClass} flex items-baseline gap-1`}>
                    <span className={currencyClass}>£</span>
                    <span className={stats.netBalance >= 0 ? 'text-emerald-600' : 'text-rose-600'}>
                      {Math.round(stats.netBalance).toLocaleString()}
                    </span>
                  </p>
                </div>
              </div>

              <div className={cardClass}>
                <h2 className="text-sm font-semibold text-slate-900 mb-6">Account Balance Over Time</h2>
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
                        tickSize: 0,
                        tickPadding: 8,
                        tickRotation: 0,
                        legend: '',
                        legendOffset: 36,
                        legendPosition: 'middle'
                      }}
                      axisLeft={{
                        tickSize: 0,
                        tickPadding: 8,
                        tickRotation: 0,
                        legend: '',
                        legendOffset: -50,
                        legendPosition: 'middle'
                      }}
                      colors={['#3b82f6']}
                      lineWidth={2}
                      pointSize={0}
                      enableGridX={false}
                      enableGridY={true}
                      gridYValues={5}
                      useMesh={true}
                      legends={[]}
                      theme={{
                        axis: {
                          ticks: {
                            text: {
                              fontSize: 11,
                              fill: '#94a3b8'
                            }
                          }
                        },
                        grid: {
                          line: {
                            stroke: '#f1f5f9',
                            strokeWidth: 1
                          }
                        }
                      }}
                      tooltip={(point) => (
                        <div className="bg-white px-3 py-2 shadow-sm rounded-lg ring-1 ring-black/5">
                          <div className="text-xs font-medium text-slate-900">{String(point.point.data.x)}</div>
                          <div className="text-xs text-slate-600 mt-1">
                            <span className="tabular-nums font-semibold">£{Math.round(Number(point.point.data.y)).toLocaleString()}</span>
                          </div>
                        </div>
                      )}
                    />
                  </div>
                ) : (
                  <p className="text-sm text-slate-500 text-center py-12">
                    No transaction history available
                  </p>
                )}
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div className={cardClass}>
                  <h2 className="text-sm font-semibold text-slate-900 mb-6">Monthly Overview</h2>
                  {monthlyData.length > 0 ? (
                    <div style={{ height: 300 }}>
                      <ResponsiveBar
                        data={monthlyData}
                        keys={['income', 'expenses']}
                        indexBy="month"
                        margin={{ top: 20, right: 100, bottom: 50, left: 60 }}
                        padding={0.25}
                        groupMode="grouped"
                        valueScale={{ type: 'linear' }}
                        indexScale={{ type: 'band', round: true }}
                        colors={({ id }) => id === 'income' ? '#10b981' : '#ef4444'}
                        borderRadius={4}
                        borderColor={{
                          from: 'color',
                          modifiers: [['darker', 0.2]]
                        }}
                        axisTop={null}
                        axisRight={null}
                        axisBottom={{
                          tickSize: 0,
                          tickPadding: 8,
                          tickRotation: 0,
                          legend: '',
                          legendPosition: 'middle',
                          legendOffset: 40
                        }}
                        axisLeft={{
                          tickSize: 0,
                          tickPadding: 8,
                          tickRotation: 0,
                          legend: '',
                          legendPosition: 'middle',
                          legendOffset: -50,
                          format: (value) => `£${Math.round(value)}`
                        }}
                        enableGridY={true}
                        enableLabel={false}
                        legends={[
                          {
                            dataFrom: 'keys',
                            anchor: 'bottom-right',
                            direction: 'column',
                            justify: false,
                            translateX: 90,
                            translateY: 0,
                            itemsSpacing: 4,
                            itemWidth: 80,
                            itemHeight: 18,
                            itemDirection: 'left-to-right',
                            itemOpacity: 1,
                            symbolSize: 10,
                            symbolShape: 'circle',
                            itemTextColor: '#475569',
                            effects: []
                          }
                        ]}
                        theme={{
                          axis: {
                            ticks: {
                              text: {
                                fontSize: 11,
                                fill: '#94a3b8'
                              }
                            }
                          },
                          grid: {
                            line: {
                              stroke: '#f1f5f9',
                              strokeWidth: 1
                            }
                          },
                          legends: {
                            text: {
                              fontSize: 11,
                              fontWeight: 500
                            }
                          }
                        }}
                        tooltip={({ id, value, indexValue, color }) => (
                          <div className="bg-white px-3 py-2 shadow-sm rounded-lg ring-1 ring-black/5">
                            <div className="text-xs font-medium text-slate-900">{indexValue}</div>
                            <div className="flex items-center gap-2 text-xs text-slate-600 mt-1">
                              <div className="w-2 h-2 rounded-full" style={{ backgroundColor: color }} />
                              <span className="capitalize">{id}:</span>
                              <span className="tabular-nums font-semibold">£{Math.round(Number(value)).toLocaleString()}</span>
                            </div>
                          </div>
                        )}
                      />
                    </div>
                  ) : (
                    <p className="text-sm text-slate-500 text-center py-12">No monthly data available</p>
                  )}
                </div>

                <div className={cardClass}>
                  <h2 className="text-sm font-semibold text-slate-900 mb-6">Spending by Category</h2>
                  {categoryData.length > 0 ? (
                    <div style={{ height: 300 }}>
                      <ResponsivePie
                        data={categoryData}
                        margin={{ top: 20, right: 20, bottom: 20, left: 20 }}
                        innerRadius={0.5}
                        padAngle={1}
                        cornerRadius={4}
                        activeOuterRadiusOffset={8}
                        colors={COLORS}
                        borderWidth={0}
                        arcLinkLabelsSkipAngle={10}
                        arcLinkLabelsTextColor="#64748b"
                        arcLinkLabelsThickness={1}
                        arcLinkLabelsColor={{ from: 'color', modifiers: [['opacity', 0.4]] }}
                        arcLabelsSkipAngle={15}
                        arcLabelsTextColor="#ffffff"
                        arcLabel={(d) => `${((d.value / categoryData.reduce((sum, c) => sum + c.value, 0)) * 100).toFixed(0)}%`}
                        theme={{
                          labels: {
                            text: {
                              fontSize: 11,
                              fontWeight: 600
                            }
                          }
                        }}
                        tooltip={({ datum }) => (
                          <div className="bg-white px-3 py-2 shadow-sm rounded-lg ring-1 ring-black/5">
                            <div className="flex items-center gap-2">
                              <div className="w-2 h-2 rounded-full" style={{ backgroundColor: datum.color }} />
                              <span className="text-xs font-medium text-slate-900">{datum.label}</span>
                            </div>
                            <div className="text-xs text-slate-600 mt-1">
                              <span className="tabular-nums font-semibold">£{Math.round(Number(datum.value)).toLocaleString()}</span>
                              <span className="text-slate-500 ml-1">({Math.round((datum.value / categoryData.reduce((sum, c) => sum + c.value, 0)) * 100)}%)</span>
                            </div>
                          </div>
                        )}
                      />
                    </div>
                  ) : (
                    <p className="text-sm text-slate-500 text-center py-12">
                      No categorized expenses yet
                    </p>
                  )}
                </div>
              </div>
            </div>
          ) : (
            <p className="text-sm text-slate-600">No data available</p>
          )}
        </div>
      </main>
    </div>
  );
}
