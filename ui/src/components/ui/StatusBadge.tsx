import { cn } from '@/shared/lib';
import { getStatusBadgeColor } from '@/shared/utils';

interface StatusBadgeProps {
  status: string;
  className?: string;
}

export default function StatusBadge({ status, className }: StatusBadgeProps) {
  // Handle edge cases: empty string, numbers, or null/undefined
  const normalizedStatus = status?.toString().toLowerCase() || 'unknown';
  const displayText = normalizedStatus.charAt(0).toUpperCase() + normalizedStatus.slice(1);
  
  return (
    <span
      className={cn(
        'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border',
        getStatusBadgeColor(normalizedStatus),
        className
      )}
    >
      {displayText}
    </span>
  );
}

