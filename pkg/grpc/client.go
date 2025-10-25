package grpc

import (
	"fmt"
	"log/slog"

	pbAdmin "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/admin"
	pbClient "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/client"
	pbManager "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Clients содержит все gRPC клиенты
type Clients struct {
	Admin   pbAdmin.ClientServiceClient
	Client  pbClient.ClientServiceClient
	Manager pbManager.ManagerServiceClient
	conns   []*grpc.ClientConn
}

// NewClients создает новые gRPC клиенты с интерцепторами
func NewClients(apiAddress string) (*Clients, error) {
	// Создаем соединения с интерцепторами
	adminConn, err := grpc.NewClient(apiAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(UserIDInterceptor()), // Пустой user_id, будет заменен в хендлерах
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to admin service: %w", err)
	}

	clientConn, err := grpc.NewClient(apiAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(UserIDInterceptor()),
	)
	if err != nil {
		adminConn.Close()
		return nil, fmt.Errorf("failed to connect to client service: %w", err)
	}

	managerConn, err := grpc.NewClient(apiAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(UserIDInterceptor()),
	)
	if err != nil {
		adminConn.Close()
		clientConn.Close()
		return nil, fmt.Errorf("failed to connect to manager service: %w", err)
	}

	return &Clients{
		Admin:   pbAdmin.NewClientServiceClient(adminConn),
		Client:  pbClient.NewClientServiceClient(clientConn),
		Manager: pbManager.NewManagerServiceClient(managerConn),
		conns:   []*grpc.ClientConn{adminConn, clientConn, managerConn},
	}, nil
}

// Close закрывает все соединения
func (c *Clients) Close() error {
	for _, conn := range c.conns {
		if err := conn.Close(); err != nil {
			slog.Error("failed to close gRPC connection", "error", err)
		}
	}
	return nil
}
