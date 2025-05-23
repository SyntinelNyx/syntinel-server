package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/SyntinelNyx/syntinel-server/internal/action"
	"github.com/SyntinelNyx/syntinel-server/internal/asset"
	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/environment"
	"github.com/SyntinelNyx/syntinel-server/internal/limiter"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/SyntinelNyx/syntinel-server/internal/role"
	"github.com/SyntinelNyx/syntinel-server/internal/scan"
	"github.com/SyntinelNyx/syntinel-server/internal/snapshots"
	"github.com/SyntinelNyx/syntinel-server/internal/telemetry"
	"github.com/SyntinelNyx/syntinel-server/internal/terminal"
	"github.com/SyntinelNyx/syntinel-server/internal/user"
	"github.com/SyntinelNyx/syntinel-server/internal/vuln"
)

type Router struct {
	router      *chi.Mux
	queries     *query.Queries
	logger      *zap.Logger
	rateLimiter *limiter.RateLimiter
}

func SetupRouter(q *query.Queries, origins []string) *Router {
	router := chi.NewRouter()

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	zlogger, _ := zap.NewProduction()
	defer zlogger.Sync()

	rl := limiter.New()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Use(logger.Middleware(zlogger))

	r := Router{
		router:      router,
		queries:     q,
		logger:      zlogger,
		rateLimiter: rl,
	}

	r.router.Route("/v1/api", func(apiRouter chi.Router) {
		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.Middleware(rate.Every(1*time.Second), 30))

			assetHandler := asset.NewHandler(r.queries)

			subRouter.Get("/coffee", func(w http.ResponseWriter, r *http.Request) {
				response.RespondWithJSON(w, http.StatusTeapot, map[string]string{"error": "I'm A Teapot!"})
			})
			subRouter.Post("/agent/enroll", assetHandler.Enroll)
		})

		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.Middleware(rate.Every(1*time.Second), 3))

			authHandler := auth.NewHandler(r.queries)

			subRouter.Post("/auth/login", authHandler.Login)
			subRouter.Post("/auth/register", authHandler.Register)
		})

		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.Middleware(rate.Every(1*time.Second), 5))

			roleHandler := role.NewHandler(r.queries)
			authHandler := auth.NewHandler(r.queries)
			actionHandler := action.NewHandler(r.queries)
			scanHandler := scan.NewHandler(r.queries)
			vulnHandler := vuln.NewHandler(r.queries)
			assetHandler := asset.NewHandler(r.queries)
			snapshotsHandler := snapshots.NewHandler(r.queries)
			telemetryHandler := telemetry.NewHandler(r.queries)
			terminal := terminal.NewHandler(r.queries)
			userHandler := user.NewHandler(r.queries)
			envHandler := environment.NewHandler(r.queries)

			subRouter.Use(authHandler.JWTMiddleware)
			subRouter.Use(authHandler.CSRFMiddleware)
			subRouter.Use(roleHandler.PermissionsMiddleware)

			subRouter.Get("/assets", assetHandler.Retrieve)
			subRouter.Get("/assets/min", assetHandler.RetrieveMin)
			subRouter.Get("/assets/{id}", assetHandler.RetrieveData)
			subRouter.Post("/assets/create-snapshot/{assetID}", snapshotsHandler.CreateSnapshot)
			subRouter.Get("/assets/snapshots/{assetID}", snapshotsHandler.ListSnapshots)

			subRouter.Get("/action/retrieve", actionHandler.Retrieve)
			subRouter.Post("/action/create", actionHandler.Create)
			subRouter.Post("/action/run", actionHandler.Run)

			subRouter.Get("/env/retrieve", envHandler.Retrieve)
			subRouter.Post("/env/create", envHandler.Create)
			subRouter.Post("/env/add-asset", envHandler.AddAsset)

			subRouter.Post("/assets/{assetID}/terminal", terminal.Terminal)
			subRouter.Get("/assets/{assetID}/telemetry-usage", telemetryHandler.LatestUsage)

			subRouter.Get("/telemetry-uptime", telemetryHandler.Uptime)
			subRouter.Get("/telemetry-usage-all", telemetryHandler.LatestUsageAll)

			subRouter.Get("/role/retrieve", roleHandler.Retrieve)
			subRouter.Get("/role/retrieve-data/{roleID}", roleHandler.RetrieveData)
			subRouter.Post("/role/create", roleHandler.Create)
			subRouter.Post("/role/update", roleHandler.Update)
			subRouter.Post("/role/delete", roleHandler.DeleteRole)

			subRouter.Post("/scan/launch", scanHandler.Launch)
			subRouter.Post("/scan/update-notes", scanHandler.UpdateNotes)
			subRouter.Get("/scan/retrieve", scanHandler.Retrieve)
			subRouter.Get("/scan/retrieve-scan-parameters", scanHandler.RetrieveScanParameters)

			subRouter.Get("/vuln/retrieve", vulnHandler.Retrieve)
			subRouter.Get("/vuln/retrieve-data/{vulnID}", vulnHandler.RetrieveData)
			subRouter.Get("/vuln/retrieve-scan/{scanID}", vulnHandler.RetrieveScan)

			subRouter.Post("/user/create", userHandler.CreateUser)
			subRouter.Get("/user/retrieve", userHandler.Retrieve)
			subRouter.Post("/user/delete", userHandler.DeleteUser)
			subRouter.Post("/user/update", userHandler.UpdateUser)
		})

		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.Middleware(rate.Every(1*time.Second), 10))

			authHandler := auth.NewHandler(r.queries)

			subRouter.Use(authHandler.JWTMiddleware)
			subRouter.Use(authHandler.CSRFMiddleware)

			subRouter.Post("/auth/logout", authHandler.Logout)
			subRouter.Get("/auth/validate", func(w http.ResponseWriter, r *http.Request) {
				account := auth.GetClaims(r.Context())
				val, err := account.AccountID.Value()
				if err != nil {
					response.RespondWithError(w, r, http.StatusInternalServerError, "Failed to parse UUID", err)
					return
				}
				response.RespondWithJSON(w, http.StatusOK,
					map[string]string{"accountId": val.(string), "accountType": account.AccountType, "accountUser": account.AccountUser})
			})
		})
	})

	return &r
}

func (r *Router) GetRouter() *chi.Mux {
	return r.router
}
