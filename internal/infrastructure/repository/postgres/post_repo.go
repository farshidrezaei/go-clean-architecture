package postgres

import (
	"context"
	"errors"

	"clean_architecture/internal/domain/entities"
	"clean_architecture/internal/domain/repositories"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostRepository struct {
	pool *pgxpool.Pool
}

func NewPostRepository(pool *pgxpool.Pool) *PostRepository {
	return &PostRepository{pool: pool}
}

func (r *PostRepository) Create(ctx context.Context, post *entities.Post) error {
	q := querierFromContext(ctx, r.pool)
	_, err := q.Exec(ctx, `
		INSERT INTO posts (id, author_id, title, content, status, likes_count, created_at, updated_at, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, post.ID, post.AuthorID, post.Title, post.Content, post.Status, post.LikesCount, post.CreatedAt, post.UpdatedAt, post.PublishedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entities.ErrConflict
		}
		return err
	}
	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id string) (*entities.Post, error) {
	q := querierFromContext(ctx, r.pool)
	var post entities.Post
	err := q.QueryRow(ctx, `
		SELECT id, author_id, title, content, status, likes_count, created_at, updated_at, published_at
		FROM posts
		WHERE id = $1
	`, id).Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.Status, &post.LikesCount, &post.CreatedAt, &post.UpdatedAt, &post.PublishedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) Update(ctx context.Context, post *entities.Post) error {
	q := querierFromContext(ctx, r.pool)
	tag, err := q.Exec(ctx, `
		UPDATE posts
		SET title = $2, content = $3, status = $4, likes_count = $5, updated_at = $6, published_at = $7
		WHERE id = $1
	`, post.ID, post.Title, post.Content, post.Status, post.LikesCount, post.UpdatedAt, post.PublishedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}
	return nil
}

func (r *PostRepository) Delete(ctx context.Context, id string) error {
	q := querierFromContext(ctx, r.pool)
	tag, err := q.Exec(ctx, `DELETE FROM posts WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}
	return nil
}

func (r *PostRepository) AddLike(ctx context.Context, postID, userID string) error {
	q := querierFromContext(ctx, r.pool)
	_, err := q.Exec(ctx, `
		INSERT INTO post_likes (post_id, user_id)
		VALUES ($1, $2)
	`, postID, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entities.ErrAlreadyLiked
		}
		return err
	}
	return nil
}

func (r *PostRepository) List(ctx context.Context, pagination repositories.Pagination) (repositories.PageResult[*entities.Post], error) {
	q := querierFromContext(ctx, r.pool)
	rows, err := q.Query(ctx, `
		SELECT id, author_id, title, content, status, likes_count, created_at, updated_at, published_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, pagination.Limit, pagination.Offset)
	if err != nil {
		return repositories.PageResult[*entities.Post]{}, err
	}
	defer rows.Close()

	items := make([]*entities.Post, 0, pagination.Limit)
	for rows.Next() {
		var post entities.Post
		if err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.Status, &post.LikesCount, &post.CreatedAt, &post.UpdatedAt, &post.PublishedAt); err != nil {
			return repositories.PageResult[*entities.Post]{}, err
		}
		copyValue := post
		items = append(items, &copyValue)
	}

	var total int
	if err := q.QueryRow(ctx, `SELECT COUNT(1) FROM posts`).Scan(&total); err != nil {
		return repositories.PageResult[*entities.Post]{}, err
	}

	return repositories.PageResult[*entities.Post]{Items: items, Page: pagination.Page, Limit: pagination.Limit, Total: total}, nil
}
