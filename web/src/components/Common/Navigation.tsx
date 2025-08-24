import { useLocation } from 'react-router-dom';
import { Link } from 'react-router-dom';
import { useApp } from '@/context/AppContext';
import { useAuth } from '@/context/AuthContext';
import { HelpButton } from './HelpModal';

export default function Navigation() {
  const { state } = useApp();
  const { user, logout } = useAuth();
  const location = useLocation();

  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  const navItems = [
    { id: 'dashboard', label: 'Dashboard', path: '/' },
    { id: 'queue', label: 'Queue Management', path: '/queue' },
    { id: 'logs', label: 'Log Monitoring', path: '/logs' },
    { id: 'reports', label: 'Reports', path: '/reports' },
  ];

  const isActive = (path: string) => {
    if (path === '/') {
      return location.pathname === '/';
    }
    return location.pathname.startsWith(path);
  };

  return (
    <nav className="bg-white shadow-sm border-b">
      <div className="container mx-auto px-4">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <div className="flex items-center">
            <h1 className="text-xl font-bold text-gray-900">
              Exim-Pilot
            </h1>
          </div>

          {/* Navigation Links */}
          <div className="hidden md:block">
            <div className="ml-10 flex items-baseline space-x-4">
              {navItems.map((item) => (
                <Link
                  key={item.id}
                  to={item.path}
                  className={`px-3 py-2 rounded-md text-sm font-medium transition-colors ${
                    isActive(item.path)
                      ? 'bg-blue-100 text-blue-700'
                      : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                  }`}
                >
                  {item.label}
                </Link>
              ))}
            </div>
          </div>

          {/* User Info */}
          <div className="flex items-center space-x-4">
            {/* Help Button */}
            <HelpButton 
              section={
                location.pathname.startsWith('/queue') ? 'queue' :
                location.pathname.startsWith('/logs') ? 'logs' :
                location.pathname.startsWith('/reports') ? 'reports' :
                'dashboard'
              }
            />

            {/* Connection Status */}
            <div className="flex items-center space-x-2">
              <div
                className={`w-2 h-2 rounded-full ${
                  state.connectionStatus === 'connected'
                    ? 'bg-green-500'
                    : state.connectionStatus === 'connecting'
                    ? 'bg-yellow-500'
                    : 'bg-red-500'
                }`}
              />
              <span className="text-sm text-gray-600">
                {state.connectionStatus === 'connected'
                  ? 'Connected'
                  : state.connectionStatus === 'connecting'
                  ? 'Connecting'
                  : 'Disconnected'}
              </span>
            </div>

            {/* User */}
            {user && (
              <div className="flex items-center space-x-3">
                <div className="text-sm text-gray-600">
                  Welcome, {user.username}
                </div>
                <button
                  onClick={handleLogout}
                  className="text-sm text-gray-600 hover:text-gray-900 px-3 py-1 rounded-md hover:bg-gray-50"
                >
                  Logout
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Mobile Navigation */}
        <div className="md:hidden">
          <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3">
            {navItems.map((item) => (
              <Link
                key={item.id}
                to={item.path}
                className={`block px-3 py-2 rounded-md text-base font-medium ${
                  isActive(item.path)
                    ? 'bg-blue-100 text-blue-700'
                    : 'text-gray-600 hover:text-gray-900 hover:bg-gray-50'
                }`}
              >
                {item.label}
              </Link>
            ))}
          </div>
        </div>
      </div>
    </nav>
  );
}