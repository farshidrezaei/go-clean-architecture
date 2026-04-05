package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"clean_architecture/internal/infrastructure/auth"
	"clean_architecture/internal/infrastructure/database/postgres"
	"clean_architecture/internal/infrastructure/health"
	"clean_architecture/internal/infrastructure/logger"
	pgrepo "clean_architecture/internal/infrastructure/repository/postgres"
	delivery "clean_architecture/internal/interface/http"
	ginadapter "clean_architecture/internal/interface/http/adapter/gin"
	"clean_architecture/internal/interface/http/handlers"
	httpmiddleware "clean_architecture/internal/interface/http/middleware"
	commentuc "clean_architecture/internal/usecase/comment"
	postuc "clean_architecture/internal/usecase/post"
	"clean_architecture/internal/usecase/shared"
	useruc "clean_architecture/internal/usecase/user"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgresAuthAndPostFlow(t *testing.T) {
	t.Parallel()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("NewPool() error = %v", err)
	}
	defer pool.Close()

	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	resetPostgresState(t, pool)

	users := pgrepo.NewUserRepository(pool)
	posts := pgrepo.NewPostRepository(pool)
	comments := pgrepo.NewCommentRepository(pool)
	refreshTokens := pgrepo.NewRefreshTokenRepository(pool)
	uow := pgrepo.NewUnitOfWork(pool)

	idGenerator := auth.UUIDGenerator{}
	clock := shared.SystemClock{}
	hasher := auth.BcryptHasher{}
	jwtService := auth.NewJWTService("integration-secret", 60*time.Minute, 24*time.Hour)
	events := shared.NoopPublisher{}

	registerUser := useruc.NewRegisterUser(users, hasher, idGenerator, clock)
	loginUser := useruc.NewLoginUser(users, refreshTokens, hasher, jwtService, idGenerator, clock, jwtService, uow, events)
	refreshSession := useruc.NewRefreshSession(users, refreshTokens, hasher, jwtService, idGenerator, clock, jwtService, uow, events)
	logoutSession := useruc.NewLogoutSession(refreshTokens, jwtService, clock, uow, events)
	logoutAllSessions := useruc.NewLogoutAllSessions(refreshTokens, clock, events)
	listSessions := useruc.NewListSessions(refreshTokens)
	revokeSession := useruc.NewRevokeSession(refreshTokens, clock, uow, events)
	createPost := postuc.NewCreatePost(posts, users, uow, idGenerator, clock)
	updatePost := postuc.NewUpdatePost(posts, clock)
	deletePost := postuc.NewDeletePost(posts)
	likePost := postuc.NewLikePost(posts, uow)
	getPost := postuc.NewGetPost(posts)
	listPosts := postuc.NewListPosts(posts)
	addComment := commentuc.NewAddComment(comments, posts, users, uow, idGenerator, clock)
	updateComment := commentuc.NewUpdateComment(comments, clock)
	deleteComment := commentuc.NewDeleteComment(comments)
	listComments := commentuc.NewListComments(comments)
	metrics := httpmiddleware.NewMetrics()

	router := delivery.NewRouter(delivery.RouterDependencies{
		UserHandler:    handlers.NewUserHandler(registerUser, loginUser, refreshSession, logoutSession, logoutAllSessions, listSessions, revokeSession),
		PostHandler:    handlers.NewPostHandler(createPost, updatePost, deletePost, likePost, getPost, listPosts),
		CommentHandler: handlers.NewCommentHandler(addComment, updateComment, deleteComment, listComments),
		AuthMiddleware: delivery.NewAuthMiddleware(jwtService),
		RequestContext: ginadapter.RequestContext(logger.New("development", "INFO")),
		RequestLogger:  ginadapter.RequestLogger(logger.New("development", "INFO")),
		HealthChecker:  health.NewPostgresChecker(pool),
		Metrics:        metrics.Middleware(),
		MetricsHandler: metrics.Handler(),
		Tracing:        httpmiddleware.Tracing("blog-api-test"),
	})

	registerResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/register", "", map[string]string{
		"name":     "Postgres User",
		"email":    "postgres@example.com",
		"password": "supersecret",
	})
	if registerResp.Code != http.StatusCreated {
		t.Fatalf("register status = %d body=%s", registerResp.Code, registerResp.Body.String())
	}

	loginResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/login", "", map[string]string{
		"email":    "postgres@example.com",
		"password": "supersecret",
	})
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", loginResp.Code, loginResp.Body.String())
	}

	var authPayload struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &authPayload); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	if authPayload.AccessToken == "" {
		t.Fatal("expected access token")
	}

	postResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/posts", authPayload.AccessToken, map[string]any{
		"title":       "postgres hello",
		"content":     "postgres world",
		"publish_now": true,
	})
	if postResp.Code != http.StatusCreated {
		t.Fatalf("create post status = %d body=%s", postResp.Code, postResp.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/posts?page=1&limit=10", bytes.NewReader(nil))
	listRecorder := httptest.NewRecorder()
	router.ServeHTTP(listRecorder, listReq)
	if listRecorder.Code != http.StatusOK {
		t.Fatalf("list posts status = %d body=%s", listRecorder.Code, listRecorder.Body.String())
	}
	if !bytes.Contains(listRecorder.Body.Bytes(), []byte("postgres hello")) {
		t.Fatalf("expected created post in listing, got %s", listRecorder.Body.String())
	}
}

func resetPostgresState(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `
		TRUNCATE TABLE refresh_tokens, comments, post_likes, posts, users RESTART IDENTITY CASCADE
	`); err != nil {
		t.Fatalf("truncate tables: %v", err)
	}
}
