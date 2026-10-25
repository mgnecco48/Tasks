from typing import Annotated, Optional
from sqlmodel import SQLModel, create_engine, Session, select, Field, Relationship
from datetime import datetime, UTC


def utc_now() -> datetime:
    return datetime.now(UTC)


# =========
# = Lists =
# =========
class List(SQLModel):
    __table__ = "lists"  # type: ignore

    id: int = Field(default=None, primary_key=True)
    name: str | None = "New List"
    description: str | None = "No particular description provided"
    created_at: datetime = Field(default_factory=utc_now)
    updated_at: datetime = Field(default_factory=utc_now)


class ListCreate(List):
    name: str | None = None
    description: str | None = None


class ListModify(List):
    name: str | None = None
    description: str | None = None


# =========
# = Tasks =
# =========
class TaskBase(SQLModel):
    body: str
    extra_details: str | None = None
    priority: int = Field(default=3, ge=1, le=3)
    list_id: int = 1


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
    list_id: int = Field(default=1, foreign_key="lists.id")


class TaskPublic(TaskBase):
    id: int
    is_completed: bool
    created_at: datetime
    updated_at: datetime
    completed_at: datetime | None

    parent_id: int | None


class TaskInsert(TaskBase):
    parent_id: int | None = None
    list_id: int = 1


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
