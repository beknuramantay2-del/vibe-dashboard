import { vitePreprocess } from '@sveltejs/vite-plugin-svelte'

export default {
  preprocess: vitePreprocess(),
  compilerOptions: {
    // All components use Svelte 4 syntax (export let, $:, on:click)
    // Force legacy mode for full compatibility
    runes: false
  }
}
