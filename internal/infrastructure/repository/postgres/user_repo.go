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

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	q := querierFromContext(ctx, r.pool)
	_, err := q.Exec(ctx, `
		INSERT INTO users (id, name, email, role, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.Name, user.Email, user.Role, user.PasswordHash, user.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entities.ErrConflict
		}
		return err
	}
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	q := querierFromContext(ctx, r.pool)
	var user entities.User
	err := q.QueryRow(ctx, `
		SELECT id, name, email, role, password_hash, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	q := querierFromContext(ctx, r.pool)
	var user entities.User
	err := q.QueryRow(ctx, `
		SELECT id, name, email, role, password_hash, created_at
		FROM users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, entities.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) List(ctx context.Context, pagination repositories.Pagination) (repositories.PageResult[*entities.User], error) {
	q := querierFromContext(ctx, r.pool)
	rows, err := q.Query(ctx, `
		SELECT id, name, email, role, password_hash, created_at
		FROM users
		ORDER BY created_at ASC
		LIMIT $1 OFFSET $2
	`, pagination.Limit, pagination.Offset)
	if err != nil {
		return repositories.PageResult[*entities.User]{}, err
	}
	defer rows.Close()

	items := make([]*entities.User, 0, pagination.Limit)
	for rows.Next() {
		var user entities.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.PasswordHash, &user.CreatedAt); err != nil {
			return repositories.PageResult[*entities.User]{}, err
		}
		copyValue := user
		items = append(items, &copyValue)
	}
	var total int
	if err := q.QueryRow(ctx, `SELECT COUNT(1) FROM users`).Scan(&total); err != nil {
		return repositories.PageResult[*entities.User]{}, err
	}
	return repositories.PageResult[*entities.User]{Items: items, Page: pagination.Page, Limit: pagination.Limit, Total: total}, nil
}

func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	q := querierFromContext(ctx, r.pool)
	tag, err := q.Exec(ctx, `
		UPDATE users
		SET name = $2, email = $3, role = $4, password_hash = $5
		WHERE id = $1
	`, user.ID, user.Name, user.Email, user.Role, user.PasswordHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return entities.ErrNotFound
	}
	return nil
}
