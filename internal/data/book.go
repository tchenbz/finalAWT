package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/tchenbz/test3AWT/internal/validator"
)

type Book struct {
    ID            int64     `json:"id"`
    Title         string    `json:"title"`
    Authors       []string  `json:"authors"`
    ISBN          string    `json:"isbn"`
    Publication   time.Time `json:"publication_date"`
    Genre         string    `json:"genre"`
    Description   string    `json:"description"`
    AverageRating float32   `json:"average_rating"`
    CreatedAt     time.Time `json:"-"`
    Version       int32     `json:"version"`
}

func (b *Book) MarshalJSON() ([]byte, error) {
    type Alias Book
    return json.Marshal(&struct {
        Publication string `json:"publication_date"`
        *Alias
    }{
        Publication: b.Publication.Format("2006-01-02"),
        Alias:       (*Alias)(b),
    })
}

type BookModel struct {
	DB *sql.DB
}

func (m BookModel) Insert(book *Book) error {
    query := `
        INSERT INTO books (title, authors, isbn, publication_date, genre, description)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, version`

    args := []interface{}{book.Title, pq.Array(book.Authors), book.ISBN, book.Publication, book.Genre, book.Description}
    return m.DB.QueryRowContext(context.Background(), query, args...).Scan(&book.ID, &book.CreatedAt, &book.Version)
}

func ValidateBook(v *validator.Validator, book *Book) {
	v.Check(book.Title != "", "title", "must be provided")
	v.Check(len(book.Authors) > 0, "authors", "must include at least one author")
	v.Check(book.ISBN != "", "isbn", "must be provided")
	v.Check(book.Publication != time.Time{}, "publication_date", "must be provided")
	v.Check(book.Genre != "", "genre", "must be provided")
	v.Check(book.Description != "", "description", "must be provided")
}

func (m BookModel) Get(id int64) (*Book, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, title, authors, isbn, publication_date, genre, description, average_rating, created_at, version
		FROM books
		WHERE id = $1`

	var book Book

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&book.ID,
		&book.Title,
		pq.Array(&book.Authors),
		&book.ISBN,
		&book.Publication,
		&book.Genre,
		&book.Description,
		&book.AverageRating,
		&book.CreatedAt,
		&book.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &book, nil
}

func (m BookModel) Update(book *Book) error {
    query := `
        UPDATE books
        SET title = $1, authors = $2, isbn = $3, publication_date = $4, genre = $5, description = $6, version = version + 1
        WHERE id = $7
        RETURNING version`

    args := []interface{}{
        book.Title,
        pq.Array(book.Authors),
        book.ISBN,
        book.Publication, 
        book.Genre,
        book.Description,
        book.ID,
    }

    return m.DB.QueryRowContext(context.Background(), query, args...).Scan(&book.Version)
}

func (m BookModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM books
		WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m BookModel) GetAll(title, author, genre string, filters Filters) ([]*Book, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, title, authors, isbn, publication_date, genre, description, average_rating, created_at, version
		FROM books
		WHERE (title ILIKE $1 OR $1 = '')
		AND (authors @> ARRAY[$2]::TEXT[] OR $2 = '')
		AND (genre ILIKE $3 OR $3 = '')
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	args := []interface{}{
		"%" + title + "%",
		author,
		"%" + genre + "%",
		filters.limit(),
		filters.offset(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	books := []*Book{}

	for rows.Next() {
		var book Book
		err := rows.Scan(
			&totalRecords,
			&book.ID,
			&book.Title,
			pq.Array(&book.Authors),
			&book.ISBN,
			&book.Publication,
			&book.Genre,
			&book.Description,
			&book.AverageRating,
			&book.CreatedAt,
			&book.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		books = append(books, &book)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return books, metadata, nil
}

