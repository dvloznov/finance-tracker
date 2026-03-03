'use client';

import { cardClass, currencyClass, statLabelClass, statValueClass } from '@/lib/ui';
import { useTransactions } from '@/features/transactions/hooks/useTransactions';
import { AppNav } from '@/shared/ui/AppNav';
import { useMemo, useState } from 'react';
import { formatCurrency, formatCurrencyWithCode } from '@/shared/formatters/currency';
import { formatShortDate } from '@/shared/formatters/date';
import { format, parse, subMonths, lastDayOfMonth } from 'date-fns';
import { getBalanceSeries } from '@/features/dashboard/analytics/balance';
import { getCategoryTotals, getSubcategoryTotals } from '@/features/dashboard/analytics/categories';
import { getMonthlyTotals, getDailyTotals, filterTransactionsByBar, type BarSelection } from '@/features/dashboard/analytics/monthly';
import { getStatsSummary, type StatsSummary } from '@/features/dashboard/analytics/stats';
import { detectTransferIds } from '@/features/dashboard/analytics/transfers';
import { useAccountScope } from '@/shared/account-scope/context';
import { useAccountOptions } from '@/shared/account-scope/useAccountOptions';
import { AccountScopeSelect } from '@/shared/ui/AccountScopeSelect';
import type { TransactionVM } from '@/features/transactions/types';
import { ResponsiveLine } from '@nivo/line';
import { ResponsiveBar } from '@nivo/bar';
import { ResponsivePie } from '@nivo/pie';
import { Inbox, FileQuestion } from 'lucide-react';

const CHART_COLORS = ['#475569', '#10b981', '#f43f5e', '#f59e0b', '#64748b', '#94a3b8'];

function getMonthOptions(): { value: string; label: string }[] {
  const options: { value: string; label: string }[] = [{ value: 'all', label: 'All time' }];
  const now = new Date();
  for (let i = 0; i < 12; i++) {
    const d = subMonths(now, i);
    const value = format(d, 'yyyy-MM');
    const label = format(d, 'MMMM yyyy');
    options.push({ value, label });
  }
  return options;
}


type StatCardsProps = {
  stats: StatsSummary;
};

function StatCards({ stats }: StatCardsProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div className={`${cardClass} flex flex-col gap-3 border-t-4 border-t-emerald-500 bg-emerald-50/30`}>
        <p className={statLabelClass}>Total Income</p>
        <p className={`${statValueClass} text-emerald-600 flex items-baseline gap-1`}>
          <span className={currencyClass}>£</span>
          <span>{formatCurrency(stats.totalIncome).replace('£', '')}</span>
        </p>
      </div>
      <div className={`${cardClass} flex flex-col gap-3 border-t-4 border-t-rose-500 bg-rose-50/30`}>
        <p className={statLabelClass}>Total Expenses</p>
        <p className={`${statValueClass} text-rose-600 flex items-baseline gap-1`}>
          <span className={currencyClass}>£</span>
          <span>{formatCurrency(stats.totalExpenses).replace('£', '')}</span>
        </p>
      </div>
      <div className={`${cardClass} flex flex-col gap-3 border-t-4 border-t-slate-500 bg-slate-50/50`}>
        <p className={statLabelClass}>Net Worth</p>
        <p className={`${statValueClass} flex items-baseline gap-1`}>
          <span className={currencyClass}>£</span>
          <span className={stats.netBalance >= 0 ? 'text-emerald-600' : 'text-rose-600'}>
            {formatCurrency(stats.netBalance).replace('£', '')}
          </span>
        </p>
      </div>
    </div>
  );
}

type BalanceChartCardProps = {
  balanceData: Array<{ id: string; data: Array<{ x: string; y: number; breakdown?: Record<string, number> }> }>;
  institutions: Array<{ institution_id: string; name: string }>;
};

