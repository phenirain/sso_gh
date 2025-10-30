package client

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/internal/dto/response"
	api "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api"
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/client"
)

type ClientHandler struct {
	s pb.ClientServiceClient
}

func NewClientHandler(clientService pb.ClientServiceClient) *ClientHandler {
	return &ClientHandler{s: clientService}
}

// RegisterClient - регистрация клиента
// @Summary Register client
// @Tags client
// @Accept json
// @Produce json
// @Param request body any true "Client request"
// @Success 200 {object} response.Response[api.ClientResponse]
// @Router /client/register [post]
func (h *ClientHandler) RegisterClient(c echo.Context) (err error) {
	var req api.ClientRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	var result *api.ClientResponse
	result, err = h.s.RegisterClient(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка регистрации клиента", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// FillClientProfile - заполнение профиля клиента
// @Summary Fill client profile
// @Tags client
// @Accept json
// @Produce json
// @Param request body any true "Client request"
// @Success 200 {object} response.Response[api.ClientResponse]
// @Router /client/profile [post]
func (h *ClientHandler) FillClientProfile(c echo.Context) (err error) {
	var req api.ClientRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	var result *api.ClientResponse
	result, err = h.s.FillClientProfile(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка заполнения профиля клиента", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetClientProfile - получение профиля клиента
// @Summary Get client profile
// @Tags client
// @Produce json
// @Param id path int true "Client ID"
// @Success 200 {object} response.Response[api.ClientResponse]
// @Router /client/profile/{id} [get]
func (h *ClientHandler) GetClientProfile(c echo.Context) (err error) {
	var req api.ActionByIdRequest
	if parsed, errConv := strconv.ParseInt(c.Param("id"), 10, 64); errConv == nil {
		req.Id = parsed
	} else {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный идентификатор", errConv.Error()))
	}

	var result *api.ClientResponse
	result, err = h.s.GetClientProfile(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения профиля клиента", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// DeleteClient - удаление клиента
// @Summary Delete client
// @Tags client
// @Produce json
// @Param id path int true "Client ID"
// @Success 200 {object} response.Response[string]
// @Router /client/{id} [delete]
func (h *ClientHandler) DeleteClient(c echo.Context) (err error) {
	id := c.Param("id")
	req := api.ActionByIdRequest{}
	// convert id string to int64
	if parsed, convErr := strconv.ParseInt(id, 10, 64); convErr == nil {
		req.Id = parsed
	} else {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный идентификатор", convErr.Error()))
	}

	_, err = h.s.DeleteClient(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка удаления клиента", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Клиент успешно удален"))
}
