package repository

import (
	"database/sql"
	"note-api/internal/model"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// setupTestRepository はテスト用のデータベースとリポジトリをセットアップする
func setupTestRepository(t *testing.T) *NoteRepository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(query)
	if err != nil {
		t.Fatalf("failed to create notes table: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return NewNoteRepository(db)
}

func TestGetNoteByID_ExistingNote_ReturnsNote(t *testing.T) {
	repo := setupTestRepository(t)

	now := time.Now().Format(time.RFC3339)

	result, err := repo.db.Exec(
		`INSERT INTO notes (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		"テストタイトル",
		"テスト本文",
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to insert test note: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get last insert id: %v", err)
	}

	note, err := repo.GetNoteByID(int(id))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if note.Title != "テストタイトル" {
		t.Fatalf("expected title %q, got %q", "テストタイトル", note.Title)
	}

	if note.Content != "テスト本文" {
		t.Fatalf("expected content %q, got %q", "テスト本文", note.Content)
	}
}

// TestGetNoteByID_NotFound_ReturnsErrNoRows は存在しないメモを取得した場合にErrNoRowsを返すことを確認する
func TestGetNoteByID_NotFound_ReturnsErrNoRows(t *testing.T) {
	repo := setupTestRepository(t)
	_, err := repo.GetNoteByID(99999)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected ErrNoRows, got %v", err)
	}
}

// TestGetAllNotes_WithOneNote_ReturnsNotes は1件のメモがある場合にメモを返すことを確認する
func TestGetAllNotes_WithOneNote_ReturnsNotes(t *testing.T) {
	repo := setupTestRepository(t)

	now := time.Now().Format(time.RFC3339)

	_, err := repo.db.Exec(
		`INSERT INTO notes (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		"一覧タイトル",
		"一覧本文",
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to insert test note: %v", err)
	}

	notes, err := repo.GetAllNotes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(notes) != 1 {
		t.Fatalf("expected notes length %d, got %d", 1, len(notes))
	}

	if notes[0].Title != "一覧タイトル" {
		t.Fatalf("expected title %q, got %q", "一覧タイトル", notes[0].Title)
	}

	if notes[0].Content != "一覧本文" {
		t.Fatalf("expected content %q, got %q", "一覧本文", notes[0].Content)
	}
}

// TestGetAllNotes_EmptyDB_ReturnsEmptySlice はデータベースにメモがない場合に空スライスを返すことを確認する
func TestGetAllNotes_EmptyDB_ReturnsEmptySlice(t *testing.T) {
	repo := setupTestRepository(t)

	notes, err := repo.GetAllNotes()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(notes) != 0 {
		t.Fatalf("expected notes length %d, got %d", 0, len(notes))
	}
}

// TestCreateNote_ValidNote_SavesNote は有効なメモを保存できることを確認する
func TestCreateNote_ValidNote_SavesNote(t *testing.T) {
	repo := setupTestRepository(t)

	now := time.Now()
	note := model.Note{
		Title:     "作成タイトル",
		Content:   "作成本文",
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := repo.CreateNote(note)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var count int
	err = repo.db.QueryRow(`SELECT COUNT(*) FROM notes`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count notes: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected note count %d, got %d", 1, count)
	}

	var savedTitle string
	var savedContent string
	err = repo.db.QueryRow(`SELECT title, content FROM notes LIMIT 1`).Scan(&savedTitle, &savedContent)
	if err != nil {
		t.Fatalf("failed to fetch saved note: %v", err)
	}

	if savedTitle != "作成タイトル" {
		t.Fatalf("expected title %q, got %q", "作成タイトル", savedTitle)
	}

	if savedContent != "作成本文" {
		t.Fatalf("expected content %q, got %q", "作成本文", savedContent)
	}
}

// TestUpdateNote_ExistingNote_UpdatesNote は既存のメモを更新できることを確認する
func TestUpdateNote_ExistingNote_UpdatesNote(t *testing.T) {
	repo := setupTestRepository(t)

	now := time.Now().Format(time.RFC3339)

	result, err := repo.db.Exec(
		`INSERT INTO notes (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		"更新前タイトル",
		"更新前本文",
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to insert test note: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get last insert id: %v", err)
	}

	updatedNote := model.Note{
		ID:        int(id),
		Title:     "更新後タイトル",
		Content:   "更新後本文",
		UpdatedAt: time.Now(),
	}

	_, err = repo.UpdateNote(updatedNote)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var title string
	var content string
	err = repo.db.QueryRow(
		`SELECT title, content FROM notes WHERE id = ?`,
		int(id),
	).Scan(&title, &content)
	if err != nil {
		t.Fatalf("failed to fetch updated note: %v", err)
	}

	if title != "更新後タイトル" {
		t.Fatalf("expected title %q, got %q", "更新後タイトル", title)
	}

	if content != "更新後本文" {
		t.Fatalf("expected content %q, got %q", "更新後本文", content)
	}
}

// TestUpdateNote_NotFound_ReturnsErrNoRows は存在しないメモを更新した場合にErrNoRowsを返すことを確認する
func TestUpdateNote_NotFound_ReturnsErrNoRows(t *testing.T) {
	repo := setupTestRepository(t)

	note := model.Note{
		ID:        999,
		Title:     "更新後タイトル",
		Content:   "更新後本文",
		UpdatedAt: time.Now(),
	}

	_, err := repo.UpdateNote(note)
	if err != sql.ErrNoRows {
		t.Fatalf("expected error %v, got %v", sql.ErrNoRows, err)
	}
}

// TestDeleteNote_ExistingNote_DeletesNote は既存のメモを削除できることを確認する
func TestDeleteNote_ExistingNote_DeletesNote(t *testing.T) {
	repo := setupTestRepository(t)

	now := time.Now().Format(time.RFC3339)

	result, err := repo.db.Exec(
		`INSERT INTO notes (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		"削除前タイトル",
		"削除前本文",
		now,
		now,
	)
	if err != nil {
		t.Fatalf("failed to insert test note: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to get last insert id: %v", err)
	}

	_, err = repo.DeleteNote(model.Note{ID: int(id)})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var count int
	err = repo.db.QueryRow(
		`SELECT COUNT(*) FROM notes WHERE id = ?`,
		int(id),
	).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count notes: %v", err)
	}

	if count != 0 {
		t.Fatalf("expected note count %d, got %d", 0, count)
	}
}

// TestDeleteNote_NotFound_ReturnsErrNoRows は存在しないメモを削除した場合にErrNoRowsを返すことを確認する
func TestDeleteNote_NotFound_ReturnsErrNoRows(t *testing.T) {
	repo := setupTestRepository(t)

	_, err := repo.DeleteNote(model.Note{ID: 999})
	if err != sql.ErrNoRows {
		t.Fatalf("expected error %v, got %v", sql.ErrNoRows, err)
	}
}
