from typing import Annotated
from sqlmodel import SQLModel, create_engine, Session, select
from fastapi import Depends, FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from sqlalchemy import event
from pathlib import Path
from . import models
from .models import utc_now

# {{{ SQLite setup
project_root = Path(__file__).resolve().parent.parent
sqlite_path = project_root / "data" / "tasks.db"
sqlite_url = f"sqlite:///{sqlite_path}"
engine = create_engine(sqlite_url, echo=True)


# this enables the foreign key support for all connections for sqlite. Just copied it from the internet though.
@event.listens_for(engine, "connect")
def set_sqlite_pragma(dbapi_connection, _connection_record):
    cursor = dbapi_connection.cursor()
    cursor.execute("PRAGMA foreign_keys=ON")
    cursor.close()


def create_everything():
    SQLModel.metadata.create_all(engine)


def init_database():
    list1 = models.List(name="Main", description="test list")

    task1 = models.Task(
        body="task1",
        extra_details="Some extra details",
    )
    task2 = models.Task(
        body="task2",
        extra_details="Some extra details",
    )
    task3 = models.Task(
        body="task3",
        extra_details="Some extra details",
    )

    data = [list1, task1, task2, task3]

    with Session(engine) as session:
        for record in data:
            session.add(record)

        session.commit()

        for record in data:
            session.refresh(record)

        for record in data:
            print(record)


def get_session():
    with Session(engine) as session:
        yield session


def add_data():
    with Session(engine) as session:
        data = session.exec(select(models.Task)).all()
        if not data:
            init_database()


SessionDep = Annotated[Session, Depends(get_session)]

# {{{ FastAPI part
app = FastAPI()

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


# Makes it so that the database functions run when the servers is started
@app.on_event("startup")
def on_startup():
    create_everything()
    add_data()


@app.get("/")
def get_root():
    return {"Hello": "This is My app"}


# Get the tasks from a specific list
@app.get("/{list_id}/tasks/")
async def get_list_tasks(session: SessionDep, list_id: int):
    list = session.get(models.List, list_id)
    if not list:
        raise HTTPException(
            status_code=404, detail=f"List with id {list_id} does not exist"
        )
    query = session.exec(
        select(models.Task).where(models.Task.list_id == list_id)
    ).all()
    if not query:
        return {"Uh oh": "There are not tasks in this list"}
    return query


# Get all the available lists
@app.get("/lists/")
async def get_all_lists(session: SessionDep):
    query = session.exec(select(models.List)).all()
    return query


# Create a new list
@app.post("/lists/", response_model=models.List)
async def create_list(session: SessionDep, list: models.ListCreate):
    new_list = models.List.model_validate(list)
    session.add(new_list)
    session.commit()
    session.refresh(new_list)
    return list


# Update the name of an existing list
@app.patch("/lists/{list_id}/", response_model=models.List)
async def update_list(session: SessionDep, list_id: int, list: models.ListModify):
    list_to_update = session.get(models.List, list_id)
    if not list_to_update:
        raise HTTPException(status_code=404, detail="List does not exist")

    new_data = list.model_dump(exclude_unset=True)
    list_to_update.sqlmodel_update(new_data)
    list_to_update.updated_at = utc_now()

    session.add(list_to_update)
    session.commit()
    session.refresh(list_to_update)
    return list_to_update


# Delete an existing list
# TODO: See what happens when a list is deleted with active tasks
@app.delete("/list/{list_id}/")
async def delete_list(session: SessionDep, list_id: int):
    list_to_delete = session.get(models.List, list_id)
    if not list_to_delete:
        raise HTTPException(
            status_code=404, detail=f"List with id {list_id} does not exist."
        )
    list_name = list_to_delete.name
    session.delete(list_to_delete)
    session.commit
    return {"deleted": f"ID: {list_id}, Name: {list_name}"}


# Get all the existing tasks
@app.get("/tasks/", response_model=list[models.TaskPublic])
async def get_all_tasks(session: SessionDep):
    query = session.exec(select(models.Task)).all()
    return query


# Cerate a new task
@app.post("/tasks/", response_model=models.TaskPublic)
async def create_task(session: SessionDep, task: models.TaskCreate):
    # Checking if a parent id was passed, and if that parent actually exists
    if task.parent_id is not None:
        parent = session.get(models.Task, task.parent_id)
        if parent is None:
            raise HTTPException(
                status_code=404,
                detail=f"Parent task with id {task.parent_id} is not stored in the database",
            )

    new_task = models.Task.model_validate(task)

    session.add(new_task)
    session.commit()
    session.refresh(new_task)
    return new_task


# Handle the rendering by grouping the children and parent tasks. It also takes care of the sorting, by created time, etc
@app.get("/tasks/tree/", response_model=list[models.TaskTreeNode])
async def get_task_tree(session: SessionDep):
    flat_tasks = session.exec(select(models.Task)).all()

    nodes = {
        task.id: models.TaskTreeNode(
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

    def recursive_sort(tasks: list["models.TaskTreeNode"]):
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


# Delete an existing task
@app.delete("/tasks/{task_id}/")
async def delete_task(session: SessionDep, task_id: int):
    task_to_delete = session.get(models.Task, task_id)
    if not task_to_delete:
        raise HTTPException(
            status_code=404, detail=f"Task with id {task_id} does not exist"
        )
    task_body = task_to_delete.body
    session.delete(task_to_delete)
    session.commit()
    return {"deleted": f"Id: {task_id}, Body: {task_body}"}


# Toggle completion status of a task. Handle completion for parent tasks when all
# children tasks are completed.
@app.patch("/tasks/{task_id}/completion", response_model=models.TaskPublic)
async def update_completed(
    session: SessionDep, task_id: int, task: models.TaskUpdateCompleted
):
    task_to_update = session.get(models.Task, task_id)

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


# Change an existing task
@app.patch("/tasks/{task_id}/", response_model=models.TaskPublic)
async def update_task(session: SessionDep, task_id: int, task: models.TaskUpdate):
    task_to_update = session.get(models.Task, task_id)
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


def main():
    create_everything()


if __name__ == "__main__":
    main()
