export function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

export function formatExpiry(iso?: string): string {
  if (!iso) return 'Never'
  return new Date(iso).toLocaleString()
}

// Turn a number of seconds into a friendly phrase: 900 -> "15 minutes".
export function humanizeDuration(seconds: number): string {
  if (!seconds) return 'never'
  const units: [number, string][] = [
    [86400, 'day'],
    [3600, 'hour'],
    [60, 'minute'],
    [1, 'second'],
  ]
  for (const [size, name] of units) {
    if (seconds % size === 0) {
      const n = seconds / size
      return `${n} ${name}${n > 1 ? 's' : ''}`
    }
  }
  return `${seconds} seconds`
}
