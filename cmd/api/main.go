package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"clean_architecture/internal/domain/repositories"
	"clean_architecture/internal/infrastructure/audit"
	"clean_architecture/internal/infrastructure/auth"
	"clean_architecture/internal/infrastructure/config"
	databasepostgres "clean_architecture/internal/infrastructure/database/postgres"
	"clean_architecture/internal/infrastructure/health"
	"clean_architecture/internal/infrastructure/logger"
	"clean_architecture/internal/infrastructure/ratelimit"
	memoryrepo "clean_architecture/internal/infrastructure/repository/memory"
	pgrepo "clean_architecture/internal/infrastructure/repository/postgres"
	"clean_architecture/internal/infrastructure/telemetry"
	delivery "clean_architecture/internal/interface/http"
	ginadapter "clean_architecture/internal/interface/http/adapter/gin"
	"clean_architecture/internal/interface/http/handlers"
	httpmiddleware "clean_architecture/internal/interface/http/middleware"
	adminuc "clean_architecture/internal/usecase/admin"
	commentuc "clean_architecture/internal/usecase/comment"
	postuc "clean_architecture/internal/usecase/post"
	"clean_architecture/internal/usecase/shared"
	useruc "clean_architecture/internal/usecase/user"
	"github.com/gin-gonic/gin"
)

