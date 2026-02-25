# OpsMon Frontend

React frontend for OpsMon Security Operations Monitor.

## Features

- **Dashboard**: Real-time metrics and statistics
- **Analytics**: Time-series charts for log trends
- **Alerts**: Security alert management with CRUD operations

## Tech Stack

- React 18
- Vite
- React Router
- Recharts
- Pure CSS (no UI libraries)

## Design

- Black background (`#000000`)
- Dark blue borders (`#1e3a8a`)
- No gradients or glows
- Clean, professional interface

## Setup

### 1. Install dependencies
```bash
npm install
```

### 2. Start development server
```bash
npm run dev
```

The app will run on `http://localhost:3000`

### 3. Backend connection

The frontend proxies API calls to the backend at `http://localhost:8080`.

Make sure the backend server is running before starting the frontend.

## API Endpoints

### Dashboard
- `GET /api/v1/dashboard/metrics?range=24h|all`

### Analytics
- `GET /api/v1/analytics/ingestion?range=24h|all`
- `GET /api/v1/analytics/errors?range=24h|all`
- `GET /api/v1/analytics/warnings?range=24h|all`
- `GET /api/v1/analytics/alerts?range=24h|all`

### Alerts
- `GET /api/v1/alerts`
- `PATCH /api/v1/alerts/{id}`
- `DELETE /api/v1/alerts/{id}`

## Project Structure

```
src/
├── api/
│   └── client.js          # API client
├── components/
│   ├── Navigation.jsx     # Navigation bar
│   └── Navigation.css
├── pages/
│   ├── Dashboard.jsx      # Dashboard page
│   ├── Dashboard.css
│   ├── Analytics.jsx      # Analytics page
│   ├── Analytics.css
│   ├── Alerts.jsx         # Alerts page
│   └── Alerts.css
├── App.jsx                # Main app with routing
├── App.css
├── main.jsx               # Entry point
└── index.css              # Global styles
```

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build

## Color Palette

```css
--bg-black: #000000        /* Main background */
--bg-dark: #0a0a0a         /* Card backgrounds */
--blue-border: #1e3a8a     /* Borders */
--blue-light: #3b82f6      /* Accents */
--text-primary: #ffffff    /* Primary text */
--text-secondary: #9ca3af  /* Secondary text */
--error: #ef4444           /* Error state */
--warning: #f59e0b         /* Warning state */
--success: #10b981         /* Success state */
```

## Alert Severity Colors

- Critical: `#dc2626`
- High: `#f97316`
- Medium: `#eab308`
- Low: `#6b7280`

## Alert Status

- New
- Investigating
- Resolved
- False Positive

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)
