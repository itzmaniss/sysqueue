package metrics

type Metrics struct {
	JobsProcessed int64
	JobsFailed    int64
}

func NewMetrics() *Metrics {
	return &Metrics{
		JobsProcessed: 0,
		JobsFailed:    0,
	}
}
