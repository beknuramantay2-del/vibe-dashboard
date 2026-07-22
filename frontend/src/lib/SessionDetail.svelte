<script>
  import { onMount } from 'svelte'
  import { GetFileChanges, KillSession, CreateSnapshot } from '../../wailsjs/go/main/App'

  export let session = null
  export let onToast = () => {}

  let fileChanges = []
  let repoPath = ''
  let killing = false
  let snapping = false

  $: if (session) loadChanges()

  async function loadChanges() {
    if (!session) return
    try {
      const result = await GetFileChanges(session.id, session.agent)
      fileChanges = result || []
    } catch (err) {
      onToast('Failed to load file changes: ' + err.message, 'error')
    }
  }

  async function handleKill() {
    if (!confirm('Kill this session? The agent process will be interrupted.')) return
    killing = true
    try {
      const result = await KillSession(session.id, session.agent)
      if (result.ok) {
        onToast('Session terminated', 'success')
      } else {
        onToast('Kill failed: ' + result.message, 'error')
      }
    } catch (err) {
      onToast('Error: ' + err.message, 'error')
    }
    killing = false
  }

  async function handleSnapshot() {
    if (!repoPath.trim()) {
      onToast('Enter a repository path first', 'error')
      return
    }
    snapping = true
    try {
      const result = await CreateSnapshot(session.id, repoPath)
      if (result.ok) {
        onToast(result.message, 'success')
        repoPath = ''
      } else {
        onToast('Snapshot failed: ' + result.message, 'error')
      }
    } catch (err) {
      onToast('Error: ' + err.message, 'error')
    }
    snapping = false
  }

  function cacheColor(rate) {
    if (rate >= 80) return 'var(--green)'
    if (rate >= 50) return 'var(--yellow)'
    return 'var(--red)'
  }

  function agentColor(agent) {
    if (agent === 'Claude Code') return '#a78bfa'
    if (agent === 'OpenCode') return '#34d399'
    if (agent === 'Codex CLI') return '#60a5fa'
    return '#999'
  }
</script>

