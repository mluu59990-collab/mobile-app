package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"expense-app/backend/internal/apierror"
	"expense-app/backend/internal/auth"
	"expense-app/backend/internal/config"
	"expense-app/backend/internal/handler"
	"expense-app/backend/internal/repository"
	"expense-app/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	authConfig, err := config.LoadAuth()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	db, err := repository.OpenDatabase(ctx, cfg.DatabaseURL)
	cancel()
	if err != nil {
		return fmt.Errorf("database startup: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("close database: %v", err)
		}
	}()
	tokens := auth.Tokens{Config: authConfig}
	authService, err := service.NewAuth(repository.NewAuth(db), tokens, service.GoogleIDTokenVerifier{})
	if err != nil {
		return err
	}
	authService.SetFacebookVerifier(service.FacebookGraphVerifier{
		AppID: authConfig.FacebookAppID, AppSecret: authConfig.FacebookAppSecret,
	})
	router, err := newRouter(
		authService,
		service.NewExpenseService(repository.NewExpenses(db)),
		service.NewMonthlyReportService(repository.NewReports(db)),
		tokens,
	)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	return nil
}

func newRouter(authService *service.Auth, expenseService *service.ExpenseService, reportService *service.MonthlyReportService, tokens auth.Tokens) (*gin.Engine, error) {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.HandleMethodNotAllowed = true
	router.NoRoute(func(c *gin.Context) {
		apierror.Write(c, http.StatusNotFound, "route_not_found", "Route not found")
	})
	router.NoMethod(func(c *gin.Context) {
		apierror.Write(c, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
	})
	if err := router.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	handler.RegisterAuth(router, authService, tokens)
	handler.RegisterExpenseRoutes(router, expenseService, tokens)
	handler.RegisterReportRoutes(router, reportService, tokens)
	router.GET("/api/v1/health", handler.Health)
	return router, nil
}
