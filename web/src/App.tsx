
import { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AppProvider } from './context/AppContext';
import { AuthProvider, useAuth } from './context/AuthContext';
import { ErrorBoundary, Layout } from './components/Common';
import { Login } from './components/Auth';
import { webSocketService } from './services/websocket';
import Dashboard from './components/Dashboard/Dashboard';
import Queue from './components/Queue/Queue';
import BulkActionsTest from './components/Queue/BulkActionsTest';
import { Logs } from './components/Logs';
import LogViewerTest from './components/Logs/LogViewerTest';
import LogsTestPage from './components/Logs/LogsTestPage';
import { Reports } from './components/Reports';
import { MessageTrace } from './components/MessageTrace';

function AppContent() {
  const { isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    if (isAuthenticated) {
      // Initialize WebSocket connection only when authenticated
      webSocketService.connect().catch(error => {
        console.error('Failed to connect to WebSocket:', error);
      });

      return () => {
        webSocketService.disconnect();
      };
    }
  }, [isAuthenticated]);

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
        <Route path="/logs" element={<Logs />} />
        <Route path="/reports" element={<Reports />} />
        <Route path="/messages/:messageId/trace" element={<MessageTrace />} />
        <Route path="/logs-test" element={<LogViewerTest />} />
        <Route path="/logs-full-test" element={<LogsTestPage />} />
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
