package models

type MetricEnvelope struct {
	Metric       Metric            `json:"metric"`
	Meta         map[string]string `json:"meta"`
	CreationTime int64             `json:"creation_time"`
}
