// All requests are same-origin and go through a proxy to the Go API:
//   dev  -> Vite proxies /api to http://localhost:8080 (vite.config.ts)
//   prod -> nginx proxies /api to the api service (nginx.conf)
// So we never hardcode the API origin and never need CORS.

export interface FileMeta {
  id: string
  filename: string
  size: number
  mimeType: string
  hasPassword: boolean
  createdAt: string
  expiresAt?: string
  downloadCount: number
}

export interface UploadResult {
  id: string
  shareUrl: string
}

export interface ServerConfig {
  maxUploadBytes: number
  maxUploadMB: number
  expirySeconds: number // 0 = never expires
}

// The Go server writes errors as {"error":"..."} bodies. Pull that message out,
// falling back to a generic one.
async function errorMessage(res: Response, fallback: string): Promise<string> {
  try {
    const body = await res.json()
    return body?.error ?? fallback
  } catch {
    return fallback
  }
}

export async function uploadFile(file: File, password: string): Promise<UploadResult> {
  const form = new FormData()
  form.append('file', file)
  if (password) form.append('password', password)

  const res = await fetch('/api/upload', { method: 'POST', body: form })
  if (!res.ok) throw new Error(await errorMessage(res, 'Upload failed'))
  return res.json()
}

export async function getMeta(id: string): Promise<FileMeta> {
  const res = await fetch(`/api/file/${id}`)
  if (!res.ok) throw new Error(await errorMessage(res, 'File not found'))
  return res.json()
}

// Downloads are a POST (password may be in the body), so we can't use a plain
// <a href>. Fetch the bytes as a blob, then trigger a browser download.
export async function downloadFile(meta: FileMeta, password: string): Promise<void> {
  const res = await fetch(`/api/file/${meta.id}/download`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ password }),
  })
  if (!res.ok) throw new Error(await errorMessage(res, 'Download failed'))

  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = meta.filename
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(url)
}

// Delete a file. For password-protected files the password is required.
export async function deleteFile(id: string, password: string): Promise<void> {
  const res = await fetch(`/api/file/${id}`, {
    method: 'DELETE',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ password }),
  })
  if (!res.ok) throw new Error(await errorMessage(res, 'Delete failed'))
}

export async function getConfig(): Promise<ServerConfig> {
  const res = await fetch('/api/config')
  if (!res.ok) throw new Error('Could not load server config')
  return res.json()
}

export function qrUrl(id: string): string {
  return `/api/qr/${id}`
}
