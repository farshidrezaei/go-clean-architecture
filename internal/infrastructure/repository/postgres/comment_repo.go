package postgres

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CommentRepository struct {
	pool *pgxpool.Pool
}

func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{pool: pool}
}

func (r *CommentRepository) Create(ctx context.Context, comment *entities.Comment) error {
	q := querierFromContext(ctx, r.pool)
	_, err := q.Exec(ctx, `
		INSERT INTO comments (id, post_id, author_id, body, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, comment.ID, comment.PostID, comment.AuthorID, comment.Body, comment.CreatedAt, comment.UpdatedAt)
	return err
}

func (r *CommentRepository) GetByID(ctx context.Context, id string) (*entities.Comment, error) {
	q := querierFromContext(ctx, r.pool)
	var comment entities.Comment
	err := q.QueryRow(ctx, `
		SELECT id, post_id, author_id, body, created_at, updated_at
		FROM comments
		WHERE id = $1
	`, id).Scan(&comment.ID, &comment.PostID, &comment.AuthorID, &comment.Body, &comment.CreatedAt, &comment.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) Update(ctx context.Context, comment *entities.Comment) error {
	q := querierFromContext(ctx, r.pool)
	tag, err := q.Exec(ctx, `
		UPDATE comments
		SET body = $2, updated_at = $3
		WHERE id = $1
	`, comment.ID, comment.Body, comment.UpdatedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}
	return nil
}

func (r *CommentRepository) Delete(ctx context.Context, id string) error {
	q := querierFromContext(ctx, r.pool)
	tag, err := q.Exec(ctx, `DELETE FROM comments WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}
	return nil
}

func (r *CommentRepository) ListByPostID(ctx context.Context, postID string, pagination repositories.Pagination) (repositories.PageResult[*entities.Comment], error) {
	q := querierFromContext(ctx, r.pool)
	rows, err := q.Query(ctx, `
		SELECT id, post_id, author_id, body, created_at, updated_at
		FROM comments
		WHERE post_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`, postID, pagination.Limit, pagination.Offset)
	if err != nil {
		return repositories.PageResult[*entities.Comment]{}, err
	}
	defer rows.Close()

	items := make([]*entities.Comment, 0, pagination.Limit)
	for rows.Next() {
		var comment entities.Comment
		if err := rows.Scan(&comment.ID, &comment.PostID, &comment.AuthorID, &comment.Body, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			return repositories.PageResult[*entities.Comment]{}, err
		}
		copyValue := comment
		items = append(items, &copyValue)
	}

	var total int
	if err := q.QueryRow(ctx, `SELECT COUNT(1) FROM comments WHERE post_id = $1`, postID).Scan(&total); err != nil {
		return repositories.PageResult[*entities.Comment]{}, err
	}

	return repositories.PageResult[*entities.Comment]{Items: items, Page: pagination.Page, Limit: pagination.Limit, Total: total}, nil
}
