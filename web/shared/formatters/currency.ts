export function formatCurrency(amount: number, currencySymbol = '£'): string {
  return `${currencySymbol}${Math.round(amount).toLocaleString()}`;
}

export function formatCurrencyWithCode(amount: number, currency: string): string {
  return `${amount.toFixed(2)} ${currency}`;
}
