package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"

	"clean_architecture/internal/infrastructure/auth"
	"clean_architecture/internal/infrastructure/config"
	databasepostgres "clean_architecture/internal/infrastructure/database/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	_ = config.LoadDotEnv()
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	pool, err := databasepostgres.NewPool(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	if err := databasepostgres.Migrate(context.Background(), pool); err != nil {
		log.Fatalf("migrate db: %v", err)
	}

	if err := seed(context.Background(), pool); err != nil {
		log.Fatalf("seed db: %v", err)
	}

	log.Println("seed complete")
	log.Println("sample admin: admin@example.com / password123")
	log.Println("sample users: sara@example.com / password123, omid@example.com / password123")
}

func seed(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	hasher := auth.BcryptHasher{}
	passwordHash, err := hasher.Hash("password123")
	if err != nil {
		return err
	}

	now := time.Date(2026, 4, 5, 9, 0, 0, 0, time.UTC)
	userAdminID := mustUUID("11111111-1111-1111-1111-111111111111")
	userSaraID := mustUUID("22222222-2222-2222-2222-222222222222")
	userOmidID := mustUUID("33333333-3333-3333-3333-333333333333")
	postOneID := mustUUID("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	postTwoID := mustUUID("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")
	postThreeID := mustUUID("cccccccc-cccc-cccc-cccc-cccccccccccc")
	commentOneID := mustUUID("dddddddd-dddd-dddd-dddd-dddddddddddd")
	commentTwoID := mustUUID("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee")
	commentThreeID := mustUUID("ffffffff-ffff-ffff-ffff-ffffffffffff")
	sessionOneID := mustUUID("12345678-1234-1234-1234-1234567890ab")
	sessionTwoID := mustUUID("22345678-1234-1234-1234-1234567890ab")
	sessionThreeID := mustUUID("32345678-1234-1234-1234-1234567890ab")
	users := []struct {
		ID    string
		Name  string
		Email string
		Role  string
	}{
		{userAdminID, "Admin User", "admin@example.com", "admin"},
		{userSaraID, "Sara Rahimi", "sara@example.com", "user"},
		{userOmidID, "Omid Karimi", "omid@example.com", "user"},
	}

	for i, user := range users {
		_, err := tx.Exec(ctx, `
			INSERT INTO users (id, name, email, role, password_hash, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE
			SET name = EXCLUDED.name, email = EXCLUDED.email, role = EXCLUDED.role, password_hash = EXCLUDED.password_hash
		`, user.ID, user.Name, user.Email, user.Role, passwordHash, now.Add(time.Duration(i)*time.Minute))
		if err != nil {
			return err
		}
	}

	posts := []struct {
		ID          string
		AuthorID    string
		Title       string
		Content     string
		Status      string
		LikesCount  int
		CreatedAt   time.Time
		UpdatedAt   time.Time
		PublishedAt *time.Time
	}{
		{
			ID:          postOneID,
			AuthorID:    userSaraID,
			Title:       "Clean Architecture in Practice",
			Content:     "A sample seeded post about boundaries, use cases, and pragmatic layering.",
			Status:      "published",
			LikesCount:  2,
			CreatedAt:   now.Add(2 * time.Hour),
			UpdatedAt:   now.Add(2 * time.Hour),
			PublishedAt: timePtr(now.Add(2 * time.Hour)),
		},
		{
			ID:          postTwoID,
			AuthorID:    userOmidID,
			Title:       "Designing Use Cases That Stay Small",
			Content:     "Another seeded article with realistic but simple demo content.",
			Status:      "published",
			LikesCount:  1,
			CreatedAt:   now.Add(3 * time.Hour),
			UpdatedAt:   now.Add(3 * time.Hour),
			PublishedAt: timePtr(now.Add(3 * time.Hour)),
		},
		{
			ID:         postThreeID,
			AuthorID:   userAdminID,
			Title:      "Starter Kit Roadmap",
			Content:    "A draft post to show mixed post states in the sample dataset.",
			Status:     "draft",
			LikesCount: 0,
			CreatedAt:  now.Add(4 * time.Hour),
			UpdatedAt:  now.Add(4 * time.Hour),
		},
	}

	for _, post := range posts {
		_, err := tx.Exec(ctx, `
			INSERT INTO posts (id, author_id, title, content, status, likes_count, created_at, updated_at, published_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (id) DO UPDATE
			SET title = EXCLUDED.title, content = EXCLUDED.content, status = EXCLUDED.status, likes_count = EXCLUDED.likes_count, updated_at = EXCLUDED.updated_at, published_at = EXCLUDED.published_at
		`, post.ID, post.AuthorID, post.Title, post.Content, post.Status, post.LikesCount, post.CreatedAt, post.UpdatedAt, post.PublishedAt)
		if err != nil {
			return err
		}
	}

	comments := []struct {
		ID       string
		PostID   string
		AuthorID string
		Body     string
		Created  time.Time
	}{
		{commentOneID, postOneID, userOmidID, "This seeded example is clear and easy to extend.", now.Add(5 * time.Hour)},
		{commentTwoID, postOneID, userAdminID, "Admin note: keep business rules inside entities and use cases.", now.Add(6 * time.Hour)},
		{commentThreeID, postTwoID, userSaraID, "Nice example of explicit orchestration in the application layer.", now.Add(7 * time.Hour)},
	}

	for _, comment := range comments {
		_, err := tx.Exec(ctx, `
			INSERT INTO comments (id, post_id, author_id, body, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO UPDATE
			SET body = EXCLUDED.body, updated_at = EXCLUDED.updated_at
		`, comment.ID, comment.PostID, comment.AuthorID, comment.Body, comment.Created, comment.Created)
		if err != nil {
			return err
		}
	}

	likes := []struct {
		PostID string
		UserID string
	}{
		{postOneID, userAdminID},
		{postOneID, userOmidID},
		{postTwoID, userSaraID},
	}

	for _, like := range likes {
		_, err := tx.Exec(ctx, `
			INSERT INTO post_likes (post_id, user_id)
			VALUES ($1, $2)
			ON CONFLICT (post_id, user_id) DO NOTHING
		`, like.PostID, like.UserID)
		if err != nil {
			return err
		}
	}

	sessions := []struct {
		ID         string
		FamilyID   string
		UserID     string
		TokenPlain string
		DeviceName string
		UserAgent  string
		IPAddress  string
		CreatedAt  time.Time
	}{
		{sessionOneID, sessionOneID, userAdminID, "admin-refresh-token", "MacBook Pro", "Seeder/1.0", "127.0.0.1", now.Add(8 * time.Hour)},
		{sessionTwoID, sessionTwoID, userSaraID, "sara-refresh-token", "iPhone", "Seeder/1.0", "127.0.0.2", now.Add(9 * time.Hour)},
		{sessionThreeID, sessionThreeID, userOmidID, "omid-refresh-token", "Linux Laptop", "Seeder/1.0", "127.0.0.3", now.Add(10 * time.Hour)},
	}

	for _, session := range sessions {
		_, err := tx.Exec(ctx, `
			INSERT INTO refresh_tokens (
				id, family_id, user_id, token_hash, device_name, user_agent, ip_address,
				expires_at, revoked_at, replaced_by_id, compromised_at, created_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NULL, NULL, NULL, $9)
			ON CONFLICT (id) DO UPDATE
			SET device_name = EXCLUDED.device_name, user_agent = EXCLUDED.user_agent, ip_address = EXCLUDED.ip_address, expires_at = EXCLUDED.expires_at
		`, session.ID, session.FamilyID, session.UserID, hashToken(session.TokenPlain), session.DeviceName, session.UserAgent, session.IPAddress, session.CreatedAt.Add(7*24*time.Hour), session.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func hashToken(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func mustUUID(value string) string {
	return uuid.MustParse(value).String()
}

func timePtr(value time.Time) *time.Time {
	return &value
}
