

import { useEffect } from 'react';
import { AppProvider } from './context/AppContext';
import { ErrorBoundary, Layout } from './components/Common';
import { webSocketService } from './services/websocket';
import Dashboard from './components/Dashboard/Dashboard';

function AppContent() {
  useEffect(() => {
    // Initialize WebSocket connection
    webSocketService.connect().catch(error => {
      console.error('Failed to connect to WebSocket:', error);
    });

    return () => {
      webSocketService.disconnect();
    };
  }, []);

  return (
    <Layout>
      <Dashboard />
    </Layout>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <AppProvider>
        <AppContent />
      </AppProvider>
    </ErrorBoundary>
  );
}

export default App
