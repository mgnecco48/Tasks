import { useEffect, useState } from 'react'

const API_URL = (import.meta.env.VITE_TASKS_API_URL || 'http://127.0.0.1:8000').replace(/\/$/, '')

function App() {
    const [tasks, setTasks] = useState([])
    const [cursor, setCursor] = useState(0)
    const [mode, setMode] = useState('normal')
    const [draft, setDraft] = useState({ body: '', parentId: null, taskId: null })
    const [isLoading, setIsLoading] = useState(true)
    const [isSaving, setIsSaving] = useState(false)
    const [error, setError] = useState('')

    const rows = taskRows(tasks)
    const selected = rows[cursor]?.task ?? null

    async function loadTasks() {
        setIsLoading(true)
        setError('')

        try {
            setTasks(await readJSON(await fetch(`${API_URL}/tasks/tree/`)))
        } catch (err) {
            setError(err.message)
        } finally {
            setIsLoading(false)
        }
    }

    useEffect(() => {
        loadTasks()
    }, [])

    useEffect(() => {
        if (rows.length === 0 && cursor !== 0) {
            setCursor(0)
        } else if (rows.length > 0 && cursor >= rows.length) {
            setCursor(rows.length - 1)
        }
    }, [cursor, rows.length])

    useEffect(() => {
        function handleKeyDown(event) {
            const isTyping = event.target instanceof HTMLElement && event.target.matches('input, textarea')

            if (mode !== 'normal') {
                if (event.key === 'Escape') {
                    event.preventDefault()
                    cancelMode()
                }
                return
            }

            if (isTyping) return


            if (event.key === 'ArrowUp' || event.key === 'k') {
                event.preventDefault()
                moveCursor(-1, rows.length)
            } else if (event.key === 'ArrowDown' || event.key === 'j') {
                event.preventDefault()
                moveCursor(1, rows.length)
            } else if (event.key === 'Enter') {
                event.preventDefault()
                if (selected) toggleTask(selected)
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
                if (selected) deleteTask(selected)
            } else if (event.key === 'r') {
                event.preventDefault()
                loadTasks()
            }
        }

        window.addEventListener('keydown', handleKeyDown)
        return () => window.removeEventListener('keydown', handleKeyDown)
    }, [mode, rows.length, selected])

    function moveCursor(direction, rowCount) {
        setCursor((current) => Math.min(Math.max(current + direction, 0), Math.max(rowCount - 1, 0)))
    }

    function startRootInsert() {
        setMode('insert')
        setDraft({ body: '', parentId: null, taskId: null })
    }

    function startChildInsert(task) {
        if (!task) {
            startRootInsert()
            return
        }

        setMode('insert')
        setDraft({ body: '', parentId: task.parent_id ?? task.id, taskId: null })
    }

    function startEdit(task) {
        setMode('edit')
        setDraft({ body: task.body, parentId: null, taskId: task.id })
    }

    function cancelMode() {
        setMode('normal')
        setDraft({ body: '', parentId: null, taskId: null })
    }

    async function submitDraft(event) {
        event.preventDefault()

        const body = draft.body.trim()
        if (!body) return

        setIsSaving(true)
        setError('')

        try {
            if (mode === 'edit') {
                await readJSON(
                    await fetch(`${API_URL}/tasks/${draft.taskId}/`, {
                        method: 'PATCH',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ body }),
                    }),
                )
            } else {
                await readJSON(
                    await fetch(`${API_URL}/tasks/`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ body, parent_id: draft.parentId }),
                    }),
                )
            }

            cancelMode()
            await loadTasks()
        } catch (err) {
            setError(err.message)
        } finally {
            setIsSaving(false)
        }
    }

    async function toggleTask(task) {
        setIsSaving(true)
        setError('')

        try {
            await readJSON(
                await fetch(`${API_URL}/tasks/${task.id}/completion`, {
                    method: 'PATCH',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ is_completed: !task.is_completed }),
                }),
            )

            await loadTasks()
        } catch (err) {
            setError(err.message)
        } finally {
            setIsSaving(false)
        }
    }

    async function deleteTask(task) {
        setIsSaving(true)
        setError('')

        try {
            await readJSON(
                await fetch(`${API_URL}/tasks/${task.id}/`, {
                    method: 'DELETE',
                }),
            )

            await loadTasks()
        } catch (err) {
            setError(err.message)
        } finally {
            setIsSaving(false)
        }
    }

    return (
        <main className="terminal-shell">
            <section className="terminal-window" aria-busy={isLoading || isSaving}>
                <header className="terminal-chrome" aria-hidden="true">
                    <span></span>
                    <span></span>
                    <span></span>
                    <p>tasks --today</p>
                </header>

                <div className="terminal-body">
                    <div className="title-line">
                        <h1>Today's Tasks:</h1>
                        <span className={`mode-light ${mode !== 'normal' ? 'active' : ''}`}></span>
                    </div>

                    <TaskList
                        rows={rows}
                        cursor={cursor}
                        mode={mode}
                        draft={draft}
                        setCursor={setCursor}
                        setDraft={setDraft}
                        submitDraft={submitDraft}
                        toggleTask={toggleTask}
                    />

                    {rows.length === 0 && !isLoading && mode === 'normal' && (
                        <p className="empty-line"><span>:</span> no tasks yet. press <kbd>n</kbd></p>
                    )}

                    {error && <p className="error-line"><span>!</span> {error}</p>}

                    <footer className="status-line">
                        <span className={mode === 'normal' ? 'mode normal' : 'mode insert'}>
                            {mode === 'normal' ? 'N' : mode === 'insert' ? 'I' : 'C'}
                        </span>
                        <span>{isLoading ? 'loading tree' : isSaving ? 'syncing db' : 'synced'}</span>
                        <span className="api-url">{API_URL}</span>
                    </footer>

                    <nav className="action-bar" aria-label="Task actions">
                        <button type="button" onClick={startRootInsert}>n new</button>
                        <button type="button" onClick={() => startChildInsert(selected)} disabled={!selected}>a child</button>
                        <button type="button" onClick={() => selected && startEdit(selected)} disabled={!selected}>C edit</button>
                        <button type="button" onClick={() => selected && toggleTask(selected)} disabled={!selected}>enter done</button>
                        <button type="button" onClick={() => selected && deleteTask(selected)} disabled={!selected}>D delete</button>
                        <button type="button" onClick={loadTasks}>r reload</button>
                    </nav>

                    <p className="help-line">j/k move · enter toggle · n root · a child · C edit · D delete · esc cancel</p>
                </div>
            </section>
        </main>
    )
}

