package repositories

import (
	"time"

	"github.com/spksupakorn/Currency-Converter/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type RateRepository interface {
	UpsertRates(rates map[string]float64, now time.Time) error
	GetAllRates() (map[string]float64, time.Time, error)
}

type rateRepository struct {
	db *gorm.DB
}

func NewRateRepository(db *gorm.DB) RateRepository {
	return &rateRepository{db: db}
}

func (r *rateRepository) UpsertRates(rates map[string]float64, now time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for cur, val := range rates {
			rt := models.Rate{
				Currency:  cur,
				Rate:      val,
				UpdatedAt: now,
			}
			if err := tx.Clauses(
				clause.OnConflict{
					Columns:   []clause.Column{{Name: "currency"}},
					DoUpdates: clause.AssignmentColumns([]string{"rate", "updated_at"}),
				},
			).Create(&rt).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *rateRepository) GetAllRates() (map[string]float64, time.Time, error) {
	var rows []models.Rate
	if err := r.db.Find(&rows).Error; err != nil {
		return nil, time.Time{}, err
	}
	out := make(map[string]float64, len(rows))
	var latest time.Time
	for _, rr := range rows {
		out[rr.Currency] = rr.Rate
		if rr.UpdatedAt.After(latest) {
			latest = rr.UpdatedAt
		}
	}
	return out, latest, nil
}
