package db

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
)

var pgQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "pg_query_duration_seconds",
	Help:    "PostgreSQL query duration in seconds",
	Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
}, []string{"operation"})

// RegisterMetrics registers PostgreSQL connection pool metrics and GORM query
// duration callbacks. Call once after NewDb.
func (db *Db) RegisterMetrics() {
	sqlDB, err := db.DB.DB()
	if err == nil {
		prometheus.MustRegister(collectors.NewDBStatsCollector(sqlDB, "link_generator"))
	}

	db.Callback().Query().Before("gorm:query").Register("metrics:before_query", startTimer)
	db.Callback().Query().After("gorm:query").Register("metrics:after_query", makeObserver("query"))

	db.Callback().Create().Before("gorm:create").Register("metrics:before_create", startTimer)
	db.Callback().Create().After("gorm:create").Register("metrics:after_create", makeObserver("create"))

	db.Callback().Update().Before("gorm:update").Register("metrics:before_update", startTimer)
	db.Callback().Update().After("gorm:update").Register("metrics:after_update", makeObserver("update"))

	db.Callback().Delete().Before("gorm:delete").Register("metrics:before_delete", startTimer)
	db.Callback().Delete().After("gorm:delete").Register("metrics:after_delete", makeObserver("delete"))
}

func startTimer(d *gorm.DB) {
	d.InstanceSet("metrics:start", time.Now())
}

func makeObserver(op string) func(*gorm.DB) {
	return func(d *gorm.DB) {
		v, ok := d.InstanceGet("metrics:start")
		if !ok {
			return
		}
		pgQueryDuration.WithLabelValues(op).Observe(time.Since(v.(time.Time)).Seconds())
	}
}
