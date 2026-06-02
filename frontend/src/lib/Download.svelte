<script lang="ts">
  import { getMeta, downloadFile, deleteFile, type FileMeta } from './api'
  import { formatSize, formatExpiry } from './format'

  let { id }: { id: string } = $props()

  let meta = $state<FileMeta | null>(null)
  let loadError = $state('')
  let loading = $state(true)

  let password = $state('')
  let downloading = $state(false)
  let downloadError = $state('')

  let deleting = $state(false)
  let deleted = $state(false)

  $effect(() => {
    loading = true
    getMeta(id)
      .then((m) => (meta = m))
      .catch((e) => (loadError = e instanceof Error ? e.message : 'File not found'))
      .finally(() => (loading = false))
  })

  async function handleDownload() {
    if (!meta) return
    downloadError = ''
    if (meta.hasPassword && !password) {
      downloadError = 'This file is password protected.'
      return
    }
    downloading = true
    try {
      await downloadFile(meta, password)
    } catch (e) {
      downloadError = e instanceof Error ? e.message : 'Download failed'
    } finally {
      downloading = false
    }
  }

  async function handleDelete() {
    if (!meta) return
    downloadError = ''
    if (meta.hasPassword && !password) {
      downloadError = 'Enter the password to delete this file.'
      return
    }
    if (!confirm('Delete this file for everyone? This cannot be undone.')) return
    deleting = true
    try {
      await deleteFile(meta.id, password)
      deleted = true
    } catch (e) {
      downloadError = e instanceof Error ? e.message : 'Delete failed'
    } finally {
      deleting = false
    }
  }
</script>

<div class="card rise p-6 sm:p-8">
  {#if loading}
    <div class="flex items-center justify-center gap-3 py-10 text-sm text-muted">
      <span class="size-4 animate-spin rounded-full border-2 border-line border-t-accent"></span>
      Loading…
    </div>
  {:else if loadError}
    <div class="space-y-3 py-6 text-center">
      <span class="mx-auto grid size-14 place-items-center rounded-full bg-accent-soft text-accent-deep">
        <svg viewBox="0 0 24 24" fill="none" class="size-7" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9" /><path d="M12 8v4" /><path d="M12 16h.01" /></svg>
      </span>
      <h1 class="text-2xl">{loadError}</h1>
      <a href="/" class="btn-ghost mx-auto">← Share your own file</a>
    </div>
  {:else if deleted}
    <div class="space-y-3 py-6 text-center">
      <span class="mx-auto grid size-14 place-items-center rounded-full bg-accent-soft text-accent-deep">
        <svg viewBox="0 0 24 24" fill="none" class="size-7" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18" /><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" /></svg>
      </span>
      <h1 class="text-2xl">File deleted</h1>
      <p class="text-sm text-muted">This file is gone and the link no longer works.</p>
      <a href="/" class="btn-ghost mx-auto">← Share your own file</a>
    </div>
  {:else if meta}
    <div class="space-y-6">
      <div class="text-center">
        <span class="mx-auto grid size-14 place-items-center rounded-2xl bg-accent-soft text-accent-deep">
          <svg viewBox="0 0 24 24" fill="none" class="size-7" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><path d="M14 3v4a1 1 0 0 0 1 1h4" /><path d="M17 21H7a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h7l5 5v11a2 2 0 0 1-2 2Z" /></svg>
        </span>
        <h1 class="mt-3 break-all text-2xl">{meta.filename}</h1>
      </div>

      <dl class="grid grid-cols-2 overflow-hidden rounded-2xl border border-line text-sm">
        <div class="border-r border-line bg-[#fffdfb] p-4">
          <dt class="text-xs font-medium uppercase tracking-wide text-muted">Size</dt>
          <dd class="mt-1 font-semibold text-ink">{formatSize(meta.size)}</dd>
        </div>
        <div class="bg-[#fffdfb] p-4">
          <dt class="text-xs font-medium uppercase tracking-wide text-muted">Expires</dt>
          <dd class="mt-1 font-semibold text-ink">{formatExpiry(meta.expiresAt)}</dd>
        </div>
      </dl>

      {#if meta.hasPassword}
        <div>
          <label for="dlpw" class="mb-1.5 flex items-center gap-1.5 text-sm font-medium text-ink">
            <svg viewBox="0 0 24 24" fill="none" class="size-4 text-accent-deep" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="11" width="18" height="11" rx="2" /><path d="M7 11V7a5 5 0 0 1 10 0v4" /></svg>
            This file requires a password
          </label>
          <input id="dlpw" type="password" bind:value={password} placeholder="Enter password" class="field" />
        </div>
      {/if}

      {#if downloadError}
        <p class="alert">{downloadError}</p>
      {/if}

      <div class="space-y-3">
        <button onclick={handleDownload} disabled={downloading || deleting} class="btn-primary w-full">
          {#if !downloading}
            <svg viewBox="0 0 24 24" fill="none" class="size-4" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v12" /><path d="m7 12 5 5 5-5" /><path d="M5 21h14" /></svg>
          {/if}
          {downloading ? 'Downloading…' : 'Download'}
        </button>
        <button onclick={handleDelete} disabled={downloading || deleting} class="btn-ghost w-full">
          {deleting ? 'Deleting…' : 'Delete file'}
        </button>
      </div>
    </div>
  {/if}
</div>
