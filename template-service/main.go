package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Template struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // email, push
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateTemplateRequest struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

var db *sql.DB

func main() {
	var err error
	dbHost := getEnv("DB_HOST", "template-db")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "templatedb")

	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbUser, dbPass, dbName)

	// Retry connection
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", connStr)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}
		log.Printf("Waiting for database... (%d/10)", i+1)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	initSchema()

	r := mux.NewRouter()
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/templates", createTemplateHandler).Methods("POST")
	r.HandleFunc("/templates/{id}", getTemplateHandler).Methods("GET")
	r.HandleFunc("/templates", listTemplatesHandler).Methods("GET")

	port := getEnv("PORT", "8083")
	log.Printf("Template service running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func initSchema() {
	schema := `
	CREATE TABLE IF NOT EXISTS templates (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		type VARCHAR(50) NOT NULL,
		subject TEXT,
		body TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Insert default templates
	INSERT INTO templates (name, type, subject, body) VALUES 
	('welcome_email', 'email', 'Welcome to Our Platform!', 'Hello {{name}}, welcome aboard!'),
	('password_reset', 'email', 'Reset Your Password', 'Click here to reset: {{link}}'),
	('order_confirmation', 'push', 'Order Confirmed', 'Your order #{{order_id}} has been confirmed!')
	ON CONFLICT (name) DO NOTHING;
	`
	_, err := db.Exec(schema)
	if err != nil {
		log.Fatal("Failed to create schema:", err)
	}
	log.Println("Database schema initialized with default templates")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	err := db.Ping()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{"status": "unhealthy"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy", "service": "template-service"})
}

func createTemplateHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var templateID int
	err := db.QueryRow(
		"INSERT INTO templates (name, type, subject, body) VALUES ($1, $2, $3, $4) RETURNING id",
		req.Name, req.Type, req.Subject, req.Body,
	).Scan(&templateID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	template := Template{
		ID:        templateID,
		Name:      req.Name,
		Type:      req.Type,
		Subject:   req.Subject,
		Body:      req.Body,
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(template)
}

func getTemplateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var template Template
	err := db.QueryRow(
		"SELECT id, name, type, subject, body, created_at FROM templates WHERE id = $1",
		id,
	).Scan(&template.ID, &template.Name, &template.Type, &template.Subject, &template.Body, &template.CreatedAt)

	if err == sql.ErrNoRows {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(template)
}

func listTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	templateType := r.URL.Query().Get("type")

	query := "SELECT id, name, type, subject, body, created_at FROM templates"
	args := []interface{}{}

	if templateType != "" {
		query += " WHERE type = $1"
		args = append(args, templateType)
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	templates := []Template{}
	for rows.Next() {
		var t Template
		if err := rows.Scan(&t.ID, &t.Name, &t.Type, &t.Subject, &t.Body, &t.CreatedAt); err != nil {
			continue
		}
		templates = append(templates, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}