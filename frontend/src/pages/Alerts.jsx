import { useState, useEffect } from 'react';
import { fetchAlerts, updateAlertStatus, deleteAlert } from '../api/client';
import './Alerts.css';

function Alerts() {
  const [alerts, setAlerts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filter, setFilter] = useState('all');

  useEffect(() => {
    loadAlerts();
  }, []);

  const loadAlerts = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await fetchAlerts();
      setAlerts(data);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleStatusChange = async (id, newStatus) => {
    try {
      await updateAlertStatus(id, newStatus);
      await loadAlerts();
    } catch (err) {
      alert('Failed to update alert: ' + err.message);
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Are you sure you want to delete this alert?')) {
      return;
    }

    try {
      await deleteAlert(id);
      await loadAlerts();
    } catch (err) {
      alert('Failed to delete alert: ' + err.message);
    }
  };

  const getSeverityClass = (severity) => {
    return `severity-badge severity-${severity}`;
  };

  const getStatusClass = (status) => {
    return `status-badge status-${status.replace('_', '-')}`;
  };

  const filteredAlerts = filter === 'all' 
    ? alerts 
    : alerts.filter(alert => alert.status === filter);

  if (loading) {
    return (
      <div className="page-container">
        <div className="loading">Loading alerts...</div>
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
        <h2>Security Alerts</h2>
        <div className="filter-buttons">
          <button
            className={`filter-btn ${filter === 'all' ? 'active' : ''}`}
            onClick={() => setFilter('all')}
          >
            All ({alerts.length})
          </button>
          <button
            className={`filter-btn ${filter === 'new' ? 'active' : ''}`}
            onClick={() => setFilter('new')}
          >
            New ({alerts.filter(a => a.status === 'new').length})
          </button>
          <button
            className={`filter-btn ${filter === 'investigating' ? 'active' : ''}`}
            onClick={() => setFilter('investigating')}
          >
            Investigating ({alerts.filter(a => a.status === 'investigating').length})
          </button>
          <button
            className={`filter-btn ${filter === 'resolved' ? 'active' : ''}`}
            onClick={() => setFilter('resolved')}
          >
            Resolved ({alerts.filter(a => a.status === 'resolved').length})
          </button>
        </div>
      </div>

      {filteredAlerts.length === 0 ? (
        <div className="empty-state">No alerts found</div>
      ) : (
        <div className="alerts-list">
          {filteredAlerts.map((alert) => (
            <div key={alert.id} className="alert-card">
              <div className="alert-header">
                <div className="alert-title-row">
                  <h3>{alert.title}</h3>
                  <div className="alert-badges">
                    <span className={getSeverityClass(alert.severity)}>
                      {alert.severity}
                    </span>
                    <span className={getStatusClass(alert.status)}>
                      {alert.status.replace('_', ' ')}
                    </span>
                  </div>
                </div>
                <div className="alert-meta">
                  <span>Host: {alert.host}</span>
                  {alert.source_ip && <span>Source IP: {alert.source_ip}</span>}
                  <span>Rule: {alert.rule_name}</span>
                  <span>{new Date(alert.timestamp).toLocaleString()}</span>
                </div>
              </div>

              <div className="alert-body">
                <p>{alert.description}</p>
              </div>

              <div className="alert-actions">
                <select
                  value={alert.status}
                  onChange={(e) => handleStatusChange(alert.id, e.target.value)}
                  className="status-select"
                >
                  <option value="new">New</option>
                  <option value="investigating">Investigating</option>
                  <option value="resolved">Resolved</option>
                  <option value="false_positive">False Positive</option>
                </select>
                <button
                  onClick={() => handleDelete(alert.id)}
                  className="delete-btn"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default Alerts;
