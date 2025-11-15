package order

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/internal/dto/response"
	pbApi "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api"
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/client"
	msg "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/client/messages/order"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderHandler struct {
	s pb.OrderServiceClient
}

func NewOrderHandler(orderService pb.OrderServiceClient) *OrderHandler {
	return &OrderHandler{s: orderService}
}

// CreateOrder - создание заказа
// @Summary Create order
// @Tags client-order
// @Accept json
// @Produce json
// @Param request body any true "Create order request"
// @Success 200 {object} response.Response[msg.ClientOrderResponse]
// @Security BearerAuth
// @Router /client/order [post]
func (h *OrderHandler) CreateOrder(c echo.Context) error {
	var req msg.CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.CreateOrder(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка создания заказа", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// CompleteOrder - оформление заказа
// @Summary Complete order
// @Tags client-order
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.Response[string]
// @Security BearerAuth
// @Router /client/order/{id}/complete [post]
func (h *OrderHandler) CompleteOrder(c echo.Context) error {
	id := c.Param("id")
	var req msg.CompleteOrderRequest
	if parsed, errConv := strconv.ParseInt(id, 10, 64); errConv == nil {
		req.OrderId = parsed
	} else {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный идентификатор", errConv.Error()))
	}

	_, err := h.s.CompleteOrder(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка оформления заказа", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Заказ успешно оформлен"))
}

// AddProductToOrder - добавить продукт в заказ
// @Summary Add product to order
// @Tags client-order
// @Accept json
// @Produce json
// @Param request body any true "Product into order request"
// @Success 200 {object} response.Response[msg.ClientOrderResponse]
// @Security BearerAuth
// @Router /client/order/add-product [post]
func (h *OrderHandler) AddProductToOrder(c echo.Context) error {
	var req msg.ProductIntoOrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.AddProductToOrder(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка добавления продукта в заказ", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetClientOrders - получить заказы клиента
// @Summary Get client orders
// @Tags client-order
// @Produce json
// @Success 200 {object} response.Response[pbApi.OrdersResponse]
// @Security BearerAuth
// @Router /client/order [get]
func (h *OrderHandler) GetClientOrders(c echo.Context) error {
	result, err := h.s.GetClientOrders(c.Request().Context(), &emptypb.Empty{})
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения заказов клиента", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetOrderById - получить заказ по id
// @Summary Get order by id
// @Tags client-order
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.Response[msg.ClientOrderResponse]
// @Security BearerAuth
// @Router /client/order/{id} [get]
func (h *OrderHandler) GetOrderById(c echo.Context) error {
	id := c.Param("id")
	var req pbApi.ActionByIdRequest
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

// CancelOrder - отменить заказ
// @Summary Cancel order
// @Tags client-order
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} response.Response[string]
// @Security BearerAuth
// @Router /client/order/{id}/cancel [post]
func (h *OrderHandler) CancelOrder(c echo.Context) error {
	id := c.Param("id")
	var req pbApi.ActionByIdRequest
	if parsed, errConv := strconv.ParseInt(id, 10, 64); errConv == nil {
		req.Id = parsed
	} else {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный идентификатор", errConv.Error()))
	}

	_, err := h.s.CancelOrder(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка отмены заказа", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Заказ успешно отменен"))
}
