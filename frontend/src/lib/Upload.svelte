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
  let dragging = $state(false)
  let fileInput = $state<HTMLInputElement>()

  // Load the server's limits once, on mount, to display them.
  $effect(() => {
    getConfig()
      .then((c) => (config = c))
      .catch(() => {}) // non-critical: just skip the notice if it fails
  })

  const selected = $derived(files?.[0] ?? null)

  // Drag & drop. dragover must preventDefault so the browser fires drop instead
  // of navigating to the file. On drop we copy dataTransfer.files into `files`
  // (the same FileList the hidden input binds to), so the rest of the flow is
  // identical to clicking and picking.
  function onDragOver(event: DragEvent) {
    event.preventDefault()
    dragging = true
  }

  function onDragLeave() {
    dragging = false
  }

  function onDrop(event: DragEvent) {
    event.preventDefault()
    dragging = false
    const dropped = event.dataTransfer?.files
    if (dropped && dropped.length > 0) {
      files = dropped
    }
  }

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

<div class="card rise p-6 sm:p-8">
  {#if !result}
    <form onsubmit={submit} class="space-y-6">
      <div>
        <p class="eyebrow">Send anything</p>
        <h1 class="mt-1.5 text-3xl">Share a file</h1>
        <p class="mt-2 text-sm text-muted">Pick a file, optionally lock it with a password.</p>
        {#if config}
          <div class="mt-4 flex flex-wrap gap-2">
            <span class="chip">
              <svg viewBox="0 0 24 24" fill="none" class="size-3.5" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M12 19V6" /><path d="m5 12 7-7 7 7" /></svg>
              Max {config.maxUploadMB} MB
            </span>
            <span class="chip">
              {#if config.expirySeconds > 0}
                <svg viewBox="0 0 24 24" fill="none" class="size-3.5" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></svg>
                Expires in {humanizeDuration(config.expirySeconds)}
              {:else}
                No expiry
              {/if}
            </span>
          </div>
        {/if}
      </div>

      <button
        type="button"
        onclick={() => fileInput?.click()}
        ondragover={onDragOver}
        ondragleave={onDragLeave}
        ondrop={onDrop}
        class="group flex w-full cursor-pointer flex-col items-center justify-center gap-3 rounded-2xl border border-dashed px-6 py-11 text-center transition {dragging
          ? 'scale-[1.01] border-accent bg-accent-soft/60'
          : 'border-line bg-[#fffdfb] hover:border-accent hover:bg-accent-soft/40'}"
      >
        <span class="grid size-12 place-items-center rounded-2xl bg-accent-soft text-accent-deep transition {dragging ? 'scale-110' : 'group-hover:scale-105'}">
          <svg viewBox="0 0 24 24" fill="none" class="size-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 16V4" /><path d="m7 9 5-5 5 5" /><path d="M5 16v2a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-2" /></svg>
        </span>
        {#if dragging}
          <span class="font-medium text-accent-deep">Drop your file to upload</span>
          <span class="text-xs text-muted">release anywhere in this box</span>
        {:else if selected}
          <span class="max-w-full truncate px-2 font-medium text-ink">{selected.name}</span>
          <span class="text-xs font-medium text-muted">{formatSize(selected.size)} · click to change</span>
        {:else}
          <span class="font-medium text-ink">Click to choose a file</span>
          <span class="text-xs text-muted">or drag &amp; drop it here</span>
        {/if}
      </button>
      <input bind:this={fileInput} type="file" class="hidden" bind:files />

      <div>
        <label for="pw" class="mb-1.5 block text-sm font-medium text-ink">
          Password <span class="font-normal text-muted">· optional</span>
        </label>
        <input id="pw" type="password" bind:value={password} placeholder="Leave blank for no password" class="field" />
      </div>

      {#if error}
        <p class="alert">{error}</p>
      {/if}

      <button type="submit" disabled={uploading} class="btn-primary w-full">
        {uploading ? 'Uploading…' : 'Upload & get link'}
      </button>
    </form>
  {:else if deleted}
    <div class="space-y-4 py-6 text-center">
      <span class="mx-auto grid size-14 place-items-center rounded-full bg-accent-soft text-accent-deep">
        <svg viewBox="0 0 24 24" fill="none" class="size-7" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 6h18" /><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" /></svg>
      </span>
      <h2 class="text-2xl">File deleted</h2>
      <p class="text-sm text-muted">The share link no longer works.</p>
      <button onclick={reset} class="btn-ghost mx-auto">← Share another file</button>
    </div>
  {:else}
    <div class="space-y-6 text-center">
      <div>
        <span class="mx-auto grid size-14 place-items-center rounded-full bg-accent-soft text-accent-deep">
          <svg viewBox="0 0 24 24" fill="none" class="size-7" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5" /></svg>
        </span>
        <h2 class="mt-3 text-2xl">Ready to share</h2>
        <p class="mt-1 text-sm text-muted">Scan the code or copy the link.</p>
      </div>

      <div class="flex flex-col items-center gap-4">
        <div class="rounded-2xl border border-line bg-white p-3 shadow-sm">
          <img src={qrUrl(result.id)} alt="QR code for the share link" class="size-40 rounded-lg" />
        </div>
        <div class="flex w-full items-center gap-2">
          <input readonly value={result.shareUrl} class="field flex-1 truncate text-left text-sm" />
          <button onclick={copyLink} class="btn-primary shrink-0 px-4">{copied ? 'Copied!' : 'Copy'}</button>
        </div>
      </div>

      {#if error}
        <p class="alert text-left">{error}</p>
      {/if}

      <div class="flex items-center justify-between border-t border-line pt-4">
        <button onclick={reset} class="text-sm font-semibold text-accent-deep transition hover:text-accent">
          ← Share another
        </button>
        <button
          onclick={deleteCurrent}
          disabled={deleting}
          class="text-sm font-semibold text-muted transition hover:text-[#b4341a] disabled:opacity-50"
        >
          {deleting ? 'Deleting…' : 'Delete file'}
        </button>
      </div>
    </div>
  {/if}
</div>
