# Real-Time Sales & Financial KPI Streaming Pipeline

A full-stack streaming architecture built to ingest, aggregate, and visualize live transactional data. This project bridges software engineering and **Business Informatics**, demonstrating how to move data from a point-of-sale system, process it in-memory for zero-latency analytics, and deliver actionable insights to an executive dashboard.

## The Business Problem
Business leaders often rely on stale data (daily or weekly batch jobs) to make decisions. This system provides a **sub-second, real-time window** into operational health. It dynamically calculates rolling revenue, gross margins, and utilizes statistical anomaly detection to immediately flag highly unusual transactions (e.g., enterprise bulk orders or massive refund spikes) to prevent operational bottlenecks.

## System Architecture

```mermaid
graph LR
    A[Python Simulator<br/>Event Producer] -->|HTTP POST| B(Go Backend<br/>REST API)
    B --> C[(In-Memory<br/>Ring Buffer)]
    C -->|Aggregates KPIs| D{Gorilla WebSocket<br/>Hub Broadcast}
    D -->|Push JSON State| E[React + TypeScript<br/>Executive Dashboard]

### Core Components
* **The Data Generator (Python):** Simulates a high-throughput e-commerce backend, injecting realistic baseline sales data and probabilistically triggering massive statistical anomalies.
* **The Ingestion Engine (Go):** Receives payloads concurrently and manages state.
* **The State Manager (Go Ring Buffer):** Uses a thread-safe (`sync.RWMutex`) circular buffer to store the last $N$ transactions. This provides an **$O(1)$ memory footprint** and enables microsecond calculations of moving averages and rolling margins without querying a database.
* **The Real-Time Broadcaster (Go WebSockets):** Instantly multiplexes the newly calculated KPI state to all connected dashboard clients simultaneously.
* **The Executive Dashboard (React/Vite):** A responsive, component-driven UI built with Tailwind CSS and Recharts to visualize the live data stream and maintain an active alert feed.

## 🛠️ Tech Stack
* **Backend:** Go (Golang), Gorilla WebSockets
* **Frontend:** React, TypeScript, Vite, Tailwind CSS v4, Recharts, Lucide Icons
* **Data Simulation:** Python (`requests`)

## 🚀 Quick Start: Run Locally
To run this pipeline on your machine, you will need to open three separate terminal windows.

### 1. Start the Go Backend
The Go server listens for incoming transactions and handles the WebSocket broadcasting.
```bash
# Terminal 1
cd Real-timesales
go run cmd/server/main.go

Expected Output: Server starting on http://localhost:8080

### 2. Start the React Frontend
The Vite server hosts the dashboard and connects to the Go WebSocket.
```bash
# Terminal 2
cd Real-timesales/frontend
npm install
npm run dev

Expected Output: Server starting on http://localhost:8080Expected Output: Open http://localhost:5173 in your browser.

### 3. Start the Python Simulator
The Python script acts as the checkout system, firing synthetic data into the pipeline.
```bash
# Terminal 3
cd Real-timesales
pip install requests
python simulator.py

Expected Output: You will see JSON payloads being sent, and the React dashboard will immediately come to life.

## Key Metric Definitions
* **Rolling Revenue:** The total gross revenue of the active window (last $N$ transactions).
* **Gross Margin %:** `(Total Revenue - Cost of Goods Sold) / Total Revenue`.
* **Dynamic Anomaly Detection:** Flags any transaction where the revenue exceeds 3x the current rolling Average Order Value (AOV).

##  Future Enhancements
- [ ] **Dockerization:** Wrap the entire system in a `docker-compose.yml` for single-command startup.
- [ ] **Persistent Storage:** Attach PostgreSQL or DuckDB to the Go ingestion route for historical querying.
- [ ] **Isolation Forests:** Move from a simple moving average threshold to a Machine Learning anomaly detection algorithm for multi-variate outliers.



