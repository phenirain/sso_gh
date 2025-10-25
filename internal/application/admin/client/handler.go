package client

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/internal/dto/response"
	"gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api"
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/admin"
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
func (h *ClientHandler) GetUsers(c echo.Context) error {
    
    result, err := h.s.GetUsers(c.Request().Context(), &emptypb.Empty{})
    if err != nil {
        return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения пользователей", err.Error()))
    }
    
    return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// CreateClient - создание клиента
func (h *ClientHandler) CreateClient(c echo.Context) error {
    var req api.ClientRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
    }
    
    result, err := h.s.CreateClient(c.Request().Context(), &req)
    if err != nil {
        return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка создания клиента", err.Error()))
    }
    
    return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetClients - получение всех клиентов
func (h *ClientHandler) GetClients(c echo.Context) error {
    result, err := h.s.GetClients(c.Request().Context(), &emptypb.Empty{})
    if err != nil {
        return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения клиентов", err.Error()))
    }
    
    return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// DeleteClient - удаление клиента
func (h *ClientHandler) DeleteClient(c echo.Context) error {
    var req api.ActionByIdRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
    }
    
    _, err := h.s.DeleteClient(c.Request().Context(), &req)
    if err != nil {
        return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка удаления клиента", err.Error()))
    }
    
    return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Клиент успешно удален"))
}