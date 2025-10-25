package grpc

import (
	"context"

	"github.com/phenirain/sso/pkg/contextkeys"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UserIDInterceptor создает интерцептор для добавления user_id в метаданные gRPC запросов
func UserIDInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Добавляем user_id в метаданные
		userID := ctx.Value(contextkeys.UserIDCtxKey).(string)
		md := metadata.Pairs(contextkeys.UserIDCtxKey, userID)
		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
