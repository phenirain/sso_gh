package grpc

import (
	"context"

	pbApi "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api"
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/manager"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ManagerServiceWrapper оборачивает gRPC клиент для работы с manager сервисом
type ManagerServiceWrapper struct {
	managerService pb.ManagerServiceClient
}

// NewManagerServiceWrapper создает новый wrapper для manager сервиса
func NewManagerServiceWrapper(managerService pb.ManagerServiceClient) *ManagerServiceWrapper {
	return &ManagerServiceWrapper{
		managerService: managerService,
	}
}

func (s *ManagerServiceWrapper) GetAllOrders(ctx context.Context) (interface{}, error) {
	return s.managerService.GetAllOrders(ctx, &emptypb.Empty{})
}

func (s *ManagerServiceWrapper) GetOrderById(ctx context.Context, request interface{}) (interface{}, error) {
	req, err := convertToProto[pbApi.ActionByIdRequest](request)
	if err != nil {
		return nil, err
	}
	return s.managerService.GetOrderById(ctx, req)
}

func (s *ManagerServiceWrapper) GiveOrder(ctx context.Context, request interface{}) error {
	req, err := convertToProto[pb.PaidOrderRequest](request)
	if err != nil {
		return err
	}
	_, err = s.managerService.GiveOrder(ctx, req)
	return err
}

func (s *ManagerServiceWrapper) CancelOrder(ctx context.Context, request interface{}) error {
	req, err := convertToProto[pbApi.ActionByIdRequest](request)
	if err != nil {
		return err
	}
	_, err = s.managerService.CancelOrder(ctx, req)
	return err
}
