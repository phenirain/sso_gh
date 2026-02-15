package product

import (
	"io"
	"net/http"
	"strconv"

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
// @Security BearerAuth
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
// @Produce json
// @Param baseModelName path string true "Base model name"
// @Success 200 {object} response.Response[api.BaseModelsResponse]
// @Security BearerAuth
// @Router /admin/product/base-models/{baseModelName} [get]
func (h *ProductHandler) GetAllBaseModels(c echo.Context) error {
	var req api.GetBaseModelsRequest

	baseModelName := c.Param("baseModelName")
	if baseModelName == "" {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный baseModelName", "baseModelName is required"))
	}
	req.BaseModel = baseModelName

	result, err := h.s.GetAllBaseModels(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения базовых моделей", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// DeleteBaseModel - удаление базовой модели
// @Summary Delete base model
// @Tags admin-product
// @Produce json
// @Param baseModelName path string true "Base model name"
// @Param id path int true "Base model ID"
// @Success 200 {object} response.Response[string]
// @Security BearerAuth
// @Router /admin/product/base-model/{baseModelName}/{id} [delete]
func (h *ProductHandler) DeleteBaseModel(c echo.Context) error {
	var req product.DeleteBaseModelRequest

	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный id", "id is required"))
	}
	id, errConv := strconv.ParseInt(idStr, 10, 64)
	if errConv != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный id", errConv.Error()))
	}
	req.Id = id

	baseModelName := c.Param("baseModelName")
	if baseModelName == "" {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный baseModelName", "baseModelName is required"))
	}
	req.BaseModel = baseModelName

	_, err := h.s.DeleteBaseModel(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка удаления базовой модели", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Базовая модель успешно удалена"))
}

// CreateOrUpdateProduct - создание или обновление продукта
// @Summary Create or update product
// @Tags admin-product
// @Accept multipart/form-data
// @Produce json
// @Param article formData string false "Product article"
// @Param article_old formData string false "Old article"
// @Param name formData string false "Product name"
// @Param description formData string false "Product description"
// @Param price formData number false "Product price"
// @Param quantity formData int false "Product quantity"
// @Param brand_id formData int false "Brand ID"
// @Param product_type_id formData int false "Product type ID"
// @Param texture_id formData int false "Texture ID"
// @Param volume formData int false "Volume"
// @Param volume_id formData int false "Volume ID"
// @Param is_archived formData bool false "Is archived"
// @Param image formData file false "Product image"
// @Success 200 {object} response.Response[api.ExtendedProductResponse]
// @Security BearerAuth
// @Router /admin/product [post]
func (h *ProductHandler) CreateOrUpdateProduct(c echo.Context) error {
	var req product.ProductRequest

	// Получаем текстовые поля из формы
	if article := c.FormValue("article"); article != "" {
		req.Article = article
	}
	if articleOld := c.FormValue("article_old"); articleOld != "" {
		req.ArticleOld = &articleOld
	}
	if name := c.FormValue("name"); name != "" {
		req.Name = name
	}
	if description := c.FormValue("description"); description != "" {
		req.Description = &description
	}
	if priceStr := c.FormValue("price"); priceStr != "" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
			req.Price = &price
		}
	}
	if quantityStr := c.FormValue("quantity"); quantityStr != "" {
		if quantity, err := strconv.ParseInt(quantityStr, 10, 32); err == nil {
			qty := int32(quantity)
			req.Quantity = &qty
		}
	}
	if brandIdStr := c.FormValue("brand_id"); brandIdStr != "" {
		if brandId, err := strconv.ParseInt(brandIdStr, 10, 64); err == nil {
			req.BrandId = &brandId
		}
	}
	if productTypeIdStr := c.FormValue("product_type_id"); productTypeIdStr != "" {
		if productTypeId, err := strconv.ParseInt(productTypeIdStr, 10, 64); err == nil {
			req.ProductTypeId = &productTypeId
		}
	}
	if textureIdStr := c.FormValue("texture_id"); textureIdStr != "" {
		if textureId, err := strconv.ParseInt(textureIdStr, 10, 64); err == nil {
			req.TextureId = &textureId
		}
	}
	if volumeStr := c.FormValue("volume"); volumeStr != "" {
		if volume, err := strconv.ParseInt(volumeStr, 10, 32); err == nil {
			vol := int32(volume)
			req.Volume = &vol
		}
	}
	if volumeIdStr := c.FormValue("volume_id"); volumeIdStr != "" {
		if volumeId, err := strconv.ParseInt(volumeIdStr, 10, 64); err == nil {
			req.VolumeId = &volumeId
		}
	}
	if isArchivedStr := c.FormValue("is_archived"); isArchivedStr != "" {
		if isArchived, err := strconv.ParseBool(isArchivedStr); err == nil {
			req.IsArchived = &isArchived
		}
	}

	// Получаем файл изображения
	file, err := c.FormFile("image")
	if err == nil && file != nil {
		// Открываем файл
		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка открытия файла изображения", err.Error()))
		}
		defer func() {
			_ = src.Close()
		}()

		// Читаем содержимое файла в байты
		imageBytes, err := io.ReadAll(src)
		if err != nil {
			return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка чтения файла изображения", err.Error()))
		}

		if len(imageBytes) > 0 {
			req.Image = imageBytes
		}
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
// @Produce json
// @Param article path string true "Product article"
// @Success 200 {object} response.Response[api.ExtendedProductResponse]
// @Security BearerAuth
// @Router /admin/product/{article} [get]
func (h *ProductHandler) GetProductByArticle(c echo.Context) error {
	var req api.ProductArticleRequest
	req.Article = c.Param("article")

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
// @Security BearerAuth
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
// @Produce json
// @Param article path string true "Product article"
// @Success 200 {object} response.Response[string]
// @Security BearerAuth
// @Router /admin/product/{article} [delete]
func (h *ProductHandler) DeleteProduct(c echo.Context) error {
	var req api.ProductArticleRequest
	req.Article = c.Param("article")

	_, err := h.s.DeleteProduct(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка удаления продукта", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponseEmpty("Продукт успешно удален"))
}