function BalanceChartCard({ balanceData, institutions }: BalanceChartCardProps) {
  const institutionLookup = useMemo(
    () => new Map(institutions.map((i) => [i.institution_id, i.name])),
    [institutions]
  );
  return (
    <div className={cardClass}>
      <h2 className="text-base font-semibold text-slate-900 mb-6">Account Balance Over Time</h2>
      {balanceData.length > 0 && balanceData[0].data.length > 0 ? (
        <div style={{ height: 300 }}>
          <ResponsiveLine
            data={balanceData}
            margin={{ top: 20, right: 20, bottom: 60, left: 60 }}
            xScale={{ type: 'point' }}
            yScale={{ type: 'linear', min: 'auto', max: 'auto' }}
            curve="monotoneX"
            axisTop={null}
            axisRight={null}
            axisBottom={{
              tickSize: 0,
              tickPadding: 10,
              tickRotation: -45,
              legend: '',
              legendOffset: 50,
              legendPosition: 'middle',
              tickValues: balanceData[0]?.data
                .filter((_, i, arr) => {
                  if (arr.length <= 8) return true;
                  const step = Math.ceil(arr.length / 8);
                  return i === 0 || i === arr.length - 1 || i % step === 0;
                })
                .map((d) => d.x) ?? [],
            }}
            axisLeft={{
              tickSize: 0,
              tickPadding: 8,
              tickRotation: 0,
              legend: '',
              legendOffset: -50,
              legendPosition: 'middle'
            }}
            colors={['#475569']}
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
            tooltip={(point) => {
              const data = point.point.data as { x: string; y: number; breakdown?: Record<string, number> };
              const value = Number(data.y);
              const breakdown = data.breakdown;
              return (
                <div className="bg-white px-3 py-2 shadow-sm rounded-lg ring-1 ring-black/5 min-w-[140px]">
                  <div className="text-xs font-medium text-slate-900">{String(data.x)}</div>
                  <div className="text-sm text-slate-700 mt-1 tabular-nums font-semibold">
                    £{value.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                  </div>
                  {breakdown && Object.keys(breakdown).length > 0 && (
                    <div className="mt-2 pt-2 border-t border-slate-100 space-y-1">
                      {Object.entries(breakdown)
                        .filter(([, v]) => v !== 0)
                        .sort(([, a], [, b]) => b - a)
                        .map(([instId, bal]) => (
                          <div key={instId} className="flex justify-between gap-4 text-xs">
                            <span className="text-slate-500 truncate max-w-[100px]">
                              {instId === '__unknown__' ? 'Unknown' : institutionLookup.get(instId) ?? instId}
                            </span>
                            <span className={`tabular-nums font-medium shrink-0 ${bal >= 0 ? 'text-emerald-600' : 'text-rose-600'}`}>
                              £{bal.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 })}
                            </span>
                          </div>
                        ))}
                    </div>
                  )}
                </div>
              );
            }}
          />
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <Inbox className="w-10 h-10 text-slate-300 mb-3" />
          <p className="text-base text-slate-500">No transaction history available</p>
        </div>
      )}
    </div>
  );
}

type OverviewChartData = Array<{ month?: string; day?: string; income: number; expenses: number }>;

type MonthlyOverviewCardProps = {
  data: OverviewChartData;
  indexBy: 'month' | 'day';
  title: string;
  onBarClick?: (indexValue: string, type: 'income' | 'expenses', day?: number) => void;
};

