<script lang="ts">
  import { uploadFile, deleteFile, qrUrl, getConfig, type UploadResult, type ServerConfig } from './api'
  import { formatSize, humanizeDuration } from './format'

  let files = $state<FileList | null>(null)
  let password = $state('')
  let uploading = $state(false)
  let error = $state('')
  let result = $state<UploadResult | null>(null)
  let copied = $state(false)
  let config = $state<ServerConfig | null>(null)
  let deleting = $state(false)
  let deleted = $state(false)

  // Load the server's limits once, on mount, to display them.
  $effect(() => {
    getConfig()
      .then((c) => (config = c))
      .catch(() => {}) // non-critical: just skip the notice if it fails
  })

  const selected = $derived(files?.[0] ?? null)

  async function submit(event: Event) {
    event.preventDefault()
    error = ''
    if (!selected) {
      error = 'Please choose a file first.'
      return
    }
    if (config && selected.size > config.maxUploadBytes) {
      error = `File is too large — the limit is ${config.maxUploadMB} MB.`
      return
    }
    uploading = true
    try {
      result = await uploadFile(selected, password)
    } catch (err) {
      error = err instanceof Error ? err.message : 'Upload failed'
    } finally {
      uploading = false
    }
  }

  async function copyLink() {
    if (!result) return
    await navigator.clipboard.writeText(result.shareUrl)
    copied = true
    setTimeout(() => (copied = false), 1500)
  }

  async function deleteCurrent() {
    if (!result) return
    if (!confirm('Delete this file? The share link will stop working immediately.')) return
    error = ''
    deleting = true
    try {
      await deleteFile(result.id, password)
      deleted = true
    } catch (err) {
      error = err instanceof Error ? err.message : 'Delete failed'
    } finally {
      deleting = false
    }
  }

  function reset() {
    result = null
    files = null
    password = ''
    error = ''
    deleted = false
  }
</script>

<div class="rounded-2xl bg-white p-6 shadow-xl shadow-slate-200/60 dark:bg-slate-900 dark:shadow-none dark:ring-1 dark:ring-slate-800 sm:p-8">
  {#if !result}
    <form onsubmit={submit} class="space-y-5">
      <div>
        <h1 class="text-xl font-semibold text-slate-800 dark:text-slate-100">Share a file</h1>
        <p class="mt-1 text-sm text-slate-500">Pick a file, optionally lock it with a password.</p>
        {#if config}
          <div class="mt-3 flex flex-wrap gap-2 text-xs">
            <span class="inline-flex items-center gap-1 rounded-full bg-slate-100 px-2.5 py-1 font-medium text-slate-600 dark:bg-slate-800 dark:text-slate-300">
              ⬆️ Max {config.maxUploadMB} MB
            </span>
            <span class="inline-flex items-center gap-1 rounded-full bg-slate-100 px-2.5 py-1 font-medium text-slate-600 dark:bg-slate-800 dark:text-slate-300">
              {#if config.expirySeconds > 0}
                ⏱️ Expires after {humanizeDuration(config.expirySeconds)}
              {:else}
                ♾️ No expiry
              {/if}
            </span>
          </div>
        {/if}
      </div>

      <label
        class="flex cursor-pointer flex-col items-center justify-center gap-2 rounded-xl border-2 border-dashed border-slate-300 px-6 py-10 text-center transition hover:border-indigo-400 hover:bg-indigo-50/50 dark:border-slate-700 dark:hover:border-indigo-500 dark:hover:bg-indigo-950/30"
      >
        <span class="text-3xl">📁</span>
        {#if selected}
          <span class="font-medium text-slate-700 dark:text-slate-200">{selected.name}</span>
          <span class="text-sm text-slate-400">{formatSize(selected.size)}</span>
        {:else}
          <span class="font-medium text-slate-600 dark:text-slate-300">Click to choose a file</span>
          <span class="text-sm text-slate-400">or drag it here</span>
        {/if}
        <input type="file" class="hidden" bind:files />
      </label>

      <div>
        <label for="pw" class="mb-1 block text-sm font-medium text-slate-600 dark:text-slate-300">
          Password <span class="font-normal text-slate-400">(optional)</span>
        </label>
        <input
          id="pw"
          type="password"
          bind:value={password}
          placeholder="Leave blank for no password"
          class="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-slate-800 outline-none transition focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/30 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100"
        />
      </div>

      {#if error}
        <p class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">{error}</p>
      {/if}

      <button
        type="submit"
        disabled={uploading}
        class="w-full rounded-lg bg-indigo-600 px-4 py-2.5 font-semibold text-white shadow-lg shadow-indigo-600/30 transition hover:bg-indigo-700 disabled:cursor-not-allowed disabled:opacity-60"
      >
        {uploading ? 'Uploading…' : 'Upload & get link'}
      </button>
    </form>
  {:else if deleted}
    <div class="space-y-4 text-center">
      <span class="text-4xl">🗑️</span>
      <h2 class="text-xl font-semibold text-slate-800 dark:text-slate-100">File deleted</h2>
      <p class="text-sm text-slate-500">The share link no longer works.</p>
      <button
        onclick={reset}
        class="text-sm font-medium text-indigo-600 hover:text-indigo-700 dark:text-indigo-400"
      >
        ← Share another file
      </button>
    </div>
  {:else}
    <div class="space-y-5 text-center">
      <div>
        <span class="text-4xl">✅</span>
        <h2 class="mt-2 text-xl font-semibold text-slate-800 dark:text-slate-100">Your file is ready to share</h2>
      </div>

      <div class="flex flex-col items-center gap-3">
        <img
          src={qrUrl(result.id)}
          alt="QR code for the share link"
          class="size-44 rounded-xl bg-white p-2 ring-1 ring-slate-200 dark:ring-slate-700"
        />
        <div class="flex w-full items-center gap-2">
          <input
            readonly
            value={result.shareUrl}
            class="flex-1 truncate rounded-lg border border-slate-300 bg-slate-50 px-3 py-2 text-sm text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-200"
          />
          <button
            onclick={copyLink}
            class="shrink-0 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-700"
          >
            {copied ? 'Copied!' : 'Copy'}
          </button>
        </div>
      </div>

      {#if error}
        <p class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-950/40 dark:text-red-400">{error}</p>
      {/if}

      <div class="flex items-center justify-between pt-1">
        <button
          onclick={reset}
          class="text-sm font-medium text-indigo-600 hover:text-indigo-700 dark:text-indigo-400"
        >
          ← Share another file
        </button>
        <button
          onclick={deleteCurrent}
          disabled={deleting}
          class="text-sm font-medium text-red-600 hover:text-red-700 disabled:opacity-60 dark:text-red-400"
        >
          {deleting ? 'Deleting…' : '🗑️ Delete file'}
        </button>
      </div>
    </div>
  {/if}
</div>
