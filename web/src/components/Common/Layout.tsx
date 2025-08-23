import { ReactNode } from 'react';
import Navigation from './Navigation';
import { useApp } from '@/context/AppContext';

interface LayoutProps {
  children: ReactNode;
}

export default function Layout({ children }: LayoutProps) {
  const { state } = useApp();

  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />
      
      {/* Connection Status Indicator */}
      {state.connectionStatus !== 'connected' && (
        <div className={`px-4 py-2 text-sm text-center ${
          state.connectionStatus === 'connecting' 
            ? 'bg-yellow-100 text-yellow-800' 
            : 'bg-red-100 text-red-800'
        }`}>
          {state.connectionStatus === 'connecting' 
            ? 'Connecting to server...' 
            : 'Disconnected from server'}
        </div>
      )}

      {/* Error Banner */}
      {state.error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3">
          <span className="block sm:inline">{state.error}</span>
        </div>
      )}

      {/* Main Content */}
      <main className="container mx-auto px-4 py-6">
        {children}
      </main>

      {/* Notifications */}
      <div className="fixed top-4 right-4 space-y-2 z-50">
        {state.notifications.map((notification) => (
          <div
            key={notification.id}
            className={`px-4 py-3 rounded-md shadow-lg max-w-sm ${
              notification.type === 'success' ? 'bg-green-100 text-green-800' :
              notification.type === 'error' ? 'bg-red-100 text-red-800' :
              notification.type === 'warning' ? 'bg-yellow-100 text-yellow-800' :
              'bg-blue-100 text-blue-800'
            }`}
          >
            {notification.message}
          </div>
        ))}
      </div>
    </div>
  );
}