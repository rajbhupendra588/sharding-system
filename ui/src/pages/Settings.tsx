import { useState } from 'react';
import { Save, RefreshCw } from 'lucide-react';
import Button from '@/components/ui/Button';
import Input from '@/components/ui/Input';
import { toast } from 'react-hot-toast';

export default function Settings() {
  const [managerUrl, setManagerUrl] = useState(localStorage.getItem('manager_url') || 'http://localhost:8081');
  const [routerUrl, setRouterUrl] = useState(localStorage.getItem('router_url') || 'http://localhost:8080');
  const [refreshInterval, setRefreshInterval] = useState(
    localStorage.getItem('refresh_interval') || '10000'
  );

  const handleSave = () => {
    localStorage.setItem('manager_url', managerUrl);
    localStorage.setItem('router_url', routerUrl);
    localStorage.setItem('refresh_interval', refreshInterval);
    toast.success('Settings saved successfully');
    // Reload to apply changes
    window.location.reload();
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-gray-900">Settings</h1>
        <p className="text-sm text-gray-500 mt-1">
          Configure application settings and API endpoints
        </p>
      </div>

      <div className="card space-y-6">
        <div>
          <h2 className="text-lg font-semibold text-gray-900 mb-4">API Configuration</h2>
          <div className="space-y-4">
            <Input
              label="Manager API URL (Control Plane)"
              value={managerUrl}
              onChange={(e) => setManagerUrl(e.target.value)}
              placeholder="http://localhost:8081"
              helperText="Base URL for the Manager API"
            />
            <Input
              label="Router API URL (Data Plane)"
              value={routerUrl}
              onChange={(e) => setRouterUrl(e.target.value)}
              placeholder="http://localhost:8080"
              helperText="Base URL for the Router API"
            />
          </div>
        </div>

        <div className="border-t pt-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Refresh Settings</h2>
          <div className="space-y-4">
            <Input
              label="Auto-refresh Interval (ms)"
              type="number"
              value={refreshInterval}
              onChange={(e) => setRefreshInterval(e.target.value)}
              placeholder="10000"
              helperText="How often to automatically refresh data (in milliseconds)"
            />
          </div>
        </div>

        <div className="flex items-center justify-end space-x-2 pt-4 border-t">
          <Button variant="secondary" onClick={() => window.location.reload()}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Reset
          </Button>
          <Button onClick={handleSave}>
            <Save className="h-4 w-4 mr-2" />
            Save Settings
          </Button>
        </div>
      </div>

      <div className="card bg-yellow-50 border-yellow-200">
        <h3 className="text-sm font-medium text-yellow-800 mb-2">Note</h3>
        <p className="text-sm text-yellow-700">
          Changing these settings will reload the application. Make sure the URLs are correct
          and accessible from your browser.
        </p>
      </div>
    </div>
  );
}

