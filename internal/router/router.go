package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"3za-digital/infrastructure/database"
	permissioncache "3za-digital/internal/cache/permission"
	domaincatalog "3za-digital/internal/domain/catalog"
	domainprovider "3za-digital/internal/domain/provider"
	appConfigHandler "3za-digital/internal/handlers/http/appconfig"
	auditHandler "3za-digital/internal/handlers/http/audit"
	dashboardHandler "3za-digital/internal/handlers/http/dashboard"
	locationHandler "3za-digital/internal/handlers/http/location"
	menuHandler "3za-digital/internal/handlers/http/menu"
	permissionHandler "3za-digital/internal/handlers/http/permission"
	providerHandler "3za-digital/internal/handlers/http/provider"
	roleHandler "3za-digital/internal/handlers/http/role"
	sessionHandler "3za-digital/internal/handlers/http/session"
	smmHandler "3za-digital/internal/handlers/http/smm"
	userHandler "3za-digital/internal/handlers/http/user"
	walletHandler "3za-digital/internal/handlers/http/wallet"
	"3za-digital/internal/integrations/h2h"
	"3za-digital/internal/integrations/qrisly"
	interfaceprovider "3za-digital/internal/interfaces/provider"
	interfacesession "3za-digital/internal/interfaces/session"
	appConfigRepo "3za-digital/internal/repositories/appconfig"
	auditRepo "3za-digital/internal/repositories/audit"
	authRepo "3za-digital/internal/repositories/auth"
	catalogRepo "3za-digital/internal/repositories/catalog"
	dashboardRepo "3za-digital/internal/repositories/dashboard"
	locationRepo "3za-digital/internal/repositories/location"
	menuRepo "3za-digital/internal/repositories/menu"
	orderRepo "3za-digital/internal/repositories/order"
	otpRepo "3za-digital/internal/repositories/otp"
	permissionRepo "3za-digital/internal/repositories/permission"
	providerRepo "3za-digital/internal/repositories/provider"
	resetRepo "3za-digital/internal/repositories/reset"
	roleRepo "3za-digital/internal/repositories/role"
	sessionRepo "3za-digital/internal/repositories/session"
	userRepo "3za-digital/internal/repositories/user"
	walletRepo "3za-digital/internal/repositories/wallet"
	appConfigSvc "3za-digital/internal/services/appconfig"
	auditSvc "3za-digital/internal/services/audit"
	catalogSvc "3za-digital/internal/services/catalog"
	dashboardSvc "3za-digital/internal/services/dashboard"
	locationSvc "3za-digital/internal/services/location"
	menuSvc "3za-digital/internal/services/menu"
	orderSvc "3za-digital/internal/services/order"
	otpSvc "3za-digital/internal/services/otp"
	permissionSvc "3za-digital/internal/services/permission"
	providerSvc "3za-digital/internal/services/provider"
	resetSvc "3za-digital/internal/services/reset"
	roleSvc "3za-digital/internal/services/role"
	sessionSvc "3za-digital/internal/services/session"
	userSvc "3za-digital/internal/services/user"
	walletSvc "3za-digital/internal/services/wallet"
	"3za-digital/middlewares"
	"3za-digital/pkg/config"
	"3za-digital/pkg/logger"
	"3za-digital/pkg/mailer"
	"3za-digital/pkg/security"
	"3za-digital/utils"
)

type Routes struct {
	App *gin.Engine
	DB  *gorm.DB
}

func NewRoutes() *Routes {
	app := gin.Default()

	app.Use(middlewares.CORS())
	app.Use(gin.CustomRecovery(middlewares.ErrorHandler))
	app.Use(middlewares.SetContextId())
	app.Use(middlewares.RequestLogger())

	app.GET("/healthcheck", func(ctx *gin.Context) {
		logger.WriteLogWithContext(ctx, logger.LogLevelDebug, "ClientIP: "+ctx.ClientIP())
		ctx.JSON(http.StatusOK, gin.H{
			"message": "OK!!",
		})
	})

	return &Routes{
		App: app,
	}
}

