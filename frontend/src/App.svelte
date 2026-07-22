<script>
  import '@fontsource/geist-sans'
  import '@fontsource/geist-mono'
  import { onMount, onDestroy } from 'svelte'
  import Sidebar from './lib/Sidebar.svelte'
  import SessionList from './lib/SessionList.svelte'
  import SessionDetail from './lib/SessionDetail.svelte'
  import DiffViewer from './lib/DiffViewer.svelte'
  import ConfigPanel from './lib/ConfigPanel.svelte'
  import SnapshotPanel from './lib/SnapshotPanel.svelte'
  import Toast from './lib/Toast.svelte'
  import { GetSessions, GetConnectedAgents, GetAggregatedCost } from '../wailsjs/go/main/App'
  import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'

  let sessions = []
  let agents = []
  let selectedTab = 'sessions'
  let selectedSession = null
  let totalCost = 0
  let theme = localStorage.getItem('vibe-theme') || 'dark'
  let loading = true
  let toasts = []

  // Apply theme on init
  document.documentElement.setAttribute('data-theme', theme)

  function addToast(message, type = 'info') {
    const id = Date.now()
    toasts = [...toasts, { id, message, type }]
    setTimeout(() => {
      toasts = toasts.filter(t => t.id !== id)
    }, 3500)
  }

  onMount(async () => {
    await loadData()
    loading = false
    EventsOn('sessions-updated', (data) => {
      if (data) {
        sessions = data
        // Update selected session if it exists
        if (selectedSession) {
          const updated = data.find(s => s.id === selectedSession.id)
          if (updated) selectedSession = updated
        }
      }
    })
  })

  onDestroy(() => {
    EventsOff('sessions-updated')
  })

  async function loadData() {
    try {
      const [s, a, c] = await Promise.all([
        GetSessions(),
        GetConnectedAgents(),
        GetAggregatedCost()
      ])
      sessions = s || []
      agents = a || []
      totalCost = c || 0
    } catch (err) {
      addToast('Failed to load data: ' + err.message, 'error')
    }
  }

  function selectSession(s) {
    selectedSession = s
    selectedTab = 'detail'
  }

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark'
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem('vibe-theme', theme)
  }

  async function handleRefresh() {
    loading = true
    await loadData()
    loading = false
    addToast('Data refreshed', 'success')
  }
</script>

<div class="app">
  <Sidebar
    {agents}
    {totalCost}
    {selectedTab}
    onTabChange={(t) => selectedTab = t}
    onRefresh={handleRefresh}
    {theme}
    onToggleTheme={toggleTheme}
  />
  <main class="main-content">
    {#if loading}
      <div class="loading-overlay">
        <div class="spinner"></div>
        <span>Loading...</span>
      </div>
    {:else if selectedTab === 'sessions'}
      <SessionList {sessions} onSelect={selectSession} />
    {:else if selectedTab === 'detail'}
      <SessionDetail session={selectedSession} onToast={addToast} />
    {:else if selectedTab === 'diff'}
      <DiffViewer />
    {:else if selectedTab === 'snapshots'}
      <SnapshotPanel onToast={addToast} />
    {:else if selectedTab === 'config'}
      <ConfigPanel {agents} {theme} onToggleTheme={toggleTheme} />
    {/if}
  </main>
</div>

<Toast {toasts} />

<style>
  .app {
    display: flex;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
    font-family: var(--font-sans);
  }
  .main-content {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
    transition: background-color var(--transition-speed) var(--ease-out);
  }
</style>
