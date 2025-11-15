package order

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/internal/dto/response"
	"gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api"
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/admin"
	"gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/admin/messages/order"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderHandler struct {
	s pb.OrderServiceClient
}

func NewOrderHandler(orderService pb.OrderServiceClient) *OrderHandler {
	return &OrderHandler{
		s: orderService,
	}
}

// GetOrderStatuses - получение всех статусов заказов
// @Summary Get order statuses
// @Tags admin-order
// @Produce json
// @Success 200 {object} response.Response[order.OrderStatusesResponse]
// @Security BearerAuth
// @Router /admin/order/statuses [get]
func (h *OrderHandler) GetOrderStatuses(c echo.Context) (err error) {
	var result *order.OrderStatusesResponse
	result, err = h.s.GetOrderStatuses(c.Request().Context(), &emptypb.Empty{})
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения статусов заказов", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetOrderClients - получение всех клиентов
// @Summary Get order clients
// @Tags admin-order
// @Produce json
// @Success 200 {object} response.Response[order.OrderClientsResponse]
// @Security BearerAuth
// @Router /admin/order/clients [get]
func (h *OrderHandler) GetOrderClients(c echo.Context) error {
	result, err := h.s.GetClients(c.Request().Context(), &emptypb.Empty{})
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения клиентов", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetOrderProducts - получение всех продуктов для заказов
// @Summary Get order products
// @Tags admin-order
// @Produce json
// @Success 200 {object} response.Response[order.OrderProductsResponse]
// @Security BearerAuth
// @Router /admin/order/products [get]
func (h *OrderHandler) GetOrderProducts(c echo.Context) error {
	result, err := h.s.GetProducts(c.Request().Context(), &emptypb.Empty{})
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения продуктов", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// CreateOrUpdateOrder - создание или обновление заказа
// @Summary Create or update order
// @Tags admin-order
// @Accept json
// @Produce json
// @Param request body order.OrderRequest true "Order request"
// @Success 200 {object} response.Response[api.ExtendedOrderResponse]
// @Security BearerAuth
// @Router /admin/order [post]
func (h *OrderHandler) CreateOrUpdateOrder(c echo.Context) (err error) {
	var req order.OrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	var result *api.ExtendedOrderResponse
	result, err = h.s.CreateOrUpdateOrder(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка создания/обновления заказа", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetOrders - получение списка заказов
// @Summary Get orders
// @Tags admin-order
// @Produce json
// @Param statusId path int true "Status ID"
// @Success 200 {object} response.Response[api.OrdersResponse]
// @Security BearerAuth
// @Router /admin/orders/status/{statusId} [get]
func (h *OrderHandler) GetOrders(c echo.Context) (err error) {
	var req order.GetOrdersRequest

	statusIdStr := c.Param("statusId")
	statusId, errConv := strconv.ParseInt(statusIdStr, 10, 64)
	if errConv != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный statusId", errConv.Error()))
	}
	req.StatusId = statusId

	var result *api.OrdersResponse
	result, err = h.s.GetOrders(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения заказов", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetOrderById - получение заказа по ID
// @Summary Get order by ID
// @Tags admin-order
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.Response[api.ExtendedOrderResponse]
// @Security BearerAuth
// @Router /admin/order/{id} [get]
func (h *OrderHandler) GetOrderById(c echo.Context) error {
	id := c.Param("id")
	var req api.ActionByIdRequest
	if parsed, errConv := strconv.ParseInt(id, 10, 64); errConv == nil {
		req.Id = parsed
	} else {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный идентификатор", errConv.Error()))
	}

	result, err := h.s.GetOrderById(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения заказа", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// DeleteOrder - удаление заказа
// @Summary Delete order
// @Tags admin-order
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.Response[string]
// @Security BearerAuth
// @Router /admin/order/{id} [delete]
func (h *OrderHandler) DeleteOrder(c echo.Context) error {
	id := c.Param("id")
	var req api.ActionByIdRequest
	if parsed, errConv := strconv.ParseInt(id, 10, 64); errConv == nil {
		req.Id = parsed
	} else {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный идентификатор", errConv.Error()))
	}

	_, err := h.s.DeleteOrder(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка удаления заказа", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Заказ успешно удален"))
}
