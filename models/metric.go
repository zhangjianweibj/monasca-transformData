package models

type Metric struct {
	Name       string            `json:"name"`
	Dimensions map[string]string `json:"dimensions"`
	Timestamp  float64           `json:"timestamp"`
	Value      float64           `json:"value"`
	ValueMeta  map[string]string `json:"value_meta"`
}
