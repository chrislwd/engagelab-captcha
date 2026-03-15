"use client";

interface StatCardProps {
  title: string;
  value: string | number;
  subtitle?: string;
  trend?: {
    value: number;
    direction: "up" | "down" | "flat";
  };
}

export default function StatCard({ title, value, subtitle, trend }: StatCardProps) {
  const trendColor =
    trend?.direction === "up"
      ? "text-green-600"
      : trend?.direction === "down"
        ? "text-red-600"
        : "text-gray-500";

  const trendIcon =
    trend?.direction === "up"
      ? "\u2191"
      : trend?.direction === "down"
        ? "\u2193"
        : "\u2192";

  return (
    <div className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm transition-shadow hover:shadow-md">
      <p className="text-sm font-medium text-gray-500">{title}</p>
      <div className="mt-2 flex items-baseline gap-2">
        <p className="text-3xl font-bold tracking-tight text-gray-900">
          {value}
        </p>
        {trend && (
          <span className={`flex items-center text-sm font-medium ${trendColor}`}>
            {trendIcon} {Math.abs(trend.value)}%
          </span>
        )}
      </div>
      {subtitle && (
        <p className="mt-1 text-sm text-gray-500">{subtitle}</p>
      )}
    </div>
  );
}
