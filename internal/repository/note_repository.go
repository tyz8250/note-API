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
