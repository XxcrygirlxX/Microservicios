const fs = require("fs");
const path = require("path");
const express = require("express");
const Database = require("better-sqlite3");

const dbPath = process.env.DB_PATH || "/data/products.db";
fs.mkdirSync(path.dirname(dbPath), { recursive: true });

const db = new Database(dbPath);
db.exec(`
  CREATE TABLE IF NOT EXISTS products (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    price REAL NOT NULL,
    description TEXT,
    stock INTEGER NOT NULL DEFAULT 0
  )
`);

const app = express();
app.use(express.json());

app.post("/products", (req, res) => {
  const { name, price, description = "", stock = 0 } = req.body;
  if (!name || price === undefined) {
    return res.status(400).json({ error: "name y price son obligatorios" });
  }
  const info = db
    .prepare("INSERT INTO products (name, price, description, stock) VALUES (?, ?, ?, ?)")
    .run(name, price, description, stock);
  const product = db.prepare("SELECT * FROM products WHERE id = ?").get(info.lastInsertRowid);
  res.status(201).json(product);
});

app.get("/products", (req, res) => {
  res.json(db.prepare("SELECT * FROM products").all());
});

app.get("/products/:id", (req, res) => {
  const product = db.prepare("SELECT * FROM products WHERE id = ?").get(req.params.id);
  if (!product) return res.status(404).json({ error: "Producto no encontrado" });
  res.json(product);
});

app.put("/products/:id", (req, res) => {
  const { name, price, description = "", stock = 0 } = req.body;
  const info = db
    .prepare("UPDATE products SET name = ?, price = ?, description = ?, stock = ? WHERE id = ?")
    .run(name, price, description, stock, req.params.id);
  if (info.changes === 0) return res.status(404).json({ error: "Producto no encontrado" });
  res.json(db.prepare("SELECT * FROM products WHERE id = ?").get(req.params.id));
});

app.delete("/products/:id", (req, res) => {
  const info = db.prepare("DELETE FROM products WHERE id = ?").run(req.params.id);
  if (info.changes === 0) return res.status(404).json({ error: "Producto no encontrado" });
  res.json({ deleted: true, id: Number(req.params.id) });
});

app.listen(8003, "0.0.0.0", () => {
  console.log("Products API running on port 8003");
});