function MonthlyOverviewCard({ data, indexBy, title, onBarClick }: MonthlyOverviewCardProps) {
  const indexKey = indexBy === 'month' ? 'month' : 'day';
  const chartData = data.map((d) => ({ ...d, [indexKey]: d[indexKey] ?? '' }));
  const isDaily = indexBy === 'day';
  const axisBottom = {
    tickSize: 0,
    tickPadding: 8,
    tickRotation: isDaily ? -45 : 0,
    legend: '',
    legendPosition: 'middle' as const,
    legendOffset: 40,
    ...(isDaily && chartData.length > 10 && {
      tickValues: chartData
        .filter((_, i) => i % Math.ceil(chartData.length / 10) === 0 || i === chartData.length - 1)
        .map((d) => d.day),
    }),
  };

  return (
    <div className={`${cardClass} hover:shadow-md`}>
      <h2 className="text-base font-semibold text-slate-900 mb-6">{title}</h2>
      {chartData.length > 0 ? (
        <div style={{ height: 300 }} className={onBarClick ? 'cursor-pointer' : ''}>
          <ResponsiveBar
            data={chartData}
            keys={['income', 'expenses']}
            indexBy={indexKey}
            onClick={(datum) => {
              if (onBarClick && datum.indexValue != null && datum.id != null) {
                const day = indexBy === 'day' ? parseInt(String(datum.indexValue), 10) : undefined;
                onBarClick(String(datum.indexValue), datum.id as 'income' | 'expenses', day);
              }
            }}
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
            axisBottom={axisBottom}
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
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <FileQuestion className="w-10 h-10 text-slate-300 mb-3" />
          <p className="text-base text-slate-500">No data available</p>
        </div>
      )}
    </div>
  );
}

type SpendingByCategoryCardProps = {
  categoryData: Array<{ id: string; label: string; value: number }>;
  subcategoryData: Array<{ id: string; label: string; value: number }>;
  selectedCategory: string | null;
  selectedSubcategory: string | null;
  onSelectCategory: (value: string) => void;
  onSelectSubcategory: (value: string) => void;
  onClearCategory: () => void;
};

