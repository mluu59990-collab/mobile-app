package main

import (
	"context"
	"fmt"
	"log"
	"time"

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
	router := gin.Default()
	if err := router.SetTrustedProxies(nil); err != nil {
		return err
	}
	handler.RegisterAuth(router, authService, tokens)
	handler.RegisterExpenseRoutes(router, service.NewExpenseService(repository.NewExpenses(db)), tokens)
	handler.RegisterReportRoutes(router, service.NewMonthlyReportService(repository.NewReports(db)), tokens)
	router.GET("/api/v1/health", handler.Health)

	if err := router.Run(cfg.HTTPAddr); err != nil {
		return fmt.Errorf("start server: %w", err)
	}
	return nil
}
