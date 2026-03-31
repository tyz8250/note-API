package repository

import (
	"database/sql"
	"time"
)

type Note struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NoteRepository struct {
	db *sql.DB
}

// 新しいノートリポジトリを作成
func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// データベースからすべてのメモを取得
func (r *NoteRepository) GetAllNotes() ([]Note, error) {
	query := `SELECT id, title, content, created_at, updated_at FROM notes`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []Note{}

	for rows.Next() {
		var note Note
		var createdAtStr string
		var updatedAtStr string

		err := rows.Scan(
			&note.ID,
			&note.Title,
			&note.Content,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, err
		}
		note.CreatedAt, err = time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			return nil, err
		}

		note.UpdatedAt, err = time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			return nil, err
		}

		notes = append(notes, note)
	}

	return notes, nil
}

// データベースから特定のメモを取得
func (r *NoteRepository) GetNoteByID(id int) (Note, error) {
	query := `SELECT id, title, content, created_at, updated_at FROM notes WHERE id = ?`

	// 受け皿を作る。時刻は文字列のため、一旦文字列で受ける
	var note Note
	var createdAt string
	var updatedAt string

	// DBから1件のメモを取得
	err := r.db.QueryRow(query, id).Scan(
		&note.ID,
		&note.Title,
		&note.Content,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return Note{}, err
	}

	note.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return Note{}, err
	}

	note.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return Note{}, err
	}

	return note, nil
}
