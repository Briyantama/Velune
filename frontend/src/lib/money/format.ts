export function formatMoneyMinor(input: { amountMinor: number; currency: string; locale?: string }): string {
  const { amountMinor, currency } = input;
  const locale = input.locale ?? "en";
  const { fractionDigits } = currencyMeta(currency);

  const sign = amountMinor < 0 ? "-" : "";
  const abs = Math.abs(amountMinor);
  const value = abs / Math.pow(10, fractionDigits);

  return (
    sign +
    new Intl.NumberFormat(locale, {
      style: "currency",
      currency: currency.toUpperCase(),
      minimumFractionDigits: fractionDigits,
      maximumFractionDigits: fractionDigits
    }).format(value)
  );
}

export function currencyMeta(currency: string): { fractionDigits: number } {
  // Conservative default for most ISO currencies.
  const cur = (currency || "USD").toUpperCase();
  if (cur === "JPY" || cur === "KRW") return { fractionDigits: 0 };
  return { fractionDigits: 2 };
}

