export const DEFAULT_PAYMENT_CURRENCY = 'CNY'
const DISPLAY_PAYMENT_CURRENCY_SYMBOL = '￥'

export function normalizePaymentCurrency(currency?: string | null): string {
  const normalized = String(currency || '').trim().toUpperCase()
  return /^[A-Z]{3}$/.test(normalized) ? normalized : DEFAULT_PAYMENT_CURRENCY
}

function paymentCurrencyFractionDigits(currency: string): number {
  try {
    return new Intl.NumberFormat(undefined, {
      style: 'currency',
      currency,
    }).resolvedOptions().maximumFractionDigits ?? 2
  } catch {
    return 2
  }
}

export function formatPaymentAmount(amount: number, currency?: string | null, locale?: string): string {
  const normalized = normalizePaymentCurrency(currency)
  const fractionDigits = paymentCurrencyFractionDigits(normalized)
  void locale
  return `${DISPLAY_PAYMENT_CURRENCY_SYMBOL}${(Number.isFinite(amount) ? amount : 0).toFixed(fractionDigits)}`
}
