package product

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/internal/dto/response"
	api "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api"
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/client"
	msg "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/client/messages/product"
)

type ProductHandler struct {
	s pb.ProductServiceClient
}

func NewProductHandler(productService pb.ProductServiceClient) *ProductHandler {
	return &ProductHandler{s: productService}
}

// GetAllBaseModels - получить базовые модели
// @Summary Get all base models
// @Tags client-product
// @Accept json
// @Produce json
// @Param request body any true "Get base models request"
// @Success 200 {object} response.Response[api.BaseModelsResponse]
// @Router /client/product/base-models [post]
func (h *ProductHandler) GetAllBaseModels(c echo.Context) error {
	var req api.GetBaseModelsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.GetAllBaseModels(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения базовых моделей", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetProducts - получить продукты
// @Summary Get products
// @Tags client-product
// @Accept json
// @Produce json
// @Param request body any true "Get products request"
// @Success 200 {object} response.Response[api.ProductsResponse]
// @Router /client/product [post]
func (h *ProductHandler) GetProducts(c echo.Context) error {
	var req msg.GetProductsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.GetProducts(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения продуктов", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetProduct - получить один продукт
// @Summary Get product
// @Tags client-product
// @Produce json
// @Param article path string true "Product article"
// @Success 200 {object} response.Response[api.ExtendedProductResponse]
// @Router /client/product/{article} [get]
func (h *ProductHandler) GetProduct(c echo.Context) error {
	var req api.ProductArticleRequest
	req.Article = c.Param("article")

	result, err := h.s.GetProduct(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения продукта", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// ActionProductToFavorites - действие с избранным
// @Summary Action product to favorites
// @Tags client-product
// @Accept json
// @Produce json
// @Param request body any true "Product into favorites request"
// @Success 200 {object} response.Response[string]
// @Router /client/product/favorites [post]
func (h *ProductHandler) ActionProductToFavorites(c echo.Context) error {
	var req msg.ProductIntoFavoritesRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	_, err := h.s.ActionProductToFavorites(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка действия с избранным", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Действие с избранным выполнено успешно"))
}

// GetFavoriteProducts - получить избранные продукты
// @Summary Get favorite products
// @Tags client-product
// @Produce json
// @Param id path int true "Client ID"
// @Success 200 {object} response.Response[api.ProductsResponse]
// @Router /client/product/{id}/favorites [get]
func (h *ProductHandler) GetFavoriteProducts(c echo.Context) error {
	var req api.ActionByIdRequest
	if parsed, errConv := strconv.ParseInt(c.Param("id"), 10, 64); errConv == nil {
		req.Id = parsed
	} else {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный идентификатор", errConv.Error()))
	}

	result, err := h.s.GetFavoriteProducts(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения избранных продуктов", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}
