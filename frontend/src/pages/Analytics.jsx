import { useState, useEffect } from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { 
  fetchIngestionTrend, 
  fetchErrorTrend, 
  fetchWarningTrend, 
  fetchAlertTrend 
} from '../api/client';
import './Analytics.css';

function Analytics() {
  const [range, setRange] = useState('24h');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [data, setData] = useState({
    ingestion: [],
    errors: [],
    warnings: [],
    alerts: []
  });

  useEffect(() => {
    loadAnalytics();
  }, [range]);

  const loadAnalytics = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const [ingestion, errors, warnings, alerts] = await Promise.all([
        fetchIngestionTrend(range),
        fetchErrorTrend(range),
        fetchWarningTrend(range),
        fetchAlertTrend(range)
      ]);

      setData({
        ingestion: formatChartData(ingestion),
        errors: formatChartData(errors),
        warnings: formatChartData(warnings),
        alerts: formatChartData(alerts)
      });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const formatChartData = (data) => {
    return data.map(point => ({
      time: new Date(point.time).toLocaleTimeString('en-US', { 
        hour: '2-digit', 
        minute: '2-digit' 
      }),
      count: point.count
    }));
  };

  if (loading) {
    return (
      <div className="page-container">
        <div className="loading">Loading analytics...</div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="page-container">
        <div className="error">Error: {error}</div>
      </div>
    );
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <h2>Analytics</h2>
        <div className="range-selector">
          <button
            className={`range-btn ${range === '24h' ? 'active' : ''}`}
            onClick={() => setRange('24h')}
          >
            Last 24 Hours
          </button>
          <button
            className={`range-btn ${range === 'all' ? 'active' : ''}`}
            onClick={() => setRange('all')}
          >
            All Time
          </button>
        </div>
      </div>

      <div className="charts-grid">
        {/* Ingestion Trend */}
        <div className="chart-card">
          <h3>Log Ingestion Trend</h3>
          <ResponsiveContainer width="100%" height={250}>
            <LineChart data={data.ingestion}>
              <CartesianGrid strokeDasharray="3 3" stroke="#1e3a8a" />
              <XAxis 
                dataKey="time" 
                stroke="#9ca3af" 
                tick={{ fill: '#9ca3af', fontSize: 12 }}
              />
              <YAxis 
                stroke="#9ca3af" 
                tick={{ fill: '#9ca3af', fontSize: 12 }}
              />
              <Tooltip 
                contentStyle={{ 
                  backgroundColor: '#0a0a0a', 
                  border: '1px solid #1e3a8a',
                  borderRadius: '4px'
                }}
                labelStyle={{ color: '#9ca3af' }}
                itemStyle={{ color: '#3b82f6' }}
              />
              <Line 
                type="monotone" 
                dataKey="count" 
                stroke="#3b82f6" 
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Error Trend */}
        <div className="chart-card">
          <h3>Error Trend</h3>
          <ResponsiveContainer width="100%" height={250}>
            <LineChart data={data.errors}>
              <CartesianGrid strokeDasharray="3 3" stroke="#1e3a8a" />
              <XAxis 
                dataKey="time" 
                stroke="#9ca3af" 
                tick={{ fill: '#9ca3af', fontSize: 12 }}
              />
              <YAxis 
                stroke="#9ca3af" 
                tick={{ fill: '#9ca3af', fontSize: 12 }}
              />
              <Tooltip 
                contentStyle={{ 
                  backgroundColor: '#0a0a0a', 
                  border: '1px solid #1e3a8a',
                  borderRadius: '4px'
                }}
                labelStyle={{ color: '#9ca3af' }}
                itemStyle={{ color: '#ef4444' }}
              />
              <Line 
                type="monotone" 
                dataKey="count" 
                stroke="#ef4444" 
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Warning Trend */}
        <div className="chart-card">
          <h3>Warning Trend</h3>
          <ResponsiveContainer width="100%" height={250}>
            <LineChart data={data.warnings}>
              <CartesianGrid strokeDasharray="3 3" stroke="#1e3a8a" />
              <XAxis 
                dataKey="time" 
                stroke="#9ca3af" 
                tick={{ fill: '#9ca3af', fontSize: 12 }}
              />
              <YAxis 
                stroke="#9ca3af" 
                tick={{ fill: '#9ca3af', fontSize: 12 }}
              />
              <Tooltip 
                contentStyle={{ 
                  backgroundColor: '#0a0a0a', 
                  border: '1px solid #1e3a8a',
                  borderRadius: '4px'
                }}
                labelStyle={{ color: '#9ca3af' }}
                itemStyle={{ color: '#f59e0b' }}
              />
              <Line 
                type="monotone" 
                dataKey="count" 
                stroke="#f59e0b" 
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Alert Trend */}
        <div className="chart-card">
          <h3>Alert Trend</h3>
          <ResponsiveContainer width="100%" height={250}>
            <LineChart data={data.alerts}>
              <CartesianGrid strokeDasharray="3 3" stroke="#1e3a8a" />
              <XAxis 
                dataKey="time" 
                stroke="#9ca3af" 
                tick={{ fill: '#9ca3af', fontSize: 12 }}
              />
              <YAxis 
                stroke="#9ca3af" 
                tick={{ fill: '#9ca3af', fontSize: 12 }}
              />
              <Tooltip 
                contentStyle={{ 
                  backgroundColor: '#0a0a0a', 
                  border: '1px solid #1e3a8a',
                  borderRadius: '4px'
                }}
                labelStyle={{ color: '#9ca3af' }}
                itemStyle={{ color: '#10b981' }}
              />
              <Line 
                type="monotone" 
                dataKey="count" 
                stroke="#10b981" 
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
}

export default Analytics;
