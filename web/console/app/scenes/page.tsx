"use client";

import { useEffect, useState } from "react";
import { listScenes, createScene, deleteScene, type Scene } from "@/lib/api";

const DEMO_SCENES: Scene[] = [
  {
    id: "scene_1",
    app_id: "app_1",
    name: "User Login",
    scene_type: "login",
    policy_id: "pol_1",
    policy_name: "Standard Protection",
    challenge_mode: "managed",
    status: "active",
    created_at: "2026-01-20T09:00:00Z",
  },
  {
    id: "scene_2",
    app_id: "app_1",
    name: "User Registration",
    scene_type: "register",
    policy_id: "pol_2",
    policy_name: "Strict Protection",
    challenge_mode: "interactive",
    status: "active",
    created_at: "2026-01-22T14:30:00Z",
  },
  {
    id: "scene_3",
    app_id: "app_1",
    name: "Checkout Payment",
    scene_type: "payment",
    policy_id: "pol_2",
    policy_name: "Strict Protection",
    challenge_mode: "interactive",
    status: "active",
    created_at: "2026-02-05T10:15:00Z",
  },
  {
    id: "scene_4",
    app_id: "app_2",
    name: "Comment Submission",
    scene_type: "comment",
    policy_id: "pol_1",
    policy_name: "Standard Protection",
    challenge_mode: "invisible",
    status: "inactive",
    created_at: "2026-02-18T16:45:00Z",
  },
];

function statusBadge(status: string) {
  return status === "active" ? (
    <span className="badge-green">Active</span>
  ) : (
    <span className="badge-gray">Inactive</span>
  );
}

function sceneTypeBadge(type: string) {
  const colors: Record<string, string> = {
    login: "badge-green",
    register: "badge-yellow",
    payment: "badge-red",
    comment: "badge-gray",
    custom: "badge-gray",
  };
  return <span className={colors[type] || "badge-gray"}>{type}</span>;
}

function modeBadge(mode: string) {
  switch (mode) {
    case "invisible":
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-purple-100 px-2.5 py-0.5 text-xs font-medium text-purple-800">
          <span className="h-1.5 w-1.5 rounded-full bg-purple-500" />
          Invisible
        </span>
      );
    case "managed":
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-blue-100 px-2.5 py-0.5 text-xs font-medium text-blue-800">
          <span className="h-1.5 w-1.5 rounded-full bg-blue-500" />
          Managed
        </span>
      );
    case "interactive":
      return (
        <span className="inline-flex items-center gap-1 rounded-full bg-orange-100 px-2.5 py-0.5 text-xs font-medium text-orange-800">
          <span className="h-1.5 w-1.5 rounded-full bg-orange-500" />
          Interactive
        </span>
      );
    default:
      return <span className="badge-gray">{mode}</span>;
  }
}

export default function ScenesPage() {
  const [scenes, setScenes] = useState<Scene[]>(DEMO_SCENES);
  const [showCreate, setShowCreate] = useState(false);
  const [form, setForm] = useState({
    name: "",
    scene_type: "login" as Scene["scene_type"],
    challenge_mode: "managed" as Scene["challenge_mode"],
    policy_id: "",
  });
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    listScenes()
      .then(setScenes)
      .catch(() => {});
  }, []);

  async function handleCreate() {
    if (!form.name) return;
    setSaving(true);
    try {
      const scene = await createScene({
        name: form.name,
        scene_type: form.scene_type,
        challenge_mode: form.challenge_mode,
        policy_id: form.policy_id || "pol_1",
        app_id: "app_1",
      });
      setScenes((prev) => [scene, ...prev]);
      setShowCreate(false);
      setForm({ name: "", scene_type: "login", challenge_mode: "managed", policy_id: "" });
    } catch {
      const fake: Scene = {
        id: `scene_${Date.now()}`,
        app_id: "app_1",
        name: form.name,
        scene_type: form.scene_type,
        policy_id: form.policy_id || "pol_1",
        policy_name: "Standard Protection",
        challenge_mode: form.challenge_mode,
        status: "active",
        created_at: new Date().toISOString(),
      };
      setScenes((prev) => [fake, ...prev]);
      setShowCreate(false);
      setForm({ name: "", scene_type: "login", challenge_mode: "managed", policy_id: "" });
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this scene?")) return;
    try {
      await deleteScene(id);
    } catch {
      // demo
    }
    setScenes((prev) => prev.filter((s) => s.id !== id));
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Scenes</h1>
          <p className="mt-1 text-sm text-gray-500">
            Configure challenge scenarios for different user interactions
          </p>
        </div>
        <button className="btn-primary" onClick={() => setShowCreate(true)}>
          + Create Scene
        </button>
      </div>

      {/* Create modal */}
      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h2 className="text-lg font-semibold text-gray-900 mb-4">
              Create New Scene
            </h2>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Scene Name
                </label>
                <input
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="User Login"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Scene Type
                </label>
                <select
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
                  value={form.scene_type}
                  onChange={(e) =>
                    setForm({ ...form, scene_type: e.target.value as Scene["scene_type"] })
                  }
                >
                  <option value="login">Login</option>
                  <option value="register">Register</option>
                  <option value="payment">Payment</option>
                  <option value="comment">Comment</option>
                  <option value="custom">Custom</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Challenge Mode
                </label>
                <select
                  className="w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-primary-500 focus:outline-none focus:ring-1 focus:ring-primary-500"
                  value={form.challenge_mode}
                  onChange={(e) =>
                    setForm({
                      ...form,
                      challenge_mode: e.target.value as Scene["challenge_mode"],
                    })
                  }
                >
                  <option value="invisible">Invisible</option>
                  <option value="managed">Managed</option>
                  <option value="interactive">Interactive</option>
                </select>
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

      {/* Scene cards */}
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        {scenes.map((scene) => (
          <div
            key={scene.id}
            className="rounded-xl border border-gray-200 bg-white p-6 shadow-sm hover:shadow-md transition-shadow"
          >
            <div className="flex items-start justify-between">
              <div>
                <h3 className="font-semibold text-gray-900">{scene.name}</h3>
                <p className="mt-1 text-xs text-gray-500">
                  ID: {scene.id}
                </p>
              </div>
              {statusBadge(scene.status)}
            </div>
            <div className="mt-4 flex flex-wrap gap-2">
              {sceneTypeBadge(scene.scene_type)}
              {modeBadge(scene.challenge_mode)}
            </div>
            <div className="mt-4 space-y-1 text-sm text-gray-600">
              <p>
                <span className="font-medium text-gray-700">Policy:</span>{" "}
                {scene.policy_name || scene.policy_id}
              </p>
              <p>
                <span className="font-medium text-gray-700">Created:</span>{" "}
                {new Date(scene.created_at).toLocaleDateString()}
              </p>
            </div>
            <div className="mt-4 flex gap-3 border-t border-gray-100 pt-4">
              <button className="text-sm text-primary-600 hover:text-primary-800 font-medium">
                Edit
              </button>
              <button
                className="text-sm text-red-600 hover:text-red-800 font-medium"
                onClick={() => handleDelete(scene.id)}
              >
                Delete
              </button>
            </div>
          </div>
        ))}
      </div>

      {scenes.length === 0 && (
        <div className="rounded-xl border border-gray-200 bg-white py-12 text-center">
          <p className="text-sm text-gray-500">
            No scenes configured. Create one to get started.
          </p>
        </div>
      )}
    </div>
  );
}
