import { Link, useLocation } from 'react-router-dom';
import './Navigation.css';

function Navigation() {
  const location = useLocation();

  const isActive = (path) => {
    return location.pathname === path;
  };

  return (
    <nav className="navigation">
      <div className="nav-brand">
        <h1>OpsMon</h1>
        <span className="nav-subtitle">Security Operations Monitor</span>
      </div>
      <div className="nav-links">
        <Link 
          to="/" 
          className={`nav-link ${isActive('/') ? 'active' : ''}`}
        >
          Dashboard
        </Link>
        <Link 
          to="/analytics" 
          className={`nav-link ${isActive('/analytics') ? 'active' : ''}`}
        >
          Analytics
        </Link>
        <Link 
          to="/alerts" 
          className={`nav-link ${isActive('/alerts') ? 'active' : ''}`}
        >
          Alerts
        </Link>
      </div>
    </nav>
  );
}

export default Navigation;
