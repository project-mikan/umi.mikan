package diary

import "github.com/prometheus/client_golang/prometheus"

var (
	// semanticSearchRequestsCounter セマンティック検索リクエスト総数
	semanticSearchRequestsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "backend_semantic_search_requests_total",
			Help: "Total number of semantic search requests",
		},
		[]string{"status"},
	)
	// semanticSearchDuration セマンティック検索処理時間
	semanticSearchDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "backend_semantic_search_duration_seconds",
			Help:    "Duration of semantic search requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)
	// semanticSearchResultsCount セマンティック検索結果件数
	semanticSearchResultsCount = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "backend_semantic_search_results_count",
			Help:    "Number of results returned per semantic search request",
			Buckets: []float64{0, 1, 3, 5, 10, 20, 30, 50},
		},
		[]string{},
	)
)

func init() {
	prometheus.MustRegister(semanticSearchRequestsCounter)
	prometheus.MustRegister(semanticSearchDuration)
	prometheus.MustRegister(semanticSearchResultsCount)
}
