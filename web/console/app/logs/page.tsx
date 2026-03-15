"use client";

import { useEffect, useState, useCallback } from "react";
import { listLogs, type ChallengeLog } from "@/lib/api";

const DEMO_LOGS: ChallengeLog[] = [
  { id: "1", session_id: "sess_a1b2c3d4", app_id: "app_1", scene_id: "scene_1", ip: "203.0.113.42", user_agent: "Mozilla/5.0 Chrome/122", risk_score: 0.12, challenge_type: "none", result: "pass", timestamp: "2026-03-15T10:42:18Z", duration_ms: 0 },
  { id: "2", session_id: "sess_e5f6g7h8", app_id: "app_1", scene_id: "scene_1", ip: "198.51.100.7", user_agent: "Mozilla/5.0 Firefox/123", risk_score: 0.67, challenge_type: "slider", result: "pass", timestamp: "2026-03-15T10:41:55Z", duration_ms: 3200 },
  { id: "3", session_id: "sess_i9j0k1l2", app_id: "app_2", scene_id: "scene_2", ip: "192.0.2.100", user_agent: "Bot/1.0", risk_score: 0.95, challenge_type: "click", result: "fail", timestamp: "2026-03-15T10:41:30Z", duration_ms: 15000 },
  { id: "4", session_id: "sess_m3n4o5p6", app_id: "app_1", scene_id: "scene_3", ip: "203.0.113.88", user_agent: "Mozilla/5.0 Safari/17", risk_score: 0.34, challenge_type: "slider", result: "pass", timestamp: "2026-03-15T10:40:12Z", duration_ms: 2100 },
  { id: "5", session_id: "sess_q7r8s9t0", app_id: "app_2", scene_id: "scene_1", ip: "198.51.100.55", user_agent: "Mozilla/5.0 Edge/122", risk_score: 0.81, challenge_type: "rotate", result: "timeout", timestamp: "2026-03-15T10:39:48Z", duration_ms: 30000 },
  { id: "6", session_id: "sess_u1v2w3x4", app_id: "app_1", scene_id: "scene_2", ip: "192.0.2.200", user_agent: "Python-requests/2.31", risk_score: 0.92, challenge_type: "click", result: "fail", timestamp: "2026-03-15T10:38:22Z", duration_ms: 12000 },
  { id: "7", session_id: "sess_y5z6a7b8", app_id: "app_1", scene_id: "scene_1", ip: "203.0.113.15", user_agent: "Mozilla/5.0 Chrome/122", risk_score: 0.08, challenge_type: "none", result: "pass", timestamp: "2026-03-15T10:37:55Z", duration_ms: 0 },
  { id: "8", session_id: "sess_c9d0e1f2", app_id: "app_2", scene_id: "scene_3", ip: "198.51.100.33", user_agent: "curl/8.4.0", risk_score: 0.88, challenge_type: "slider", result: "fail", timestamp: "2026-03-15T10:36:10Z", duration_ms: 5000 },
  { id: "9", session_id: "sess_g3h4i5j6", app_id: "app_1", scene_id: "scene_1", ip: "203.0.113.77", user_agent: "Mozilla/5.0 Chrome/122", risk_score: 0.21, challenge_type: "none", result: "pass", timestamp: "2026-03-15T10:35:44Z", duration_ms: 0 },
  { id: "10", session_id: "sess_k7l8m9n0", app_id: "app_1", scene_id: "scene_2", ip: "192.0.2.150", user_agent: "Mozilla/5.0 Firefox/123", risk_score: 0.55, challenge_type: "slider", result: "pass", timestamp: "2026-03-15T10:34:18Z", duration_ms: 2800 },
  { id: "11", session_id: "sess_o1p2q3r4", app_id: "app_2", scene_id: "scene_1", ip: "198.51.100.99", user_agent: "Scrapy/2.11", risk_score: 0.97, challenge_type: "click", result: "fail", timestamp: "2026-03-15T10:33:02Z", duration_ms: 20000 },
  { id: "12", session_id: "sess_s5t6u7v8", app_id: "app_1", scene_id: "scene_1", ip: "203.0.113.60", user_agent: "Mozilla/5.0 Safari/17", risk_score: 0.15, challenge_type: "none", result: "pass", timestamp: "2026-03-15T10:32:30Z", duration_ms: 0 },
];