func (r *Routes) UserRoutes() {
	blacklistRepo := authRepo.NewBlacklistRepo(r.DB)
	repo := userRepo.NewUserRepo(r.DB)
	rRepo := roleRepo.NewRoleRepo(r.DB)
	pRepo := permissionRepo.NewPermissionRepo(r.DB)
	redisClient := database.GetRedisClient()
	permissionInvalidator := permissioncache.NewInvalidator(redisClient)
	uc := userSvc.NewUserService(repo, blacklistRepo, rRepo, pRepo, permissionInvalidator)
	var userSessionSvc interfacesession.ServiceSessionInterface
	repoAudit := auditRepo.NewAuditRepo(r.DB)
	svcAudit := auditSvc.NewAuditService(repoAudit)
	repoAppConfig := appConfigRepo.NewAppConfigRepo(r.DB)
	svcAppConfig := appConfigSvc.NewAppConfigService(repoAppConfig)

	// Setup login limiter if Redis is available
	var loginLimiter security.LoginLimiter
	var registerOTPService = otpSvc.NewOTPService(nil, nil, config.LoadOTPConfig())
	var passwordResetService = resetSvc.NewPasswordResetService(nil, nil, config.LoadPasswordResetConfig())
	if redisClient != nil {
		loginLimiter = security.NewRedisLoginLimiter(
			redisClient,
			utils.GetEnv("LOGIN_ATTEMPT_LIMIT", 5),
			time.Duration(utils.GetEnv("LOGIN_ATTEMPT_WINDOW_SECONDS", 60))*time.Second,
			time.Duration(utils.GetEnv("LOGIN_BLOCK_DURATION_SECONDS", 300))*time.Second,
		)

		sRepo := sessionRepo.NewSessionRepository(redisClient)
		userSessionSvc = sessionSvc.NewSessionService(sRepo)

		sender, err := mailer.NewBrevoSenderFromEnv()
		if err != nil {
			logger.WriteLog(logger.LogLevelWarn, "Email sender not configured: ", err)
		} else {
			registerOTPService = otpSvc.NewOTPService(otpRepo.NewOTPRepository(redisClient), sender, config.LoadOTPConfig())
			passwordResetService = resetSvc.NewPasswordResetService(resetRepo.NewPasswordResetRepository(redisClient), sender, config.LoadPasswordResetConfig())
		}
	}

	h := userHandler.NewUserHandler(uc, blacklistRepo, userSessionSvc, loginLimiter, svcAudit, svcAppConfig, registerOTPService, passwordResetService)
	mdw := middlewares.NewMiddleware(blacklistRepo, pRepo)

	// Setup register rate limiter
	registerLimit := utils.GetEnv("REGISTER_RATE_LIMIT", 5)
	registerWindowSeconds := utils.GetEnv("REGISTER_RATE_WINDOW_SECONDS", 60)
	if registerWindowSeconds <= 0 {
		registerWindowSeconds = 60
	}
	registerLimiter := middlewares.IPRateLimitMiddleware(
		redisClient,
		"user_register",
		registerLimit,
		time.Duration(registerWindowSeconds)*time.Second,
	)

	user := r.App.Group("/api/user")
	{
		user.GET("/register/status", h.GetRegisterStatus)
		user.POST("/register", registerLimiter, h.Register)
		user.POST("/register/otp/send", registerLimiter, h.SendRegisterOTP)
		user.POST("/login", h.Login)
		user.POST("/google/login", h.GoogleLogin)
		user.POST("/refresh-token", h.RefreshToken)
		user.POST("/forgot-password", h.ForgotPassword)
		user.POST("/reset-password", h.ResetPassword)

		userPriv := user.Group("").Use(mdw.AuthMiddleware())
		{
			userPriv.POST("/logout", h.Logout)
			userPriv.GET("", h.GetUserByAuth)
			userPriv.GET("/:id", mdw.PermissionMiddleware("users", "view"), h.GetUserById)
			userPriv.POST("/:id/impersonate", mdw.PermissionMiddleware("users", "impersonate"), h.ImpersonateUser)
			userPriv.POST("/stop-impersonation", h.StopImpersonation)
			userPriv.PUT("", h.Update)
			userPriv.PUT("/:id", mdw.PermissionMiddleware("users", "update"), h.UpdateUserById)
			userPriv.PUT("/change/password", h.ChangePassword)
			userPriv.DELETE("", h.Delete)
			userPriv.DELETE("/:id", mdw.PermissionMiddleware("users", "delete"), h.DeleteUserById)

			// Admin create user endpoint (with role selection)
			userPriv.POST("", mdw.PermissionMiddleware("users", "create"), h.AdminCreateUser)
		}
	}

	r.App.GET("/api/users", mdw.AuthMiddleware(), mdw.PermissionMiddleware("users", "list"), h.GetAllUsers)
}

