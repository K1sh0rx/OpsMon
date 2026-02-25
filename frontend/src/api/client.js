const API_BASE = '/api/v1';

// Dashboard API
export const fetchDashboardMetrics = async (range = '24h') => {
  const response = await fetch(`${API_BASE}/dashboard/metrics?range=${range}`);
  if (!response.ok) throw new Error('Failed to fetch dashboard metrics');
  return response.json();
};

// Analytics API
export const fetchIngestionTrend = async (range = '24h') => {
  const response = await fetch(`${API_BASE}/analytics/ingestion?range=${range}`);
  if (!response.ok) throw new Error('Failed to fetch ingestion trend');
  return response.json();
};

export const fetchErrorTrend = async (range = '24h') => {
  const response = await fetch(`${API_BASE}/analytics/errors?range=${range}`);
  if (!response.ok) throw new Error('Failed to fetch error trend');
  return response.json();
};

export const fetchWarningTrend = async (range = '24h') => {
  const response = await fetch(`${API_BASE}/analytics/warnings?range=${range}`);
  if (!response.ok) throw new Error('Failed to fetch warning trend');
  return response.json();
};

export const fetchAlertTrend = async (range = '24h') => {
  const response = await fetch(`${API_BASE}/analytics/alerts?range=${range}`);
  if (!response.ok) throw new Error('Failed to fetch alert trend');
  return response.json();
};

// Alerts API
export const fetchAlerts = async () => {
  const response = await fetch(`${API_BASE}/alerts`);
  if (!response.ok) throw new Error('Failed to fetch alerts');
  return response.json();
};

export const updateAlertStatus = async (id, status) => {
  const response = await fetch(`${API_BASE}/alerts/${id}`, {
    method: 'PATCH',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ status }),
  });
  if (!response.ok) throw new Error('Failed to update alert');
  return response.json();
};

export const deleteAlert = async (id) => {
  const response = await fetch(`${API_BASE}/alerts/${id}`, {
    method: 'DELETE',
  });
  if (!response.ok) throw new Error('Failed to delete alert');
  return response.json();
};
