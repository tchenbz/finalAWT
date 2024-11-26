package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrRecordNotFound = errors.New("record not found")

type Review struct {
	ID           int64     `json:"id"`
	ProductID    int64     `json:"product_id"`
	Content      string    `json:"content"`
	Author       string    `json:"author"`
	Rating       int       `json:"rating"`         
	HelpfulCount int       `json:"helpful_count"`  
	CreatedAt    time.Time `json:"created_at"`
	Version      int32     `json:"version"`
}

type ReviewModel struct {
	DB *sql.DB
}

func (m ReviewModel) Insert(review *Review) error {
	query := `
		INSERT INTO reviews (product_id, content, author, rating, helpful_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, version`

	args := []interface{}{review.ProductID, review.Content, review.Author, review.Rating, review.HelpfulCount}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&review.ID, &review.CreatedAt, &review.Version)
}

func (m ReviewModel) Get(productID, reviewID int64) (*Review, error) {
	if productID < 1 || reviewID < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
		SELECT id, product_id, content, author, rating, helpful_count, created_at, version
		FROM reviews
		WHERE product_id = $1 AND id = $2`

	var review Review

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, productID, reviewID).Scan(
		&review.ID, &review.ProductID, &review.Content, &review.Author, 
		&review.Rating, &review.HelpfulCount, &review.CreatedAt, &review.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &review, nil
}

func (m ReviewModel) Update(review *Review) error {
	query := `
		UPDATE reviews
		SET content = $1, author = $2, rating = $3, helpful_count = $4, version = version + 1
		WHERE product_id = $5 AND id = $6
		RETURNING version`

	args := []interface{}{review.Content, review.Author, review.Rating, review.HelpfulCount, review.ProductID, review.ID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&review.Version)
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

func (m ReviewModel) Delete(productID, reviewID int64) error {
	if productID < 1 || reviewID < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM reviews
		WHERE product_id = $1 AND id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, productID, reviewID)
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

func (m ReviewModel) GetAll(content, author string, rating int, filters Filters) ([]*Review, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, product_id, content, author, rating, helpful_count, created_at, version
		FROM reviews
		WHERE (content ILIKE $1 OR $1 = '')
		AND (author ILIKE $2 OR $2 = '')
		AND (rating = $3 OR $3 = 0)
		ORDER BY %s %s, id ASC
		LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	args := []interface{}{
		"%" + content + "%",
		"%" + author + "%",
		rating,
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
	reviews := []*Review{}

	for rows.Next() {
		var review Review
		err := rows.Scan(
			&totalRecords,
			&review.ID,
			&review.ProductID,
			&review.Content,
			&review.Author,
			&review.Rating,
			&review.HelpfulCount,
			&review.CreatedAt,
			&review.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return reviews, metadata, nil
}

func (m ReviewModel) GetAllForProduct(productID int64, content, author string, rating int, filters Filters) ([]*Review, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), id, product_id, content, author, rating, helpful_count, created_at, version
		FROM reviews
		WHERE product_id = $1
		AND (content ILIKE $2 OR $2 = '')
		AND (author ILIKE $3 OR $3 = '')
		AND (rating = $4 OR $4 = 0)
		ORDER BY %s %s, id ASC
		LIMIT $5 OFFSET $6`, filters.sortColumn(), filters.sortDirection())

	args := []interface{}{
		productID,
		"%" + content + "%",
		"%" + author + "%",
		rating,
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
	reviews := []*Review{}

	for rows.Next() {
		var review Review
		err := rows.Scan(
			&totalRecords,
			&review.ID,
			&review.ProductID,
			&review.Content,
			&review.Author,
			&review.Rating,
			&review.HelpfulCount,
			&review.CreatedAt,
			&review.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}
		reviews = append(reviews, &review)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)
	return reviews, metadata, nil
}