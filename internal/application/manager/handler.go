package order

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/internal/dto/response"
	api "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api"
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/manager"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderHandler struct {
	s pb.ManagerServiceClient
}

func NewOrderHandler(managerService pb.ManagerServiceClient) *OrderHandler {
	return &OrderHandler{s: managerService}
}

// GetAllOrders - получить все заказы
// @Summary Get all orders
// @Tags manager
// @Produce json
// @Success 200 {object} response.Response[api.OrdersResponse]
// @Security BearerAuth
// @Router /manager/order [get]
func (h *OrderHandler) GetAllOrders(c echo.Context) (err error) {
	var result *api.OrdersResponse
	result, err = h.s.GetAllOrders(c.Request().Context(), &emptypb.Empty{})
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения заказов", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetOrderById - получить заказ по id
// @Summary Get order by id
// @Tags manager
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.Response[api.ExtendedOrderResponse]
// @Security BearerAuth
// @Router /manager/order/{id} [get]
func (h *OrderHandler) GetOrderById(c echo.Context) (err error) {
	id := c.Param("id")
	var req api.ActionByIdRequest
	if parsed, errConv := strconv.ParseInt(id, 10, 64); errConv == nil {
		req.Id = parsed
	} else {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный идентификатор", errConv.Error()))
	}

	var result *api.ExtendedOrderResponse
	result, err = h.s.GetOrderById(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения заказа", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GiveOrder - выдать заказ
// @Summary Give order
// @Tags manager
// @Accept json
// @Produce json
// @Param request body pb.PaidOrderRequest true "Paid order request"
// @Success 200 {object} response.Response[string]
// @Security BearerAuth
// @Router /manager/order/give [post]
func (h *OrderHandler) GiveOrder(c echo.Context) (err error) {
	var req pb.PaidOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	_, err = h.s.GiveOrder(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка передачи заказа", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Заказ успешно передан"))
}

// CancelOrder - отменить заказ
// @Summary Cancel order
// @Tags manager
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.Response[string]
// @Security BearerAuth
// @Router /manager/order/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c echo.Context) (err error) {
	id := c.Param("id")
	var req api.ActionByIdRequest
	if parsed, errConv := strconv.ParseInt(id, 10, 64); errConv == nil {
		req.Id = parsed
	} else {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный идентификатор", errConv.Error()))
	}

	_, err = h.s.CancelOrder(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка отмены заказа", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Заказ успешно отменен"))
}
