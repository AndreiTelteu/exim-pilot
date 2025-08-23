
import { useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AppProvider } from './context/AppContext';
import { ErrorBoundary, Layout } from './components/Common';
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
      <AppProvider>
        <Router>
          <AppContent />
        </Router>
      </AppProvider>
    </ErrorBoundary>
  );
}

export default App
