package main

import (
	"api_gateway/internal/config"
	v1 "api_gateway/internal/transport/v1"
	"api_gateway/pkg/logger"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	// Добавляем логгер в контекст
	ctx := context.Background()
	ctx, err := logger.New(ctx)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "logger error", zap.Error(err))
	}

	// Загрузка конфигурации
	cfg, err := config.ParseConfig("./config/config.env")
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Error(ctx, "cfg error", zap.Error(err))
	}

	server := v1.NewServer(cfg.Server.Port)
	// Регистрация обработчиков
	server.RegisterHandler(ctx, cfg)
	wg := sync.WaitGroup{}
	wg.Add(1)
	// Запуск HTTP сервера
	go func() {
		defer wg.Done()
		log.Printf("server starting on localhost:%d", cfg.Server.Port)
		if err := server.Start(); !errors.Is(err, http.ErrServerClosed) {
			logger.GetLoggerFromCtx(ctx).Error(ctx, "server error", zap.Error(err))
		}
	}()

	// Ожидание сигнала завершения работы
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan
	shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.TimeOut)
	defer cancel()
	if err := server.Stop(shutdownCtx); err != nil {
		log.Printf("failed to stop http server")
	}
	wg.Wait()
}
