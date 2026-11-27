import os
import sqlite3
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

DB_PATH = os.getenv("DB_PATH", "/data/users.db")
os.makedirs(os.path.dirname(DB_PATH), exist_ok=True)

app = FastAPI(title="Users API")


def get_conn():
    conn = sqlite3.connect(DB_PATH)
    conn.row_factory = sqlite3.Row
    return conn


@app.on_event("startup")
def init_db():
    with get_conn() as conn:
        conn.execute(
            """
            CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                email TEXT NOT NULL UNIQUE,
                password TEXT NOT NULL
            )
            """
        )


class UserIn(BaseModel):
    name: str
    email: str
    password: str


@app.post("/users", status_code=201)
def create_user(user: UserIn):
    with get_conn() as conn:
        try:
            cur = conn.execute(
                "INSERT INTO users (name, email, password) VALUES (?, ?, ?)",
                (user.name, user.email, user.password),
            )
            conn.commit()
            row = conn.execute(
                "SELECT id, name, email FROM users WHERE id = ?", (cur.lastrowid,)
            ).fetchone()
            return dict(row)
        except sqlite3.IntegrityError:
            raise HTTPException(status_code=400, detail="Email ya registrado")


@app.get("/users")
def list_users():
    with get_conn() as conn:
        rows = conn.execute("SELECT id, name, email FROM users").fetchall()
        return [dict(r) for r in rows]


@app.get("/users/{user_id}")
def get_user(user_id: int):
    with get_conn() as conn:
        row = conn.execute(
            "SELECT id, name, email FROM users WHERE id = ?", (user_id,)
        ).fetchone()
        if not row:
            raise HTTPException(status_code=404, detail="Usuario no encontrado")
        return dict(row)


@app.put("/users/{user_id}")
def update_user(user_id: int, user: UserIn):
    with get_conn() as conn:
        cur = conn.execute(
            "UPDATE users SET name = ?, email = ?, password = ? WHERE id = ?",
            (user.name, user.email, user.password, user_id),
        )
        conn.commit()
        if cur.rowcount == 0:
            raise HTTPException(status_code=404, detail="Usuario no encontrado")
        row = conn.execute(
            "SELECT id, name, email FROM users WHERE id = ?", (user_id,)
        ).fetchone()
        return dict(row)


@app.delete("/users/{user_id}")
def delete_user(user_id: int):
    with get_conn() as conn:
        cur = conn.execute("DELETE FROM users WHERE id = ?", (user_id,))
        conn.commit()
        if cur.rowcount == 0:
            raise HTTPException(status_code=404, detail="Usuario no encontrado")
        return {"deleted": True, "id": user_id}