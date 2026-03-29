package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// TestPostNotes_EmptyTitle_ReturnsBadRequest は、タイトルが空のときに400を返すことを確認するテスト
func TestPostNotes_EmptyTitle_ReturnsBadRequest(t *testing.T) {
	setupTestDB(t)

	body := `{"title":"","content":"本文あり"}`
	// テスト用のリクエスト作成(POST /notes にJSONを送信)
	req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(body))
	// Content-Typeを設定
	req.Header.Set("Content-Type", "application/json")
	// レスポンスを受け取る変数
	w := httptest.NewRecorder()
	// ハンドラを呼び出し
	postNotes(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestPostNotes_EmptyContent_ReturnsBadRequest は、コンテンツが空のときに400を返すことを確認するテスト
func TestPostNotes_EmptyContent_ReturnsBadRequest(t *testing.T) {
	setupTestDB(t)

	body := `{"title":"タイトル","content":""}`
	// テスト用のリクエスト作成(POST /notes にJSONを送信)
	req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(body))
	// Content-Typeを設定
	req.Header.Set("Content-Type", "application/json")
	// レスポンスを受け取る変数
	w := httptest.NewRecorder()
	// ハンドラを呼び出し
	postNotes(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestPostNotes_ValidRequest_ReturnsCreated は、有効なリクエストに対して201を返すことを確認するテスト
func TestPostNotes_ValidRequest_ReturnsCreated(t *testing.T) {
	setupTestDB(t)

	body := `{"title":"正常タイトル","content":"正常な本文"}`

	req := httptest.NewRequest(http.MethodPost, "/notes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	postNotes(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var count int

	err := db.QueryRow("SELECT COUNT(*) FROM notes").Scan(&count)
	// エラー処理
	if err != nil {
		t.Fatalf("failed to count notes: %v", err)
	}
	// データベースに1件登録されていることを確認
	if count != 1 {
		t.Fatalf("expected 1 note, got %d", count)
	}

	var savedTitle string
	var savedContent string

	err = db.QueryRow("SELECT title, content FROM notes").Scan(&savedTitle, &savedContent)
	if err != nil {
		t.Fatalf("failed to query saved note: %v", err)
	}

	if savedTitle != "正常タイトル" {
		t.Fatalf("expected title '正常タイトル', got '%s'", savedTitle)
	}

	if savedContent != "正常な本文" {
		t.Fatalf("expected content '正常な本文', got '%s'", savedContent)
	}

}

func setupTestDB(t *testing.T) {
	t.Helper()

	var err error
	db, err = sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	query := `CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	);
	`

	_, err = db.Exec(query)
	if err != nil {
		t.Fatalf("failed to create test table: %v", err)
	}
	// テスト終了時にデータベースを閉じる
	t.Cleanup(func() {
		db.Close()
	})
}

func TestGetNotesId_NotFound_ReturnsNotFound(t *testing.T) {
	setupTestDB(t)

	// ID=999のノートが存在しないことを確認
	req := httptest.NewRequest(http.MethodGet, "/notes/999", nil)
	req.SetPathValue("id", "999")
	w := httptest.NewRecorder()

	getNotesId(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}

}

func TestGetNotesId_ValidID_ReturnsOK(t *testing.T) {
	setupTestDB(t)

	_, err := db.Exec(
		`INSERT INTO notes (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		"テストタイトル",
		"テスト本文",
		"2026-03-29T10:00:00Z",
		"2026-03-29T10:00:00Z",
	)
	if err != nil {
		t.Fatalf("failed to insert test note: %v", err)
	}

	// 存在するIDでテスト
	req := httptest.NewRequest(http.MethodGet, "/notes/1", nil)
	req.SetPathValue("id", "1")
	w := httptest.NewRecorder()

	getNotesId(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "テストタイトル") {
		t.Fatalf("expected response body to contain title, got %s", body)
	}

	if !strings.Contains(body, "テスト本文") {
		t.Fatalf("expected response body to contain content, got %s", body)
	}
}

func TestGetNotes_EmptyDB_ReturnsEmptyArray(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
	w := httptest.NewRecorder()

	getNotes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := strings.TrimSpace(w.Body.String())
	if body != "[]" {
		t.Fatalf("expected response body %q, got %q", "[]", body)
	}
}

func TestGetNotes_WithOneNote_ReturnsNotes(t *testing.T) {
	setupTestDB(t)

	_, err := db.Exec(
		`INSERT INTO notes (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		"一覧テストタイトル",
		"一覧テスト本文",
		"2026-03-29T10:00:00Z",
		"2026-03-29T10:00:00Z",
	)
	if err != nil {
		t.Fatalf("failed to insert test note: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
	w := httptest.NewRecorder()

	getNotes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "一覧テストタイトル") {
		t.Fatalf("expected response body to contain title, got %s", body)
	}

	if !strings.Contains(body, "一覧テスト本文") {
		t.Fatalf("expected response body to contain content, got %s", body)
	}
}

func TestPutNotesID_ValidRequest_ReturnsOKAndUpdatesNote(t *testing.T) {
	setupTestDB(t)

	_, err := db.Exec(
		`INSERT INTO notes (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		"更新前タイトル",
		"更新前本文",
		"2026-03-29T10:00:00Z",
		"2026-03-29T10:00:00Z",
	)
	if err != nil {
		t.Fatalf("failed to insert test note: %v", err)
	}

	body := `{"title":"更新後タイトル","content":"更新後本文"}`
	req := httptest.NewRequest(http.MethodPut, "/notes/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	putNotesID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var title string
	var content string

	err = db.QueryRow(`SELECT title, content FROM notes WHERE id = ?`, 1).Scan(&title, &content)
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

func TestPutNotesID_NotFound_ReturnsNotFound(t *testing.T) {
	setupTestDB(t)

	body := `{"title":"更新タイトル","content":"更新本文"}`
	req := httptest.NewRequest(http.MethodPut, "/notes/999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "999")

	w := httptest.NewRecorder()

	putNotesID(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
func TestDeleteNotesID_ValidID_ReturnsNoContentAndDeletesNote(t *testing.T) {
	setupTestDB(t)

	_, err := db.Exec(
		`INSERT INTO notes (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		"削除前タイトル",
		"削除前本文",
		"2026-03-29T10:00:00Z",
		"2026-03-29T10:00:00Z",
	)
	if err != nil {
		t.Fatalf("failed to insert test note: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/notes/1", nil)
	req.SetPathValue("id", "1")

	w := httptest.NewRecorder()

	deleteNotesID(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM notes WHERE id = ?`, 1).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count notes: %v", err)
	}

	if count != 0 {
		t.Fatalf("expected note count %d, got %d", 0, count)
	}
}

func TestDeleteNotesID_NotFound_ReturnsNotFound(t *testing.T) {
	setupTestDB(t)

	req := httptest.NewRequest(http.MethodDelete, "/notes/999", nil)
	req.SetPathValue("id", "999")

	w := httptest.NewRecorder()

	deleteNotesID(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}
