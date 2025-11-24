import { useState, useEffect } from 'react';
import { Activity, ArrowUp, ArrowDown } from 'lucide-react';

export default function LiveMetrics() {
    const [rps, setRps] = useState(45);
    const [errorRate, setErrorRate] = useState(0.02);
    const [latency, setLatency] = useState(12);

    useEffect(() => {
        const interval = setInterval(() => {
            // Simulate live data fluctuations
            setRps(prev => Math.max(10, Math.floor(prev + (Math.random() - 0.5) * 10)));
            setErrorRate(prev => Math.max(0, Math.min(1, prev + (Math.random() - 0.5) * 0.01)));
            setLatency(prev => Math.max(5, Math.floor(prev + (Math.random() - 0.5) * 5)));
        }, 2000);

        return () => clearInterval(interval);
    }, []);

    return (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="bg-white dark:bg-gray-800 p-4 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm">
                <div className="flex items-center justify-between">
                    <div>
                        <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Live RPS</p>
                        <div className="flex items-baseline mt-1">
                            <p className="text-2xl font-semibold text-gray-900 dark:text-white">{rps}</p>
                            <span className="ml-2 text-sm font-medium text-green-600 flex items-center">
                                <ArrowUp className="h-3 w-3 mr-0.5" />
                                12%
                            </span>
                        </div>
                    </div>
                    <div className="p-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
                        <Activity className="h-5 w-5 text-blue-600 dark:text-blue-400" />
                    </div>
                </div>
            </div>

            <div className="bg-white dark:bg-gray-800 p-4 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm">
                <div className="flex items-center justify-between">
                    <div>
                        <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Error Rate</p>
                        <div className="flex items-baseline mt-1">
                            <p className="text-2xl font-semibold text-gray-900 dark:text-white">{errorRate.toFixed(2)}%</p>
                            <span className="ml-2 text-sm font-medium text-green-600 flex items-center">
                                <ArrowDown className="h-3 w-3 mr-0.5" />
                                0.01%
                            </span>
                        </div>
                    </div>
                    <div className="p-2 bg-red-50 dark:bg-red-900/20 rounded-lg">
                        <Activity className="h-5 w-5 text-red-600 dark:text-red-400" />
                    </div>
                </div>
            </div>

            <div className="bg-white dark:bg-gray-800 p-4 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm">
                <div className="flex items-center justify-between">
                    <div>
                        <p className="text-sm font-medium text-gray-500 dark:text-gray-400">Avg Latency</p>
                        <div className="flex items-baseline mt-1">
                            <p className="text-2xl font-semibold text-gray-900 dark:text-white">{latency}ms</p>
                            <span className="ml-2 text-sm font-medium text-yellow-600 flex items-center">
                                <ArrowUp className="h-3 w-3 mr-0.5" />
                                2ms
                            </span>
                        </div>
                    </div>
                    <div className="p-2 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
                        <Activity className="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
                    </div>
                </div>
            </div>
        </div>
    );
}
