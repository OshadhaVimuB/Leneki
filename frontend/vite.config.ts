import {defineConfig} from 'vite'
import {svelte} from '@sveltejs/vite-plugin-svelte'

export default defineConfig({
  plugins: [svelte()],
  build: {
    // Emptying dist would delete the tracked gitkeep that go:embed needs.
    emptyOutDir: false
  }
})
