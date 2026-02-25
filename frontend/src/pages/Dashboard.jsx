import { useState, useEffect } from 'react';
import { fetchDashboardMetrics } from '../api/client';
import './Dashboard.css';

function Dashboard() {
  const [metrics, setMetrics] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [range, setRange] = useState('24h');

  useEffect(() => {
    loadMetrics();
  }, [range]);

  const loadMetrics = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await fetchDashboardMetrics(range);
      setMetrics(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="page-container">
        <div className="loading">Loading dashboard...</div>
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
        <h2>Dashboard</h2>
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

      {/* Metrics Cards */}
      <div className="metrics-grid">
        <div className="metric-card">
          <div className="metric-label">Total Logs</div>
          <div className="metric-value">{metrics.total_logs.toLocaleString()}</div>
        </div>
        <div className="metric-card">
          <div className="metric-label">Logs/Second</div>
          <div className="metric-value">{metrics.logs_per_sec.toFixed(2)}</div>
        </div>
        <div className="metric-card error-card">
          <div className="metric-label">Errors</div>
          <div className="metric-value">{metrics.error_count.toLocaleString()}</div>
        </div>
        <div className="metric-card warning-card">
          <div className="metric-label">Warnings</div>
          <div className="metric-value">{metrics.warning_count.toLocaleString()}</div>
        </div>
      </div>

      {/* Top Lists */}
      <div className="lists-grid">
        <div className="list-card">
          <h3>Top Hosts</h3>
          <div className="list-items">
            {metrics.top_hosts.map((item, idx) => (
              <div key={idx} className="list-item">
                <span className="item-name">{item.name}</span>
                <span className="item-count">{item.count.toLocaleString()}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="list-card">
          <h3>Top Processes</h3>
          <div className="list-items">
            {metrics.top_processes.map((item, idx) => (
              <div key={idx} className="list-item">
                <span className="item-name">{item.name}</span>
                <span className="item-count">{item.count.toLocaleString()}</span>
              </div>
            ))}
          </div>
        </div>

        <div className="list-card">
          <h3>Transport Split</h3>
          <div className="list-items">
            {metrics.transport_split.map((item, idx) => (
              <div key={idx} className="list-item">
                <span className="item-name">{item.name}</span>
                <span className="item-count">{item.count.toLocaleString()}</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

export default Dashboard;
