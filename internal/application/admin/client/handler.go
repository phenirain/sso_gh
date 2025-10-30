package client

import (
	"net/http"

	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/internal/dto/response"
	"gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api"
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/admin"
	"gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/admin/messages/client"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ClientHandler struct {
	s pb.ClientServiceClient
}

func NewClientHandler(clientService pb.ClientServiceClient) *ClientHandler {
	return &ClientHandler{
		s: clientService,
	}
}

// GetUsers - получение всех пользователей
// @Summary Get users
// @Tags admin-client
// @Produce json
// @Success 200 {object} response.Response[client.ClientUsersResponse]
// @Router /admin/client/users [get]
func (h *ClientHandler) GetUsers(c echo.Context) (err error) {
	var result *client.ClientUsersResponse
	result, err = h.s.GetUsers(c.Request().Context(), &emptypb.Empty{})
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения пользователей", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// CreateClient - создание клиента
// @Summary Create client
// @Tags admin-client
// @Accept json
// @Produce json
// @Param request body api.ClientRequest true "Client request"
// @Success 200 {object} response.Response[api.ClientResponse]
// @Router /admin/client [post]
func (h *ClientHandler) CreateClient(c echo.Context) (err error) {
	var req api.ClientRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	var result *api.ClientResponse
	result, err = h.s.CreateClient(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка создания клиента", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetClients - получение всех клиентов
// @Summary Get all clients
// @Tags admin-client
// @Produce json
// @Success 200 {object} response.Response[client.ClientsResponse]
// @Router /admin/client [get]
func (h *ClientHandler) GetClients(c echo.Context) (err error) {
	var result *client.ClientsResponse
	result, err = h.s.GetClients(c.Request().Context(), &emptypb.Empty{})
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения клиентов", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// DeleteClient - удаление клиента
// @Summary Delete client
// @Tags admin-client
// @Produce json
// @Param id path int true "Client ID"
// @Success 200 {object} response.Response[string]
// @Router /admin/client/{id} [delete]
func (h *ClientHandler) DeleteClient(c echo.Context) error {
	id := c.Param("id")
	req := api.ActionByIdRequest{}
	if parsed, convErr := strconv.ParseInt(id, 10, 64); convErr == nil {
		req.Id = parsed
	} else {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный идентификатор", convErr.Error()))
	}

	_, err := h.s.DeleteClient(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка удаления клиента", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Клиент успешно удален"))
}
