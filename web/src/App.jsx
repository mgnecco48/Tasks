import { useEffect, useState } from 'react'

const API_URL = (import.meta.env.VITE_TASKS_API_URL || 'http://127.0.0.1:8000').replace(/\/$/, '')

function App() {
    const [tasks, setTasks] = useState([])
    const [rootBody, setRootBody] = useState('')
    const [childDraft, setChildDraft] = useState(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState('')

    async function loadTasks() {
        setIsLoading(true)
        setError('')

        try {
            const response = await fetch(`${API_URL}/tasks/`)
            const data = await readJSON(response)
            setTasks(data)
        } catch (err) {
            setError(err.message)
        } finally {
            setIsLoading(false)
        }
    }

    useEffect(() => {
        loadTasks()
    }, [])

    async function addTask(body, parentId = null) {
        const trimmedBody = body.trim()
        if (!trimmedBody) return

        setIsLoading(true)
        setError('')

        try {
            await readJSON(
                await fetch(`${API_URL}/tasks/`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        body: trimmedBody,
                        extra_details: null,
                        parent_id: parentId,
                    }),
                }),
            )

            setRootBody('')
            setChildDraft(null)
            await loadTasks()
        } catch (err) {
            setError(err.message)
            setIsLoading(false)
        }
    }

    async function toggleTask(task) {
        setIsLoading(true)
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
            setIsLoading(false)
        }
    }

    const tree = buildTaskTree(tasks)

    return (
        <main className="app-shell">
            <section className="hero">
                <p className="eyebrow">Tasks Tracker</p>
                <h1>Small task client</h1>
                <p className="api-url">API: {API_URL}</p>
            </section>

            <form
                className="add-card"
                onSubmit={(event) => {
                    event.preventDefault()
                    addTask(rootBody)
                }}
            >
                <input
                    value={rootBody}
                    onChange={(event) => setRootBody(event.target.value)}
                    placeholder="Add a root task..."
                />
                <button type="submit">Add</button>
            </form>

            <div className="toolbar">
                <button type="button" onClick={loadTasks} disabled={isLoading}>
                    Refresh
                </button>
                {isLoading && <span>Loading...</span>}
            </div>

            {error && <p className="error">{error}</p>}

            <section className="task-list">
                {tree.length === 0 && !isLoading ? (
                    <p className="empty">No tasks yet.</p>
                ) : (
                    tree.map((task) => (
                        <TaskItem
                            key={task.id}
                            task={task}
                            childDraft={childDraft}
                            setChildDraft={setChildDraft}
                            addTask={addTask}
                            toggleTask={toggleTask}
                        />
                    ))
                )}
            </section>
        </main>
    )
}

function TaskItem({ task, childDraft, setChildDraft, addTask, toggleTask }) {
    const isAddingChild = childDraft?.parentId === task.id

    return (
        <article className={`task-card ${task.is_completed ? 'completed' : ''}`}>
            <div className="task-row">
                <button
                    className="checkbox"
                    type="button"
                    aria-label={`Mark ${task.body} as ${task.is_completed ? 'incomplete' : 'complete'}`}
                    onClick={() => toggleTask(task)}
                >
                    {task.is_completed ? '✓' : ''}
                </button>

                <div className="task-main">
                    <p>{task.body}</p>
                    {task.extra_details && <small>{task.extra_details}</small>}
                </div>

                <button
                    className="secondary-button"
                    type="button"
                    onClick={() => setChildDraft({ parentId: task.id, body: '' })}
                >
                    Add child
                </button>
            </div>

            {isAddingChild && (
                <form
                    className="child-form"
                    onSubmit={(event) => {
                        event.preventDefault()
                        addTask(childDraft.body, task.id)
                    }}
                >
                    <input
                        autoFocus
                        value={childDraft.body}
                        onChange={(event) => setChildDraft({ parentId: task.id, body: event.target.value })}
                        placeholder={`Child task for ${task.body}`}
                    />
                    <button type="submit">Save</button>
                    <button type="button" onClick={() => setChildDraft(null)}>
                        Cancel
                    </button>
                </form>
            )}

            {task.children.length > 0 && (
                <div className="children">
                    {task.children.map((child) => (
                        <TaskItem
                            key={child.id}
                            task={child}
                            childDraft={childDraft}
                            setChildDraft={setChildDraft}
                            addTask={addTask}
                            toggleTask={toggleTask}
                        />
                    ))}
                </div>
            )}
        </article>
    )
}

async function readJSON(response) {
    const text = await response.text()
    const data = text ? JSON.parse(text) : null

    if (!response.ok) {
        throw new Error(`${response.status} ${response.statusText}: ${text}`)
    }

    return data
}

function buildTaskTree(tasks) {
    const nodesById = new Map()
    const roots = []

    for (const task of tasks) {
        nodesById.set(task.id, { ...task, children: [] })
    }

    for (const task of nodesById.values()) {
        if (task.parent_id && nodesById.has(task.parent_id)) {
            nodesById.get(task.parent_id).children.push(task)
        } else {
            roots.push(task)
        }
    }

    sortTree(roots)
    return roots
}

function sortTree(tasks) {
    tasks.sort((a, b) => a.id - b.id)
    for (const task of tasks) {
        sortTree(task.children)
    }
}

export default App
