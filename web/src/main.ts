import './styles.css'

const API_URL = (import.meta.env.VITE_TASKS_API_URL || 'http://100.99.22.62:8000').replace(/\/$/, '')

type Task = {
  id: number
  body: string
  priority: number
  is_completed: boolean
  parent_id: number | null
  children?: Task[]
}

type Row = {
  task: Task
  indent: number
}

type Mode = 'normal' | 'insert' | 'edit'

const state: {
  tasks: Task[]
  cursor: number
  mode: Mode
  draft: { body: string; parentId: number | null; taskId: number | null }
  isLoading: boolean
  isSaving: boolean
  error: string
} = {
  tasks: [],
  cursor: 0,
  mode: 'normal',
  draft: { body: '', parentId: null, taskId: null },
  isLoading: true,
  isSaving: false,
  error: '',
}

const app = document.querySelector<HTMLDivElement>('#root')

if (!app) {
  throw new Error('Missing #root element')
}

app.innerHTML = `
  <div class="sky" aria-hidden="true">
    <span></span><span></span><span></span><span></span><span></span>
  </div>
  <main class="app-shell">
    <section class="task-card" aria-busy="true">
      <header class="app-header">
        <p>Tasks</p>
        <span id="syncState">loading</span>
      </header>
      <section id="taskList" class="task-list" aria-label="Tasks"></section>
      <p id="emptyLine" class="empty-line" hidden>No tasks yet.</p>
      <p id="errorLine" class="error-line" hidden></p>
      <footer class="app-footer">
        <span id="apiUrl"></span>
      </footer>
    </section>
    <nav class="action-island" aria-label="Task actions">
      <button type="button" data-action="root">Parent</button>
      <button type="button" data-action="child">Child</button>
      <button type="button" data-action="edit">Edit</button>
      <button type="button" data-action="done">Done</button>
      <button type="button" data-action="delete">Delete</button>
      <button type="button" data-action="refresh">Refresh</button>
    </nav>
  </main>
`

const taskList = document.querySelector<HTMLElement>('#taskList')!
const emptyLine = document.querySelector<HTMLParagraphElement>('#emptyLine')!
const errorLine = document.querySelector<HTMLParagraphElement>('#errorLine')!
const syncState = document.querySelector<HTMLSpanElement>('#syncState')!
const apiUrl = document.querySelector<HTMLSpanElement>('#apiUrl')!
const card = document.querySelector<HTMLElement>('.task-card')!
const actionIsland = document.querySelector<HTMLElement>('.action-island')!

apiUrl.textContent = API_URL

void loadTasks()

window.addEventListener('keydown', (event) => {
  const isTyping = event.target instanceof HTMLElement && event.target.matches('input, textarea')

  if (state.mode !== 'normal') {
    if (event.key === 'Escape') {
      event.preventDefault()
      cancelMode()
    }
    return
  }

  if (isTyping) return

  const rows = taskRows(state.tasks)
  const selected = rows[state.cursor]?.task ?? null

  if (event.key === 'ArrowUp' || event.key === 'k') {
    event.preventDefault()
    moveCursor(-1)
  } else if (event.key === 'ArrowDown' || event.key === 'j') {
    event.preventDefault()
    moveCursor(1)
  } else if (event.key === 'Enter') {
    event.preventDefault()
    if (selected) void toggleTask(selected)
  } else if (event.key === 'n') {
    event.preventDefault()
    startRootInsert()
  } else if (event.key === 'a') {
    event.preventDefault()
    startChildInsert(selected)
  } else if (event.key === 'C') {
    event.preventDefault()
    if (selected) startEdit(selected)
  } else if (event.key === 'D') {
    event.preventDefault()
    if (selected) void deleteTask(selected)
  } else if (event.key === 'r') {
    event.preventDefault()
    void loadTasks()
  }
})

actionIsland.addEventListener('click', (event) => {
  const button = (event.target as HTMLElement).closest<HTMLButtonElement>('button[data-action]')
  if (!button) return

  const selected = getSelectedTask()

  if (button.dataset.action === 'root') startRootInsert()
  if (button.dataset.action === 'child') startChildInsert(selected)
  if (button.dataset.action === 'edit' && selected) startEdit(selected)
  if (button.dataset.action === 'done' && selected) void toggleTask(selected)
  if (button.dataset.action === 'delete' && selected) void deleteTask(selected)
  if (button.dataset.action === 'refresh') void loadTasks()
})

async function loadTasks() {
  state.isLoading = true
  state.error = ''
  render()

  try {
    state.tasks = await readJSON<Task[]>(await fetch(`${API_URL}/tasks/tree/`))
    clampCursor()
  } catch (err) {
    state.error = err instanceof Error ? err.message : String(err)
  } finally {
    state.isLoading = false
    render()
  }
}

function render() {
  const rows = taskRows(state.tasks)
  const insertIndex = rows.findIndex((row) => row.task.id === state.draft.parentId)

  taskList.replaceChildren()

  if (state.mode === 'insert' && state.draft.parentId === null) {
    taskList.append(createInputRow(0))
  }

  rows.forEach((row, index) => {
    if (state.mode === 'edit' && state.draft.taskId === row.task.id) {
      taskList.append(createInputRow(row.indent))
    } else {
      taskList.append(createTaskRow(row, index))
    }

    if (state.mode === 'insert' && index === insertIndex) {
      taskList.append(createInputRow(row.indent + 1))
    }
  })

  emptyLine.hidden = rows.length !== 0 || state.isLoading || state.mode !== 'normal'
  errorLine.hidden = !state.error
  errorLine.textContent = state.error
  syncState.textContent = state.isLoading ? 'loading' : state.isSaving ? 'syncing' : state.mode
  card.setAttribute('aria-busy', String(state.isLoading || state.isSaving))

  const hasSelection = Boolean(rows[state.cursor])
  actionIsland.querySelectorAll<HTMLButtonElement>('[data-action="child"], [data-action="edit"], [data-action="done"], [data-action="delete"]').forEach((button) => {
    button.disabled = !hasSelection
  })
}

