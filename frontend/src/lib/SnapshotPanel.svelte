<script>
  import { onMount } from 'svelte'
  import { ListSnapshots, Rollback, DeleteSnapshot } from '../../wailsjs/go/main/App'

  export let onToast = () => {}

  let snapshots = []
  let loading = true
  let repoPath = ''
  let rollingBack = null

  onMount(loadSnapshots)

  async function loadSnapshots() {
    loading = true
    try {
      const result = await ListSnapshots()
      snapshots = result || []
    } catch (err) {
      onToast('Failed to load snapshots: ' + err.message, 'error')
    }
    loading = false
  }

  async function handleRollback(snapId) {
    if (!repoPath.trim()) {
      onToast('Enter the repository path to rollback', 'error')
      return
    }
    if (!confirm(`Rollback snapshot ${snapId.slice(0, 20)}...? This will apply the stashed changes to ${repoPath}.`)) return
    rollingBack = snapId
    try {
      const result = await Rollback(snapId, repoPath)
      if (result.ok) {
        onToast('Rollback successful!', 'success')
      } else {
        onToast('Rollback failed: ' + result.message, 'error')
      }
    } catch (err) {
      onToast('Error: ' + err.message, 'error')
    }
    rollingBack = null
  }

  async function handleDelete(snapId) {
    if (!confirm('Delete this snapshot metadata?')) return
    try {
      const result = await DeleteSnapshot(snapId)
      if (result.ok) {
        onToast('Snapshot deleted', 'success')
        await loadSnapshots()
      } else {
        onToast('Delete failed: ' + result.message, 'error')
      }
    } catch (err) {
      onToast('Error: ' + err.message, 'error')
    }
  }

  function formatDate(iso) {
    try {
      return new Date(iso).toLocaleString()
    } catch {
      return iso
    }
  }
</script>

<div class="snapshots">
  <div class="header">
    <h2>Snapshots</h2>
    <span class="count">{snapshots.length}</span>
    <button class="btn btn-sm" on:click={loadSnapshots}>⟳ Refresh</button>
  </div>

  <div class="repo-bar">
    <label for="snap-repo">Repository path for rollback:</label>
    <input id="snap-repo" class="repo-input" bind:value={repoPath} placeholder="/path/to/your/repo" />
  </div>

  {#if loading}
    <div class="loading-overlay">
      <div class="spinner"></div>
      <span>Loading snapshots...</span>
    </div>
  {:else if snapshots.length === 0}
    <div class="empty">
      <div class="empty-icon">⏱</div>
      <p>No snapshots yet</p>
      <p class="sub">Create a snapshot from a session's detail view to enable rollback.</p>
    </div>
  {:else}
    <div class="snap-list">
      {#each snapshots as snap (snap.id)}
        <div class="snap-card">
          <div class="snap-info">
            <div class="snap-id">{snap.id}</div>
            <div class="snap-meta">
              <span>Session: {snap.sessionId || '—'}</span>
              <span>•</span>
              <span>{formatDate(snap.createdAt)}</span>
            </div>
            {#if snap.message}
              <div class="snap-msg">{snap.message}</div>
            {/if}
          </div>
          <div class="snap-actions">
            <button
              class="btn btn-primary"
              on:click={() => handleRollback(snap.id)}
              disabled={rollingBack === snap.id}
            >
              {rollingBack === snap.id ? 'Rolling back...' : '↩ Rollback'}
            </button>
            <button class="btn btn-ghost" on:click={() => handleDelete(snap.id)}>✕</button>
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .snapshots { max-width: 800px; }
  .header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
  .header h2 { font-size: 20px; color: var(--text); }
  .count {
    font-size: 12px; background: var(--bg3); color: var(--text2);
    padding: 2px 8px; border-radius: var(--radius-sm);
  }
  .repo-bar { margin-bottom: 20px; }
  .repo-bar label { display: block; font-size: 12px; color: var(--text2); margin-bottom: 6px; }
  .repo-input {
    width: 100%; padding: 8px 12px; border-radius: var(--radius-md);
    border: 1px solid var(--border); background: var(--bg2); color: var(--text);
    font-size: 13px; font-family: var(--font-mono); outline: none;
    transition: border-color 0.2s var(--ease-out), box-shadow 0.2s var(--ease-out);
  }
  .repo-input:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15); }

  .snap-list { display: flex; flex-direction: column; gap: 8px; }
  .snap-card {
    display: flex; justify-content: space-between; align-items: center; gap: 16px;
    background: var(--bg2); border: 1px solid var(--border); border-radius: var(--radius-lg);
    padding: 14px 16px;
    transition: background-color 0.15s var(--ease-out), border-color 0.15s var(--ease-out);
  }
  .snap-card:hover { border-color: var(--accent); background: var(--bg3); }
  .snap-info { flex: 1; min-width: 0; }
  .snap-id { font-family: var(--font-mono); font-size: 12px; color: var(--accent); margin-bottom: 4px; }
  .snap-meta { font-size: 12px; color: var(--text2); display: flex; gap: 6px; flex-wrap: wrap; }
  .snap-msg { font-size: 12px; color: var(--text3); margin-top: 4px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .snap-actions { display: flex; gap: 6px; flex-shrink: 0; }

  .btn {
    padding: 6px 14px; border-radius: var(--radius-md); border: 1px solid var(--border);
    background: var(--bg2); color: var(--text); font-size: 12px; cursor: pointer;
    transition: all 0.2s var(--ease-out); font-family: var(--font-sans);
    white-space: nowrap;
  }
  .btn:hover:not(:disabled) { background: var(--bg3); }
  .btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .btn-sm { padding: 4px 10px; font-size: 12px; }
  .btn-primary { background: var(--accent); color: #fff; border-color: var(--accent); }
  .btn-primary:hover:not(:disabled) { background: var(--accent-hover); }
  .btn-ghost { background: transparent; border-color: transparent; color: var(--text3); padding: 6px 8px; }
  .btn-ghost:hover { color: var(--red); background: var(--red-bg); }

  .empty {
    display: flex; flex-direction: column; align-items: center; justify-content: center;
    padding: 60px 20px; color: var(--text2);
  }
  .empty-icon { font-size: 48px; margin-bottom: 16px; opacity: 0.3; }
  .empty p { font-size: 14px; margin-bottom: 4px; }
  .empty .sub { font-size: 12px; color: var(--text3); }
</style>