func (r *Routes) RoleRoutes() {
	repoRole := roleRepo.NewRoleRepo(r.DB)
	repoPermission := permissionRepo.NewPermissionRepo(r.DB)
	repoMenu := menuRepo.NewMenuRepo(r.DB)
	permissionInvalidator := permissioncache.NewInvalidator(database.GetRedisClient())
	svc := roleSvc.NewRoleService(repoRole, repoPermission, repoMenu, permissionInvalidator)
	repoAudit := auditRepo.NewAuditRepo(r.DB)
	svcAudit := auditSvc.NewAuditService(repoAudit)
	h := roleHandler.NewRoleHandler(svc, svcAudit)
	blacklistRepo := authRepo.NewBlacklistRepo(r.DB)
	mdw := middlewares.NewMiddleware(blacklistRepo, repoPermission)

	// List endpoints
	r.App.GET("/api/roles", mdw.AuthMiddleware(), mdw.PermissionMiddleware("roles", "list"), h.GetAll)

	// CRUD endpoints
	role := r.App.Group("/api/role").Use(mdw.AuthMiddleware())
	{
		role.POST("", mdw.PermissionMiddleware("roles", "create"), h.Create)
		role.GET("/:id", mdw.PermissionMiddleware("roles", "view"), h.GetByID)
		role.PUT("/:id", mdw.PermissionMiddleware("roles", "update"), h.Update)
		role.DELETE("/:id", mdw.PermissionMiddleware("roles", "delete"), h.Delete)

		// Permission and menu assignment
		role.POST("/:id/permissions", mdw.PermissionMiddleware("roles", "assign_permissions"), h.AssignPermissions)
	}
}

func (r *Routes) PermissionRoutes() {
	repo := permissionRepo.NewPermissionRepo(r.DB)
	permissionInvalidator := permissioncache.NewInvalidator(database.GetRedisClient())
	svc := permissionSvc.NewPermissionService(repo, permissionInvalidator)
	repoAudit := auditRepo.NewAuditRepo(r.DB)
	svcAudit := auditSvc.NewAuditService(repoAudit)
	h := permissionHandler.NewPermissionHandler(svc, svcAudit)
	blacklistRepo := authRepo.NewBlacklistRepo(r.DB)
	mdw := middlewares.NewMiddleware(blacklistRepo, repo)

	// List endpoints
	r.App.GET("/api/permissions", mdw.AuthMiddleware(), mdw.PermissionMiddleware("permissions", "list"), h.GetAll)

	// Get current user's permissions
	r.App.GET("/api/permissions/me", mdw.AuthMiddleware(), h.GetUserPermissions)

	// CRUD endpoints
	permission := r.App.Group("/api/permission").Use(mdw.AuthMiddleware())
	{
		permission.POST("", mdw.PermissionMiddleware("permissions", "create"), h.Create)
		permission.GET("/:id", mdw.PermissionMiddleware("permissions", "view"), h.GetByID)
		permission.PUT("/:id", mdw.PermissionMiddleware("permissions", "update"), h.Update)
		permission.DELETE("/:id", mdw.PermissionMiddleware("permissions", "delete"), h.Delete)
	}
}

