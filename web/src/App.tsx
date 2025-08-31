
import { useEffect, useRef } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AppProvider, useApp } from './context/AppContext';
import { AuthProvider, useAuth } from './context/AuthContext';
import { ErrorBoundary, Layout } from './components/Common';
import { Login } from './components/Auth';
import { webSocketService } from './services/websocket';
import Dashboard from './components/Dashboard/Dashboard';
import Queue from './components/Queue/Queue';
import BulkActionsTest from './components/Queue/BulkActionsTest';
// import { Logs } from './components/Logs';
// import LogViewerTest from './components/Logs/LogViewerTest';
// import LogsTestPage from './components/Logs/LogsTestPage';
import { Reports } from './components/Reports';
import { MessageTrace } from './components/MessageTrace';

function AppContent() {
  const { isAuthenticated, isLoading } = useAuth();
  const { actions } = useApp();
  const connectionInitialized = useRef(false);
  const hasShownDisconnectedNotification = useRef(false);

  useEffect(() => {
    console.log('AppContent useEffect triggered', { isAuthenticated, connectionInitialized: connectionInitialized.current });
    
    if (isAuthenticated && !connectionInitialized.current) {
      connectionInitialized.current = true;
      
      // Set up connection status callback
      webSocketService.setConnectionStatusCallback((status) => {
        console.log('WebSocket status changed to:', status);
        actions.setConnectionStatus(status);
        if (status === 'disconnected' && !hasShownDisconnectedNotification.current) {
          hasShownDisconnectedNotification.current = true;
          actions.addNotification({
            type: 'warning',
            message: 'Real-time updates unavailable. Data will refresh periodically.'
          });
        } else if (status === 'connected') {
          hasShownDisconnectedNotification.current = false;
        }
      });

      // Initialize WebSocket connection only when authenticated
      console.log('Attempting to connect WebSocket...');
      webSocketService.connect().catch(error => {
        console.error('Failed to connect to WebSocket:', error);
      });

      return () => {
        console.log('AppContent cleanup - disconnecting WebSocket');
        connectionInitialized.current = false;
        webSocketService.setConnectionStatusCallback(null);
        webSocketService.disconnect();
      };
    } else if (!isAuthenticated && connectionInitialized.current) {
      // Clear connection when not authenticated
      console.log('User not authenticated - clearing WebSocket connection');
      connectionInitialized.current = false;
      webSocketService.setConnectionStatusCallback(null);
      webSocketService.disconnect();
      actions.setConnectionStatus('disconnected');
    }
  }, [isAuthenticated]); // Only depend on isAuthenticated

  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-indigo-600"></div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Login />;
  }

  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/queue" element={<Queue />} />
        {/* <Route path="/logs" element={<Logs />} /> */}
        <Route path="/reports" element={<Reports />} />
        <Route path="/messages/:messageId/trace" element={<MessageTrace />} />
        {/* <Route path="/logs-test" element={<LogViewerTest />} />
        <Route path="/logs-full-test" element={<LogsTestPage />} /> */}
        <Route path="/bulk-test" element={<BulkActionsTest />} />
      </Routes>
    </Layout>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <AuthProvider>
        <AppProvider>
          <Router>
            <AppContent />
          </Router>
        </AppProvider>
      </AuthProvider>
    </ErrorBoundary>
  );
}

export default App
