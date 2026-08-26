import { useEffect, useState } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { AlertCircle, DollarSign, ShoppingCart, Percent } from 'lucide-react';

interface KPISnapshot {
  total_revenue: number;
  gross_margin_percentage: number;
  order_count: number;
  average_order_value: number;
  alert_msg?: string;
}

interface ChartDataPoint {
  time: string;
  revenue: number;
}

// 1. New interface for our anomaly history
interface AnomalyLog {
  id: string;
  time: string;
  msg: string;
}

export default function App() {
  const [kpi, setKpi] = useState<KPISnapshot | null>(null);
  const [chartData, setChartData] = useState<ChartDataPoint[]>([]);
  
  // 2. New state to hold the history of anomalies
  const [anomalies, setAnomalies] = useState<AnomalyLog[]>([]);
  const [isConnected, setIsConnected] = useState(false);

  useEffect(() => {
    const ws = new WebSocket('ws://localhost:8080/ws');

    ws.onopen = () => setIsConnected(true);
    ws.onclose = () => setIsConnected(false);

    ws.onmessage = (event) => {
      const data: KPISnapshot = JSON.parse(event.data);
      setKpi(data);

      setChartData((prev) => {
        const newData = [...prev, { 
          time: new Date().toLocaleTimeString(), 
          revenue: data.total_revenue 
        }];
        return newData.slice(-20); 
      });

      // 3. If there is an alert, add it to our history
      if (data.alert_msg) {
        setAnomalies((prev) => {
          // Prevent spam: only add if it's a DIFFERENT message than the most recent one
          if (prev.length > 0 && prev[0].msg === data.alert_msg) {
            return prev;
          }
          const newLog = {
          id: Date.now().toString(),
          time: new Date().toLocaleTimeString(),
          msg: data.alert_msg as string // <-- Tell TypeScript this is definitely a string
        };
          // Keep the last 50 logs so the browser doesn't run out of memory
          return [newLog, ...prev].slice(0, 50);
        });
      }
    };

    return () => ws.close();
  }, []);

  if (!kpi) {
    return <div className="p-8 text-slate-500">Waiting for live sales data...</div>;
  }

  return (
    <div className="min-h-screen bg-slate-50 p-8 font-sans">
      <header className="mb-8 flex items-center justify-between">
        <h1 className="text-3xl font-bold text-slate-800">Live Sales Operations</h1>
        <div className={`px-3 py-1 rounded-full text-sm font-medium ${isConnected ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
          {isConnected ? 'Live (WebSocket connected)' : 'Disconnected'}
        </div>
      </header>

      {/* KPI Cards */}
      <div className="mb-8 grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <KpiCard title="Rolling Revenue" value={`$${kpi.total_revenue.toFixed(2)}`} icon={<DollarSign />} />
        <KpiCard title="Gross Margin" value={`${kpi.gross_margin_percentage.toFixed(1)}%`} icon={<Percent />} />
        <KpiCard title="Active Orders" value={kpi.order_count.toString()} icon={<ShoppingCart />} />
        <KpiCard title="Avg Order Value" value={`$${kpi.average_order_value.toFixed(2)}`} icon={<DollarSign />} />
      </div>

      {/* 4. New Grid Layout: Chart takes up 2/3, Anomaly Log takes 1/3 */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        
        {/* Live Chart */}
        <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm lg:col-span-2">
          <h2 className="mb-6 text-lg font-semibold text-slate-700">Real-Time Revenue Trend</h2>
          <div className="h-72">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e2e8f0" />
                <XAxis dataKey="time" stroke="#64748b" fontSize={12} tickLine={false} />
                <YAxis stroke="#64748b" fontSize={12} tickLine={false} axisLine={false} tickFormatter={(val) => `$${val}`} />
                <Tooltip contentStyle={{ borderRadius: '8px', border: 'none', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)' }} />
                <Line type="monotone" dataKey="revenue" stroke="#3b82f6" strokeWidth={3} dot={false} animationDuration={300} />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Anomaly Log Feed */}
        <div className="flex h-full flex-col rounded-xl border border-slate-200 bg-white shadow-sm overflow-hidden">
          <div className="border-b border-slate-100 bg-slate-50/50 p-4">
            <h2 className="text-lg font-semibold text-slate-700">Anomaly Feed</h2>
          </div>
          
          <div className="flex-1 overflow-y-auto p-4" style={{ maxHeight: '312px' }}>
            {anomalies.length === 0 ? (
              <div className="flex h-full items-center justify-center text-sm text-slate-400">
                No anomalies detected recently.
              </div>
            ) : (
              <div className="space-y-3">
                {anomalies.map((log) => (
                  <div key={log.id} className="flex flex-col gap-1 rounded-lg bg-red-50 p-3 text-sm border border-red-100">
                    <div className="flex items-center justify-between text-red-500">
                      <div className="flex items-center gap-1 font-semibold">
                        <AlertCircle className="h-4 w-4" />
                        <span>Alert</span>
                      </div>
                      <span className="text-xs opacity-80">{log.time}</span>
                    </div>
                    <p className="text-red-700 font-medium">{log.msg}</p>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>

      </div>
    </div>
  );
}

function KpiCard({ title, value, icon }: { title: string, value: string, icon: React.ReactNode }) {
  return (
    <div className="flex items-center gap-4 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-blue-50 text-blue-600">
        {icon}
      </div>
      <div>
        <p className="text-sm font-medium text-slate-500">{title}</p>
        <p className="text-2xl font-bold text-slate-800">{value}</p>
      </div>
    </div>
  );
}