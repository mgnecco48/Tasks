from typing import Annotated
from sqlmodel import SQLModel, create_engine, Session, select, Field, Relationship
from datetime import datetime, UTC
from fastapi import Depends, FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware


# {{{ Models
def utc_now() -> datetime:
    return datetime.now(UTC)


# TODO: modify the schema, to have tasks and children tasks, that way, i can create the relationships that will update the status of the parent when completing all the children tasks. Do not create another table, but rather have the same model reference itself.
class TaskBase(SQLModel):
    body: str
    extra_details: str | None = None


class Task(TaskBase, table=True):
    __tablename__ = "tasks"  # type: ignore

    id: int | None = Field(default=None, primary_key=True)
    is_completed: bool = False
    created_at: datetime = Field(default_factory=utc_now)
    updated_at: datetime = Field(default_factory=utc_now)
    completed_at: datetime | None = None
    parent_id: int | None = Field(default=None, foreign_key="tasks.id")

    parent: Task = Relationship(back_populates="children")
    children: list["Task"] | None = Relationship(back_populates="parent")


class TaskInsert(TaskBase):
    parent_id: int | None = None


class TaskUpdateCompleted(SQLModel):
    is_completed: bool


class TaskUpdateBody(SQLModel):
    body: str
    extra_details: str | None = None


# }}}

# {{{ SQLite setup # TODO: Need to create the databases with something like Alembic, so that when i add a column to the model it automatically updates the databse.
sqlite_filename = "tasks.db"
sqlite_url = f"sqlite:///{sqlite_filename}"
engine = create_engine(sqlite_url, echo=True)


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
async def get_taks(session: SessionDep):
    query = session.exec(select(Task)).all()
    return query


@app.post("/tasks/")
async def add_task(session: SessionDep, task: TaskInsert):
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


# TODO: make it so that when a root task is marked as complete, all the child tasks aswell. Also, when the main task is marked as incomplete, then the children ones get marked as incompleted.
@app.patch("/tasks/{task_id}/completion")
async def update_completed(
    session: SessionDep, task_id: int, task: TaskUpdateCompleted
):
    task_to_update = session.get(Task, task_id)

    if task_to_update is None:
        raise HTTPException(status_code=404, detail="Task not found§")

    task_to_update.is_completed = task.is_completed

    now = utc_now()
    task_to_update.completed_at = now if task_to_update.is_completed else None
    task_to_update.updated_at = now

    session.add(task_to_update)
    session.commit()
    session.refresh(task_to_update)
    return task_to_update


@app.patch("/tasks/{task_id}/")
async def update_task(session: SessionDep, task_id: int, task: TaskUpdateBody):
    task_to_update = session.get(Task, task_id)
    if not task_to_update:
        raise HTTPException(status_code=404, detail="Task not found")
    new_data = task.model_dump(exclude_unset=True)
    task_to_update.sqlmodel_update(new_data)

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
