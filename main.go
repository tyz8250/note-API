package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"note-api/config"
	"note-api/internal/model"
	"note-api/internal/repository"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

var noteRepo *repository.NoteRepository

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// GET /notes - 一覧を取得
func getNotes(w http.ResponseWriter, r *http.Request) {
	// メソッドチェック
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	notes, err := noteRepo.GetAllNotes()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

// GET /notes/{id} - IDでメモを取得
func getNotesId(w http.ResponseWriter, r *http.Request) {
	// メソッドチェック
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := r.PathValue("id")

	targetID, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	// メモを取得
	note, err := noteRepo.GetNoteByID(targetID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSONError(w, http.StatusNotFound, "not found")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(note)
}

// POST /notes- 新規メモを作成
func postNotes(w http.ResponseWriter, r *http.Request) {
	// メソッドチェック
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// POSTではサーバが材料を受け取ってからNoteを作成する
	type NoteRequest struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	// リクエストボディをデコードする
	var request NoteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// タイトルと内容が空でないかチェック
	if strings.TrimSpace(request.Title) == "" {
		writeJSONError(w, http.StatusBadRequest, "title is required")
		return
	}

	if strings.TrimSpace(request.Content) == "" {
		writeJSONError(w, http.StatusBadRequest, "content is required")
		return
	}

	now := time.Now()
	note := model.Note{
		Title:     request.Title,
		Content:   request.Content,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := noteRepo.CreateNote(note)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated) // 201 Created
}

// PUT /notes/{id} - IDでメモを更新
func putNotesID(w http.ResponseWriter, r *http.Request) {
	// メソッドチェック
	if r.Method != http.MethodPut {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	// 更新したい内容を受ける構造体
	type NoteRequest struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	// リクエストボディをデコードする
	var request NoteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	// タイトルと内容が空でないかチェック
	if strings.TrimSpace(request.Title) == "" {
		writeJSONError(w, http.StatusBadRequest, "title is required")
		return
	}
	if strings.TrimSpace(request.Content) == "" {
		writeJSONError(w, http.StatusBadRequest, "content is required")
		return
	}

	now := time.Now()

	note := model.Note{
		ID:        id,
		Title:     request.Title,
		Content:   request.Content,
		UpdatedAt: now,
	}

	_, err = noteRepo.UpdateNote(note)
	if err == sql.ErrNoRows {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DELETE /notes/{id} - IDでメモを削除
func deleteNotesID(w http.ResponseWriter, r *http.Request) {
	// メソッドチェック
	if r.Method != http.MethodDelete {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// idを取得する
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	_, err = noteRepo.DeleteNote(model.Note{ID: id})
	if err == sql.ErrNoRows {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

var db *sql.DB

func openDB(cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	return db, nil
}

func ensureSchema(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("create notes table: %w", err)
	}
	return nil
}

// GET /healthz - ヘルスチェック
func healthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func setupRoutes() {
	http.HandleFunc("GET /notes", getNotes)
	http.HandleFunc("GET /notes/{id}", getNotesId)
	http.HandleFunc("POST /notes", postNotes)
	http.HandleFunc("PUT /notes/{id}", putNotesID)
	http.HandleFunc("DELETE /notes/{id}", deleteNotesID)
	http.HandleFunc("GET /healthz", healthz)
}
func runServer(cfg config.Config) error {
	return http.ListenAndServe(":"+cfg.Port, nil)
}

func main() {
	cfg := config.Load()

	var err error
	db, err = openDB(cfg)
	// DB接続に失敗した場合、プログラムを終了する
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := ensureSchema(db); err != nil {
		log.Fatal(err)
	}

	noteRepo = repository.NewNoteRepository(db)

	setupRoutes()

	if err = runServer(cfg); err != nil {
		log.Fatal(err)
	}
}
