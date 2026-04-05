package repository

import (
	"database/sql"
	"note-api/internal/model"
	"time"
)

type NoteRepository struct {
	db *sql.DB
}

// 新しいノートリポジトリを作成
func NewNoteRepository(db *sql.DB) *NoteRepository {
	return &NoteRepository{db: db}
}

// データベースからすべてのメモを取得
func (r *NoteRepository) GetAllNotes() ([]model.Note, error) {
	query := `SELECT id, title, content, created_at, updated_at FROM notes`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []model.Note{}

	for rows.Next() {
		var note model.Note
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
func (r *NoteRepository) GetNoteByID(id int) (model.Note, error) {
	query := `SELECT id, title, content, created_at, updated_at FROM notes WHERE id = ?`

	// 受け皿を作る。時刻は文字列のため、一旦文字列で受ける
	var note model.Note
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
		return model.Note{}, err
	}

	note.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return model.Note{}, err
	}

	note.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return model.Note{}, err
	}

	return note, nil
}

func (r *NoteRepository) CreateNote(note model.Note) (model.Note, error) {
	query := `
	INSERT INTO notes (title, content, created_at, updated_at)
	VALUES (?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, note.Title, note.Content, note.CreatedAt.Format(time.RFC3339), note.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return model.Note{}, err
	}
	return note, nil
}
func (r *NoteRepository) UpdateNote(note model.Note) (model.Note, error) {
	query := `
	UPDATE notes
	SET title = ?, content = ?, updated_at = ?
	WHERE id = ?
	`
	// 実行
	result, err := r.db.Exec(
		query,
		note.Title,
		note.Content,
		note.UpdatedAt.Format(time.RFC3339),
		note.ID,
	)
	if err != nil {
		return model.Note{}, err
	}

	// 影響を受けた行数を取得
	affected, err := result.RowsAffected()
	if err != nil {
		return model.Note{}, err
	}

	// 影響を受けた行数が0なら、該当するメモが存在しない
	if affected == 0 {
		return model.Note{}, sql.ErrNoRows
	}

	return note, nil
}

func (r *NoteRepository) DeleteNote(note model.Note) (model.Note, error) {
	query := "DELETE FROM notes WHERE id = ?"
	result, err := r.db.Exec(query, note.ID)
	if err != nil {
		return model.Note{}, err
	}

	// 削除された行数を取得
	affected, err := result.RowsAffected()
	if err != nil {
		return model.Note{}, err
	}

	if affected == 0 {
		return model.Note{}, sql.ErrNoRows
	}

	return note, nil
}
