package models

import "time"

//map out structs here

// what a normal transaction consists of
type Transaction struct {
	ID        string    `json:"id"`
	SKU       string    `json:"sku"`
	Region    string    `json:"region"`
	Revenue   float64   `json:"revenue"`
	COGS      float64   `json:"cogs"` //cost of goods sold
	Timestamp time.Time `json:"timestamp"`
}

// KPISnapshot is the aggregated payload sent down the WebSocket
type KPISnapshot struct {
	TotalRevenue float64 `json:"total_revenue"`
	TotalCOGS    float64 `json:"total_cogs"`
	GrossMargin  float64 `json:"gross_margin_percentage"`
	OrderCount   int     `json:"order_count"`
	AverageOrder float64 `json:"average_order_value"`
	AlertMsg     string  `json:"alert_msg,omitempty"` // populated in case of an anomaly being detected
}
