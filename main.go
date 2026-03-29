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

// メモリ上に保存するための配列
var notes []Note

// ダミーデータを2つ用意する
func init() {
	notes = append(notes, Note{
		ID:        1,
		Title:     "初めてのメモ",
		Content:   "初めてのメモです",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	notes = append(notes, Note{
		ID:        2,
		Title:     "2回目のメモ",
		Content:   "2回目のメモです",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
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

	for _, note := range notes {
		if note.ID == targetID {
			json.NewEncoder(w).Encode(note)
			return
		}
	}
	// 該当するIDが見つからない場合は404を返す
	http.Error(w, "not found", http.StatusNotFound)
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
	// 新しいメモを作成する
	newNote := Note{
		ID:        len(notes) + 1,
		Title:     request.Title,
		Content:   request.Content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	notes = append(notes, newNote)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newNote)
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

	// notesの中からidが一致するものを探し、更新する
	for i, note := range notes {
		if note.ID == id {
			notes[i].Title = request.Title
			notes[i].Content = request.Content
			notes[i].UpdatedAt = time.Now()
			json.NewEncoder(w).Encode(notes[i])
			return
		}
	}

	// 該当するidがない場合は404を返す
	http.Error(w, "not found", http.StatusNotFound)
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

	// 一致するnoteを探し、見つかったら削除する
	for i, note := range notes {
		if note.ID == id {
			// 削除する
			notes = append(notes[:i], notes[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	// 該当するidがない場合は404を返す
	http.Error(w, "not found", http.StatusNotFound)
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
