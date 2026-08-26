package state

import (
	"Real-timesales/internal/models"
	"fmt"
	"sync"
)

//ring buffer and its functionality defined here

type RingBuffer struct {
	mu    sync.RWMutex         //control and locks
	data  []models.Transaction //whats in the ring buffer
	size  int                  //capacity
	head  int                  //where the NEXT transaction will go
	count int                  //number of items in the buffer
}

// pre allocates space for the ring buffer
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]models.Transaction, size),
		size: size,
	}
}

// Add inserts a new transaction and overwrites the oldest element if the ring buffer is full
func (b *RingBuffer) Add(t models.Transaction) {
	b.mu.Lock()         //put a write lock on during adding
	defer b.mu.Unlock() //schedule the unlock once everything is over

	b.data[b.head] = t             //insert at head (itll overwrite if full)
	b.head = (b.head + 1) % b.size //increment and wrap

	if b.count < b.size { //if we're within bounds just increment normally
		b.count++
	}
}

// GetMetrics will lock the buffer for reading and calculates live aggregated KPIs across the window
func (b *RingBuffer) GetMetrics() models.KPISnapshot {
	//put on a read lock and defer the unlock while we read to get metrics
	b.mu.RLock()
	defer b.mu.RUnlock()

	var snapshot models.KPISnapshot

	if b.count == 0 {
		return snapshot //if theres nothing return an empty snapshot
	}

	//when it comes to calculating the aggregates we can just count through b.count items since the chronological order doesnt matter
	for i := 0; i < b.count; i++ {
		tx := b.data[i]
		snapshot.TotalRevenue += tx.Revenue
		snapshot.TotalCOGS += tx.COGS
	}

	snapshot.OrderCount = b.count

	//find Average order value AOV = total revenue/total orders
	if snapshot.OrderCount > 0 {
		snapshot.AverageOrder = snapshot.TotalRevenue / float64(snapshot.OrderCount)
	}

	//find Gross Margin Percentage
	//make sure about division against 0
	// Gross profit = revenue - COGS
	// margin percentage = (gross profit / revenue) *100
	if snapshot.TotalRevenue > 0 {
		snapshot.GrossMargin = ((snapshot.TotalRevenue - snapshot.TotalCOGS) / snapshot.TotalRevenue) * 100
	}

	//locate most recent transaction
	latestindex := (b.head - 1 + b.size) % b.size
	latestTx := b.data[latestindex]

	//anomaly detection (spike alert)
	//check if the latest order is an outlier(>3x the window average) and requires ordercount >10 for meaningful sample size
	if snapshot.OrderCount > 10 && latestTx.Revenue > (snapshot.AverageOrder*3) {
		snapshot.AlertMsg = fmt.Sprintf(" Anomaly!!!!: Unusually large order (%.2f) detected in %s!", latestTx.Revenue, latestTx.Region)
	}
	//anomaly detection (negative revenue/ refund alert)
	//flag the returns or chargebacks
	if latestTx.Revenue < 0 {
		snapshot.AlertMsg = fmt.Sprintf("Alert!!!!: Refund processed in %s.", latestTx.Region)
	}
	return snapshot
}
