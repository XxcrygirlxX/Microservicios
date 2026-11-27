package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Order struct {
	ID        int     `json:"id"`
	Customer  string  `json:"customer"`
	Total     float64 `json:"total"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

var db *sql.DB

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func initDB() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "/data/orders.db"
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return err
	}
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			customer TEXT NOT NULL,
			total REAL NOT NULL,
			status TEXT NOT NULL,
			created_at TEXT NOT NULL
		)
	`)
	return err
}

func ordersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, err := db.Query("SELECT id, customer, total, status, created_at FROM orders")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		defer rows.Close()

		orders := []Order{}
		for rows.Next() {
			var o Order
			if err := rows.Scan(&o.ID, &o.Customer, &o.Total, &o.Status, &o.CreatedAt); err == nil {
				orders = append(orders, o)
			}
		}
		writeJSON(w, http.StatusOK, orders)

	case http.MethodPost:
		var req struct {
			Customer string  `json:"customer"`
			Total    float64 `json:"total"`
			Status   string  `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "JSON invalido"})
			return
		}
		if req.Status == "" {
			req.Status = "PENDING"
		}
		now := time.Now().UTC().Format(time.RFC3339)
		res, err := db.Exec(
			"INSERT INTO orders (customer, total, status, created_at) VALUES (?, ?, ?, ?)",
			req.Customer, req.Total, req.Status, now,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		id, _ := res.LastInsertId()
		writeJSON(w, http.StatusCreated, Order{
			ID:        int(id),
			Customer:  req.Customer,
			Total:     req.Total,
			Status:    req.Status,
			CreatedAt: now,
		})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metodo no permitido"})
	}
}

func orderByIDHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/orders/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ID invalido"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		var o Order
		err := db.QueryRow(
			"SELECT id, customer, total, status, created_at FROM orders WHERE id = ?", id,
		).Scan(&o.ID, &o.Customer, &o.Total, &o.Status, &o.CreatedAt)
		if err == sql.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Pedido no encontrado"})
			return
		}
		writeJSON(w, http.StatusOK, o)

	case http.MethodPatch, http.MethodPut:
		var req struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Estado requerido"})
			return
		}
		res, err := db.Exec("UPDATE orders SET status = ? WHERE id = ?", req.Status, id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Pedido no encontrado"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"id": id, "status": req.Status})

	case http.MethodDelete:
		res, err := db.Exec("DELETE FROM orders WHERE id = ?", id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		rows, _ := res.RowsAffected()
		if rows == 0 {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Pedido no encontrado"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})

	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "Metodo no permitido"})
	}
}

func main() {
	if err := initDB(); err != nil {
		panic(err)
	}
	defer db.Close()

	http.HandleFunc("/orders", ordersHandler)
	http.HandleFunc("/orders/", orderByIDHandler)
	_ = http.ListenAndServe(":8002", nil)
}
