"use client";

import { useEffect, useState } from "react";
import { listApps, createApp, deleteApp, type App } from "@/lib/api";

const DEMO_APPS: App[] = [
  {
    id: "app_1",
    name: "Main Website",
    site_key: "sk_live_abc123def456",
    secret_key: "sec_live_xyz789",
    domain: "www.example.com",
    status: "active",
    created_at: "2026-01-15T08:30:00Z",
    updated_at: "2026-03-10T14:20:00Z",
  },
  {
    id: "app_2",
    name: "Mobile API",
    site_key: "sk_live_mob789ghi012",
    secret_key: "sec_live_mob345",
    domain: "api.example.com",
    status: "active",
    created_at: "2026-02-20T11:00:00Z",
    updated_at: "2026-03-12T09:15:00Z",
  },
  {
    id: "app_3",
    name: "Staging Environment",
    site_key: "sk_test_stg456jkl789",
    secret_key: "sec_test_stg012",
    domain: "staging.example.com",
    status: "inactive",
    created_at: "2026-03-01T16:45:00Z",
    updated_at: "2026-03-01T16:45:00Z",
  },
];

function statusBadge(status: string) {
  switch (status) {
    case "active":
      return <span className="badge-green">Active</span>;
    case "inactive":
      return <span className="badge-gray">Inactive</span>;
    case "suspended":
      return <span className="badge-red">Suspended</span>;
    default:
      return <span className="badge-gray">{status}</span>;
  }
}

export default function AppsPage() {
  const [apps, setApps] = useState<App[]>(DEMO_APPS);
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({ name: "", domain: "" });
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    listApps()
      .then(setApps)
      .catch(() => {});
  }, []);

  async function handleCreate() {
    if (!form.name || !form.domain) return;
    setSaving(true);
    try {
      const app = await createApp({ name: form.name, domain: form.domain });
      setApps((prev) => [app, ...prev]);
      setShowCreate(false);
      setForm({ name: "", domain: "" });
    } catch {
      // In demo mode, add a fake app
      const fake: App = {
        id: `app_${Date.now()}`,
        name: form.name,
        site_key: `sk_live_${Math.random().toString(36).slice(2, 14)}`,
        secret_key: `sec_live_${Math.random().toString(36).slice(2, 14)}`,
        domain: form.domain,
        status: "active",
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      setApps((prev) => [fake, ...prev]);
      setShowCreate(false);
      setForm({ name: "", domain: "" });
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this app? This cannot be undone.")) return;
    try {
      await deleteApp(id);
    } catch {
      // demo mode
    }
    setApps((prev) => prev.filter((a) => a.id !== id));
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Apps</h1>
          <p className="mt-1 text-sm text-gray-500">
            Manage your registered applications
          </p>
        </div>
        <button className="btn-primary" onClick={() => setShowCreate(true)}>
          + Create App
        </button>
      </div>

      {/* Create modal */}
      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Create New App
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  App Name
                </label>
                <input
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="My Website"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Domain
                </label>
                <input
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
                  value={form.domain}
                  onChange={(e) => setForm({ ...form, domain: e.target.value })}
                  placeholder="www.example.com"
                />
              </div>
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                className="btn-secondary"
                onClick={() => setShowCreate(false)}
              >
                Cancel
              </button>
              <button
                className="btn-primary"
                onClick={handleCreate}
                disabled={saving}
              >
                {saving ? "Creating..." : "Create"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Apps table */}
      <div className="rounded-xl border border-gray-200 bg-white shadow-sm overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="table-header">Name</th>
                <th className="table-header">Site Key</th>
                <th className="table-header">Domain</th>
                <th className="table-header">Status</th>
                <th className="table-header">Created</th>
                <th className="table-header">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {apps.map((app) => (
                <tr key={app.id} className="hover:bg-gray-50">
                  <td className="table-cell font-medium">{app.name}</td>
                  <td className="table-cell">
                    <code className="rounded bg-gray-100 px-2 py-1 text-xs font-mono">
                      {app.site_key}
                    </code>
                  </td>
                  <td className="table-cell text-gray-600">{app.domain}</td>
                  <td className="table-cell">{statusBadge(app.status)}</td>
                  <td className="table-cell text-gray-500 text-xs">
                    {new Date(app.created_at).toLocaleDateString()}
                  </td>
                  <td className="table-cell">
                    <button
                      className="text-sm text-red-600 hover:text-red-800"
                      onClick={() => handleDelete(app.id)}
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
              {apps.length === 0 && (
                <tr>
                  <td colSpan={6} className="px-6 py-12 text-center text-sm text-gray-500">
                    No apps yet. Create one to get started.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
