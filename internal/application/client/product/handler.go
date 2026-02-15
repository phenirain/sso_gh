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
// @Produce json
// @Param baseModelName query string false "Base model name"
// @Success 200 {object} response.Response[api.BaseModelsResponse]
// @Security BearerAuth
// @Router /client/product/base-models [get]
func (h *ProductHandler) GetAllBaseModels(c echo.Context) error {
	var req api.GetBaseModelsRequest

	baseModelName := c.QueryParam("baseModelName")
	if baseModelName != "" {
		req.BaseModel = baseModelName
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
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param brand_id query int false "Brand ID"
// @Param term query string false "Search term (article, name, brand)"
// @Success 200 {object} response.Response[api.ProductsResponse]
// @Security BearerAuth
// @Router /client/product [get]
func (h *ProductHandler) GetProducts(c echo.Context) error {
	req := msg.GetProductsRequest{}

	// Получаем параметры из query с базовыми значениями
	// Limit - по умолчанию 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if limit, err := strconv.ParseInt(limitStr, 10, 32); err == nil {
			limit32 := int32(limit)
			req.Limit = &limit32
		}
	} else {
		defaultLimit := int32(10)
		req.Limit = &defaultLimit
	}

	// Offset - по умолчанию 0
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if offset, err := strconv.ParseInt(offsetStr, 10, 32); err == nil {
			offset32 := int32(offset)
			req.Offset = &offset32
		}
	} else {
		defaultOffset := int32(0)
		req.Offset = &defaultOffset
	}

	// MinPrice - опциональный параметр
	if minPriceStr := c.QueryParam("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			req.MinPrice = &minPrice
		}
	}

	// MaxPrice - опциональный параметр
	if maxPriceStr := c.QueryParam("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			req.MaxPrice = &maxPrice
		}
	}

	// BrandId - опциональный параметр
	if brandIdStr := c.QueryParam("brand_id"); brandIdStr != "" {
		if brandId, err := strconv.ParseInt(brandIdStr, 10, 64); err == nil {
			req.BrandId = &brandId
		}
	}

	// Term - опциональный параметр поиска
	if term := c.QueryParam("term"); term != "" {
		req.Term = &term
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
// @Security BearerAuth
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
// @Produce json
// @Param article path string true "Product article"
// @Success 200 {object} response.Response[string]
// @Security BearerAuth
// @Router /client/product/{article}/favorites [post]
func (h *ProductHandler) ActionProductToFavorites(c echo.Context) error {
	article := c.Param("article")
	var req msg.ProductIntoFavoritesRequest
	req.Article = article

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
// @Security BearerAuth
// @Router /client/product/favorites [get]
func (h *ProductHandler) GetFavoriteProducts(c echo.Context) error {
	req := api.ActionByIdRequest{
		Id: 0,
	}
	result, err := h.s.GetFavoriteProducts(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения избранных продуктов", err.Error()))
	}
	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}
