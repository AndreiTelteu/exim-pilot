import { useApp } from '@/context/AppContext';
import { LoadingSpinner } from '@/components/Common';

export default function Dashboard() {
  const { state } = useApp();

  if (state.isLoading) {
    return <LoadingSpinner size="lg" className="py-12" />;
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow-md p-6">
        <h2 className="text-2xl font-semibold text-gray-800 mb-4">
          Welcome to Exim-Pilot
        </h2>
        <p className="text-gray-600 mb-6">
          Your comprehensive web-based management interface for Exim mail servers.
        </p>
        
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-blue-50 p-4 rounded-lg">
            <h3 className="font-semibold text-blue-900">Queue Management</h3>
            <p className="text-blue-700 text-sm">Monitor and manage mail queue</p>
          </div>
          <div className="bg-green-50 p-4 rounded-lg">
            <h3 className="font-semibold text-green-900">Log Monitoring</h3>
            <p className="text-green-700 text-sm">Real-time log analysis</p>
          </div>
          <div className="bg-purple-50 p-4 rounded-lg">
            <h3 className="font-semibold text-purple-900">Reports</h3>
            <p className="text-purple-700 text-sm">Deliverability analytics</p>
          </div>
        </div>
      </div>

      {/* System Status */}
      <div className="bg-white rounded-lg shadow-md p-6">
        <h3 className="text-lg font-semibold text-gray-800 mb-4">System Status</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-blue-600">--</div>
            <div className="text-sm text-gray-600">Queue Messages</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-green-600">--</div>
            <div className="text-sm text-gray-600">Delivered Today</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-yellow-600">--</div>
            <div className="text-sm text-gray-600">Deferred</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-red-600">--</div>
            <div className="text-sm text-gray-600">Frozen</div>
          </div>
        </div>
      </div>
    </div>
  );
}