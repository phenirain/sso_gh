package manager

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/internal/dto/response"
)

// ManagerService интерфейс для manager сервиса
type ManagerService interface {
	GetAllOrders(ctx context.Context) (interface{}, error)
	GetOrderById(ctx context.Context, request interface{}) (interface{}, error)
	GiveOrder(ctx context.Context, request interface{}) error
	CancelOrder(ctx context.Context, request interface{}) error
}

type Handler struct {
	s ManagerService
}

func NewHandler(managerService ManagerService) *Handler {
	return &Handler{
		s: managerService,
	}
}

// GetAllOrders godoc
// @Summary Get all orders
// @Tags manager
// @Produce json
// @Success 200 {object} response.Response[interface{}]
// @Router /manager/orders [get]
func (h *Handler) GetAllOrders(c echo.Context) error {
	ctx := c.Request().Context()

	result, err := h.s.GetAllOrders(ctx)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения заказов", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetOrderById godoc
// @Summary Get order by id
// @Tags manager
// @Accept json
// @Produce json
// @Param request body interface{} true "Action by id request"
// @Success 200 {object} response.Response[interface{}]
// @Router /manager/order/by-id [post]
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

// GiveOrder godoc
// @Summary Give order
// @Tags manager
// @Accept json
// @Produce json
// @Param request body interface{} true "Paid order request"
// @Success 200 {object} response.Response[interface{}]
// @Router /manager/order/give [post]
func (h *Handler) GiveOrder(c echo.Context) error {
	ctx := c.Request().Context()

	var req interface{}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	err := h.s.GiveOrder(ctx, req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка передачи заказа", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Заказ успешно передан"))
}

// CancelOrder godoc
// @Summary Cancel order
// @Tags manager
// @Accept json
// @Produce json
// @Param request body interface{} true "Action by id request"
// @Success 200 {object} response.Response[interface{}]
// @Router /manager/order/cancel [post]
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
