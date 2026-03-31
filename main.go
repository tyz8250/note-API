package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"note-api/internal/repository"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type Note struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

var noteRepo *repository.NoteRepository

func writeJSONError(w http.ResponseWriter, status int, message string) {
	// JSONエラーを返す方式
	w.Header().Set("Content-Type", "application/json")
	// 400や404などのステータスコードを返す
	w.WriteHeader(status)
	// エラーをJSONで返す
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}

// GET /notes - 一覧を取得
func getNotes(w http.ResponseWriter, r *http.Request) {

	notes, err := noteRepo.GetAllNotes()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

// GET /notes/{id} - IDでメモを取得
func getNotesId(w http.ResponseWriter, r *http.Request) {
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

	// 時刻を決める
	now := time.Now().Format(time.RFC3339)
	query := `
	INSERT INTO notes (title, content, created_at, updated_at)
	VALUES (?, ?, ?, ?)
	`
	_, err := db.Exec(query, request.Title, request.Content, now, now)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated) // 201 Created
}

// PUT /notes/{id} - IDでメモを更新
func putNotesID(w http.ResponseWriter, r *http.Request) {

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

	now := time.Now().Format(time.RFC3339)
	query := `
	UPDATE notes
	SET title = ?, content = ?, updated_at = ?
	WHERE id = ?
	`

	// 更新命令
	result, err := db.Exec(query, request.Title, request.Content, now, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 更新された行数を取得
	affected, err := result.RowsAffected()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if affected == 0 {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DELETE /notes/{id} - IDでメモを削除
func deleteNotesID(w http.ResponseWriter, r *http.Request) {
	// idを取得する
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}
	query := "DELETE FROM notes WHERE id = ?"
	result, err := db.Exec(query, id)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 削除された行数を取得
	affected, err := result.RowsAffected()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if affected == 0 {
		writeJSONError(w, http.StatusNotFound, "not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

var db *sql.DB

func main() {

	var err error
	db, err = sql.Open("sqlite", "./notes.db")
	// DB接続に失敗した場合、プログラムを終了する
	if err != nil {
		panic(err)
	}
	query := `CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	noteRepo = repository.NewNoteRepository(db)
	_, err = db.Exec(query) // テーブルを作成する
	if err != nil {
		panic(err)
	}
	// テーブルを作成に失敗した場合、プログラムを終了する

	defer db.Close()

	http.HandleFunc("GET /notes", getNotes)
	http.HandleFunc("GET /notes/{id}", getNotesId)
	http.HandleFunc("POST /notes", postNotes)
	http.HandleFunc("PUT /notes/{id}", putNotesID)
	http.HandleFunc("DELETE /notes/{id}", deleteNotesID)
	http.ListenAndServe(":8080", nil)
}