function resultBadge(result: string) {
  switch (result) {
    case "pass":
      return <span className="badge-green">Pass</span>;
    case "fail":
      return <span className="badge-red">Fail</span>;
    case "timeout":
      return <span className="badge-yellow">Timeout</span>;
    case "skip":
      return <span className="badge-gray">Skip</span>;
    default:
      return <span className="badge-gray">{result}</span>;
  }
}

function riskBadge(score: number) {
  if (score >= 0.7) return <span className="badge-red">{score.toFixed(2)}</span>;
  if (score >= 0.4) return <span className="badge-yellow">{score.toFixed(2)}</span>;
  return <span className="badge-green">{score.toFixed(2)}</span>;
}

function challengeTypeBadge(type: string) {
  if (type === "none") return <span className="text-gray-400 text-xs">--</span>;
  const colors: Record<string, string> = {
    slider: "bg-blue-100 text-blue-800",
    click: "bg-purple-100 text-purple-800",
    rotate: "bg-orange-100 text-orange-800",
    sms: "bg-pink-100 text-pink-800",
  };
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
        colors[type] || "bg-gray-100 text-gray-800"
      }`}
    >
      {type}
    </span>
  );
}

export default function LogsPage() {
  const [logs, setLogs] = useState<ChallengeLog[]>(DEMO_LOGS);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(DEMO_LOGS.length);
  const limit = 10;

  const fetchLogs = useCallback(async (s: string, p: number) => {
    try {
      const res = await listLogs({ search: s || undefined, page: p, limit });
      setLogs(res.logs);
      setTotal(res.total);
    } catch {
      // filter demo data locally
      const filtered = DEMO_LOGS.filter(
        (l) =>
          !s ||
          l.session_id.includes(s) ||
          l.ip.includes(s) ||
          l.challenge_type.includes(s) ||
          l.result.includes(s)
      );
      setLogs(filtered.slice((p - 1) * limit, p * limit));
      setTotal(filtered.length);
    }
  }, []);

  useEffect(() => {
    fetchLogs(search, page);
  }, [fetchLogs, search, page]);

  function handleSearchSubmit(e: React.FormEvent) {
    e.preventDefault();
    setPage(1);
    fetchLogs(search, 1);
  }

  const totalPages = Math.max(1, Math.ceil(total / limit));

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Challenge Logs</h1>
        <p className="mt-1 text-sm text-gray-500">
          Detailed log of all CAPTCHA challenge sessions
        </p>
      </div>

      {/* Search */}
      <form onSubmit={handleSearchSubmit} className="flex gap-3">
        <input
          className="flex-1 max-w-md rounded-lg border border-gray-300 px-4 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
          placeholder="Search by session ID, IP, type, or result..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <button type="submit" className="btn-primary">
          Search
        </button>
        {search && (
          <button
            type="button"
            className="btn-secondary"
            onClick={() => {
              setSearch("");
              setPage(1);
              fetchLogs("", 1);
            }}
          >
            Clear
          </button>
        )}
      </form>

      {/* Table */}
      <div className="rounded-xl border border-gray-200 bg-white shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="table-header">Session ID</th>
                <th className="table-header">IP Address</th>
                <th className="table-header">Risk Score</th>
                <th className="table-header">Challenge Type</th>
                <th className="table-header">Result</th>
                <th className="table-header">Duration</th>
                <th className="table-header">Timestamp</th>
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
                  <td className="table-cell">
                    {challengeTypeBadge(log.challenge_type)}
                  </td>
                  <td className="table-cell">{resultBadge(log.result)}</td>
                  <td className="table-cell text-gray-500 text-xs">
                    {log.duration_ms > 0 ? `${(log.duration_ms / 1000).toFixed(1)}s` : "--"}
                  </td>
                  <td className="table-cell text-gray-500 text-xs">
                    {new Date(log.timestamp).toLocaleString()}
                  </td>
                </tr>
              ))}
              {logs.length === 0 && (
                <tr>
                  <td
                    colSpan={7}
                    className="px-6 py-12 text-center text-sm text-gray-500"
                  >
                    No logs found.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        <div className="flex items-center justify-between border-t border-gray-200 px-6 py-3">
          <p className="text-sm text-gray-500">
            Showing {(page - 1) * limit + 1}-{Math.min(page * limit, total)} of{" "}
            {total} results
          </p>
          <div className="flex gap-2">
            <button
              className="btn-secondary text-xs"
              disabled={page <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              Previous
            </button>
            <span className="flex items-center px-3 text-sm text-gray-600">
              {page} / {totalPages}
            </span>
            <button
              className="btn-secondary text-xs"
              disabled={page >= totalPages}
              onClick={() => setPage((p) => p + 1)}
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
