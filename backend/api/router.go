// Purpose: Gin router wiring (CORS, public/protected routes).

package api

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"radiokpowka/backend/auth"
	"radiokpowka/backend/config"
	"radiokpowka/backend/player"
	"radiokpowka/backend/websocket"
	"radiokpowka/backend/youtube"
)

type RouterDeps struct {
	Cfg    config.Config
	DB     *gorm.DB
	Player *player.Controller
	Hub    *websocket.Hub
	YT     *youtube.Client
}

func NewRouter(cfg config.Config, database *gorm.DB) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())

	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-RK-Webhook-Secret"},
		AllowCredentials: false,
		MaxAge:           3600,
	}))

	hub := websocket.NewHub()
	go hub.Run()

	yt := youtube.NewClient(youtube.Config{
		YTDLPPath:          cfg.YTDLPPath,
		FFMPEGPath:         cfg.FFMPEGPath,
		CookiesFromBrowser: cfg.YTDLPCookiesFromBrowser,
	})

	ctrl := player.NewController(player.ControllerDeps{
		DB:  database,
		Hub: hub,
		YT:  yt,
	})

	deps := RouterDeps{
		Cfg:    cfg,
		DB:     database,
		Player: ctrl,
		Hub:    hub,
		YT:     yt,
	}

	// Public
	r.GET("/api/health", HealthHandler)
	r.POST("/api/auth/login", LoginHandler(deps))

	r.GET("/api/player/state", GetPlayerStateHandler(deps))
	r.POST("/api/playlist/add", auth.OptionalJWT(cfg.JWTSecret), PlaylistAddHandler(deps))
	r.GET("/api/playlist", PlaylistListHandler(deps))

	r.GET("/stream", StreamHandler(deps))
	r.GET("/ws", WSHandler(deps))

	// Webhook (protected by shared secret header if WEBHOOK_SECRET set)
	r.POST("/api/webhook/donation", DonationWebhookHandler(deps))

	// Protected
	protected := r.Group("/api")
	protected.Use(auth.JWTMiddleware(cfg.JWTSecret))
	protected.GET("/me", MeHandler(deps))

	owner := protected.Group("/")
	owner.Use(auth.RequireRole("owner"))
	owner.POST("/player/play", PlayerPlayHandler(deps))
	owner.POST("/player/pause", PlayerPauseHandler(deps))
	owner.POST("/player/next", PlayerNextHandler(deps))
	owner.POST("/player/prev", PlayerPrevHandler(deps))
	owner.POST("/player/volume", PlayerVolumeHandler(deps))

	owner.POST("/integrations/donationalerts/connect", DonAlertsConnectHandler(deps))
	owner.POST("/integrations/donx/connect", DonXConnectHandler(deps))

	return r
}
