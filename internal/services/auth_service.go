package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spksupakorn/Currency-Converter/config"
	"github.com/spksupakorn/Currency-Converter/pkg/logger"

	"github.com/spksupakorn/Currency-Converter/internal/repositories"
)

type RateService interface {
	StartBackgroundRefresh(ctx context.Context)
	GetRates(base string) (baseOut string, rates map[string]float64, updatedAt time.Time, err error)
	Convert(from, to string, amount float64) (rate float64, result float64, updatedAt time.Time, err error)
}

type rateService struct {
	cfg    config.Config
	repo   repositories.RateRepository
	log    *logger.Logger
	client *http.Client

	mu         sync.RWMutex
	cacheBase  string
	cacheRates map[string]float64
	cacheAt    time.Time
}

func NewRateService(cfg config.Config, repo repositories.RateRepository, log *logger.Logger) RateService {
	return &rateService{
		cfg:    cfg,
		repo:   repo,
		log:    log,
		client: &http.Client{Timeout: cfg.HTTPClientTimeout},
	}
}

func (s *rateService) StartBackgroundRefresh(ctx context.Context) {
	// Initial load
	go func() {
		if err := s.refresh(ctx); err != nil {
			s.log.Error("initial rate refresh failed", logger.Fields{"error": err.Error()})
		}
	}()

	ticker := time.NewTicker(s.cfg.RateRefreshInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := s.refresh(ctx); err != nil {
					s.log.Error("rate refresh failed", logger.Fields{"error": err.Error()})
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (s *rateService) refresh(ctx context.Context) error {
	base := s.cfg.RateBaseCurrency
	base = strings.ToUpper(strings.TrimSpace(base))
	if base == "" {
		base = "USD"
	}

	url := fmt.Sprintf("https://api.exchangerate.host/latest?base=%s", base)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("rates api error: %s", resp.Status)
	}

	var out struct {
		Base  string             `json:"base"`
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}

	// Ensure base has rate 1.0 in the map
	if out.Rates == nil {
		out.Rates = map[string]float64{}
	}
	out.Rates[out.Base] = 1.0

	now := time.Now().UTC()
	if err := s.repo.UpsertRates(out.Rates, now); err != nil {
		return err
	}

	s.mu.Lock()
	s.cacheBase = out.Base
	s.cacheRates = out.Rates
	s.cacheAt = now
	s.mu.Unlock()

	s.log.Info("rates refreshed", logger.Fields{"base": out.Base, "count": len(out.Rates)})
	return nil
}

func (s *rateService) GetRates(base string) (string, map[string]float64, time.Time, error) {
	base = normalizeCurrency(base)
	s.mu.RLock()
	rates := s.cacheRates
	cacheBase := s.cacheBase
	cacheAt := s.cacheAt
	s.mu.RUnlock()

	if len(rates) == 0 {
		// Fallback to DB
		dbRates, updatedAt, err := s.repo.GetAllRates()
		if err != nil || len(dbRates) == 0 {
			return "", nil, time.Time{}, errors.New("rates are not available yet")
		}
		rates = dbRates
		cacheBase = s.cfg.RateBaseCurrency
		cacheAt = updatedAt
	}

	// If requested base equals cache base, return as-is
	if base == "" || base == cacheBase {
		// Copy to avoid mutation
		out := make(map[string]float64, len(rates))
		for k, v := range rates {
			out[k] = v
		}
		out[cacheBase] = 1.0
		return cacheBase, out, cacheAt, nil
	}

	// Derive rates for requested base: rate(base->X) = rate(cacheBase->X) / rate(cacheBase->base)
	baseRate, ok := rates[base]
	if !ok || baseRate == 0 {
		return "", nil, time.Time{}, fmt.Errorf("unsupported base currency: %s", base)
	}
	out := make(map[string]float64, len(rates))
	for cur, r := range rates {
		out[cur] = r / baseRate
	}
	out[base] = 1.0
	return base, out, cacheAt, nil
}

func (s *rateService) Convert(from, to string, amount float64) (float64, float64, time.Time, error) {
	from = normalizeCurrency(from)
	to = normalizeCurrency(to)
	if from == "" || to == "" {
		return 0, 0, time.Time{}, errors.New("from and to currencies are required")
	}
	if amount < 0 {
		return 0, 0, time.Time{}, errors.New("amount must be non-negative")
	}

	s.mu.RLock()
	rates := s.cacheRates
	cacheAt := s.cacheAt
	s.mu.RUnlock()

	if len(rates) == 0 {
		dbRates, updatedAt, err := s.repo.GetAllRates()
		if err != nil || len(dbRates) == 0 {
			return 0, 0, time.Time{}, errors.New("exchange rates are not available")
		}
		rates = dbRates
		cacheAt = updatedAt
	}

	// Convert via cacheBase: rate(from->to) = (rate(cacheBase->to) / rate(cacheBase->from))
	rFrom, okFrom := rates[from]
	rTo, okTo := rates[to]
	if !okFrom || rFrom == 0 {
		return 0, 0, time.Time{}, fmt.Errorf("unsupported currency: %s", from)
	}
	if !okTo {
		return 0, 0, time.Time{}, fmt.Errorf("unsupported currency: %s", to)
	}

	rate := rTo / rFrom
	result := amount * rate
	return rate, result, cacheAt, nil
}

func normalizeCurrency(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}
