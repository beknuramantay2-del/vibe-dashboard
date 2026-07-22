<script>
  import '@fontsource/geist-sans'
  import '@fontsource/geist-mono'
  import { onMount, onDestroy } from 'svelte'
  import Sidebar from './lib/Sidebar.svelte'
  import SessionList from './lib/SessionList.svelte'
  import SessionDetail from './lib/SessionDetail.svelte'
  import DiffViewer from './lib/DiffViewer.svelte'
  import ConfigPanel from './lib/ConfigPanel.svelte'
  import { GetSessions, GetConnectedAgents, GetAggregatedCost } from '../wailsjs/go/main/App'
  import { EventsOn, EventsOff } from '../wailsjs/runtime/runtime'

  let sessions = []
  let agents = []
  let selectedTab = 'sessions'
  let selectedSession = null
  let totalCost = 0
  let theme = 'dark'

  onMount(() => {
    loadData()
    EventsOn('sessions-updated', (data) => {
      sessions = data
    })
  })

  onDestroy(() => {
    EventsOff('sessions-updated')
  })

  async function loadData() {
    sessions = await GetSessions()
    agents = await GetConnectedAgents()
    totalCost = await GetAggregatedCost()
  }

  function selectSession(s) {
    selectedSession = s
  }

  function toggleTheme() {
    theme = theme === 'dark' ? 'light' : 'dark'
    document.documentElement.setAttribute('data-theme', theme)
  }
</script>

<div class="app" class:dark={theme === 'dark'} class:light={theme === 'light'}>
  <Sidebar {agents} {totalCost} {selectedTab} onTabChange={(t) => selectedTab = t} onRefresh={loadData} {theme} onToggleTheme={toggleTheme} />
  <main class="main-content">
    {#if selectedTab === 'sessions'}
      <SessionList {sessions} onSelect={selectSession} />
    {:else if selectedTab === 'detail' && selectedSession}
      <SessionDetail session={selectedSession} />
    {:else if selectedTab === 'diff'}
      <DiffViewer />
    {:else if selectedTab === 'config'}
      <ConfigPanel {agents} {theme} onToggleTheme={toggleTheme} />
    {/if}
  </main>
</div>

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
    transition: background-color 0.3s var(--ease-out);
  }
</style>