func (r *Routes) MenuRoutes() {
	repo := menuRepo.NewMenuRepo(r.DB)
	pRepo := permissionRepo.NewPermissionRepo(r.DB)
	svc := menuSvc.NewMenuService(repo, pRepo)
	repoAudit := auditRepo.NewAuditRepo(r.DB)
	svcAudit := auditSvc.NewAuditService(repoAudit)
	h := menuHandler.NewMenuHandler(svc, svcAudit)
	blacklistRepo := authRepo.NewBlacklistRepo(r.DB)
	mdw := middlewares.NewMiddleware(blacklistRepo, pRepo)

	// Public endpoints for authenticated users
	r.App.GET("/api/menus/active", mdw.AuthMiddleware(), h.GetActiveMenus)
	r.App.GET("/api/menus/me", mdw.AuthMiddleware(), h.GetUserMenus)

	// List endpoints
	r.App.GET("/api/menus", mdw.AuthMiddleware(), mdw.PermissionMiddleware("menus", "list"), h.GetAll)

	// CRUD endpoints
	menu := r.App.Group("/api/menu").Use(mdw.AuthMiddleware())
	{
		menu.GET("/:id", mdw.PermissionMiddleware("menus", "view"), h.GetByID)
		menu.PUT("/:id", mdw.PermissionMiddleware("menus", "update"), h.Update)
	}
}

func (r *Routes) AppConfigRoutes() {
	repo := appConfigRepo.NewAppConfigRepo(r.DB)
	svc := appConfigSvc.NewAppConfigService(repo)
	repoAudit := auditRepo.NewAuditRepo(r.DB)
	svcAudit := auditSvc.NewAuditService(repoAudit)
	h := appConfigHandler.NewAppConfigHandler(svc, svcAudit)
	blacklistRepo := authRepo.NewBlacklistRepo(r.DB)
	pRepo := permissionRepo.NewPermissionRepo(r.DB)
	mdw := middlewares.NewMiddleware(blacklistRepo, pRepo)

	r.App.GET("/api/configs", mdw.AuthMiddleware(), mdw.PermissionMiddleware("configs", "list"), h.GetAll)

	appConfig := r.App.Group("/api/config").Use(mdw.AuthMiddleware())
	{
		appConfig.GET("/:id", mdw.PermissionMiddleware("configs", "view"), h.GetByID)
		appConfig.PUT("/:id", mdw.PermissionMiddleware("configs", "update"), h.Update)
	}
}

func (r *Routes) AuditRoutes() {
	repo := auditRepo.NewAuditRepo(r.DB)
	svc := auditSvc.NewAuditService(repo)
	h := auditHandler.NewAuditHandler(svc)
	blacklistRepo := authRepo.NewBlacklistRepo(r.DB)
	pRepo := permissionRepo.NewPermissionRepo(r.DB)
	mdw := middlewares.NewMiddleware(blacklistRepo, pRepo)

	r.App.GET("/api/audits", mdw.AuthMiddleware(), mdw.PermissionMiddleware("audits", "list"), h.GetAll)

	audit := r.App.Group("/api/audit").Use(mdw.AuthMiddleware())
	{
		audit.GET("/:id", mdw.PermissionMiddleware("audits", "view"), h.GetByID)
	}
}