func main() {
	bootstrapLogger := logger.New("development", "INFO")
	if err := config.LoadDotEnv(); err != nil {
		bootstrapLogger.Error("failed to load .env", "error", err)
		os.Exit(1)
	}

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		bootstrapLogger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	log := logger.New(cfg.AppEnv, cfg.LogLevel)
	shutdownTracing, err := telemetry.Setup(context.Background(), telemetry.Config{
		Enabled:        cfg.OTELEnabled,
		ServiceName:    cfg.OTELServiceName,
		OTLPEndpoint:   cfg.OTLPTraceEndpoint,
		OTLPInsecure:   cfg.OTLPInsecure,
		OTLPHeaders:    cfg.OTLPHeaders,
		OTLPTimeout:    cfg.OTLPTimeout,
		OTLPCertFile:   cfg.OTLPCertFile,
		OTLPServerName: cfg.OTLPServerName,
	}, log)
	if err != nil {
		log.Error("failed to initialize tracing", "error", err)
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := shutdownTracing(shutdownCtx); err != nil {
			log.Error("failed to shutdown tracing", "error", err)
		}
	}()

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	storeBundle, cleanup, err := buildStorage(context.Background(), cfg)
	if err != nil {
		log.Error("failed to initialize storage", "error", err)
		os.Exit(1)
	}
	defer cleanup()
	events := audit.NewCompositePublisher(audit.NewLogPublisher(log), storeBundle.auditPublisher)

	idGenerator := auth.UUIDGenerator{}
	clock := shared.SystemClock{}
	passwordHasher := auth.BcryptHasher{}
	jwtService := auth.NewJWTService(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	registerUser := useruc.NewRegisterUser(storeBundle.users, passwordHasher, idGenerator, clock)
	loginUser := useruc.NewLoginUser(storeBundle.users, storeBundle.refreshTokens, passwordHasher, jwtService, idGenerator, clock, jwtService, storeBundle.uow, events)
	refreshSession := useruc.NewRefreshSession(storeBundle.users, storeBundle.refreshTokens, passwordHasher, jwtService, idGenerator, clock, jwtService, storeBundle.uow, events)
	logoutSession := useruc.NewLogoutSession(storeBundle.refreshTokens, jwtService, clock, storeBundle.uow, events)
	logoutAllSessions := useruc.NewLogoutAllSessions(storeBundle.refreshTokens, clock, events)
	listSessions := useruc.NewListSessions(storeBundle.refreshTokens)
	revokeSession := useruc.NewRevokeSession(storeBundle.refreshTokens, clock, storeBundle.uow, events)

	createPost := postuc.NewCreatePost(storeBundle.posts, storeBundle.users, storeBundle.uow, idGenerator, clock)
	updatePost := postuc.NewUpdatePost(storeBundle.posts, clock)
	deletePost := postuc.NewDeletePost(storeBundle.posts)
	likePost := postuc.NewLikePost(storeBundle.posts, storeBundle.uow)
	getPost := postuc.NewGetPost(storeBundle.posts)
	listPosts := postuc.NewListPosts(storeBundle.posts)

	addComment := commentuc.NewAddComment(storeBundle.comments, storeBundle.posts, storeBundle.users, storeBundle.uow, idGenerator, clock)
	updateComment := commentuc.NewUpdateComment(storeBundle.comments, clock)
	deleteComment := commentuc.NewDeleteComment(storeBundle.comments)
	listComments := commentuc.NewListComments(storeBundle.comments)
	listUsers := adminuc.NewListUsers(storeBundle.users)
	promoteUser := adminuc.NewPromoteUser(storeBundle.users, events)

	metrics := httpmiddleware.NewMetrics()

	router := delivery.NewRouter(delivery.RouterDependencies{
		UserHandler:      handlers.NewUserHandler(registerUser, loginUser, refreshSession, logoutSession, logoutAllSessions, listSessions, revokeSession),
		PostHandler:      handlers.NewPostHandler(createPost, updatePost, deletePost, likePost, getPost, listPosts),
		CommentHandler:   handlers.NewCommentHandler(addComment, updateComment, deleteComment, listComments),
		AdminHandler:     handlers.NewAdminHandler(listUsers, promoteUser),
		AuthMiddleware:   delivery.NewAuthMiddleware(jwtService),
		LoginRateLimit:   httpmiddleware.LoginRateLimit(storeBundle.rateLimiter),
		RefreshRateLimit: httpmiddleware.RefreshRateLimit(storeBundle.rateLimiter),
		RequestContext:   ginadapter.RequestContext(log),
		RequestLogger:    ginadapter.RequestLogger(log),
		HealthChecker:    storeBundle.healthChecker,
		Metrics:          metrics.Middleware(),
		MetricsHandler:   metrics.Handler(),
		Tracing:          httpmiddleware.Tracing(cfg.OTELServiceName),
	})

	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	go func() {
		log.Info("server starting", "port", cfg.HTTPPort, "storage", cfg.StorageDriver, "env", cfg.AppEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server crashed", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
	}
	log.Info("server stopped")
}

type storageBundle struct {
	users          repositories.UserRepository
	posts          repositories.PostRepository
	comments       repositories.CommentRepository
	refreshTokens  repositories.RefreshTokenRepository
	uow            repositories.UnitOfWork
	healthChecker  health.Checker
	rateLimiter    ratelimit.Service
	auditPublisher shared.EventPublisher
}

func buildStorage(ctx context.Context, cfg config.Config) (*storageBundle, func(), error) {
	switch cfg.StorageDriver {
	case "postgres":
		pool, err := databasepostgres.NewPool(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, nil, err
		}
		if err := databasepostgres.Migrate(ctx, pool); err != nil {
			pool.Close()
			return nil, nil, err
		}

		users := pgrepo.NewUserRepository(pool)
		posts := pgrepo.NewPostRepository(pool)
		comments := pgrepo.NewCommentRepository(pool)
		refreshTokens := pgrepo.NewRefreshTokenRepository(pool)
		uow := pgrepo.NewUnitOfWork(pool)
		rateLimiter := ratelimit.NewPostgresService(pool)

		return &storageBundle{
			users:          users,
			posts:          posts,
			comments:       comments,
			refreshTokens:  refreshTokens,
			uow:            uow,
			healthChecker:  health.NewPostgresChecker(pool),
			rateLimiter:    rateLimiter,
			auditPublisher: audit.NewDBPublisher(pool),
		}, pool.Close, nil
	default:
		store := memoryrepo.NewStore()
		users := memoryrepo.NewUserRepository(store)
		posts := memoryrepo.NewPostRepository(store)
		comments := memoryrepo.NewCommentRepository(store)
		refreshTokens := memoryrepo.NewRefreshTokenRepository(store)
		uow := memoryrepo.NewUnitOfWork(store)
		rateLimiter := ratelimit.NewMemoryService()

		return &storageBundle{
			users:          users,
			posts:          posts,
			comments:       comments,
			refreshTokens:  refreshTokens,
			uow:            uow,
			healthChecker:  health.MemoryChecker{},
			rateLimiter:    rateLimiter,
			auditPublisher: nil,
		}, func() {}, nil
	}
}
