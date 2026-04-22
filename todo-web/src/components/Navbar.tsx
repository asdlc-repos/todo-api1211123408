import { NavLink, useNavigate } from 'react-router-dom';
import { useAuth } from '../auth';

export function Navbar() {
  const { email, logout } = useAuth();
  const navigate = useNavigate();

  const handleLogout = async () => {
    await logout();
    navigate('/login', { replace: true });
  };

  return (
    <nav className="navbar">
      <div className="navbar-brand">Todo App</div>
      <div className="navbar-links">
        <NavLink to="/todos" className={({ isActive }) => (isActive ? 'active' : '')}>
          Todos
        </NavLink>
        <NavLink to="/categories" className={({ isActive }) => (isActive ? 'active' : '')}>
          Categories
        </NavLink>
      </div>
      <div className="navbar-user">
        {email && <span>{email}</span>}
        <button className="secondary" onClick={handleLogout}>
          Log out
        </button>
      </div>
    </nav>
  );
}