func (r *Routes) SessionRoutes() {
	redisClient := database.GetRedisClient()
	if redisClient == nil {
		logger.WriteLog(logger.LogLevelDebug, "Redis not available, session routes will not be registered")
		return
	}

	repo := sessionRepo.NewSessionRepository(redisClient)
	svc := sessionSvc.NewSessionService(repo)
	repoAudit := auditRepo.NewAuditRepo(r.DB)
	svcAudit := auditSvc.NewAuditService(repoAudit)
	h := sessionHandler.NewSessionHandler(svc, svcAudit)
	blacklistRepo := authRepo.NewBlacklistRepo(r.DB)
	pRepo := permissionRepo.NewPermissionRepo(r.DB)
	mdw := middlewares.NewMiddleware(blacklistRepo, pRepo)

	// Session management endpoints (authenticated users only)
	sessionGroup := r.App.Group("/api/user").Use(mdw.AuthMiddleware())
	{
		sessionGroup.GET("/sessions", h.GetActiveSessions)
		sessionGroup.DELETE("/session/:session_id", h.RevokeSession)
		sessionGroup.POST("/sessions/revoke-others", h.RevokeAllOtherSessions)
	}

	logger.WriteLog(logger.LogLevelInfo, "Session management routes registered")
}

func (r *Routes) LocationRoutes() {
	repo := locationRepo.NewLocationRepo(r.DB)
	svc := locationSvc.NewLocationService(repo, database.GetRedisClient())
	repoAudit := auditRepo.NewAuditRepo(r.DB)
	svcAudit := auditSvc.NewAuditService(repoAudit)
	h := locationHandler.NewLocationHandler(svc, svcAudit)
	blacklistRepo := authRepo.NewBlacklistRepo(r.DB)
	pRepo := permissionRepo.NewPermissionRepo(r.DB)
	mdw := middlewares.NewMiddleware(blacklistRepo, pRepo)

	location := r.App.Group("/api/location")
	{
		location.GET("/province", h.GetProvince)
		location.GET("/city", h.GetCity)
		location.GET("/district", h.GetDistrict)
		location.GET("/village", h.GetVillage)
	}

	locationPriv := r.App.Group("/api/location").Use(mdw.AuthMiddleware())
	{
		locationPriv.POST("/sync", mdw.PermissionMiddleware("locations", "sync"), h.Sync)
		locationPriv.GET("/sync/:id", mdw.PermissionMiddleware("locations", "sync"), h.GetSyncJob)
	}
}

func (r *Routes) SMMRoutes() {
	repoCatalog := catalogRepo.NewCatalogRepo(r.DB)
	repoOrder := orderRepo.NewOrderRepo(r.DB)
	repoProvider := providerRepo.NewProviderRepo(r.DB)
	repoAppConfig := appConfigRepo.NewAppConfigRepo(r.DB)
	repoAudit := auditRepo.NewAuditRepo(r.DB)
	svcAppConfig := appConfigSvc.NewAppConfigService(repoAppConfig)
	svcAudit := auditSvc.NewAuditService(repoAudit)
	providerFactory := func() (interfaceprovider.Client, error) {
		return newObservedH2HClient(repoProvider)
	}
	redisClient := database.GetRedisClient()
	syncStateStore := catalogRepo.NewCatalogSyncStateStore(r.DB, redisClient)
	svcCatalog := catalogSvc.NewCatalogService(repoCatalog, providerFactory, svcAppConfig).WithSyncStateStore(syncStateStore)
	svcOrder := orderSvc.NewOrderService(repoOrder, providerFactory, svcAppConfig).WithAuditService(svcAudit).WithCatalogService(svcCatalog)
	h := smmHandler.NewSMMHandler(svcCatalog, svcOrder, svcAudit)

	blacklistRepo := authRepo.NewBlacklistRepo(r.DB)
	pRepo := permissionRepo.NewPermissionRepo(r.DB)
	mdw := middlewares.NewMiddleware(blacklistRepo, pRepo)
	orderLimiter := middlewares.IPRateLimitMiddleware(
		redisClient,
		"smm_order_create",
		utils.GetEnv("SMM_ORDER_RATE_LIMIT", 20),
		time.Duration(utils.GetEnv("SMM_ORDER_RATE_WINDOW_SECONDS", 60))*time.Second,
	)

	smm := r.App.Group("/api/smm").Use(mdw.AuthMiddleware())
	{
		smm.GET("/services", mdw.PermissionMiddleware("smm_services", "list"), h.GetServices)
		smm.POST("/services/sync", mdw.PermissionMiddleware("smm_services", "sync"), h.SyncServices)
		smm.GET("/orders", mdw.PermissionMiddleware("smm_orders", "list"), h.GetOrders)
		smm.POST("/orders", orderLimiter, mdw.PermissionMiddleware("smm_orders", "create"), h.CreateOrder)
		smm.GET("/orders/:id", mdw.PermissionMiddleware("smm_orders", "view"), h.GetOrderByID)
		smm.GET("/orders/:id/status-logs", mdw.PermissionMiddleware("smm_orders", "view"), h.GetOrderStatusLogs)
		smm.POST("/orders/:id/refresh-status", mdw.PermissionMiddleware("smm_orders", "refresh_status"), h.RefreshOrderStatus)
	}
}

