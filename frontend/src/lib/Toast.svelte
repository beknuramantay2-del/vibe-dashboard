<script>
  export let toasts = []
</script>

{#if toasts.length > 0}
  <div class="toast-container">
    {#each toasts as toast (toast.id)}
      <div class="toast {toast.type}">
        {toast.message}
      </div>
    {/each}
  </div>
{/if}
