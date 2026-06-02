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

<div class="rounded-2xl bg-white p-6 shadow-xl shadow-slate-200/60 dark:bg-slate-900 dark:shadow-none dark:ring-1 dark:ring-slate-800 sm:p-8">
  {#if loading}
    <p class="text-center text-slate-500">Loading…</p>
  {:else if loadError}
    <div class="space-y-3 text-center">
      <span class="text-4xl">🚫</span>
      <h1 class="text-xl font-semibold text-slate-800 dark:text-slate-100">{loadError}</h1>
      <a href="/" class="inline-block text-sm font-medium text-indigo-600 hover:text-indigo-700 dark:text-indigo-400">
        ← Share your own file
      </a>
    </div>
  {:else if deleted}
    <div class="space-y-3 text-center">
      <span class="text-4xl">🗑️</span>
      <h1 class="text-xl font-semibold text-slate-800 dark:text-slate-100">File deleted</h1>
      <p class="text-sm text-slate-500">This file is gone and the link no longer works.</p>
      <a href="/" class="inline-block text-sm font-medium text-indigo-600 hover:text-indigo-700 dark:text-indigo-400">
        ← Share your own file
      </a>
    </div>
  {:else if meta}
    <div class="space-y-5">
      <div class="text-center">
        <span class="text-4xl">📄</span>
        <h1 class="mt-2 break-all text-xl font-semibold text-slate-800 dark:text-slate-100">{meta.filename}</h1>
      </div>

      <dl class="grid grid-cols-2 gap-3 rounded-xl bg-slate-50 p-4 text-sm dark:bg-slate-800/60">
        <div>
          <dt class="text-slate-400">Size</dt>
          <dd class="font-medium text-slate-700 dark:text-slate-200">{formatSize(meta.size)}</dd>
        </div>
        <div>
          <dt class="text-slate-400">Expires</dt>
          <dd class="font-medium text-slate-700 dark:text-slate-200">{formatExpiry(meta.expiresAt)}</dd>
        </div>
      </dl>

      {#if meta.hasPassword}
        <div>
          <label for="dlpw" class="mb-1 block text-sm font-medium text-slate-600 dark:text-slate-300">
            🔒 This file requires a password
          </label>
          <input
            id="dlpw"
            type="password"
            bind:value={password}
            placeholder="Enter password"
            class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-slate-800 outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/30 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100"
          />
        </div>
      {/if}

      {#if downloadError}
        <p class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">{downloadError}</p>
      {/if}

      <button
        onclick={handleDownload}
        disabled={downloading || deleting}
        class="w-full rounded-lg bg-indigo-600 px-4 py-2.5 font-semibold text-white shadow-lg shadow-indigo-600/30 transition hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-60"
      >
        {downloading ? 'Downloading…' : 'Download'}
      </button>

      <button
        onclick={handleDelete}
        disabled={downloading || deleting}
        class="w-full rounded-lg border border-red-300 px-4 py-2.5 font-semibold text-red-600 transition hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-60 dark:border-red-900/60 dark:text-red-400 dark:hover:bg-red-950/30"
      >
        {deleting ? 'Deleting…' : '🗑️ Delete file'}
      </button>
    </div>
  {/if}
</div>
