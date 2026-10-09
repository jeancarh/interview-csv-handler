const usd = new Intl.NumberFormat("en-US", { style: "currency", currency: "USD" });

export function formatCents(cents: number): string {
  return usd.format(cents / 100);
}

export function parseFixed(input: string, scale: number): number | null {
  const value = input.trim();
  const match = new RegExp(`^(\\d+)(?:\\.(\\d{1,${scale}}))?$`).exec(value);
  if (!match) {
    return null;
  }
  const whole = match[1] ?? "0";
  const fraction = (match[2] ?? "").padEnd(scale, "0");
  const result = Number(whole + fraction);
  return Number.isSafeInteger(result) ? result : null;
}

export function toFixedString(units: number, scale: number): string {
  const factor = 10 ** scale;
  const whole = Math.floor(units / factor);
  const fraction = String(units % factor).padStart(scale, "0");
  return `${whole}.${fraction}`;
}
