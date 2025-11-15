package grpc

import (
	"context"
	"fmt"

	"github.com/phenirain/sso/pkg/contextkeys"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UserIDInterceptor создает интерцептор для добавления user_id в метаданные gRPC запросов
func UserIDInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Добавляем user_id в метаданные
		userID, ok := ctx.Value(contextkeys.UserIDCtxKey).(int64)
		if !ok {
			// Если user_id не найден в контексте, продолжаем без него
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		// Конвертируем int64 в строку для передачи в метаданных
		md := metadata.Pairs(contextkeys.UserIDCtxKey, fmt.Sprintf("%d", userID))
		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
