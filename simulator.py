import time
import random
import requests
from datetime import datetime, timezone

# URL of the GO backend
URL = "http://localhost:8080/ingest"
REGIONS = ["EU-West", "US-East", "AP-South"]

print("Starting synthetic sales generator...")
print(f"Targeting backend at: {URL}")
print("Press Ctrl+C to stop.")
print("-" * 30)

while True:
    #generate realistic base data
    revenue = round(random.uniform(10.0, 150.0), 2)
    # COGSis typically 30 to 70 percent of the revenue
    cogs = round(revenue * random.uniform(0.3, 0.7), 2) 
    
    # inject a anomaly with 5 percent probability
    is_anomaly = random.random() < 0.05
    if is_anomaly:
        revenue = round(revenue * 15, 2) 
        print("\n--> 🚨 Injecting massive anomaly order!\n")

    # build the JSON payload matching the models.transaction struct
    payload = {
        "id": f"tx_{int(time.time()*1000)}",
        "sku": f"ITEM-{random.randint(100, 999)}",
        "region": random.choice(REGIONS),
        "revenue": revenue,
        "cogs": cogs,
        "timestamp": datetime.now(timezone.utc).isoformat()
    }
    
    # POST request to the GO backend
    try:
        response = requests.post(URL, json=payload)
        if response.status_code == 200:
            print(f"Sent: ${payload['revenue']:<7} from {payload['region']}")
        else:
            print(f"Server rejected payload with status: {response.status_code}")
    except requests.exceptions.ConnectionError:
        print("Failed to connect. Is the Go server running on port 8080?")
        
    # 5. Pause randomly between 0.2 and 1.5 seconds to simulate real organic traffic
    time.sleep(random.uniform(0.2, 1.5))