func (r *Routes) WalletRoutes() {
	repoWallet := walletRepo.NewWalletRepo(r.DB)
	repoProvider := providerRepo.NewProviderRepo(r.DB)
	repoAppConfig := appConfigRepo.NewAppConfigRepo(r.DB)
	svcProvider := providerSvc.NewProviderService(repoProvider, func() (interfaceprovider.Client, error) {
		return newObservedH2HClient(repoProvider)
	})
	svcAppConfig := appConfigSvc.NewAppConfigService(repoAppConfig)
	svcWallet := walletSvc.NewWalletService(repoWallet, svcProvider).WithConfigService(svcAppConfig)
	if qrisClient, err := qrisly.NewClient(qrisly.LoadConfigFromEnv()); err != nil {
		logger.WriteLog(logger.LogLevelWarn, "QRISLY client not configured: ", err)
	} else {
		svcWallet.WithQRISProvider(qrisClient)
	}
	repoAudit := auditRepo.NewAuditRepo(r.DB)
	svcAudit := auditSvc.NewAuditService(repoAudit)
	h := walletHandler.NewWalletHandler(svcWallet, svcAudit)

	blacklistRepo := authRepo.NewBlacklistRepo(r.DB)
	pRepo := permissionRepo.NewPermissionRepo(r.DB)
	mdw := middlewares.NewMiddleware(blacklistRepo, pRepo)
	redisClient := database.GetRedisClient()
	depositLimiter := middlewares.IPRateLimitMiddleware(
		redisClient,
		"deposit_create",
		utils.GetEnv("DEPOSIT_CREATE_RATE_LIMIT", 10),
		time.Duration(utils.GetEnv("DEPOSIT_CREATE_RATE_WINDOW_SECONDS", 60))*time.Second,
	)
	webhookLimiter := middlewares.IPRateLimitMiddleware(
		redisClient,
		"payment_webhook",
		utils.GetEnv("PAYMENT_WEBHOOK_RATE_LIMIT", 120),
		time.Duration(utils.GetEnv("PAYMENT_WEBHOOK_RATE_WINDOW_SECONDS", 60))*time.Second,
	)

	wallet := r.App.Group("/api/wallet").Use(mdw.AuthMiddleware())
	{
		wallet.GET("/me", mdw.PermissionMiddleware("wallet", "view"), h.GetMyWallet)
		wallet.GET("/transactions", mdw.PermissionMiddleware("wallet_transactions", "list"), h.GetMyTransactions)
	}

	adminWallet := r.App.Group("/api/admin/wallets").Use(mdw.AuthMiddleware())
	{
		adminWallet.GET("", mdw.PermissionMiddleware("wallets", "list"), h.GetWallets)
		adminWallet.POST("/:user_id/topup", mdw.PermissionMiddleware("wallets", "topup"), h.AdminTopup)
		adminWallet.POST("/:user_id/adjust", mdw.PermissionMiddleware("wallets", "adjust"), h.AdminAdjust)
	}

	adminDeposits := r.App.Group("/api/admin/deposits").Use(mdw.AuthMiddleware())
	{
		adminDeposits.GET("", mdw.PermissionMiddleware("admin_deposits", "list"), h.GetDeposits)
		adminDeposits.GET("/:id", mdw.PermissionMiddleware("admin_deposits", "view"), h.GetDepositByID)
		adminDeposits.POST("/:id/status", mdw.AnyPermissionMiddleware(
			middlewares.RequiredPermission{Resource: "admin_deposits", Action: "approve"},
			middlewares.RequiredPermission{Resource: "admin_deposits", Action: "cancel"},
		), h.AdminUpdateDepositStatus)
	}

	deposits := r.App.Group("/api/deposits").Use(mdw.AuthMiddleware())
	{
		deposits.POST("", depositLimiter, mdw.PermissionMiddleware("deposits", "create"), h.CreateDeposit)
		deposits.GET("", mdw.PermissionMiddleware("deposits", "list"), h.GetMyDeposits)
		deposits.GET("/settings", mdw.PermissionMiddleware("deposits", "create"), h.GetDepositSettings)
		deposits.GET("/:id", mdw.PermissionMiddleware("deposits", "view"), h.GetMyDepositByID)
	}

	r.App.POST("/api/webhooks/payments/:provider", webhookLimiter, h.PaymentWebhook)
}

