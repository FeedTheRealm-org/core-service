package dtos

type GemMetricsResponse struct {
	GemsBought  int64   `json:"gems_bought"`
	GemsSpent   int64   `json:"gems_spent"`
	GemsRevenue float64 `json:"gems_revenue"`
	GemsFlow    float64 `json:"gems_flow"`
}
