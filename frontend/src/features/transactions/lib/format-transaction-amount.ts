export function formatTransactionAmount(
  amount: string,
  currencyCode?: string,
) {
  const value = Number(amount);

  if (!Number.isFinite(value)) {
    return currencyCode ? `${amount} ${currencyCode}` : amount;
  }

  if (!currencyCode) {
    return new Intl.NumberFormat(undefined, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 4,
    }).format(value);
  }

  try {
    return new Intl.NumberFormat(undefined, {
      style: "currency",
      currency: currencyCode,
      minimumFractionDigits: 2,
      maximumFractionDigits: 4,
    }).format(value);
  } catch {
    return `${amount} ${currencyCode}`;
  }
}
