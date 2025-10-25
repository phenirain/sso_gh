package product

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/internal/dto/response"
	"gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api"
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/admin"
	"gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/admin/messages/product"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProductHandler struct {
	s pb.ProductServiceClient
}

func NewProductHandler(productService pb.ProductServiceClient) *ProductHandler {
	return &ProductHandler{
		s: productService,
	}
}

// CreateOrUpdateBaseModel - создание или обновление базовой модели
// @Summary Create or update base model
// @Tags admin-product
// @Accept json
// @Produce json
// @Param request body product.BaseModelRequest true "Base model request"
// @Success 200 {object} response.Response[api.BaseModelResponse]
// @Router /admin/product/base-model [post]
func (h *ProductHandler) CreateOrUpdateBaseModel(c echo.Context) error {
	var req product.BaseModelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.CreateOrUpdateBaseModel(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка создания/обновления базовой модели", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetAllBaseModels - получение всех базовых моделей
// @Summary Get all base models
// @Tags admin-product
// @Accept json
// @Produce json
// @Param request body api.GetBaseModelsRequest true "Get base models request"
// @Success 200 {object} response.Response[api.BaseModelsResponse]
// @Router /admin/product/base-models [post]
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

// DeleteBaseModel - удаление базовой модели
// @Summary Delete base model
// @Tags admin-product
// @Accept json
// @Produce json
// @Param request body product.DeleteBaseModelRequest true "Delete base model request"
// @Success 200 {object} response.Response[string]
// @Router /admin/product/base-model [delete]
func (h *ProductHandler) DeleteBaseModel(c echo.Context) error {
	var req product.DeleteBaseModelRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	_, err := h.s.DeleteBaseModel(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка удаления базовой модели", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Базовая модель успешно удалена"))
}

// CreateOrUpdateProduct - создание или обновление продукта
// @Summary Create or update product
// @Tags admin-product
// @Accept json
// @Produce json
// @Param request body product.ProductRequest true "Product request"
// @Success 200 {object} response.Response[api.ExtendedProductResponse]
// @Router /admin/product [post]
func (h *ProductHandler) CreateOrUpdateProduct(c echo.Context) error {
	var req product.ProductRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.CreateOrUpdateProduct(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка создания/обновления продукта", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetProductByArticle - получение продукта по артикулу
// @Summary Get product by article
// @Tags admin-product
// @Accept json
// @Produce json
// @Param request body api.ProductArticleRequest true "Product article request"
// @Success 200 {object} response.Response[api.ExtendedProductResponse]
// @Router /admin/product/by-article [post]
func (h *ProductHandler) GetProductByArticle(c echo.Context) error {
	var req api.ProductArticleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	result, err := h.s.GetProductByArticle(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения продукта", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetProducts - получение всех продуктов
// @Summary Get all products
// @Tags admin-product
// @Produce json
// @Success 200 {object} response.Response[api.ProductsResponse]
// @Router /admin/product [get]
func (h *ProductHandler) GetProducts(c echo.Context) error {
	result, err := h.s.GetProducts(c.Request().Context(), &emptypb.Empty{})
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения продуктов", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// DeleteProduct - удаление продукта
// @Summary Delete product
// @Tags admin-product
// @Accept json
// @Produce json
// @Param request body api.ProductArticleRequest true "Product article request"
// @Success 200 {object} response.Response[string]
// @Router /admin/product [delete]
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	var req api.ProductArticleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения json", err.Error()))
	}

	_, err := h.s.DeleteProduct(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка удаления продукта", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Продукт успешно удален"))
}
