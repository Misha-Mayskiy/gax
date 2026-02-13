package logger

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	LogKey    = "logger"
	RequestID = "request_id"
)

type Logger struct {
	l *zap.Logger
}

func New(ctx context.Context) (context.Context, error) {
	l, err := zap.NewProduction()
	if err != nil {
		return ctx, err
	}
	logger := &Logger{l: l}
	ctx = context.WithValue(ctx, LogKey, logger)
	return ctx, nil
}
func GetLoggerFromCtx(ctx context.Context) *Logger {
	loggerVal := ctx.Value(LogKey)
	if loggerVal == nil {
		ctx, _ = New(ctx)
		loggerVal = ctx.Value(LogKey)
	}
	logger, _ := loggerVal.(*Logger)
	return logger
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.l.Info(msg, fields...)
}
func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.l.Fatal(msg, fields...)
}
func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(RequestID, ctx.Value(RequestID).(string)))
	}
	l.l.Error(msg, fields...)
}

// LoggingInterceptor штука для логирования gRPC запросов
func LoggingInterceptor(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (any, error) {

	// Создаем request ID
	guid := uuid.New().String()
	ctx = context.WithValue(ctx, RequestID, guid)
	log := GetLoggerFromCtx(ctx)

	// Логируем начало запроса
	log.Info(ctx, "Request started",
		zap.String("request_id", guid),
		zap.String("method", info.FullMethod),
		zap.Time("request_time", time.Now()),
	)

	// Замеряем время выполнения
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	// Логируем завершение запроса
	if err != nil {
		log.Error(ctx, "Request failed",
			zap.String("request_id", guid),
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
	} else {
		log.Info(ctx, "Request completed",
			zap.String("request_id", guid),
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
		)
	}

	return resp, err
}
