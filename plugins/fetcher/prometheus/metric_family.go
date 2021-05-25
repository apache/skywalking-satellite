package prometheus

type MetricFamily struct {
	Name    string
	Help    string
	Type    string
	Metrics []Metric
}

type Metric struct {
	Name      string
	Labels    map[string]string
	Timestamp int64
}