<div class="detail">
  {#if session}
    <div class="header">
      <h2>Session Detail</h2>
      <span class="agent-badge" style="background: {agentColor(session.agent)}">{session.agent}</span>
      <span class="status-badge" class:active={session.status === 'active'} class:completed={session.status === 'completed'}>
        {session.status}
      </span>
    </div>

    <div class="stats-grid">
      <div class="stat">
        <div class="stat-label">Session ID</div>
        <div class="stat-value mono">{session.id}</div>
      </div>
      <div class="stat">
        <div class="stat-label">Project</div>
        <div class="stat-value">{session.project || '—'}</div>
      </div>
      <div class="stat">
        <div class="stat-label">Duration</div>
        <div class="stat-value">{session.duration}</div>
      </div>
      <div class="stat">
        <div class="stat-label">Start Time</div>
        <div class="stat-value">{session.startTime}</div>
      </div>
      <div class="stat highlight">
        <div class="stat-label">Cost</div>
        <div class="stat-value accent">${session.cost?.toFixed(4)}</div>
      </div>
      <div class="stat">
        <div class="stat-label">Input Tokens</div>
        <div class="stat-value">{session.inputTokens?.toLocaleString()}</div>
      </div>
      <div class="stat">
        <div class="stat-label">Output Tokens</div>
        <div class="stat-value">{session.outputTokens?.toLocaleString()}</div>
      </div>
      <div class="stat">
        <div class="stat-label">Cache Hit Rate</div>
        <div class="stat-value" style="color: {cacheColor(session.cacheHitRate || 0)}">
          {(session.cacheHitRate || 0).toFixed(1)}%
        </div>
      </div>
    </div>

    {#if session.pid > 0}
      <div class="pid-row">
        <span class="pid-label">PID:</span>
        <code>{session.pid}</code>
      </div>
    {/if}

    <div class="section">
      <h3>File Changes</h3>
      {#if fileChanges.length}
        {#each fileChanges as fc}
          <div class="file-row">
            <span class="file-path">{fc.path}</span>
            <span class="file-add">+{fc.additions}</span>
            <span class="file-del">-{fc.deletions}</span>
          </div>
        {/each}
      {:else}
        <div class="empty-small">No file changes recorded for this session</div>
      {/if}
    </div>

    <div class="section">
      <h3>Actions</h3>
      <div class="actions">
        {#if session.status === 'active'}
          <button class="btn btn-danger" on:click={handleKill} disabled={killing}>
            {killing ? 'Killing...' : '⚡ Kill Session'}
          </button>
        {/if}
        <div class="snapshot-row">
          <input
            class="repo-input"
            bind:value={repoPath}
            placeholder="/path/to/repo"
          />
          <button class="btn" on:click={handleSnapshot} disabled={snapping}>
            {snapping ? 'Creating...' : '📸 Snapshot'}
          </button>
        </div>
      </div>
    </div>
  {:else}
    <div class="empty">
      <div class="empty-icon">◎</div>
      <p>Select a session from the list to view details</p>
    </div>
  {/if}
</div>

<style>
  .detail { max-width: 800px; }
  .header { display: flex; align-items: center; gap: 12px; margin-bottom: 20px; flex-wrap: wrap; }
  .header h2 { font-size: 20px; color: var(--text); }
  .agent-badge {
    padding: 4px 10px; border-radius: var(--radius-sm);
    font-size: 12px; color: #fff; font-weight: 600;
  }
  .status-badge {
    padding: 4px 10px; border-radius: var(--radius-sm); font-size: 11px;
  }
  .status-badge.active { background: var(--green-bg); color: var(--green); }
  .status-badge.completed { background: rgba(100,100,100,0.15); color: var(--text2); }

  .stats-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 12px; margin-bottom: 20px; }
  .stat {
    background: var(--bg2); border-radius: var(--radius-lg); padding: 14px;
    border: 1px solid var(--border);
    transition: background-color var(--transition-speed) var(--ease-out),
                border-color var(--transition-speed) var(--ease-out);
  }
  .stat.highlight { border-color: var(--accent); }
  .stat-label { font-size: 11px; color: var(--text3); text-transform: uppercase; letter-spacing: 1px; margin-bottom: 4px; }
  .stat-value { font-size: 16px; font-weight: 600; color: var(--text); word-break: break-all; }
  .stat-value.mono { font-family: var(--font-mono); font-size: 12px; }
  .stat-value.accent { color: var(--accent); }

  .pid-row {
    display: flex; align-items: center; gap: 8px; margin-bottom: 20px;
    padding: 8px 12px; background: var(--bg2); border-radius: var(--radius-md);
    font-size: 13px;
  }
  .pid-label { color: var(--text2); }
  .pid-row code { font-family: var(--font-mono); color: var(--text); }

  .section { margin-bottom: 24px; }
  .section h3 { font-size: 14px; color: var(--text); margin-bottom: 12px; font-weight: 600; }
  .file-row {
    display: flex; align-items: center; gap: 12px; padding: 8px 12px;
    background: var(--bg2); border-radius: var(--radius-md); margin-bottom: 4px;
    font-size: 13px; transition: background-color 0.15s var(--ease-out);
  }
  .file-row:hover { background: var(--bg3); }
  .file-path { flex: 1; font-family: var(--font-mono); color: var(--text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .file-add { color: var(--green); font-family: var(--font-mono); }
  .file-del { color: var(--red); font-family: var(--font-mono); }

  .actions { display: flex; flex-wrap: wrap; gap: 8px; align-items: center; }
  .btn {
    padding: 8px 16px; border-radius: var(--radius-md); border: 1px solid var(--border);
    background: var(--bg2); color: var(--text); font-size: 13px; cursor: pointer;
    transition: all 0.2s var(--ease-out); font-family: var(--font-sans);
  }
  .btn:hover:not(:disabled) { background: var(--bg3); border-color: var(--accent); }
  .btn:active:not(:disabled) { transform: scale(0.97); }
  .btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .btn-danger { background: var(--red); color: #fff; border-color: var(--red); }
  .btn-danger:hover:not(:disabled) { background: #dc2626; border-color: #dc2626; }
  .snapshot-row { display: flex; gap: 8px; flex: 1; min-width: 280px; }
  .repo-input {
    flex: 1; padding: 8px 12px; border-radius: var(--radius-md); border: 1px solid var(--border);
    background: var(--bg2); color: var(--text); font-size: 13px; outline: none;
    font-family: var(--font-mono);
    transition: border-color 0.2s var(--ease-out), box-shadow 0.2s var(--ease-out);
  }
  .repo-input:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15); }

  .empty {
    display: flex; flex-direction: column; align-items: center; justify-content: center;
    padding: 60px 20px; color: var(--text2);
  }
  .empty-icon { font-size: 48px; margin-bottom: 16px; opacity: 0.3; }
  .empty p { font-size: 14px; }
  .empty-small { padding: 20px; text-align: center; color: var(--text3); font-size: 13px; }
</style>
