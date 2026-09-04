<script lang="ts">
  import { onMount } from 'svelte'
  import { Ping } from '../wailsjs/go/main/App.js'

  let version = $state('')
  let failed = $state(false)

  onMount(async () => {
    try {
      version = await Ping()
    } catch {
      failed = true
    }
  })
</script>

<main>
  <h1>Leneki</h1>
  {#if failed}
    <p class="status error">could not reach the Go core</p>
  {:else if version}
    <p class="status">version {version}</p>
  {:else}
    <p class="status">connecting</p>
  {/if}
</main>

<style>
  main {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100vh;
    gap: 0.25rem;
  }

  h1 {
    margin: 0;
    font-size: 2.5rem;
    font-weight: 600;
    letter-spacing: -0.02em;
  }

  .status {
    margin: 0;
    font-size: 0.9rem;
    opacity: 0.6;
  }

  .error {
    color: #ff8a80;
    opacity: 1;
  }
</style>
