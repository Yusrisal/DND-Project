/**
 * Formatting helpers for the wiki. They are pure and live apart from the panel
 * so they can be exercised without a renderer — see `wiki-format.test.mjs`.
 */

/**
 * Prices are stored in copper. 10 cp is 1 sp and 100 cp is 1 gp, so the largest
 * denomination that does not read as a fraction below one is the useful one.
 */
export function price(cp: number): string {
  if (cp <= 0) return 'free'
  if (cp >= 100) return `${trimZeros((cp / 100).toFixed(2))} gp`
  if (cp >= 10) return `${trimZeros((cp / 10).toFixed(1))} sp`
  return `${cp} cp`
}

/** Shops buy back at half, rounded down, but never for nothing. */
export function sellPrice(cp: number): number {
  if (cp <= 0) return 0
  return Math.max(1, Math.floor(cp / 2))
}

/** Case-insensitive substring match across any of the fields given. */
export function matches(query: string, ...fields: (string | number)[]): boolean {
  const q = query.trim().toLowerCase()
  if (!q) return true
  return fields.some((field) => String(field).toLowerCase().includes(q))
}

/**
 * "1.50" becomes "1.5" and "10.00" becomes "10". The guard matters only if a
 * caller ever passes an integer string: stripping trailing zeros from "1000"
 * would leave "1". `toFixed` always emits a decimal point, so today it never
 * fires.
 */
function trimZeros(value: string): string {
  if (!value.includes('.')) return value
  return value.replace(/0+$/, '').replace(/\.$/, '')
}
