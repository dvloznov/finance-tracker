'use client';

import { cardClass, currencyClass, statLabelClass, statValueClass } from '@/lib/ui';
import { useTransactions } from '@/lib/hooks/useTransactions';
import { AppNav } from '@/components/app-nav';
import { useMemo, useState } from 'react';
import { formatCurrency } from '@/shared/formatters/currency';
import { getBalanceSeries } from '@/features/dashboard/analytics/balance';
import { getCategoryTotals, getSubcategoryTotals } from '@/features/dashboard/analytics/categories';
import { getMonthlyTotals } from '@/features/dashboard/analytics/monthly';
import { getStatsSummary, type StatsSummary } from '@/features/dashboard/analytics/stats';
import { ResponsiveLine } from '@nivo/line';
import { ResponsiveBar } from '@nivo/bar';
import { ResponsivePie } from '@nivo/pie';

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82CA9D'];

type StatCardsProps = {
  stats: StatsSummary;
};

function StatCards({ stats }: StatCardsProps) {
  return (
    <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div className={`${cardClass} flex flex-col gap-3`}>
        <p className={statLabelClass}>Total Income</p>
        <p className={`${statValueClass} text-emerald-600 flex items-baseline gap-1`}>
          <span className={currencyClass}>£</span>
          <span>{formatCurrency(stats.totalIncome).replace('£', '')}</span>
        </p>
      </div>
      <div className={`${cardClass} flex flex-col gap-3`}>
        <p className={statLabelClass}>Total Expenses</p>
        <p className={`${statValueClass} text-rose-600 flex items-baseline gap-1`}>
          <span className={currencyClass}>£</span>
          <span>{formatCurrency(stats.totalExpenses).replace('£', '')}</span>
        </p>
      </div>
      <div className={`${cardClass} flex flex-col gap-3`}>
        <p className={statLabelClass}>Net Balance</p>
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
  balanceData: Array<{ id: string; data: Array<{ x: string; y: number }> }>;
};

function BalanceChartCard({ balanceData }: BalanceChartCardProps) {
  return (
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
  );
}

type MonthlyOverviewCardProps = {
  monthlyData: Array<{ month: string; income: number; expenses: number }>;
};

function MonthlyOverviewCard({ monthlyData }: MonthlyOverviewCardProps) {
  return (
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
  );
}

type SpendingByCategoryCardProps = {
  categoryData: Array<{ id: string; label: string; value: number }>;
  subcategoryData: Array<{ id: string; label: string; value: number }>;
  selectedCategory: string | null;
  onSelectCategory: (value: string) => void;
  onClearCategory: () => void;
};

function SpendingByCategoryCard({
  categoryData,
  subcategoryData,
  selectedCategory,
  onSelectCategory,
  onClearCategory,
}: SpendingByCategoryCardProps) {
  const activeData = selectedCategory ? subcategoryData : categoryData;

  return (
    <div className={cardClass}>
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-sm font-semibold text-slate-900">
          {selectedCategory ? `Spending in ${selectedCategory}` : 'Spending by Category'}
        </h2>
        {selectedCategory && (
          <button
            type="button"
            onClick={onClearCategory}
            className="text-xs font-medium text-slate-600 hover:text-slate-900"
          >
            Back to categories
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
            colors={COLORS}
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
        <p className="text-sm text-slate-500 text-center py-12">
          {selectedCategory ? 'No subcategory data available' : 'No categorized expenses yet'}
        </p>
      )}
    </div>
  );
}

export default function DashboardPage() {
  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);
  const { data: transactions, isLoading, error } = useTransactions();

  const stats = useMemo(() => {
    return getStatsSummary(transactions);
  }, [transactions]);

  const monthlyData = useMemo(() => {
    return getMonthlyTotals(transactions);
  }, [transactions]);

  const categoryData = useMemo(() => {
    return getCategoryTotals(transactions);
  }, [transactions]);

  const subcategoryData = useMemo(() => {
    return getSubcategoryTotals(transactions, selectedCategory);
  }, [transactions, selectedCategory]);

  const balanceData = useMemo(() => {
    return getBalanceSeries(transactions);
  }, [transactions]);

  return (
    <div className="min-h-screen bg-slate-50">
      <AppNav active="dashboard" />

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
              <StatCards stats={stats} />
              <BalanceChartCard balanceData={balanceData} />
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <MonthlyOverviewCard monthlyData={monthlyData} />
                <SpendingByCategoryCard
                  categoryData={categoryData}
                  subcategoryData={subcategoryData}
                  selectedCategory={selectedCategory}
                  onSelectCategory={setSelectedCategory}
                  onClearCategory={() => setSelectedCategory(null)}
                />
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
