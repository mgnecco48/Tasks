from typing import Annotated, Optional
from sqlmodel import SQLModel, create_engine, Session, select, Field, Relationship
from datetime import datetime, UTC
from fastapi import Depends, FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from sqlalchemy import event


# {{{ Models
def utc_now() -> datetime:
    return datetime.now(UTC)


class TaskBase(SQLModel):
    body: str
    extra_details: str | None = None
    priority: int = Field(default=3, ge=1, le=3)


class Task(TaskBase, table=True):
    __tablename__ = "tasks"  # type: ignore

    id: int | None = Field(default=None, primary_key=True)
    is_completed: bool = False
    created_at: datetime = Field(default_factory=utc_now)
    updated_at: datetime = Field(default_factory=utc_now)
    completed_at: datetime | None = None
    parent_id: int | None = Field(
        default=None, foreign_key="tasks.id", ondelete="CASCADE"
    )

    parent: Optional["Task"] = Relationship(
        back_populates="children", sa_relationship_kwargs={"remote_side": "Task.id"}
    )
    children: list["Task"] = Relationship(back_populates="parent", cascade_delete=True)


class TaskInsert(TaskBase):
    parent_id: int | None = None


class TaskUpdateCompleted(SQLModel):
    is_completed: bool


class TaskUpdate(SQLModel):
    body: str | None = None
    extra_details: str | None = None
    priority: int | None = Field(default=3, ge=1, le=3)


class TaskTreeNode(TaskBase):
    id: int
    is_completed: bool = False
    created_at: datetime = Field(default_factory=utc_now)
    updated_at: datetime = Field(default_factory=utc_now)
    completed_at: datetime | None = None
    priority: int = Field(default=3, ge=1, le=3)
    parent_id: int | None = None

    children: list["TaskTreeNode"] = Field(default_factory=list)


# }}}

# {{{ SQLite setup # TODO: Need to create the databases with something like Alembic, so that when i add a column to the model it automatically updates the databse.
sqlite_filename = "tasks.db"
sqlite_url = f"sqlite:///{sqlite_filename}"
engine = create_engine(sqlite_url, echo=True)


# this enables the foreign key support for all connections for sqlite. Just copied it from the internet though.
@event.listens_for(engine, "connect")
def set_sqlite_pragma(dbapi_connection, _connection_record):
    cursor = dbapi_connection.cursor()
    cursor.execute("PRAGMA foreign_keys=ON")
    cursor.close()


def create_everything():
    SQLModel.metadata.create_all(engine)


def add_tasks():
    task1 = Task(  # {{{
        body="Comprar Cosas",
        extra_details="Ir al Kiwi mejor",
        created_at=datetime.now(UTC),
    )
    task2 = Task(
        body="Hacer Cena",
        extra_details="Arroz con Pollo",
        created_at=datetime.now(UTC),
    )
    child1 = Task(
        body="Leche",
        extra_details="Tine, entera",
        created_at=datetime.now(UTC),
        parent_id=1,
    )
    child2 = Task(
        body="Huevos",
        extra_details="18 pieces, large",
        created_at=datetime.now(UTC),
        parent_id=1,
    )
    child3 = Task(
        body="cocinar arroz",
        extra_details="solo dos tazas",
        created_at=datetime.now(UTC),
        parent_id=2,
    )
    child4 = Task(
        body="cocinar pollo",
        extra_details="dos pechugas",
        created_at=datetime.now(UTC),
        parent_id=2,
    )
    tasks = [task1, task2, child1, child2, child3, child4]

    with Session(engine) as session:
        for task in tasks:
            session.add(task)

        session.commit()

        for task in tasks:
            session.refresh(task)

        for task in tasks:
            print(task)  # }}}


def get_session():
    with Session(engine) as session:
        yield session


SessionDep = Annotated[Session, Depends(get_session)]
# }}}

# {{{ FastAPI part
app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.on_event("startup")
def on_startup():
    create_everything()


@app.get("/")
def get_root():
    return {"Hello": "This is My app"}


@app.get("/tasks/")
async def get_all_tasks(session: SessionDep):
    query = session.exec(select(Task)).all()
    return query


@app.get("/tasks/tree/", response_model=list[TaskTreeNode])
async def get_task_tree(session: SessionDep):
    flat_tasks = session.exec(select(Task)).all()

    nodes = {
        task.id: TaskTreeNode(
            id=task.id,
            body=task.body,
            extra_details=task.extra_details,
            is_completed=task.is_completed,
            created_at=task.created_at,
            updated_at=task.updated_at,
            completed_at=task.completed_at,
            parent_id=task.parent_id,
            priority=task.priority,
            children=[],
        )
        for task in flat_tasks
        if task.id is not None
    }

    roots = []
    for task in flat_tasks:
        if task.id is None:
            continue

        node = nodes[task.id]

        if task.parent_id is not None and task.parent_id in nodes:
            parent_node = nodes[task.parent_id]
            parent_node.children.append(node)
        else:
            roots.append(node)

    def recursive_sort(tasks: list["TaskTreeNode"]):
        tasks.sort(
            key=lambda task: (
                task.is_completed,
                task.priority,
                task.created_at,
            )
        )

        for task in tasks:
            recursive_sort(task.children)

    recursive_sort(roots)

    return roots


@app.post("/tasks/")
async def add_task(session: SessionDep, task: TaskInsert):
    if task.parent_id is not None:
        parent = session.get(Task, task.parent_id)
        if parent is None:
            raise HTTPException(
                status_code=404,
                detail=f"Parent task with id {task.parent_id} is not stored in the database",
            )

    new_task = Task.model_validate(task)

    session.add(new_task)
    session.commit()
    session.refresh(new_task)
    return new_task


@app.delete("/tasks/{task_id}/")
async def delete_task(session: SessionDep, task_id: int):
    task_to_delete = session.get(Task, task_id)
    if not task_to_delete:
        raise HTTPException(
            status_code=404, detail=f"Task with id {task_id} does not exist"
        )
    session.delete(task_to_delete)
    session.commit()
    return {"deleted": True}


@app.patch("/tasks/{task_id}/completion")
async def update_completed(
    session: SessionDep, task_id: int, task: TaskUpdateCompleted
):
    task_to_update = session.get(Task, task_id)

    if task_to_update is None:
        raise HTTPException(status_code=404, detail="Task not found")

    task_to_update.is_completed = task.is_completed

    now = utc_now()

    task_to_update.completed_at = now if task_to_update.is_completed else None
    task_to_update.updated_at = now

    parent = task_to_update.parent
    if parent is not None:
        parent.is_completed = all(child.is_completed for child in parent.children)
        parent.completed_at = now if parent.is_completed else None
        parent.updated_at = now
        session.add(parent)
    else:
        for child in task_to_update.children:
            child.is_completed = task_to_update.is_completed
            child.completed_at = now if task_to_update.is_completed else None
            child.updated_at = now

    session.add(task_to_update)
    session.commit()
    session.refresh(task_to_update)
    return task_to_update


@app.patch("/tasks/{task_id}/")
async def update_task(session: SessionDep, task_id: int, task: TaskUpdate):
    task_to_update = session.get(Task, task_id)
    if not task_to_update:
        raise HTTPException(status_code=404, detail="Task not found")
    new_data = task.model_dump(exclude_unset=True)
    task_to_update.sqlmodel_update(new_data)
    task_to_update.updated_at = utc_now()

    session.add(task_to_update)
    session.commit()
    session.refresh(task_to_update)
    return task_to_update


# }}}


#
def main():
    create_everything()
    add_tasks()


if __name__ == "__main__":
    main()