function TaskList({ rows, cursor, mode, draft, setCursor, setDraft, submitDraft, toggleTask }) {
    const insertIndex = rows.findIndex((row) => row.task.id === draft.parentId)

    return (
        <section className="task-list" aria-label="Tasks">
            {mode === 'insert' && draft.parentId === null && (
                <TaskInputRow draft={draft} setDraft={setDraft} submitDraft={submitDraft} indent={0} />
            )}

            {rows.map((row, index) => (
                <div key={row.task.id}>
                    {mode === 'edit' && draft.taskId === row.task.id ? (
                        <TaskInputRow draft={draft} setDraft={setDraft} submitDraft={submitDraft} indent={row.indent} />
                    ) : (
                        <TaskRow
                            row={row}
                            isSelected={mode === 'normal' && cursor === index}
                            onSelect={() => setCursor(index)}
                            onToggle={() => toggleTask(row.task)}
                        />
                    )}

                    {mode === 'insert' && index === insertIndex && (
                        <TaskInputRow draft={draft} setDraft={setDraft} submitDraft={submitDraft} indent={row.indent + 1} />
                    )}
                </div>
            ))}
        </section>
    )
}

function TaskRow({ row, isSelected, onSelect, onToggle }) {
    const task = row.task
    const box = task.is_completed ? '✓' : ''

    return (
        <div
            className={`task-row ${isSelected ? 'selected' : ''} ${task.is_completed ? 'completed' : ''}`}
            style={{ '--indent': row.indent }}
            onClick={onSelect}
        >
            <span className="cursor" aria-hidden="true">{isSelected ? '>' : ''}</span>
            <span className="rail" aria-hidden="true">:</span>
            {row.indent > 0 && <span className="branch" aria-hidden="true">↳</span>}
            <button
                className="checkbox"
                type="button"
                aria-label={`Mark ${task.body} as ${task.is_completed ? 'incomplete' : 'complete'}`}
                onClick={(event) => {
                    event.stopPropagation()
                    onToggle()
                }}
            >
                {box}
            </button>
            <span className="task-body">{task.body}</span>
        </div>
    )
}

function TaskInputRow({ draft, setDraft, submitDraft, indent }) {
    return (
        <form className="task-row input-row" style={{ '--indent': indent }} onSubmit={submitDraft}>
            <span className="cursor active" aria-hidden="true">›</span>
            <span className="rail" aria-hidden="true">:</span>
            {indent > 0 && <span className="branch" aria-hidden="true">↳</span>}
            <span className="checkbox ghost" aria-hidden="true"></span>
            <input
                autoFocus
                value={draft.body}
                onChange={(event) => setDraft({ ...draft, body: event.target.value })}
                placeholder="New Task"
            />
        </form>
    )
}

function taskRows(tasks, indent = 0) {
    const rows = []

    for (const task of tasks) {
        rows.push({ task, indent })
        rows.push(...taskRows(task.children ?? [], indent + 1))
    }

    return rows
}

async function readJSON(response) {
    const text = await response.text()
    const data = text ? JSON.parse(text) : null

    if (!response.ok) {
        throw new Error(`${response.status} ${response.statusText}: ${text}`)
    }

    return data
}

export default App
