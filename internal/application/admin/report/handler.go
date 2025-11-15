package report

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/phenirain/sso/internal/dto/response"
	pb "gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/admin"
	"gitlab.com/mpt4164636/fourthcoursefirstprojectgroup/proto/generated/api/admin/messages/report"
)

type ReportHandler struct {
	s pb.ReportServiceClient
}

func NewReportHandler(reportService pb.ReportServiceClient) *ReportHandler {
	return &ReportHandler{
		s: reportService,
	}
}

// GetAmountOfOrdersByTimeOfDay - получение количества заказов по времени суток
// @Summary Get amount of orders by time of day
// @Tags admin-report
// @Produce json
// @Param period path string true "Period (today, yesterday, week, month, year)" Enums(today, yesterday, week, month, year)
// @Success 200 {object} response.Response[report.OrdersByTimeOfDayResponse]
// @Security BearerAuth
// @Router /admin/report/orders-by-time/{period} [get]
func (h *ReportHandler) GetAmountOfOrdersByTimeOfDay(c echo.Context) (err error) {
	var result *report.OrdersByTimeOfDayResponse
	var req report.ReportRequest

	period := c.Param("period")
	if period == "" {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный period", "period is required"))
	}
	req.Period = period

	result, err = h.s.GetAmountOfOrdersByTimeOfDay(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения данных по времени суток", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetPurchasesByBrands - получение покупок по брендам
// @Summary Get purchases by brands
// @Tags admin-report
// @Produce json
// @Param period path string true "Period (today, yesterday, week, month, year)" Enums(today, yesterday, week, month, year)
// @Success 200 {object} response.Response[report.PurchasesByBrandsResponse]
// @Security BearerAuth
// @Router /admin/report/purchases-by-brands/{period} [get]
func (h *ReportHandler) GetPurchasesByBrands(c echo.Context) (err error) {
	var result *report.PurchasesByBrandsResponse
	var req report.ReportRequest

	period := c.Param("period")
	if period == "" {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный period", "period is required"))
	}
	req.Period = period

	result, err = h.s.GetPurchasesByBrands(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения данных по брендам", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}

// GetAverageOrderProcessingTime - получение среднего времени обработки заказов
// @Summary Get average order processing time
// @Tags admin-report
// @Produce json
// @Param period path string true "Period (today, yesterday, week, month, year)" Enums(today, yesterday, week, month, year)
// @Success 200 {object} response.Response[report.AverageOrderProcessingTimeResponse]
// @Security BearerAuth
// @Router /admin/report/average-processing-time/{period} [get]
func (h *ReportHandler) GetAverageOrderProcessingTime(c echo.Context) (err error) {
	var result *report.AverageOrderProcessingTimeResponse
	var req report.ReportRequest

	period := c.Param("period")
	if period == "" {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Некорректный period", "period is required"))
	}
	req.Period = period

	result, err = h.s.GetAverageOrderProcessingTime(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusOK, response.NewBadResponse[any]("Ошибка получения данных по времени обработки", err.Error()))
	}

	return c.JSON(http.StatusOK, response.NewSuccessResponse(&result))
}
