<script lang="ts">
  import Upload from './lib/Upload.svelte'
  import Download from './lib/Download.svelte'

  // Minimal client-side routing: "/" shows the uploader, "/f/{id}" shows the
  // download page. No router library needed for two routes.
  let path = $state(window.location.pathname)

  $effect(() => {
    const onPop = () => (path = window.location.pathname)
    window.addEventListener('popstate', onPop)
    return () => window.removeEventListener('popstate', onPop)
  })

  const shareMatch = $derived(path.match(/^\/f\/(.+)$/))
</script>

<div class="flex min-h-screen flex-col">
  <header class="mx-auto w-full max-w-xl px-6 pt-9 pb-2">
    <a href="/" class="rise inline-flex items-center gap-3">
      <span class="grid size-10 place-items-center rounded-2xl bg-gradient-to-br from-[#ff7a4d] to-[#e04e26] text-white shadow-lg shadow-[#e04e26]/30">
        <svg viewBox="0 0 24 24" fill="none" class="size-5" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M12 19V6" /><path d="m5 12 7-7 7 7" />
        </svg>
      </span>
      <span class="font-display text-[1.7rem] font-semibold leading-none tracking-tight">Share Jer</span>
    </a>
  </header>

  <main class="mx-auto w-full max-w-xl flex-1 px-6 py-6">
    {#if shareMatch}
      <Download id={shareMatch[1]} />
    {:else}
      <Upload />
    {/if}
  </main>

  <footer class="mx-auto w-full max-w-xl px-6 py-8">
    <p class="text-center text-sm text-muted">
      Upload a file, share the link — it expires on its own.
    </p>
  </footer>
</div>
