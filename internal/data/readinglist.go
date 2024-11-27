package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type ReadingList struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   int64     `json:"created_by"` 
	Books       []int64   `json:"books"`      
	Status      string    `json:"status"`    
	CreatedAt   time.Time `json:"created_at"`
	Version     int32     `json:"version"`
}

type ReadingListModel struct {
	DB *sql.DB
}

func (m ReadingListModel) Insert(list *ReadingList) error {
	query := `
		INSERT INTO reading_lists (name, description, created_by, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	args := []interface{}{list.Name, list.Description, list.CreatedBy, list.Status}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&list.ID, &list.CreatedAt, &list.Version)
}

func (m ReadingListModel) Get(id int64) (*ReadingList, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, name, description, created_by, status, created_at, version
		FROM reading_lists
		WHERE id = $1`

	var list ReadingList

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&list.ID, &list.Name, &list.Description, &list.CreatedBy,
		&list.Status, &list.CreatedAt, &list.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	bookQuery := `SELECT book_id FROM reading_list_books WHERE reading_list_id = $1`
	rows, err := m.DB.QueryContext(ctx, bookQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var bookID int64
		err = rows.Scan(&bookID)
		if err != nil {
			return nil, err
		}
		list.Books = append(list.Books, bookID)
	}

	return &list, nil
}

func (m ReadingListModel) Update(list *ReadingList) error {
	query := `
		UPDATE reading_lists
		SET name = $1, description = $2, status = $3, version = version + 1
		WHERE id = $4
		RETURNING version`

	args := []interface{}{list.Name, list.Description, list.Status, list.ID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&list.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}

func (m ReadingListModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM reading_lists WHERE id = $1`

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

	deleteBooksQuery := `DELETE FROM reading_list_books WHERE reading_list_id = $1`
	_, err = m.DB.ExecContext(ctx, deleteBooksQuery, id)
	return err
}

func (m ReadingListModel) AddBook(listID, bookID int64) error {
	query := `
		INSERT INTO reading_list_books (reading_list_id, book_id)
		VALUES ($1, $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, listID, bookID)
	return err
}

func (m ReadingListModel) RemoveBook(listID, bookID int64) error {
	query := `
		DELETE FROM reading_list_books
		WHERE reading_list_id = $1 AND book_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, listID, bookID)
	return err
}

func (m ReadingListModel) GetAll(name string, filters Filters) ([]*ReadingList, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, name, description, created_by, status, created_at, version
		FROM reading_lists
		WHERE (name ILIKE $1 OR $1 = '')
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	args := []interface{}{
		"%" + name + "%",
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
	lists := []*ReadingList{}

	for rows.Next() {
		var list ReadingList
		err := rows.Scan(
			&totalRecords,
			&list.ID,
			&list.Name,
			&list.Description,
			&list.CreatedBy,
			&list.Status,
			&list.CreatedAt,
			&list.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		bookQuery := `SELECT book_id FROM reading_list_books WHERE reading_list_id = $1`
		bookRows, err := m.DB.QueryContext(ctx, bookQuery, list.ID)
		if err != nil {
			return nil, Metadata{}, err
		}

		for bookRows.Next() {
			var bookID int64
			err = bookRows.Scan(&bookID)
			if err != nil {
				return nil, Metadata{}, err
			}
			list.Books = append(list.Books, bookID)
		}
		bookRows.Close()

		lists = append(lists, &list)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return lists, metadata, nil
}

func (m ReadingListModel) GetAllByUser(userID int64, name string, filters Filters) ([]*ReadingList, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, name, description, created_by, status, created_at, version
		FROM reading_lists
		WHERE created_by = $1
		AND (name ILIKE $2 OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	args := []interface{}{
		userID,              
		"%" + name + "%",    
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
	lists := []*ReadingList{}

	for rows.Next() {
		var list ReadingList
		err := rows.Scan(
			&totalRecords,
			&list.ID,
			&list.Name,
			&list.Description,
			&list.CreatedBy,
			&list.Status,
			&list.CreatedAt,
			&list.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		lists = append(lists, &list)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return lists, metadata, nil
}
