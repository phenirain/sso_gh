package client

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/internal/dto/response"
)

// ClientService интерфейс для всех client сервисов
type ClientService interface {
	// ClientService методы
	RegisterClient(ctx context.Context, request interface{}) (interface{}, error)
	FillClientProfile(ctx context.Context, request interface{}) (interface{}, error)
	GetClientProfile(ctx context.Context, request interface{}) (interface{}, error)
	DeleteClient(ctx context.Context, request interface{}) error

	// ProductService методы
	GetAllBaseModels(ctx context.Context, request interface{}) (interface{}, error)
	GetProducts(ctx context.Context, request interface{}) (interface{}, error)
	GetProduct(ctx context.Context, request interface{}) (interface{}, error)
	ActionProductToFavorites(ctx context.Context, request interface{}) error
	GetFavoriteProducts(ctx context.Context, request interface{}) (interface{}, error)

	// OrderService методы
	CreateOrder(ctx context.Context, request interface{}) (interface{}, error)
	CompleteOrder(ctx context.Context, request interface{}) error
	AddProductToOrder(ctx context.Context, request interface{}) (interface{}, error)
	GetClientOrders(ctx context.Context) (interface{}, error)
	GetOrderById(ctx context.Context, request interface{}) (interface{}, error)
	CancelOrder(ctx context.Context, request interface{}) error
}

type Handler struct {
	s ClientService
}

func NewHandler(clientService ClientService) *Handler {
	return &Handler{
		s: clientService,
	}
}

// ClientService handlers

// RegisterClient godoc
// @Summary Register client
// @Tags client
// @Accept json
// @Produce json
// @Param request body interface{} true "Client request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/register [post]
func (h *Handler) RegisterClient(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.RegisterClient(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка регистрации клиента", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// FillClientProfile godoc
// @Summary Fill client profile
// @Tags client
// @Accept json
// @Produce json
// @Param request body interface{} true "Client request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/profile [post]
func (h *Handler) FillClientProfile(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.FillClientProfile(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка заполнения профиля клиента", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetClientProfile godoc
// @Summary Get client profile
// @Tags client
// @Accept json
// @Produce json
// @Param request body interface{} true "Action by id request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/profile [get]
func (h *Handler) GetClientProfile(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.GetClientProfile(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения профиля клиента", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// DeleteClient godoc
// @Summary Delete client
// @Tags client
// @Accept json
// @Produce json
// @Param request body interface{} true "Action by id request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client [delete]
func (h *Handler) DeleteClient(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	err := h.s.DeleteClient(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка удаления клиента", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Клиент успешно удален"))
}

// ProductService handlers

// GetAllBaseModels godoc
// @Summary Get all base models
// @Tags client-product
// @Accept json
// @Produce json
// @Param request body interface{} true "Get base models request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/product/base-models [post]
func (h *Handler) GetAllBaseModels(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.GetAllBaseModels(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения базовых моделей", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetProducts godoc
// @Summary Get products
// @Tags client-product
// @Accept json
// @Produce json
// @Param request body interface{} true "Get products request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/products [post]
func (h *Handler) GetProducts(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.GetProducts(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения продуктов", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetProduct godoc
// @Summary Get product
// @Tags client-product
// @Accept json
// @Produce json
// @Param request body interface{} true "Product article request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/product [post]
func (h *Handler) GetProduct(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.GetProduct(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения продукта", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// ActionProductToFavorites godoc
// @Summary Action product to favorites
// @Tags client-product
// @Accept json
// @Produce json
// @Param request body interface{} true "Product into favorites request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/product/favorites [post]
func (h *Handler) ActionProductToFavorites(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	err := h.s.ActionProductToFavorites(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка добавления/удаления продукта в избранное", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Действие с избранным выполнено успешно"))
}

// GetFavoriteProducts godoc
// @Summary Get favorite products
// @Tags client-product
// @Accept json
// @Produce json
// @Param request body interface{} true "Action by id request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/product/favorites [get]
func (h *Handler) GetFavoriteProducts(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.GetFavoriteProducts(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения избранных продуктов", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// OrderService handlers

// CreateOrder godoc
// @Summary Create order
// @Tags client-order
// @Accept json
// @Produce json
// @Param request body interface{} true "Create order request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/order [post]
func (h *Handler) CreateOrder(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.CreateOrder(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка создания заказа", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// CompleteOrder godoc
// @Summary Complete order
// @Tags client-order
// @Accept json
// @Produce json
// @Param request body interface{} true "Complete order request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/order/complete [post]
func (h *Handler) CompleteOrder(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	err := h.s.CompleteOrder(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка оформления заказа", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Заказ успешно оформлен"))
}

// AddProductToOrder godoc
// @Summary Add product to order
// @Tags client-order
// @Accept json
// @Produce json
// @Param request body interface{} true "Product into order request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/order/add-product [post]
func (h *Handler) AddProductToOrder(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.AddProductToOrder(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка добавления продукта в заказ", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetClientOrders godoc
// @Summary Get client orders
// @Tags client-order
// @Produce json
// @Success 200 {object} response.Response[interface{}]
// @Router /client/orders [get]
func (h *Handler) GetClientOrders(c echo.Context) error {
	ctx := c.Request().Context()

	result, err := h.s.GetClientOrders(ctx)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения заказов клиента", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetOrderById godoc
// @Summary Get order by id
// @Tags client-order
// @Accept json
// @Produce json
// @Param request body interface{} true "Action by id request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/order/by-id [post]
func (h *Handler) GetOrderById(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.GetOrderById(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения заказа", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// CancelOrder godoc
// @Summary Cancel order
// @Tags client-order
// @Accept json
// @Produce json
// @Param request body interface{} true "Action by id request"
// @Success 200 {object} response.Response[interface{}]
// @Router /client/order/cancel [post]
func (h *Handler) CancelOrder(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	err := h.s.CancelOrder(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка отмены заказа", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Заказ успешно отменен"))
}