function SpendingByCategoryCard({
  categoryData,
  subcategoryData,
  selectedCategory,
  selectedSubcategory,
  onSelectCategory,
  onSelectSubcategory,
  onClearCategory,
}: SpendingByCategoryCardProps) {
  const activeData = selectedCategory ? subcategoryData : categoryData;

  return (
    <div className={`${cardClass} hover:shadow-md`}>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-base font-semibold text-slate-900">
          {selectedSubcategory
            ? `${selectedCategory} › ${selectedSubcategory}`
            : selectedCategory
            ? `Spending in ${selectedCategory}`
            : 'Spending by Category'}
        </h2>
        {selectedCategory && (
          <button
            type="button"
            onClick={onClearCategory}
            className="text-xs font-medium text-slate-600 hover:text-slate-900 focus:outline-none focus:ring-2 focus:ring-slate-900/10 rounded"
          >
            ← Back to categories
          </button>
        )}
      </div>
      {activeData.length > 0 ? (
        <div style={{ height: 300 }}>
          <ResponsivePie
            data={activeData}
            margin={{ top: 20, right: 20, bottom: 20, left: 20 }}
            innerRadius={0.5}
            padAngle={1}
            cornerRadius={4}
            activeOuterRadiusOffset={8}
            colors={CHART_COLORS}
            borderWidth={0}
            arcLinkLabelsSkipAngle={10}
            arcLinkLabelsTextColor="#64748b"
            arcLinkLabelsThickness={1}
            arcLinkLabelsColor={{ from: 'color', modifiers: [['opacity', 0.4]] }}
            arcLabelsSkipAngle={15}
            arcLabelsTextColor="#ffffff"
            arcLabel={(d) => {
              const total = activeData.reduce((sum, c) => sum + c.value, 0);
              return total ? `${((d.value / total) * 100).toFixed(0)}%` : '';
            }}
            theme={{
              labels: {
                text: {
                  fontSize: 11,
                  fontWeight: 600
                }
              }
            }}
            onClick={(datum) => {
              if (!selectedCategory) {
                onSelectCategory(String(datum.id));
              } else {
                onSelectSubcategory(String(datum.id));
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
                  <span className="text-slate-500 ml-1">({Math.round((datum.value / activeData.reduce((sum, c) => sum + c.value, 0)) * 100)}%)</span>
                </div>
              </div>
            )}
          />
        </div>
      ) : (
        <div className="flex flex-col items-center justify-center py-12 text-center">
          <FileQuestion className="w-10 h-10 text-slate-300 mb-3" />
          <p className="text-base text-slate-500">
            {selectedCategory ? 'No subcategory data available' : 'No categorized expenses yet'}
          </p>
        </div>
      )}
    </div>
  );
}

type RecentTransactionsCardProps = {
  transactions: TransactionVM[];
  selectedCategory: string | null;
  selectedSubcategory: string | null;
  selectedBar: BarSelection | null;
  transferIds: Set<string>;
  institutions: Array<{ institution_id: string; name: string }>;
  onClearFilter: () => void;
  onClearBarFilter: () => void;
};

const PAGE_SIZE = 10;

function RecentTransactionsCard({
  transactions,
  selectedCategory,
  selectedSubcategory,
  selectedBar,
  transferIds,
  institutions,
  onClearFilter,
  onClearBarFilter,
}: RecentTransactionsCardProps) {
  const [page, setPage] = useState(0);
  const institutionLookup = useMemo(
    () => new Map(institutions.map((i) => [i.institution_id, i.name])),
    [institutions]
  );

  const filtered = useMemo(() => {
    let txns = [...transactions].sort(
      (a, b) => Math.abs(parseFloat(b.amount)) - Math.abs(parseFloat(a.amount))
    );
    if (selectedBar) {
      txns = filterTransactionsByBar(txns, selectedBar, transferIds);
    }
    if (selectedSubcategory) {
      txns = txns.filter(
        (t) => t.category_name === selectedCategory && t.subcategory_name === selectedSubcategory
      );
    } else if (selectedCategory) {
      txns = txns.filter((t) => t.category_name === selectedCategory);
    }
    return txns;
  }, [transactions, selectedCategory, selectedSubcategory, selectedBar, transferIds]);

  // Reset to first page whenever the filter changes.
  const prevFilter = useMemo(
    () => `${selectedCategory}|${selectedSubcategory}|${selectedBar?.month ?? ''}|${selectedBar?.type ?? ''}|${selectedBar?.day ?? ''}`,
    [selectedCategory, selectedSubcategory, selectedBar]
  );
  const [lastFilter, setLastFilter] = useState(prevFilter);
  if (prevFilter !== lastFilter) {
    setLastFilter(prevFilter);
    setPage(0);
  }

  const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  const safePage = Math.min(page, totalPages - 1);
  const pageRows = filtered.slice(safePage * PAGE_SIZE, safePage * PAGE_SIZE + PAGE_SIZE);

  const filterLabels: string[] = [];
  if (selectedBar) {
    const barLabel = selectedBar.day != null
      ? `${selectedBar.month} ${selectedBar.day} ${selectedBar.type}`
      : `${selectedBar.month} ${selectedBar.type}`;
    filterLabels.push(barLabel);
  }
  if (selectedSubcategory) {
    filterLabels.push(`${selectedCategory} › ${selectedSubcategory}`);
  } else if (selectedCategory) {
    filterLabels.push(selectedCategory);
  }
  const filterLabel = filterLabels.length > 0 ? filterLabels.join(' · ') : null;

  return (
    <div className={cardClass}>
      <div className="flex items-center justify-between mb-4">
        <div className="flex items-center gap-3 flex-wrap">
          <h2 className="text-base font-semibold text-slate-900">Recent Transactions</h2>
          {filterLabel && (
            <span className="inline-flex items-center gap-1.5 bg-slate-100 text-slate-600 text-xs font-medium px-2.5 py-1 rounded-full">
              {filterLabel}
              <button
                type="button"
                onClick={() => {
                  onClearFilter();
                  onClearBarFilter();
                }}
                className="text-slate-400 hover:text-slate-700 leading-none focus:outline-none focus:ring-2 focus:ring-slate-900/10 rounded"
                aria-label="Clear filters"
              >
                ×
              </button>
            </span>
          )}
        </div>
        <span className="text-xs text-slate-400">
          {filtered.length} transaction{filtered.length !== 1 ? 's' : ''}
        </span>
      </div>

      {filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-6 text-center">
          <Inbox className="w-10 h-10 text-slate-300 mb-3" />
          <p className="text-base text-slate-500">
            {filterLabel ? `No transactions in "${filterLabel}"` : 'No transactions yet'}
          </p>
        </div>
      ) : (
        <>
          <div className="divide-y divide-slate-100">
            {pageRows.map((txn) => {
              const amount = parseFloat(txn.amount);
              const isTransfer = transferIds.has(txn.transaction_id);
              return (
                <div key={txn.transaction_id} className="flex items-center justify-between py-3 gap-4 hover:bg-slate-50 rounded-lg -mx-2 px-2">
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-medium text-slate-900 truncate">
                      {txn.merchant_name || txn.raw_description}
                    </p>
                    <p className="text-xs text-slate-400 mt-0.5">
                      {formatShortDate(txn.transaction_date)}
                      {txn.institution_id && (
                        <> · <span className="text-slate-500">{institutionLookup.get(txn.institution_id) ?? txn.institution_id}</span></>
                      )}
                      {txn.category_name && !isTransfer && (
                        <> · <span className="text-slate-500">{txn.category_name}</span></>
                      )}
                    </p>
                  </div>
                  <div className="flex items-center gap-3 shrink-0">
                    {isTransfer ? (
                      <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-violet-100 text-violet-700">
                        ⇄ Transfer
                      </span>
                    ) : txn.category_name && !selectedCategory ? (
                      <span className="text-xs text-slate-400 bg-slate-100 px-2 py-0.5 rounded-full">
                        {txn.subcategory_name || txn.category_name}
                      </span>
                    ) : null}
                    <span className={`text-sm font-semibold tabular-nums ${amount < 0 ? 'text-rose-600' : 'text-emerald-600'}`}>
                      {formatCurrencyWithCode(amount, txn.currency)}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>

          {totalPages > 1 && (
            <div className="flex items-center justify-between border-t border-slate-100 pt-3 mt-1">
              <span className="text-xs text-slate-500">
                Page {safePage + 1} of {totalPages}
              </span>
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => setPage((p) => Math.max(0, p - 1))}
                  disabled={safePage === 0}
                  className="rounded-md border border-slate-200 px-3 py-1 text-xs text-slate-700 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-slate-900/10"
                >
                  Previous
                </button>
                <button
                  type="button"
                  onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
                  disabled={safePage === totalPages - 1}
                  className="rounded-md border border-slate-200 px-3 py-1 text-xs text-slate-700 disabled:opacity-50 disabled:cursor-not-allowed hover:bg-slate-50 focus:outline-none focus:ring-2 focus:ring-slate-900/10"
                >
                  Next
                </button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}

export default function DashboardPage() {
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const [selectedSubcategory, setSelectedSubcategory] = useState<string | null>(null);
  const [dateRange, setDateRange] = useState<string>('all');
  const { scope } = useAccountScope();
  const dateParams = useMemo(() => {
    if (dateRange === 'all') return {};
    const [year, month] = dateRange.split('-').map(Number);
    const firstDay = new Date(year, month - 1, 1);
    const lastDay = lastDayOfMonth(firstDay);
    return {
      start_date: format(firstDay, 'yyyy-MM-dd'),
      end_date: format(lastDay, 'yyyy-MM-dd'),
    };
  }, [dateRange]);
  const { data: transactions, isLoading, error } = useTransactions({ scope, ...dateParams });
  const { institutions = [] } = useAccountOptions();

  // Detect transfer pairs only when viewing all accounts together.
  // In single-account or single-institution scope there can be no cross-account transfers.
  const transferIds = useMemo(() => {
    if (scope.mode !== 'all') return new Set<string>();
    return detectTransferIds(transactions ?? []);
  }, [transactions, scope.mode]);

  const handleSelectCategory = (cat: string) => {
    setSelectedCategory(cat);
    setSelectedSubcategory(null);
  };

  const handleSelectSubcategory = (sub: string) => {
    setSelectedSubcategory(sub);
  };

  const handleClearCategory = () => {
    setSelectedCategory(null);
    setSelectedSubcategory(null);
  };

  const [selectedBar, setSelectedBar] = useState<BarSelection | null>(null);
  const handleBarClick = (indexValue: string, type: 'income' | 'expenses', day?: number) => {
    const month = dateRange === 'all' ? indexValue : format(parse(`${dateRange}-01`, 'yyyy-MM', new Date()), 'MMM yyyy');
    const newSel: BarSelection = { month, type, ...(day != null && { day }) };
    setSelectedBar((prev) =>
      prev?.month === newSel.month && prev?.type === newSel.type && prev?.day === newSel.day ? null : newSel
    );
  };
  const handleClearBarFilter = () => setSelectedBar(null);

  const stats = useMemo(() => {
    return getStatsSummary(transactions, transferIds);
  }, [transactions, transferIds]);

  const monthlyData = useMemo(() => {
    return getMonthlyTotals(transactions, transferIds);
  }, [transactions, transferIds]);

  const dailyData = useMemo(() => {
    if (dateRange === 'all') return [];
    return getDailyTotals(transactions, dateRange, transferIds);
  }, [transactions, dateRange, transferIds]);

  const overviewData = dateRange === 'all' ? monthlyData : dailyData;
  const overviewMode = dateRange === 'all' ? 'monthly' : 'daily';
  const overviewTitle = dateRange === 'all' ? 'Monthly Overview' : 'Daily Overview';
  const overviewIndexBy = dateRange === 'all' ? 'month' : 'day';

  const categoryData = useMemo(() => {
    return getCategoryTotals(transactions, transferIds);
  }, [transactions, transferIds]);

  const subcategoryData = useMemo(() => {
    return getSubcategoryTotals(transactions, selectedCategory, transferIds);
  }, [transactions, selectedCategory, transferIds]);

  const balanceData = useMemo(() => {
    return getBalanceSeries(transactions);
  }, [transactions]);

  return (
    <div className="min-h-screen bg-slate-50">
      <AppNav active="dashboard" />

      <main className="container mx-auto px-6 py-8">
        <div className="space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div className="space-y-1">
              <h1 className="text-3xl font-semibold tracking-tight text-slate-900">Dashboard</h1>
              <p className="text-sm text-slate-600">Overview of your financial activity</p>
            </div>
            <div className="bg-white rounded-xl px-4 py-2 shadow-sm flex flex-wrap items-center gap-3 ring-1 ring-black/5">
              <select
                value={dateRange}
                onChange={(e) => setDateRange(e.target.value)}
                className="px-4 py-2 border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-slate-900/10 focus:border-slate-300 text-sm font-medium text-slate-900 bg-white w-fit"
              >
                {getMonthOptions().map((opt) => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
              <div className="h-4 w-px bg-slate-200" aria-hidden="true" />
              <AccountScopeSelect />
            </div>
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
              <StatCards stats={stats} />
              <BalanceChartCard balanceData={balanceData} institutions={institutions} />
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <MonthlyOverviewCard
                  data={overviewData}
                  indexBy={overviewIndexBy}
                  title={overviewTitle}
                  onBarClick={handleBarClick}
                />
                <SpendingByCategoryCard
                  categoryData={categoryData}
                  subcategoryData={subcategoryData}
                  selectedCategory={selectedCategory}
                  selectedSubcategory={selectedSubcategory}
                  onSelectCategory={handleSelectCategory}
                  onSelectSubcategory={handleSelectSubcategory}
                  onClearCategory={handleClearCategory}
                />
              </div>
              <RecentTransactionsCard
                transactions={transactions ?? []}
                selectedCategory={selectedCategory}
                selectedSubcategory={selectedSubcategory}
                selectedBar={selectedBar}
                transferIds={transferIds}
                institutions={institutions}
                onClearFilter={handleClearCategory}
                onClearBarFilter={handleClearBarFilter}
              />
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-12 text-center">
              <FileQuestion className="w-10 h-10 text-slate-300 mb-3" />
              <p className="text-base text-slate-600">No data available</p>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
