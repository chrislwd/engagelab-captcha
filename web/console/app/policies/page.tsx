"use client";

import { useEffect, useState } from "react";
import {
  listPolicies,
  createPolicy,
  updatePolicy,
  deletePolicy,
  type Policy,
} from "@/lib/api";

const DEMO_POLICIES: Policy[] = [
  {
    id: "pol_1",
    name: "Standard Protection",
    description: "Balanced protection suitable for most scenarios",
    thresholds: { low: 0.3, medium: 0.6, high: 0.85 },
    actions: { low: "allow", medium: "challenge", high: "block" },
    challenge_types: ["slider", "click"],
    created_at: "2026-01-10T08:00:00Z",
    updated_at: "2026-03-05T12:00:00Z",
  },
  {
    id: "pol_2",
    name: "Strict Protection",
    description: "High-security policy for sensitive operations",
    thresholds: { low: 0.2, medium: 0.4, high: 0.7 },
    actions: { low: "challenge", medium: "challenge", high: "block" },
    challenge_types: ["slider", "click", "rotate", "sms"],
    created_at: "2026-01-12T10:30:00Z",
    updated_at: "2026-03-08T15:45:00Z",
  },
  {
    id: "pol_3",
    name: "Lenient",
    description: "Minimal friction for low-risk pages",
    thresholds: { low: 0.5, medium: 0.75, high: 0.9 },
    actions: { low: "allow", medium: "allow", high: "challenge" },
    challenge_types: ["slider"],
    created_at: "2026-02-01T14:00:00Z",
    updated_at: "2026-02-01T14:00:00Z",
  },
];

function actionBadge(action: string) {
  switch (action) {
    case "allow":
      return <span className="badge-green">Allow</span>;
    case "challenge":
      return <span className="badge-yellow">Challenge</span>;
    case "block":
      return <span className="badge-red">Block</span>;
    default:
      return <span className="badge-gray">{action}</span>;
  }
}

interface PolicyForm {
  name: string;
  description: string;
  thresholds: { low: number; medium: number; high: number };
  actions: { low: string; medium: string; high: string };
  challenge_types: string[];
}

const EMPTY_FORM: PolicyForm = {
  name: "",
  description: "",
  thresholds: { low: 0.3, medium: 0.6, high: 0.85 },
  actions: { low: "allow", medium: "challenge", high: "block" },
  challenge_types: ["slider", "click"],
};

