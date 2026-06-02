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

<div class="min-h-screen bg-gradient-to-br from-slate-50 to-indigo-100 dark:from-slate-950 dark:to-indigo-950">
  <header class="mx-auto flex max-w-2xl items-center gap-2 px-6 pt-10 pb-2">
    <a href="/" class="flex items-center gap-2 text-2xl font-bold text-slate-800 dark:text-slate-100">
      <span class="grid size-9 place-items-center rounded-xl bg-indigo-600 text-white shadow-lg shadow-indigo-600/30">📤</span>
      Share Jer
    </a>
  </header>

  <main class="mx-auto max-w-2xl px-6 py-6">
    {#if shareMatch}
      <Download id={shareMatch[1]} />
    {:else}
      <Upload />
    {/if}
  </main>

  <footer class="mx-auto max-w-2xl px-6 py-8 text-center text-sm text-slate-400">
    Upload a file, share the link. Files may expire automatically.
  </footer>
</div>
