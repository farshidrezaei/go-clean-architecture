package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"clean_architecture/internal/infrastructure/auth"
	"clean_architecture/internal/infrastructure/health"
	"clean_architecture/internal/infrastructure/logger"
	memoryrepo "clean_architecture/internal/infrastructure/repository/memory"
	delivery "clean_architecture/internal/interface/http"
	ginadapter "clean_architecture/internal/interface/http/adapter/gin"
	"clean_architecture/internal/interface/http/handlers"
	httpmiddleware "clean_architecture/internal/interface/http/middleware"
	commentuc "clean_architecture/internal/usecase/comment"
	postuc "clean_architecture/internal/usecase/post"
	"clean_architecture/internal/usecase/shared"
	useruc "clean_architecture/internal/usecase/user"
	"github.com/gin-gonic/gin"
)

func TestAuthRefreshAndCreatePostFlow(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	store := memoryrepo.NewStore()
	users := memoryrepo.NewUserRepository(store)
	posts := memoryrepo.NewPostRepository(store)
	comments := memoryrepo.NewCommentRepository(store)
	refreshTokens := memoryrepo.NewRefreshTokenRepository(store)
	uow := memoryrepo.NewUnitOfWork(store)

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
		HealthChecker:  health.MemoryChecker{},
		Metrics:        metrics.Middleware(),
		MetricsHandler: metrics.Handler(),
		Tracing:        httpmiddleware.Tracing("blog-api-test"),
	})

	registerBody := map[string]string{
		"name":     "Farshid",
		"email":    "farshid@example.com",
		"password": "supersecret",
	}
	registerResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/register", "", registerBody)
	if registerResp.Code != http.StatusCreated {
		t.Fatalf("register status = %d body=%s", registerResp.Code, registerResp.Body.String())
	}

	loginBody := map[string]string{
		"email":    "farshid@example.com",
		"password": "supersecret",
	}
	loginResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/login", "", loginBody)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status = %d body=%s", loginResp.Code, loginResp.Body.String())
	}

	var authPayload struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &authPayload); err != nil {
		t.Fatalf("unmarshal login response: %v", err)
	}
	if authPayload.AccessToken == "" || authPayload.RefreshToken == "" {
		t.Fatalf("expected both access and refresh tokens, got %+v", authPayload)
	}

	postResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/posts", authPayload.AccessToken, map[string]any{
		"title":       "hello",
		"content":     "world",
		"publish_now": true,
	})
	if postResp.Code != http.StatusCreated {
		t.Fatalf("create post status = %d body=%s", postResp.Code, postResp.Body.String())
	}

	var createdPost struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(postResp.Body.Bytes(), &createdPost); err != nil {
		t.Fatalf("unmarshal post response: %v", err)
	}
	if createdPost.ID == "" {
		t.Fatal("expected created post id")
	}
	if requestID := postResp.Header().Get("X-Request-ID"); requestID == "" {
		t.Fatal("expected X-Request-ID header on protected endpoint")
	}

	getPostReq := httptest.NewRequest(http.MethodGet, "/api/v1/posts/"+createdPost.ID, nil)
	getPostRecorder := httptest.NewRecorder()
	router.ServeHTTP(getPostRecorder, getPostReq)
	if getPostRecorder.Code != http.StatusOK {
		t.Fatalf("get post status = %d body=%s", getPostRecorder.Code, getPostRecorder.Body.String())
	}

	refreshResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{
		"refresh_token": authPayload.RefreshToken,
	})
	if refreshResp.Code != http.StatusOK {
		t.Fatalf("refresh status = %d body=%s", refreshResp.Code, refreshResp.Body.String())
	}

	reuseResp := performJSONRequest(t, router, http.MethodPost, "/api/v1/auth/refresh", "", map[string]string{
		"refresh_token": authPayload.RefreshToken,
	})
	if reuseResp.Code != http.StatusUnauthorized {
		t.Fatalf("expected refresh reuse to be rejected, got status = %d body=%s", reuseResp.Code, reuseResp.Body.String())
	}

	sessionsReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/sessions", nil)
	sessionsReq.Header.Set("Authorization", "Bearer "+authPayload.AccessToken)
	sessionsRecorder := httptest.NewRecorder()
	router.ServeHTTP(sessionsRecorder, sessionsReq)
	if sessionsRecorder.Code != http.StatusOK {
		t.Fatalf("sessions status = %d body=%s", sessionsRecorder.Code, sessionsRecorder.Body.String())
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRecorder := httptest.NewRecorder()
	router.ServeHTTP(metricsRecorder, metricsReq)
	if metricsRecorder.Code != http.StatusOK {
		t.Fatalf("metrics status = %d body=%s", metricsRecorder.Code, metricsRecorder.Body.String())
	}
	if !bytes.Contains(metricsRecorder.Body.Bytes(), []byte("blog_api_http_requests_total")) {
		t.Fatalf("metrics output missing request counter: %s", metricsRecorder.Body.String())
	}

	docsReq := httptest.NewRequest(http.MethodGet, "/docs", nil)
	docsRecorder := httptest.NewRecorder()
	router.ServeHTTP(docsRecorder, docsReq)
	if docsRecorder.Code != http.StatusOK {
		t.Fatalf("docs status = %d body=%s", docsRecorder.Code, docsRecorder.Body.String())
	}
}

func performJSONRequest(t *testing.T, router *gin.Engine, method, path, bearerToken string, payload any) *httptest.ResponseRecorder {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}