export default function PoliciesPage() {
  const [policies, setPolicies] = useState<Policy[]>(DEMO_POLICIES);
  const [showModal, setShowModal] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<PolicyForm>(EMPTY_FORM);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    listPolicies()
      .then(setPolicies)
      .catch(() => {});
  }, []);

  function openCreate() {
    setEditingId(null);
    setForm(EMPTY_FORM);
    setShowModal(true);
  }

  function openEdit(policy: Policy) {
    setEditingId(policy.id);
    setForm({
      name: policy.name,
      description: policy.description,
      thresholds: { ...policy.thresholds },
      actions: { ...policy.actions },
      challenge_types: [...policy.challenge_types],
    });
    setShowModal(true);
  }

  async function handleSave() {
    if (!form.name) return;
    setSaving(true);
    const payload = {
      name: form.name,
      description: form.description,
      thresholds: form.thresholds,
      actions: form.actions as Policy["actions"],
      challenge_types: form.challenge_types,
    };

    try {
      if (editingId) {
        const updated = await updatePolicy(editingId, payload);
        setPolicies((prev) =>
          prev.map((p) => (p.id === editingId ? updated : p))
        );
      } else {
        const created = await createPolicy(payload);
        setPolicies((prev) => [created, ...prev]);
      }
    } catch {
      // demo fallback
      if (editingId) {
        setPolicies((prev) =>
          prev.map((p) =>
            p.id === editingId
              ? { ...p, ...payload, updated_at: new Date().toISOString() }
              : p
          )
        );
      } else {
        const fake: Policy = {
          id: `pol_${Date.now()}`,
          ...payload,
          actions: payload.actions as Policy["actions"],
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        setPolicies((prev) => [fake, ...prev]);
      }
    } finally {
      setSaving(false);
      setShowModal(false);
      setForm(EMPTY_FORM);
      setEditingId(null);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this policy?")) return;
    try {
      await deletePolicy(id);
    } catch {
      // demo
    }
    setPolicies((prev) => prev.filter((p) => p.id !== id));
  }

  function toggleChallengeType(type: string) {
    setForm((prev) => ({
      ...prev,
      challenge_types: prev.challenge_types.includes(type)
        ? prev.challenge_types.filter((t) => t !== type)
        : [...prev.challenge_types, type],
    }));
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Policies</h1>
          <p className="mt-1 text-sm text-gray-500">
            Define risk thresholds and actions for challenge decisions
          </p>
        </div>
        <button className="btn-primary" onClick={openCreate}>
          + Create Policy
        </button>
      </div>

      {/* Modal */}
      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-lg rounded-xl bg-white p-6 shadow-xl">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              {editingId ? "Edit Policy" : "Create Policy"}
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Name
                </label>
                <input
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Description
                </label>
                <input
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
                  value={form.description}
                  onChange={(e) =>
                    setForm({ ...form, description: e.target.value })
                  }
                />
              </div>

              {/* Thresholds */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Risk Thresholds
                </label>
                <div className="grid grid-cols-3 gap-3">
                  {(["low", "medium", "high"] as const).map((level) => (
                    <div key={level}>
                      <label className="block text-xs text-gray-500 capitalize mb-1">
                        {level}
                      </label>
                      <input
                        type="number"
                        step="0.05"
                        min="0"
                        max="1"
                        className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
                        value={form.thresholds[level]}
                        onChange={(e) =>
                          setForm({
                            ...form,
                            thresholds: {
                              ...form.thresholds,
                              [level]: parseFloat(e.target.value) || 0,
                            },
                          })
                        }
                      />
                    </div>
                  ))}
                </div>
              </div>

              {/* Actions */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Actions per Risk Level
                </label>
                <div className="grid grid-cols-3 gap-3">
                  {(["low", "medium", "high"] as const).map((level) => (
                    <div key={level}>
                      <label className="block text-xs text-gray-500 capitalize mb-1">
                        {level} Risk
                      </label>
                      <select
                        className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
                        value={form.actions[level]}
                        onChange={(e) =>
                          setForm({
                            ...form,
                            actions: {
                              ...form.actions,
                              [level]: e.target.value,
                            },
                          })
                        }
                      >
                        <option value="allow">Allow</option>
                        <option value="challenge">Challenge</option>
                        <option value="block">Block</option>
                      </select>
                    </div>
                  ))}
                </div>
              </div>

              {/* Challenge types */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Challenge Types
                </label>
                <div className="flex flex-wrap gap-2">
                  {["slider", "click", "rotate", "sms"].map((type) => (
                    <button
                      key={type}
                      type="button"
                      onClick={() => toggleChallengeType(type)}
                      className={`rounded-full border px-3 py-1 text-xs font-medium transition-colors ${
                        form.challenge_types.includes(type)
                          ? "border-primary-500 bg-primary-50 text-primary-700"
                          : "border-gray-300 bg-white text-gray-500 hover:bg-gray-50"
                      }`}
                    >
                      {type}
                    </button>
                  ))}
                </div>
              </div>
            </div>
            <div className="mt-6 flex justify-end gap-3">
              <button
                className="btn-secondary"
                onClick={() => {
                  setShowModal(false);
                  setEditingId(null);
                }}
              >
                Cancel
              </button>
              <button
                className="btn-primary"
                onClick={handleSave}
                disabled={saving}
              >
                {saving ? "Saving..." : editingId ? "Update" : "Create"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Policy cards */}
      <div className="space-y-4">
        {policies.map((policy) => (
          <div
            key={policy.id}
            className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm"
          >
            <div className="flex items-start justify-between">
              <div>
                <h3 className="text-lg font-semibold text-gray-900">
                  {policy.name}
                </h3>
                <p className="mt-1 text-sm text-gray-500">
                  {policy.description}
                </p>
              </div>
              <div className="flex gap-3">
                <button
                  className="text-sm text-primary-600 hover:text-primary-800 font-medium"
                  onClick={() => openEdit(policy)}
                >
                  Edit
                </button>
                <button
                  className="text-sm text-red-600 hover:text-red-800 font-medium"
                  onClick={() => handleDelete(policy.id)}
                >
                  Delete
                </button>
              </div>
            </div>

            {/* Threshold + action table */}
            <div className="mt-4 overflow-hidden rounded-lg border border-gray-200">
              <table className="w-full text-sm">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">
                      Risk Level
                    </th>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">
                      Threshold
                    </th>
                    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500">
                      Action
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-100">
                  <tr>
                    <td className="px-4 py-2">
                      <span className="badge-green">Low</span>
                    </td>
                    <td className="px-4 py-2 text-gray-600">
                      &lt; {policy.thresholds.low}
                    </td>
                    <td className="px-4 py-2">
                      {actionBadge(policy.actions.low)}
                    </td>
                  </tr>
                  <tr>
                    <td className="px-4 py-2">
                      <span className="badge-yellow">Medium</span>
                    </td>
                    <td className="px-4 py-2 text-gray-600">
                      {policy.thresholds.low} - {policy.thresholds.medium}
                    </td>
                    <td className="px-4 py-2">
                      {actionBadge(policy.actions.medium)}
                    </td>
                  </tr>
                  <tr>
                    <td className="px-4 py-2">
                      <span className="badge-red">High</span>
                    </td>
                    <td className="px-4 py-2 text-gray-600">
                      &gt; {policy.thresholds.high}
                    </td>
                    <td className="px-4 py-2">
                      {actionBadge(policy.actions.high)}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>

            {/* Challenge types */}
            <div className="mt-4 flex items-center gap-2">
              <span className="text-xs font-medium text-gray-500">
                Challenge types:
              </span>
              {policy.challenge_types.map((t) => (
                <span
                  key={t}
                  className="rounded-full border border-gray-200 bg-gray-50 px-2 py-0.5 text-xs text-gray-600"
                >
                  {t}
                </span>
              ))}
            </div>

            <p className="mt-3 text-xs text-gray-400">
              Updated {new Date(policy.updated_at).toLocaleDateString()}
            </p>
          </div>
        ))}
      </div>
    </div>
  );
}