function createTaskRow(row: Row, index: number) {
  const task = row.task
  const item = document.createElement('button')
  item.type = 'button'
  item.className = `task-row ${state.mode === 'normal' && state.cursor === index ? 'selected' : ''} ${task.is_completed ? 'completed' : ''}`
  item.style.setProperty('--indent', String(row.indent))
  item.addEventListener('click', () => {
    state.cursor = index
    render()
  })

  const constellation = priorityGlyph(task.priority)
  item.innerHTML = `
    <span class="row-marker" aria-hidden="true">${state.cursor === index && state.mode === 'normal' ? '✦' : '·'}</span>
    <span class="checkbox" aria-hidden="true">${task.is_completed ? '✓' : ''}</span>
    <span class="task-body"></span>
    <span class="task-stars" aria-hidden="true">${constellation}</span>
  `
  item.querySelector('.task-body')!.textContent = task.body

  item.querySelector('.checkbox')!.addEventListener('click', (event) => {
    event.stopPropagation()
    void toggleTask(task)
  })

  return item
}

function createInputRow(indent: number) {
  const form = document.createElement('form')
  form.className = 'task-row input-row'
  form.style.setProperty('--indent', String(indent))

  form.innerHTML = `
    <span class="row-marker" aria-hidden="true">✦</span>
    <span class="checkbox ghost" aria-hidden="true"></span>
    <input aria-label="Task text" placeholder="New task" />
  `

  const input = form.querySelector('input')!
  input.value = state.draft.body
  input.addEventListener('input', () => {
    state.draft.body = input.value
  })
  form.addEventListener('submit', (event) => {
    event.preventDefault()
    void submitDraft()
  })

  setTimeout(() => input.focus(), 0)

  return form
}

function moveCursor(direction: number) {
  const rows = taskRows(state.tasks)
  state.cursor = Math.min(Math.max(state.cursor + direction, 0), Math.max(rows.length - 1, 0))
  render()
}

function startRootInsert() {
  state.mode = 'insert'
  state.draft = { body: '', parentId: null, taskId: null }
  render()
}

function startChildInsert(task: Task | null) {
  if (!task) {
    startRootInsert()
    return
  }

  state.mode = 'insert'
  state.draft = { body: '', parentId: task.parent_id ?? task.id, taskId: null }
  render()
}

function startEdit(task: Task) {
  state.mode = 'edit'
  state.draft = { body: task.body, parentId: null, taskId: task.id }
  render()
}

function cancelMode() {
  state.mode = 'normal'
  state.draft = { body: '', parentId: null, taskId: null }
  render()
}

async function submitDraft() {
  const body = state.draft.body.trim()
  if (!body) return

  state.isSaving = true
  state.error = ''
  render()

  try {
    if (state.mode === 'edit') {
      await readJSON(await fetch(`${API_URL}/tasks/${state.draft.taskId}/`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ body }),
      }))
    } else {
      await readJSON(await fetch(`${API_URL}/tasks/`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ body, parent_id: state.draft.parentId }),
      }))
    }

    cancelMode()
    await loadTasks()
  } catch (err) {
    state.error = err instanceof Error ? err.message : String(err)
  } finally {
    state.isSaving = false
    render()
  }
}

async function toggleTask(task: Task) {
  state.isSaving = true
  state.error = ''
  render()

  try {
    await readJSON(await fetch(`${API_URL}/tasks/${task.id}/completion`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ is_completed: !task.is_completed }),
    }))
    await loadTasks()
  } catch (err) {
    state.error = err instanceof Error ? err.message : String(err)
  } finally {
    state.isSaving = false
    render()
  }
}

async function deleteTask(task: Task) {
  state.isSaving = true
  state.error = ''
  render()

  try {
    await readJSON(await fetch(`${API_URL}/tasks/${task.id}/`, { method: 'DELETE' }))
    await loadTasks()
  } catch (err) {
    state.error = err instanceof Error ? err.message : String(err)
  } finally {
    state.isSaving = false
    render()
  }
}

function getSelectedTask() {
  return taskRows(state.tasks)[state.cursor]?.task ?? null
}

function clampCursor() {
  const rows = taskRows(state.tasks)
  state.cursor = Math.min(state.cursor, Math.max(rows.length - 1, 0))
}

function taskRows(tasks: Task[], indent = 0): Row[] {
  return tasks.flatMap((task) => [{ task, indent }, ...taskRows(task.children ?? [], indent + 1)])
}

function priorityGlyph(priority: number) {
  if (priority === 1) return '✶──✦'
  if (priority === 2) return '✧─✧'
  return '·'
}

async function readJSON<T = unknown>(response: Response): Promise<T> {
  const text = await response.text()
  const data = text ? JSON.parse(text) : null

  if (!response.ok) {
    throw new Error(`${response.status} ${response.statusText}: ${text}`)
  }

  return data as T
}
