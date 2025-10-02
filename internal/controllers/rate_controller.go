package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spksupakorn/Currency-Converter/internal/services"
	"github.com/spksupakorn/Currency-Converter/pkg/logger"
	"github.com/spksupakorn/Currency-Converter/pkg/response"
)

type RateController struct {
	rates services.RateService
	log   *logger.Logger
}

func NewRateController(rates services.RateService, log *logger.Logger) *RateController {
	return &RateController{rates: rates, log: log}
}

// GetRates godoc
// @Summary      Get Exchange Rates
// @Description  Get all exchange rates relative to a specified base currency. If no base is provided, defaults to USD.
// @Tags         Rates
// @Accept       json
// @Produce      json
// @Param        base   query     string  false  "Base currency (3-letter code, e.g., USD)"
// @Success      200    {object}  map[string]interface{}
// @Failure      400    {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /rates [get]
func (h *RateController) GetRates(c *gin.Context) {
	base := strings.ToUpper(strings.TrimSpace(c.Query("base")))
	if base != "" && !isCurrency(base) {
		response.BadRequest(c, "validation_error", "base must be a 3-letter currency code")
		return
	}
	baseOut, rates, updatedAt, err := h.rates.GetRates(base)
	if err != nil {
		response.BadRequest(c, "rates_unavailable", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"base":       baseOut,
		"rates":      rates,
		"updated_at": updatedAt,
	})
}

// ConvertCurrency godoc
// @Summary      Convert Currency
// @Description  Convert an amount from one currency to another using the latest exchange rates.
// @Tags         Rates
// @Accept       json
// @Produce      json
// @Param        from    query     string  true  "Source currency (3-letter code, e.g., USD)"
// @Param        to      query     string  true  "Target currency (3-letter code, e.g., THB)"
// @Param        amount  query     number  true  "Amount to convert (non-negative)"
// @Success      200     {object}  map[string]interface{}
// @Failure      400     {object}  response.ErrorResponse
// @Security     BearerAuth
// @Router       /convert [get]
func (h *RateController) ConvertCurrency(c *gin.Context) {
	from := strings.ToUpper(strings.TrimSpace(c.Query("from")))
	to := strings.ToUpper(strings.TrimSpace(c.Query("to")))
	amountS := c.Query("amount")

	if !isCurrency(from) || !isCurrency(to) {
		response.BadRequest(c, "validation_error", "from and to must be 3-letter currency codes")
		return
	}
	if amountS == "" {
		response.BadRequest(c, "validation_error", "amount is required")
		return
	}
	amount, err := strconv.ParseFloat(amountS, 64)
	if err != nil || amount < 0 {
		response.BadRequest(c, "validation_error", "amount must be a non-negative number")
		return
	}

	rate, result, updatedAt, err := h.rates.Convert(from, to, amount)
	if err != nil {
		response.BadRequest(c, "conversion_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"from":       from,
		"to":         to,
		"amount":     amount,
		"rate":       rate,
		"result":     result,
		"updated_at": updatedAt,
	})
}

func isCurrency(s string) bool {
	if len(s) != 3 {
		return false
	}
	for _, ch := range s {
		if ch < 'A' || ch > 'Z' {
			return false
		}
	}
	return true
}
