// ──────────────────────────────────────────────
// Types
// ──────────────────────────────────────────────

export interface App {
  id: string;
  name: string;
  site_key: string;
  secret_key: string;
  domain: string;
  status: "active" | "inactive" | "suspended";
  created_at: string;
  updated_at: string;
}

export interface Scene {
  id: string;
  app_id: string;
  name: string;
  scene_type: "login" | "register" | "payment" | "comment" | "custom";
  policy_id: string;
  policy_name?: string;
  challenge_mode: "invisible" | "managed" | "interactive";
  status: "active" | "inactive";
  created_at: string;
}

export interface Policy {
  id: string;
  name: string;
  description: string;
  thresholds: {
    low: number;
    medium: number;
    high: number;
  };
  actions: {
    low: "allow" | "challenge" | "block";
    medium: "allow" | "challenge" | "block";
    high: "allow" | "challenge" | "block";
  };
  challenge_types: string[];
  created_at: string;
  updated_at: string;
}

export interface ChallengeLog {
  id: string;
  session_id: string;
  app_id: string;
  scene_id: string;
  ip: string;
  user_agent: string;
  risk_score: number;
  challenge_type: "slider" | "click" | "rotate" | "sms" | "none";
  result: "pass" | "fail" | "timeout" | "skip";
  timestamp: string;
  duration_ms: number;
}

export interface DashboardStats {
  total_requests: number;
  challenge_rate: number;
  pass_rate: number;
  bot_blocked: number;
  avg_risk_score: number;
  false_positive_rate: number;
  requests_trend: { date: string; count: number }[];
  challenge_distribution: { type: string; count: number }[];
}

export interface PrecheckRequest {
  site_key: string;
  scene_id: string;
  client_info: Record<string, unknown>;
}

export interface PrecheckResponse {
  session_id: string;
  risk_score: number;
  challenge_required: boolean;
  challenge_type: string;
  token?: string;
}

export interface ChallengeRenderResponse {
  challenge_id: string;
  type: string;
  payload: Record<string, unknown>;
}

export interface ChallengeVerifyRequest {
  challenge_id: string;
  answer: Record<string, unknown>;
}

export interface ChallengeVerifyResponse {
  success: boolean;
  token: string;
  score: number;
}

export interface SiteVerifyRequest {
  secret: string;
  token: string;
  remote_ip?: string;
}

export interface SiteVerifyResponse {
  success: boolean;
  score: number;
  action: string;
  challenge_ts: string;
  hostname: string;
  errors: string[];
}

// ──────────────────────────────────────────────
// API Client
// ──────────────────────────────────────────────

const BASE = "/api";

async function request<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "Content-Type": "application/json", ...options?.headers },
    ...options,
  });
  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(`API ${res.status}: ${text || res.statusText}`);
  }
  return res.json();
}

// ── Captcha flow ──

export function precheck(body: PrecheckRequest) {
  return request<PrecheckResponse>("/captcha/precheck", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function challengeRender(sessionId: string) {
  return request<ChallengeRenderResponse>(
    `/captcha/challenge/render?session_id=${sessionId}`
  );
}

export function challengeVerify(body: ChallengeVerifyRequest) {
  return request<ChallengeVerifyResponse>("/captcha/challenge/verify", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function siteVerify(body: SiteVerifyRequest) {
  return request<SiteVerifyResponse>("/captcha/siteverify", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

// ── Apps ──

export function listApps() {
  return request<App[]>("/apps");
}

export function getApp(id: string) {
  return request<App>(`/apps/${id}`);
}

export function createApp(body: Partial<App>) {
  return request<App>("/apps", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function updateApp(id: string, body: Partial<App>) {
  return request<App>(`/apps/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteApp(id: string) {
  return request<void>(`/apps/${id}`, { method: "DELETE" });
}

// ── Scenes ──

export function listScenes() {
  return request<Scene[]>("/scenes");
}

export function getScene(id: string) {
  return request<Scene>(`/scenes/${id}`);
}

export function createScene(body: Partial<Scene>) {
  return request<Scene>("/scenes", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function updateScene(id: string, body: Partial<Scene>) {
  return request<Scene>(`/scenes/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deleteScene(id: string) {
  return request<void>(`/scenes/${id}`, { method: "DELETE" });
}

// ── Policies ──

export function listPolicies() {
  return request<Policy[]>("/policies");
}

export function getPolicy(id: string) {
  return request<Policy>(`/policies/${id}`);
}

export function createPolicy(body: Partial<Policy>) {
  return request<Policy>("/policies", {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export function updatePolicy(id: string, body: Partial<Policy>) {
  return request<Policy>(`/policies/${id}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export function deletePolicy(id: string) {
  return request<void>(`/policies/${id}`, { method: "DELETE" });
}

// ── Stats / Logs ──

export function getDashboardStats() {
  return request<DashboardStats>("/stats/dashboard");
}

export function listLogs(params?: {
  search?: string;
  page?: number;
  limit?: number;
}) {
  const qs = new URLSearchParams();
  if (params?.search) qs.set("search", params.search);
  if (params?.page) qs.set("page", String(params.page));
  if (params?.limit) qs.set("limit", String(params.limit));
  const q = qs.toString();
  return request<{ logs: ChallengeLog[]; total: number }>(
    `/logs${q ? `?${q}` : ""}`
  );
}
