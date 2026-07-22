<script>
  import { onMount } from 'svelte'
  import { GetFileChanges, KillSession, CreateSnapshot } from '../../wailsjs/go/main/App'

  export let session = null

  let fileChanges = []
  let repoPath = ''
  let killResult = ''
  let snapResult = ''

  onMount(async () => {
    if (session) {
      fileChanges = await GetFileChanges(session.id, session.agent)
    }
  })

  async function handleKill() {
    if (!confirm('Kill this session?')) return
    killResult = await KillSession(session.id, session.agent)
  }

  async function handleSnapshot() {
    if (!repoPath) return
    snapResult = await CreateSnapshot(session.id, repoPath)
  }
</script>

<div class="detail">
  {#if session}
    <div class="header">
      <h2>Session Detail</h2>
      <span class="agent-badge" style="background: {session.agent === 'Claude Code' ? '#a78bfa' : session.agent === 'OpenCode' ? '#34d399' : '#60a5fa'}">{session.agent}</span>
    </div>

    <div class="stats-grid">
      <div class="stat"><div class="stat-label">ID</div><div class="stat-value mono">{session.id}</div></div>
      <div class="stat"><div class="stat-label">Project</div><div class="stat-value">{session.project || '-'}</div></div>
      <div class="stat"><div class="stat-label">Status</div><div class="stat-value">{session.status}</div></div>
      <div class="stat"><div class="stat-label">Duration</div><div class="stat-value">{session.duration}</div></div>
      <div class="stat"><div class="stat-label">Cost</div><div class="stat-value accent">${session.cost?.toFixed(4)}</div></div>
      <div class="stat"><div class="stat-label">Input Tokens</div><div class="stat-value">{session.inputTokens?.toLocaleString()}</div></div>
      <div class="stat"><div class="stat-label">Output Tokens</div><div class="stat-value">{session.outputTokens?.toLocaleString()}</div></div>
      <div class="stat"><div class="stat-label">Cache Hit Rate</div><div class="stat-value">{session.cacheHitRate?.toFixed(1)}%</div></div>
    </div>

    <div class="section">
      <h3>File Changes</h3>
      {#if fileChanges?.length}
        {#each fileChanges as fc}
          <div class="file-row">
            <span class="file-path">{fc.path}</span>
            <span class="file-add">+{fc.additions}</span>
            <span class="file-del">-{fc.deletions}</span>
          </div>
        {/each}
      {:else}
        <div class="empty-small">No file changes recorded</div>
      {/if}
    </div>

    <div class="section">
      <h3>Actions</h3>
      <div class="actions">
        <button class="btn btn-danger" on:click={handleKill}>Kill Session</button>
        <div class="snapshot-row">
          <input class="repo-input" bind:value={repoPath} placeholder="Repo path for snapshot..." />
          <button class="btn" on:click={handleSnapshot}>Snapshot</button>
        </div>
      </div>
      {#if killResult}
        <div class="result">{killResult}</div>
      {/if}
      {#if snapResult}
        <div class="result success">Snapshot created</div>
      {/if}
    </div>
  {:else}
    <div class="empty">Select a session from the list</div>
  {/if}
</div>

<style>
  .detail { max-width: 800px; }
  .header { display: flex; align-items: center; gap: 12px; margin-bottom: 20px; }
  .header h2 { font-size: 20px; color: var(--text); }
  .agent-badge { padding: 4px 10px; border-radius: var(--radius-sm); font-size: 12px; color: #fff; font-weight: 600; }
  .stats-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-bottom: 24px; }
  .stat { background: var(--bg2); border-radius: var(--radius-lg); padding: 14px; border: 1px solid var(--border); transition: background-color 0.3s var(--ease-out), border-color 0.3s var(--ease-out); }
  .stat-label { font-size: 11px; color: var(--text2); text-transform: uppercase; letter-spacing: 1px; margin-bottom: 4px; }
  .stat-value { font-size: 16px; font-weight: 600; color: var(--text); }
  .stat-value.mono { font-family: var(--font-mono); font-size: 13px; }
  .stat-value.accent { color: var(--accent); }
  .section { margin-bottom: 24px; }
  .section h3 { font-size: 14px; color: var(--text); margin-bottom: 12px; }
  .file-row { display: flex; align-items: center; gap: 12px; padding: 8px 12px; background: var(--bg2); border-radius: var(--radius-md); margin-bottom: 4px; font-size: 13px; transition: background-color 0.2s var(--ease-out); }
  .file-row:hover { background: var(--bg3); }
  .file-path { flex: 1; font-family: var(--font-mono); color: var(--text); }
  .file-add { color: var(--green); font-family: var(--font-mono); }
  .file-del { color: var(--red); font-family: var(--font-mono); }
  .actions { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; }
  .btn { padding: 8px 16px; border-radius: var(--radius-md); border: 1px solid var(--border); background: var(--bg2); color: var(--text); font-size: 13px; cursor: pointer; transition: all 0.2s var(--ease-out); }
  .btn:hover { background: var(--bg3); border-color: var(--accent); }
  .btn:active { transform: scale(0.97); }
  .btn-danger { background: var(--red); color: #fff; border-color: var(--red); }
  .btn-danger:hover { background: #dc2626; border-color: #dc2626; }
  .snapshot-row { display: flex; gap: 8px; flex: 1; }
  .repo-input { flex: 1; padding: 8px 12px; border-radius: var(--radius-md); border: 1px solid var(--border); background: var(--bg2); color: var(--text); font-size: 13px; outline: none; transition: border-color 0.2s var(--ease-out), box-shadow 0.2s var(--ease-out); }
  .repo-input:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15); }
  .result { margin-top: 12px; padding: 8px 12px; border-radius: var(--radius-sm); background: var(--bg3); color: var(--red); font-size: 13px; }
  .result.success { color: var(--green); }
  .empty, .empty-small { padding: 40px 20px; text-align: center; color: var(--text2); font-size: 14px; }
  .empty-small { padding: 20px; font-size: 13px; }
</style>
