"use client";

import { useEffect, useState } from "react";
import StatCard from "@/components/stat-card";
import { getDashboardStats, type DashboardStats, type ChallengeLog, listLogs } from "@/lib/api";
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  Legend,
} from "recharts";

const DEMO_STATS: DashboardStats = {
  total_requests: 1_284_530,
  challenge_rate: 12.4,
  pass_rate: 94.7,
  bot_blocked: 68_241,
  avg_risk_score: 0.23,
  false_positive_rate: 0.8,
  requests_trend: [
    { date: "Mar 09", count: 180200 },
    { date: "Mar 10", count: 195400 },
    { date: "Mar 11", count: 172800 },
    { date: "Mar 12", count: 210600 },
    { date: "Mar 13", count: 198300 },
    { date: "Mar 14", count: 187500 },
    { date: "Mar 15", count: 140730 },
  ],
  challenge_distribution: [
    { type: "Slider", count: 45200 },
    { type: "Click", count: 32100 },
    { type: "Rotate", count: 18400 },
    { type: "SMS", count: 4300 },
  ],
};

const DEMO_RECENT_LOGS: ChallengeLog[] = [
  { id: "1", session_id: "sess_a1b2c3", app_id: "app1", scene_id: "s1", ip: "203.0.113.42", user_agent: "Chrome", risk_score: 0.12, challenge_type: "none", result: "pass", timestamp: "2026-03-15T10:42:18Z", duration_ms: 0 },
  { id: "2", session_id: "sess_d4e5f6", app_id: "app1", scene_id: "s1", ip: "198.51.100.7", user_agent: "Firefox", risk_score: 0.67, challenge_type: "slider", result: "pass", timestamp: "2026-03-15T10:41:55Z", duration_ms: 3200 },
  { id: "3", session_id: "sess_g7h8i9", app_id: "app2", scene_id: "s2", ip: "192.0.2.100", user_agent: "Bot/1.0", risk_score: 0.95, challenge_type: "click", result: "fail", timestamp: "2026-03-15T10:41:30Z", duration_ms: 15000 },
  { id: "4", session_id: "sess_j0k1l2", app_id: "app1", scene_id: "s3", ip: "203.0.113.88", user_agent: "Safari", risk_score: 0.34, challenge_type: "slider", result: "pass", timestamp: "2026-03-15T10:40:12Z", duration_ms: 2100 },
  { id: "5", session_id: "sess_m3n4o5", app_id: "app2", scene_id: "s1", ip: "198.51.100.55", user_agent: "Edge", risk_score: 0.81, challenge_type: "rotate", result: "timeout", timestamp: "2026-03-15T10:39:48Z", duration_ms: 30000 },
];

const PIE_COLORS = ["#3b82f6", "#10b981", "#f59e0b", "#ef4444"];

function resultBadge(result: string) {
  switch (result) {
    case "pass":
      return <span className="badge-green">Pass</span>;
    case "fail":
      return <span className="badge-red">Fail</span>;
    case "timeout":
      return <span className="badge-yellow">Timeout</span>;
    default:
      return <span className="badge-gray">{result}</span>;
  }
}

function riskBadge(score: number) {
  if (score >= 0.7) return <span className="badge-red">{score.toFixed(2)}</span>;
  if (score >= 0.4) return <span className="badge-yellow">{score.toFixed(2)}</span>;
  return <span className="badge-green">{score.toFixed(2)}</span>;
}

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats>(DEMO_STATS);
  const [logs, setLogs] = useState<ChallengeLog[]>(DEMO_RECENT_LOGS);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        const [s, l] = await Promise.all([
          getDashboardStats(),
          listLogs({ limit: 5 }),
        ]);
        setStats(s);
        setLogs(l.logs);
      } catch {
        // keep demo data
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-1 text-sm text-gray-500">
          Overview of your CAPTCHA platform performance
        </p>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-3">
        <StatCard
          title="Total Requests"
          value={stats.total_requests.toLocaleString()}
          subtitle="Last 7 days"
          trend={{ value: 8.2, direction: "up" }}
        />
        <StatCard
          title="Challenge Rate"
          value={`${stats.challenge_rate}%`}
          subtitle="Of total requests"
          trend={{ value: 1.3, direction: "down" }}
        />
        <StatCard
          title="Pass Rate"
          value={`${stats.pass_rate}%`}
          subtitle="Challenges passed"
          trend={{ value: 2.1, direction: "up" }}
        />
        <StatCard
          title="Bots Blocked"
          value={stats.bot_blocked.toLocaleString()}
          subtitle="Last 7 days"
          trend={{ value: 12.5, direction: "up" }}
        />
        <StatCard
          title="Avg Risk Score"
          value={stats.avg_risk_score.toFixed(2)}
          subtitle="Lower is better"
          trend={{ value: 0.5, direction: "down" }}
        />
        <StatCard
          title="False Positive Rate"
          value={`${stats.false_positive_rate}%`}
          subtitle="Legit users challenged"
          trend={{ value: 0.2, direction: "down" }}
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <div className="col-span-2 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Request Volume (7 days)
          </h2>
          <ResponsiveContainer width="100%" height={280}>
            <AreaChart data={stats.requests_trend}>
              <defs>
                <linearGradient id="colorCount" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#3b82f6" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis dataKey="date" tick={{ fontSize: 12 }} />
              <YAxis tick={{ fontSize: 12 }} tickFormatter={(v: number) => `${(v / 1000).toFixed(0)}k`} />
              <Tooltip formatter={(v: number) => v.toLocaleString()} />
              <Area
                type="monotone"
                dataKey="count"
                stroke="#3b82f6"
                strokeWidth={2}
                fill="url(#colorCount)"
              />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        <div className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Challenge Types
          </h2>
          <ResponsiveContainer width="100%" height={280}>
            <PieChart>
              <Pie
                data={stats.challenge_distribution}
                dataKey="count"
                nameKey="type"
                cx="50%"
                cy="50%"
                outerRadius={90}
                innerRadius={50}
                paddingAngle={3}
              >
                {stats.challenge_distribution.map((_, i) => (
                  <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />
                ))}
              </Pie>
              <Tooltip formatter={(v: number) => v.toLocaleString()} />
              <Legend />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Recent Challenges */}
      <div className="rounded-xl border border-gray-200 bg-white shadow-sm">
        <div className="border-b border-gray-200 px-6 py-4">
          <h2 className="text-lg font-semibold text-gray-900">
            Recent Challenges
          </h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="table-header">Session</th>
                <th className="table-header">IP</th>
                <th className="table-header">Risk Score</th>
                <th className="table-header">Type</th>
                <th className="table-header">Result</th>
                <th className="table-header">Time</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {logs.map((log) => (
                <tr key={log.id} className="hover:bg-gray-50">
                  <td className="table-cell font-mono text-xs">
                    {log.session_id}
                  </td>
                  <td className="table-cell font-mono text-xs">{log.ip}</td>
                  <td className="table-cell">{riskBadge(log.risk_score)}</td>
                  <td className="table-cell capitalize">{log.challenge_type}</td>
                  <td className="table-cell">{resultBadge(log.result)}</td>
                  <td className="table-cell text-gray-500 text-xs">
                    {new Date(log.timestamp).toLocaleTimeString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {loading && (
        <p className="text-center text-sm text-gray-400">
          Using demo data (API unavailable)
        </p>
      )}
    </div>
  );
}
