<script>
  let before = ''
  let after = ''

  $: lines = computeDiff(before, after)

  function computeDiff(b, a) {
    if (!b && !a) return [{ type: 'empty', text: 'Enter or paste before/after text to compare' }]
    const beforeLines = b.split('\n')
    const afterLines = a.split('\n')
    const result = []
    const maxLen = Math.max(beforeLines.length, afterLines.length)
    for (let i = 0; i < maxLen; i++) {
      const bl = i < beforeLines.length ? beforeLines[i] : undefined
      const al = i < afterLines.length ? afterLines[i] : undefined
      if (bl === al) {
        result.push({ type: 'equal', text: `  ${bl}`, num: i + 1 })
      } else {
        if (bl !== undefined) result.push({ type: 'remove', text: `- ${bl}`, num: i + 1 })
        if (al !== undefined) result.push({ type: 'add', text: `+ ${al}`, num: i + 1 })
      }
    }
    return result
  }

  function clearAll() {
    before = ''
    after = ''
  }

  $: addCount = lines.filter(l => l.type === 'add').length
  $: delCount = lines.filter(l => l.type === 'remove').length
</script>

<div class="diff-viewer">
  <div class="header">
    <h2>Diff Viewer</h2>
    {#if before || after}
      <button class="btn-clear" on:click={clearAll}>Clear</button>
    {/if}
  </div>

  <div class="inputs">
    <div class="input-group">
      <label for="diff-before">Before</label>
      <textarea id="diff-before" bind:value={before} placeholder="Paste original content..." class="editor"></textarea>
    </div>
    <div class="input-group">
      <label for="diff-after">After</label>
      <textarea id="diff-after" bind:value={after} placeholder="Paste new content..." class="editor"></textarea>
    </div>
  </div>

  <div class="output">
    <div class="output-header">
      <span>Diff Output</span>
      <span class="stats">
        <span class="add-count">+{addCount}</span>
        <span class="del-count">-{delCount}</span>
      </span>
    </div>
    <div class="diff-content">
      {#each lines as line, i}
        <div class="diff-line" class:add={line.type === 'add'} class:remove={line.type === 'remove'} class:equal={line.type === 'equal'} class:empty={line.type === 'empty'}>
          {#if line.type !== 'empty'}
            <span class="line-num">{line.num}</span>
          {/if}
          <code>{line.text}</code>
        </div>
      {/each}
    </div>
  </div>
</div>

<style>
  .diff-viewer { height: 100%; display: flex; flex-direction: column; }
  .header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
  .header h2 { font-size: 20px; color: var(--text); }
  .btn-clear {
    padding: 4px 12px; border-radius: var(--radius-sm); border: 1px solid var(--border);
    background: var(--bg2); color: var(--text2); font-size: 12px; cursor: pointer;
    transition: all 0.2s var(--ease-out); font-family: var(--font-sans);
  }
  .btn-clear:hover { color: var(--red); border-color: var(--red); }

  .inputs { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-bottom: 16px; }
  .input-group label {
    display: block; font-size: 11px; color: var(--text3); text-transform: uppercase;
    letter-spacing: 1px; margin-bottom: 6px;
  }
  .editor {
    width: 100%; height: 180px; padding: 12px; border-radius: var(--radius-md);
    border: 1px solid var(--border); background: var(--bg2); color: var(--text);
    font-family: var(--font-mono); font-size: 12px; resize: vertical; outline: none;
    transition: border-color 0.2s var(--ease-out), box-shadow 0.2s var(--ease-out);
  }
  .editor:focus { border-color: var(--accent); box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15); }
  .editor::placeholder { color: var(--text3); }

  .output { flex: 1; display: flex; flex-direction: column; min-height: 0; }
  .output-header {
    display: flex; justify-content: space-between; align-items: center;
    padding: 8px 12px; background: var(--bg2);
    border-radius: var(--radius-md) var(--radius-md) 0 0;
    border: 1px solid var(--border); font-size: 12px; color: var(--text2);
  }
  .stats { display: flex; gap: 12px; }
  .add-count { color: var(--green); font-weight: 600; }
  .del-count { color: var(--red); font-weight: 600; }
  .diff-content {
    flex: 1; overflow-y: auto; background: var(--bg2);
    border: 1px solid var(--border); border-top: none;
    border-radius: 0 0 var(--radius-md) var(--radius-md);
    padding: 8px 0; font-family: var(--font-mono); font-size: 12px;
  }
  .diff-line { padding: 2px 12px; white-space: pre; display: flex; gap: 8px; }
  .diff-line code { font-family: inherit; flex: 1; }
  .line-num { color: var(--text3); min-width: 32px; text-align: right; user-select: none; }
  .diff-line.add { background: var(--green-bg); }
  .diff-line.add code { color: var(--green); }
  .diff-line.remove { background: var(--red-bg); }
  .diff-line.remove code { color: var(--red); }
  .diff-line.equal code { color: var(--text); }
  .diff-line.empty { padding: 20px; text-align: center; justify-content: center; }
  .diff-line.empty code { color: var(--text3); }
</style>