func (r *Routes) ProviderRoutes() {
	repoProvider := providerRepo.NewProviderRepo(r.DB)
	svcProvider := providerSvc.NewProviderService(repoProvider, func() (interfaceprovider.Client, error) {
		return newObservedH2HClient(repoProvider)
	})
	h := providerHandler.NewProviderHandler(svcProvider)

	blacklistRepo := authRepo.NewBlacklistRepo(r.DB)
	pRepo := permissionRepo.NewPermissionRepo(r.DB)
	mdw := middlewares.NewMiddleware(blacklistRepo, pRepo)

	provider := r.App.Group("/api/provider").Use(mdw.AuthMiddleware())
	{
		provider.GET("/h2h/balance", mdw.PermissionMiddleware("provider_balance", "view"), h.GetH2HBalance)
		provider.GET("/api-logs", mdw.PermissionMiddleware("provider_api_logs", "list"), h.GetAPILogs)
	}
}

func newObservedH2HClient(repo interfaceprovider.RepoProviderInterface) (interfaceprovider.Client, error) {
	client, err := h2h.NewClient(h2h.LoadConfigFromEnv())
	if err != nil {
		return nil, err
	}
	client.SetObserver(func(ctx context.Context, event h2h.RequestLog) {
		if err := repo.StoreAPILog(ctx, domainprovider.APILog{
			Provider:       domaincatalog.ProviderH2H,
			ProductType:    event.ProductType,
			Endpoint:       event.Endpoint,
			RequestRef:     event.RequestRef,
			ResponseStatus: new(event.ResponseStatus),
			ResponseBody:   event.ResponseBody,
			DurationMS:     new(event.DurationMS),
			ErrorMessage:   event.ErrorMessage,
		}); err != nil {
			logger.WriteLog(logger.LogLevelWarn, "Failed to store provider API log: ", err)
		}
	})
	return client, nil
}

func (r *Routes) DashboardRoutes() {
	repoDashboard := dashboardRepo.NewDashboardRepo(r.DB)
	svcDashboard := dashboardSvc.NewDashboardService(repoDashboard)
	h := dashboardHandler.NewDashboardHandler(svcDashboard)

	blacklistRepo := authRepo.NewBlacklistRepo(r.DB)
	pRepo := permissionRepo.NewPermissionRepo(r.DB)
	mdw := middlewares.NewMiddleware(blacklistRepo, pRepo)

	r.App.GET("/api/dashboard/summary", mdw.AuthMiddleware(), mdw.PermissionMiddleware("dashboard", "view"), h.GetSummary)
}
