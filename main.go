package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
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

// GET /notes - 一覧を取得
func getNotes(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, title, content, created_at, updated_at FROM notes`
	// db.Query(...) で結果をもらう
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	noteList := []Note{} // 最後にJSONで返すための変数

	for rows.Next() {
		var note Note

		err := rows.Scan(
			&note.ID,
			&note.Title,
			&note.Content,
			&note.CreatedAt,
			&note.UpdatedAt,
		)
		// スキャンエラーをチェック
		if err != nil {
			http.Error(w, "failed to scan note", http.StatusInternalServerError)
			return
		}
		noteList = append(noteList, note)
	}

	json.NewEncoder(w).Encode(noteList)
}

// GET /notes/{id} - IDでメモを取得
func getNotesId(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	targetID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	query := `SELECT id, title, content, created_at, updated_at FROM notes WHERE id = ?`
	var note Note
	var createdAt string
	var updatedAt string

	// DBからメモを取得
	err = db.QueryRow(query, targetID).Scan(
		&note.ID,
		&note.Title,
		&note.Content,
		&createdAt,
		&updatedAt,
	)
	// データが見つからなかった場合
	if err == sql.ErrNoRows {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	// その他のエラーの場合
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	note.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		http.Error(w, "failed to parse created_at", http.StatusInternalServerError)
		return
	}

	note.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		http.Error(w, "failed to parse updated_at", http.StatusInternalServerError)
		return
	}

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
		http.Error(w, "invalid request body", http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated) // 201 Created
}

// PUT /notes/{id} - IDでメモを更新
func putNotesID(w http.ResponseWriter, r *http.Request) {

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
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
		http.Error(w, "invalid request body", http.StatusBadRequest)
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 更新された行数を取得
	affected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if affected == 0 {
		http.Error(w, "not found", http.StatusNotFound)
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
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	query := "DELETE FROM notes WHERE id = ?"
	result, err := db.Exec(query, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 削除された行数を取得
	affected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if affected == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("sqlite", "./notes.db")
	query := `CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

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
