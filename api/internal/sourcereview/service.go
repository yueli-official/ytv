package sourcereview

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/yueli-official/ytv/api/internal/source"
	"slices"
	"strings"
	"time"
)

// Review is one viewer's current assessment, not a verified quality guarantee.
type Review struct {
	Ads       string    `json:"ads"`
	Quality   string    `json:"quality"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Summary struct {
	Total   int `json:"total"`
	AdsFew  int `json:"adsFew"`
	AdsMany int `json:"adsMany"`
	HD      int `json:"hd"`
	Normal  int `json:"normal"`
	Blurry  int `json:"blurry"`
}

type Store interface {
	Get(context.Context, string, string) (Review, error)
	Save(context.Context, string, string, Review) error
	Summaries(context.Context) (map[string]Summary, error)
}
type Service struct{ store Store }

func New(store Store) *Service { return &Service{store: store} }
func (s *Service) Get(ctx context.Context, user, key string) (Review, error) {
	return s.store.Get(ctx, user, key)
}
func (s *Service) Summaries(ctx context.Context) (map[string]Summary, error) {
	return s.store.Summaries(ctx)
}
func (s *Service) Save(ctx context.Context, user, key, ads, quality string) (Review, error) {
	if strings.TrimSpace(user) == "" || key == "" || !slices.Contains([]string{"", "few", "many"}, ads) || !slices.Contains([]string{"", "hd", "normal", "blurry"}, quality) {
		return Review{}, fmt.Errorf("%w: invalid source assessment", source.ErrInvalid)
	}
	review := Review{Ads: ads, Quality: quality, UpdatedAt: time.Now().UTC()}
	if err := s.store.Save(ctx, user, key, review); err != nil {
		return Review{}, err
	}
	return s.store.Get(ctx, user, key)
}

type PostgresStore struct{ db *sql.DB }

func NewPostgresStore(db *sql.DB) *PostgresStore { return &PostgresStore{db: db} }
func (s *PostgresStore) Get(ctx context.Context, user, key string) (Review, error) {
	var r Review
	err := s.db.QueryRowContext(ctx, `SELECT ads,quality,updated_at FROM ytv_source_reviews WHERE user_key=$1 AND source_key=$2`, user, key).Scan(&r.Ads, &r.Quality, &r.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Review{}, nil
	}
	return r, err
}
func (s *PostgresStore) Save(ctx context.Context, user, key string, r Review) error {
	// Atomic replacement and a composite primary key prevent retries or concurrent submissions from inflating counts.
	if r.Ads == "" && r.Quality == "" {
		_, err := s.db.ExecContext(ctx, `DELETE FROM ytv_source_reviews WHERE user_key=$1 AND source_key=$2`, user, key)
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO ytv_source_reviews(user_key,source_key,ads,quality,updated_at) VALUES($1,$2,$3,$4,$5)
 ON CONFLICT(user_key,source_key) DO UPDATE SET ads=EXCLUDED.ads,quality=EXCLUDED.quality,updated_at=EXCLUDED.updated_at`, user, key, r.Ads, r.Quality, r.UpdatedAt)
	return err
}
func (s *PostgresStore) Summaries(ctx context.Context) (map[string]Summary, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT source_key,count(*),count(*) FILTER(WHERE ads='few'),count(*) FILTER(WHERE ads='many'),count(*) FILTER(WHERE quality='hd'),count(*) FILTER(WHERE quality='normal'),count(*) FILTER(WHERE quality='blurry') FROM ytv_source_reviews GROUP BY source_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]Summary{}
	for rows.Next() {
		var key string
		var v Summary
		if err := rows.Scan(&key, &v.Total, &v.AdsFew, &v.AdsMany, &v.HD, &v.Normal, &v.Blurry); err != nil {
			return nil, err
		}
		result[key] = v
	}
	return result, rows.Err()
}
