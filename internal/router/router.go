package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/SyntinelNyx/syntinel-server/internal/asset"
	"github.com/SyntinelNyx/syntinel-server/internal/auth"
	"github.com/SyntinelNyx/syntinel-server/internal/database/query"
	"github.com/SyntinelNyx/syntinel-server/internal/limiter"
	"github.com/SyntinelNyx/syntinel-server/internal/logger"
	"github.com/SyntinelNyx/syntinel-server/internal/response"
	"github.com/SyntinelNyx/syntinel-server/internal/role"
	"github.com/SyntinelNyx/syntinel-server/internal/scan"
	"github.com/SyntinelNyx/syntinel-server/internal/shell"
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
			scanHandler := scan.NewHandler(r.queries)
			vulnHandler := vuln.NewHandler(r.queries)
			assetHandler := asset.NewHandler(r.queries)
			shellHandler := shell.NewHandler(r.queries)

			subRouter.Use(authHandler.JWTMiddleware)
			subRouter.Use(authHandler.CSRFMiddleware)

			subRouter.Get("/assets", assetHandler.Retrieve)
			subRouter.Get("/assets/{id}", assetHandler.RetrieveData)

			subRouter.Post("/assets/{assetID}/shell", shellHandler.Shell)

			subRouter.Post("/role/retrieve", roleHandler.Retrieve)
			subRouter.Post("/role/create", roleHandler.Create)
			subRouter.Post("/role/delete", roleHandler.DeleteRole)
			subRouter.Post("/scan/launch", scanHandler.Launch)
			subRouter.Get("/scan/retrieve", scanHandler.Retrieve)
			subRouter.Get("/vuln/retrieve", vulnHandler.Retrieve)
			subRouter.Post("/vuln/retrieve-data", vulnHandler.RetrieveData)
		})

		apiRouter.Group(func(subRouter chi.Router) {
			subRouter.Use(r.rateLimiter.Middleware(rate.Every(1*time.Second), 10))

			authHandler := auth.NewHandler(r.queries)
			subRouter.Use(authHandler.JWTMiddleware)
			subRouter.Use(authHandler.CSRFMiddleware)

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
			subRouter.Post("/auth/logout", authHandler.Logout)
		})
	})

	return &r
}

func (r *Router) GetRouter() *chi.Mux {
	return r.router
}


useEffect(() => {
    // Initialize terminal
    if (typeof window !== 'undefined' && terminalRef.current) {
      // Clean up previous terminal instance if it exists
      if (xtermRef.current) {
        xtermRef.current.dispose();
      }
      
      // Create new terminal
      const term = new XTerm({
        cursorBlink: true,
        fontSize: 14,
        theme: {
          background: '#1a1b26',
          foreground: '#c0caf5',
          cursor: '#c0caf5',
        }
      });
      
      xtermRef.current = term;
      term.open(terminalRef.current);
      term.write('Connected to asset: \x1B[1;3;32m' + slug + '\x1B[0m\r\n$ ');
      
      term.onData((data) => {
        // Handle backspace
        if (data === '\x7F') {
          if (commandBuffer.length > 0) {
            term.write('\b \b'); // Erase character
            setCommandBuffer(prev => prev.substring(0, prev.length - 1));
          }
          return;
        }
        
        // Echo back input for visual feedback
        term.write(data);
        
        // If enter key is pressed, process the command
        if (data === '\r') {
          // Store command to be processed
          const command = commandBuffer.trim();

          console.log('Command entered:', command);
          
          // Clear the buffer for next command
          setCommandBuffer('');
          
          // Execute command
          if (command) {
            term.write('\r\n');
            executeCommand(command, term);
          } else {
            // Just show a new prompt for empty commands
            term.write('\r\n$ ');
          }
        } else {
          // Add to command buffer
          setCommandBuffer(prev => prev + data);
        }
      });
    }
    const executeCommand = async (command: string, term: XTerm) => {
      try {
        term.write(`Executing command: ${command}\r\n`);
        
        // Make API request to execute command on the asset
        const response = await apiFetch(`/assets/${slug}/shell`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ command }),
        });
        
        if (!response.ok) {
          throw new Error(`Command failed with status: ${response.status}`);
        }
        
        const result = await response.json();
        
        // Display command output
        term.write(`${result.output}\r\n`);
      } catch (error) {
        // Handle errors
        console.error('Command execution error:', error);
        term.write(`\x1B[1;3;31mError: ${error instanceof Error ? error.message : 'Unknown error'}\x1B[0m\r\n`);
      } finally {
        // Show prompt for next command
        term.write('$ ');
      }
    };

    return () => {
      // Clean up on unmount
      if (xtermRef.current) {
        xtermRef.current.dispose();
      }
    };
  }, [slug]);