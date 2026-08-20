# Working Plan

## Direction

The database and FastAPI backend should be the source of truth. Every client should read and write tasks through the API instead of accessing the database directly.

```text
SQLite/Postgres DB -> FastAPI API -> Bubble Tea TUI / React frontend / future clients
```

The original Markdown and Neovim sync idea is still possible, but it should come later.

## Current Backend State

- FastAPI app in `main.py`.
- SQLModel task model backed by SQLite.
- Self-referencing task relationship with `parent_id`, `parent`, and `children`.
- Basic endpoints for listing, creating, deleting, editing, and toggling completion.
- SQLite foreign keys enabled through a SQLAlchemy connection event listener.

## Near-Term Backend Work

- [x] Add client-friendly task responses.

  Keep `GET /tasks/` as a flat list for now. Later, add something like `GET /tasks/tree` to return root tasks with nested children.

- [ ] Validate parent tasks on create.

  If a request includes `parent_id`, check that the parent exists before inserting the child task. Return a clean `404` if it does not.

- [ ] Add Alembic.

  Set up migrations before the schema changes much more. This avoids manually recreating `tasks.db` whenever columns or relationships change.

- [x] Harden completion behavior.

  Make parent/child completion updates predictable. Eventually consider recursive behavior for deeper task trees.

## Client Plan

1. Build the Bubble Tea TUI first.

   This should be the first real client because it fits the terminal-first workflow and makes the app useful quickly.

   MVP features:

   - Fetch tasks from the FastAPI API.
   - Display tasks in a list.
   - Indent child tasks under parents.
   - Move cursor up and down.
   - Toggle task completion.
   - Add a root task.
   - Add a child task.
   - Refresh from the API.

2. Build a simple React frontend second.

   Keep it small and mobile-friendly.

   MVP features:

   - List tasks.
   - Add tasks.
   - Toggle completion.
   - Edit task body and details.
   - Show child tasks.

3. Postpone Markdown/Neovim integration.

   A safer later version would start with one-way export from the database to Markdown, then maybe add an explicit import command. Avoid automatic bidirectional sync until the API and clients are stable.

## Suggested Order

- [ ] Clean up backend response/request models.
- [ ] Add Alembic migrations.
- [ ] Build the Bubble Tea MVP.
- [ ] Build the simple React frontend.
- [ ] Deploy the API somewhere reachable.
- [ ] Add authentication before exposing personal tasks publicly.
- [ ] Revisit Markdown/Neovim sync after the core workflow works.
