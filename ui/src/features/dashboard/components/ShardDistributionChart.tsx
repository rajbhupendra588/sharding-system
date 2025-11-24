import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';
import { Shard } from '@/features/shard';

interface ShardDistributionChartProps {
    shards: Shard[];
}

const COLORS = ['#10B981', '#F59E0B', '#EF4444', '#6366F1'];

export default function ShardDistributionChart({ shards }: ShardDistributionChartProps) {
    const data = [
        { name: 'Active', value: shards.filter(s => s.status === 'active').length },
        { name: 'Migrating', value: shards.filter(s => s.status === 'migrating').length },
        { name: 'Read-only', value: shards.filter(s => s.status === 'readonly').length },
        { name: 'Inactive', value: shards.filter(s => s.status === 'inactive').length },
    ].filter(item => item.value > 0);

    if (data.length === 0) {
        return (
            <div className="flex items-center justify-center h-64 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                <p className="text-gray-500 dark:text-gray-400">No shard data available</p>
            </div>
        );
    }

    return (
        <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                    <Pie
                        data={data}
                        cx="50%"
                        cy="50%"
                        innerRadius={60}
                        outerRadius={80}
                        paddingAngle={5}
                        dataKey="value"
                    >
                        {data.map((_, index) => (
                            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                    </Pie>
                    <Tooltip
                        contentStyle={{
                            backgroundColor: 'rgba(255, 255, 255, 0.9)',
                            borderRadius: '0.5rem',
                            border: 'none',
                            boxShadow: '0 4px 6px -1px rgba(0, 0, 0, 0.1)'
                        }}
                    />
                    <Legend verticalAlign="bottom" height={36} />
                </PieChart>
            </ResponsiveContainer>
        </div>
    );
}
