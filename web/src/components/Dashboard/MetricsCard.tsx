import { MetricsCardProps } from '@/types/dashboard';

const colorClasses = {
  blue: {
    bg: 'bg-blue-50',
    text: 'text-blue-900',
    value: 'text-blue-600',
    subtitle: 'text-blue-700'
  },
  green: {
    bg: 'bg-green-50',
    text: 'text-green-900',
    value: 'text-green-600',
    subtitle: 'text-green-700'
  },
  yellow: {
    bg: 'bg-yellow-50',
    text: 'text-yellow-900',
    value: 'text-yellow-600',
    subtitle: 'text-yellow-700'
  },
  red: {
    bg: 'bg-red-50',
    text: 'text-red-900',
    value: 'text-red-600',
    subtitle: 'text-red-700'
  },
  purple: {
    bg: 'bg-purple-50',
    text: 'text-purple-900',
    value: 'text-purple-600',
    subtitle: 'text-purple-700'
  },
  gray: {
    bg: 'bg-gray-50',
    text: 'text-gray-900',
    value: 'text-gray-600',
    subtitle: 'text-gray-700'
  }
};

const trendIcons = {
  up: '↗',
  down: '↘',
  stable: '→'
};

const trendColors = {
  up: 'text-green-600',
  down: 'text-red-600',
  stable: 'text-gray-600'
};

export function MetricsCard({ 
  title, 
  value, 
  subtitle, 
  color = 'blue', 
  trend, 
  loading = false 
}: MetricsCardProps) {
  const colors = colorClasses[color];

  if (loading) {
    return (
      <div className={`${colors.bg} p-6 rounded-lg border border-gray-200`}>
        <div className="animate-pulse">
          <div className="h-4 bg-gray-300 rounded w-3/4 mb-3"></div>
          <div className="h-8 bg-gray-300 rounded w-1/2 mb-2"></div>
          <div className="h-3 bg-gray-300 rounded w-2/3"></div>
        </div>
      </div>
    );
  }

  return (
    <div className={`${colors.bg} p-6 rounded-lg border border-gray-200 transition-all duration-200 hover:shadow-md`}>
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <h3 className={`font-semibold ${colors.text} text-sm uppercase tracking-wide`}>
            {title}
          </h3>
          <div className={`text-3xl font-bold ${colors.value} mt-2 mb-1`}>
            {typeof value === 'number' ? value.toLocaleString() : value}
          </div>
          {subtitle && (
            <p className={`text-sm ${colors.subtitle}`}>
              {subtitle}
            </p>
          )}
        </div>
        
        {trend && (
          <div className={`flex items-center text-sm font-medium ${trendColors[trend.direction]}`}>
            <span className="mr-1">{trendIcons[trend.direction]}</span>
            <span>{Math.abs(trend.value)}%</span>
          </div>
        )}
      </div>
    </div>
  );
}