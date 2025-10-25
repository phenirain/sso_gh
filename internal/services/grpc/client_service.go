package grpc

import (
	"context"
	"encoding/json"

	pbApi "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api"
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/client"
	"gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/client/messages/order"
	"gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/client/messages/product"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ClientServiceWrapper оборачивает gRPC клиенты для работы с client сервисом
type ClientServiceWrapper struct {
	clientService  pb.ClientServiceClient
	productService pb.ProductServiceClient
	orderService   pb.OrderServiceClient
}

// NewClientServiceWrapper создает новый wrapper для client сервиса
func NewClientServiceWrapper(
	clientService pb.ClientServiceClient,
	productService pb.ProductServiceClient,
	orderService pb.OrderServiceClient,
) *ClientServiceWrapper {
	return &ClientServiceWrapper{
		clientService:  clientService,
		productService: productService,
		orderService:   orderService,
	}
}

// ClientService методы

func (s *ClientServiceWrapper) RegisterClient(ctx context.Context, request interface{}) (interface{}, error) {
	req, err := convertToProto[pbApi.ClientRequest](request)
	if err != nil {
		return nil, err
	}
	return s.clientService.RegisterClient(ctx, req)
}

func (s *ClientServiceWrapper) FillClientProfile(ctx context.Context, request interface{}) (interface{}, error) {
	req, err := convertToProto[pbApi.ClientRequest](request)
	if err != nil {
		return nil, err
	}
	return s.clientService.FillClientProfile(ctx, req)
}

func (s *ClientServiceWrapper) GetClientProfile(ctx context.Context, request interface{}) (interface{}, error) {
	req, err := convertToProto[pbApi.ActionByIdRequest](request)
	if err != nil {
		return nil, err
	}
	return s.clientService.GetClientProfile(ctx, req)
}

func (s *ClientServiceWrapper) DeleteClient(ctx context.Context, request interface{}) error {
	req, err := convertToProto[pbApi.ActionByIdRequest](request)
	if err != nil {
		return err
	}
	_, err = s.clientService.DeleteClient(ctx, req)
	return err
}

// ProductService методы

func (s *ClientServiceWrapper) GetAllBaseModels(ctx context.Context, request interface{}) (interface{}, error) {
	req, err := convertToProto[pbApi.GetBaseModelsRequest](request)
	if err != nil {
		return nil, err
	}
	return s.productService.GetAllBaseModels(ctx, req)
}

func (s *ClientServiceWrapper) GetProducts(ctx context.Context, request interface{}) (interface{}, error) {
	req, err := convertToProto[product.GetProductsRequest](request)
	if err != nil {
		return nil, err
	}
	return s.productService.GetProducts(ctx, req)
}

func (s *ClientServiceWrapper) GetProduct(ctx context.Context, request interface{}) (interface{}, error) {
	req, err := convertToProto[pbApi.ProductArticleRequest](request)
	if err != nil {
		return nil, err
	}
	return s.productService.GetProduct(ctx, req)
}

func (s *ClientServiceWrapper) ActionProductToFavorites(ctx context.Context, request interface{}) error {
	req, err := convertToProto[product.ProductIntoFavoritesRequest](request)
	if err != nil {
		return err
	}
	_, err = s.productService.ActionProductToFavorites(ctx, req)
	return err
}

func (s *ClientServiceWrapper) GetFavoriteProducts(ctx context.Context, request interface{}) (interface{}, error) {
	req, err := convertToProto[pbApi.ActionByIdRequest](request)
	if err != nil {
		return nil, err
	}
	return s.productService.GetFavoriteProducts(ctx, req)
}

// OrderService методы

func (s *ClientServiceWrapper) CreateOrder(ctx context.Context, request interface{}) (interface{}, error) {
	req, err := convertToProto[order.CreateOrderRequest](request)
	if err != nil {
		return nil, err
	}
	return s.orderService.CreateOrder(ctx, req)
}

func (s *ClientServiceWrapper) CompleteOrder(ctx context.Context, request interface{}) error {
	req, err := convertToProto[order.CompleteOrderRequest](request)
	if err != nil {
		return err
	}
	_, err = s.orderService.CompleteOrder(ctx, req)
	return err
}

func (s *ClientServiceWrapper) AddProductToOrder(ctx context.Context, request interface{}) (interface{}, error) {
	req, err := convertToProto[order.ProductIntoOrderRequest](request)
	if err != nil {
		return nil, err
	}
	return s.orderService.AddProductToOrder(ctx, req)
}

func (s *ClientServiceWrapper) GetClientOrders(ctx context.Context) (interface{}, error) {
	return s.orderService.GetClientOrders(ctx, &emptypb.Empty{})
}

func (s *ClientServiceWrapper) GetOrderById(ctx context.Context, request interface{}) (interface{}, error) {
	req, err := convertToProto[pbApi.ActionByIdRequest](request)
	if err != nil {
		return nil, err
	}
	return s.orderService.GetOrderById(ctx, req)
}

func (s *ClientServiceWrapper) CancelOrder(ctx context.Context, request interface{}) error {
	req, err := convertToProto[pbApi.ActionByIdRequest](request)
	if err != nil {
		return err
	}
	_, err = s.orderService.CancelOrder(ctx, req)
	return err
}

// convertToProto - вспомогательная функция для конвертации interface{} в proto структуру
func convertToProto[T any](data interface{}) (*T, error) {
	var result T

	// Сериализуем в JSON и десериализуем обратно
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonData, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
