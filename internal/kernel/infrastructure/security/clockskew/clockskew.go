package clockskew

import (
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var histogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "jwt_clock_skew_seconds",
		Help:    "Observed clock skew between token iat and server time",
		Buckets: prometheus.ExponentialBuckets(0.1, 2, 15), // 0.1s to ~1638s
	},
	[]string{"integration"},
)

func init() {
	prometheus.MustRegister(histogram)
}

// Observe records the skew between the token's IssuedAt and now.
func Observe(integration string, issuedAt time.Time) {
	skew := math.Abs(time.Since(issuedAt).Seconds())
	histogram.WithLabelValues(integration).Observe(skew)
}
