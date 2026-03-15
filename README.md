# EngageLab CAPTCHA

**English** | [中文](#engagelab-captcha-中文)

**Smart CAPTCHA & Abuse Prevention Platform**

EngageLab CAPTCHA is a risk-first, adaptive challenge platform that protects web applications from bots, credential stuffing, account fraud, and abuse — while minimizing friction for real users.

Unlike traditional CAPTCHAs that challenge every request, EngageLab assesses risk first, then decides the appropriate response: silent pass, invisible verification, interactive challenge, or outright block.

## How It Works

```
Request → Risk Engine (6 signal layers) → Policy Engine → Decision
                                                            │
                                              ┌─────────────┼─────────────┐
                                              │             │             │
                                            Pass      Challenge        Deny
                                         (score<15)    (15-70)        (>70)
                                              │             │
                                         Return token  Show CAPTCHA
                                                      (slider/click)
                                                            │
                                                    User completes
                                                            │
                                                  Server-side verify
```

## Risk Engine — 6 Signal Layers

| Layer | Signals | Score Range |
|-------|---------|-------------|
| **IP Reputation** | Blacklisted IPs, datacenter CIDRs (AWS/GCP/Azure) | 0-30 |
| **Rate Limiting** | Per-IP and per-fingerprint sliding window counters | 0-30 |
| **User-Agent** | Bot tokens (curl, python-requests, headless, selenium) | 0-15 |
| **Behavior Analysis** | Mouse entropy, page timing, keystrokes, scroll events | 0-25 |
| **Proxy/VPN/Tor** | Tor exit nodes, VPN ranges, CGNAT, extended DC (DO/Hetzner/OVH) | 0-25 |
| **Bot Patterns** | HeadlessChrome, WebDriver, automation frameworks, impossible mouse rates | 0-25 |

Score is clamped to 0-100. Risk levels: low (0-15), medium (16-40), high (41-70), critical (71+).

## API Endpoints (30)

### SDK Endpoints (site_key auth)
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/risk/precheck` | Risk assessment + challenge decision |
| `POST` | `/v1/challenge/render` | Get challenge configuration |
| `POST` | `/v1/challenge/verify` | Submit challenge answer |

### Server-side Verification
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/siteverify` | Validate challenge token (secret key) |

### Auth (no auth required)
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/auth/register` | Register user + create org + default app |
| `POST` | `/v1/auth/login` | Login with email/password |

### Console Management (API key auth)
| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/v1/apps` | Create application |
| `GET` | `/v1/apps` | List applications |
| `GET` | `/v1/apps/:id` | Get application |
| `DELETE` | `/v1/apps/:id` | Delete application |
| `POST` | `/v1/scenes` | Create scene |
| `GET` | `/v1/scenes` | List scenes |
| `GET` | `/v1/policies` | List policies |
| `POST` | `/v1/webhooks` | Create webhook |
| `GET` | `/v1/webhooks` | List webhooks |
| `DELETE` | `/v1/webhooks/:id` | Delete webhook |
| `POST` | `/v1/events/feedback` | Submit false positive/negative feedback |
| `POST` | `/v1/account/api-keys` | Generate API key |
| `GET` | `/v1/account/api-keys` | List API keys |

### Analytics & Monitoring
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/v1/stats/dashboard` | Dashboard statistics |
| `GET` | `/v1/analytics/countries` | Per-country breakdown |
| `GET` | `/v1/analytics/devices` | Browser / platform / OS breakdown |
| `GET` | `/v1/analytics/challenges` | Challenge type performance |
| `GET` | `/v1/analytics/risk-trends` | 7-day risk trends |

### Attack Monitoring
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/v1/threats` | List detected threats |
| `GET` | `/v1/threats/dashboard` | Threat dashboard (timeline, top IPs, severity) |
| `POST` | `/v1/threats/:id/mitigate` | Mark threat as mitigated |

### Infrastructure
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/v1/i18n/:lang` | Challenge UI translations (10 languages) |
| `GET` | `/health` | Health check |
| `GET` | `/metrics` | Prometheus metrics |

## Quick Start

### Prerequisites
- Go 1.23+
- Node.js 20+ (for console)
- PostgreSQL 16 & Redis 7 (optional — auto-fallback to in-memory)

### Run Backend
```bash
cd "engagelab captcha"
go mod tidy
make dev
```
Server starts on `http://localhost:8080`.

### Run Console
```bash
cd web/console
npm install
npm run dev
```
Console at `http://localhost:3000`.

### Run with Docker
```bash
docker-compose up -d
```

### Test the API
```bash
# Register a new account
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@example.com","password":"SecureP@ss1","name":"Dev","company":"MyCo"}'

# Risk precheck (normal user → pass, score=5)
curl -X POST http://localhost:8080/v1/risk/precheck \
  -H "X-Site-Key: sk_demo_site_key_abc" \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "app-001",
    "scene_id": "scene_register",
    "ip": "8.8.8.8",
    "ua": "Mozilla/5.0 Chrome/120",
    "fingerprint": "fp_user123",
    "behavior_data": {"mouse_moves":80,"clicks":4,"duration_ms":6000,"mouse_entropy":0.7}
  }'

# Risk precheck (headless bot → deny, score=87)
curl -X POST http://localhost:8080/v1/risk/precheck \
  -H "X-Site-Key: sk_demo_site_key_abc" \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "app-001",
    "ip": "159.89.1.1",
    "ua": "HeadlessChrome",
    "fingerprint": "fp_bot",
    "behavior_data": {"mouse_moves":0,"duration_ms":50,"webdriver":true}
  }'

# Server-side verification
curl -X POST http://localhost:8080/v1/siteverify \
  -H "Content-Type: application/json" \
  -d '{"token": "<token_from_precheck>", "secret": "sec_demo_secret_key_xyz"}'

# View threat dashboard
curl http://localhost:8080/v1/threats/dashboard \
  -H "X-API-Key: ak_demo_key_123456"
```

### Run Tests
```bash
make test   # 59 tests
```

## SDKs

### JavaScript (Browser)

```html
<script src="https://cdn.engagelab.com/captcha/v1/engagelab-captcha.js"></script>
<div id="captcha"></div>
<script>
  EngagelabCaptcha.init({
    siteKey: 'your_site_key',
    scene: 'register',
    container: '#captcha',
    onSuccess: (token) => {
      // Send token to your backend for server-side verification
      fetch('/api/register', {
        method: 'POST',
        body: JSON.stringify({ captcha_token: token, ...formData }),
      });
    },
  });

  // Trigger verification on form submit
  document.querySelector('form').addEventListener('submit', async (e) => {
    e.preventDefault();
    await EngagelabCaptcha.execute();
  });
</script>
```

Features: browser fingerprinting (Canvas/WebGL/screen/timezone), behavior tracking (mouse/click/keystroke/scroll/entropy), slider & click-select renderers with touch support, auto precheck-challenge-verify flow.

### iOS (Swift Package)

```swift
import EngagelabCaptcha

let captcha = EngagelabCaptcha(config: CaptchaConfig(
    siteKey: "your_site_key",
    scene: .register
))

let result = try await captcha.execute()
switch result {
case .passed(let token):
    // Send token to your backend
case .challengeRequired(let challengeID, let type):
    // Show challenge UI (handled automatically with WKWebView)
case .denied:
    // Request blocked
case .error(let err):
    // Handle error
}
```

Features: UIDevice fingerprinting, touch behavior tracking with Shannon entropy, WKWebView challenge rendering, async/await API.

### Android (Kotlin)

```kotlin
import com.engagelab.captcha.EngagelabCaptcha
import com.engagelab.captcha.CaptchaConfig
import com.engagelab.captcha.CaptchaResult
import com.engagelab.captcha.SceneType

val captcha = EngagelabCaptcha(CaptchaConfig(
    siteKey = "your_site_key",
    scene = SceneType.REGISTER
))

when (val result = captcha.execute(context)) {
    is CaptchaResult.Passed -> {
        // Send result.token to your backend
    }
    is CaptchaResult.ChallengeRequired -> {
        // Show challenge (handled automatically with WebView)
    }
    is CaptchaResult.Denied -> {
        // Request blocked
    }
    is CaptchaResult.Error -> {
        // Handle result.exception
    }
}
```

Features: emulator detection, MotionEvent behavior tracking, WebView challenge rendering with JS bridge, coroutine suspend API, kotlinx.serialization.

## Challenge Types

| Type | When Used | User Experience |
|------|-----------|----------------|
| **Invisible** | Low risk (score < 15) | Zero friction, token issued silently |
| **Slider** | Medium risk (15-50) | Drag slider to target position |
| **Click Select** | High risk (50-70) | Click items in correct order |
| **Deny** | Critical risk (> 70) | Request blocked |

## Policy Engine

Per-scene configurable thresholds and actions:

```json
{
  "scene_type": "register",
  "threshold_low": 15,
  "threshold_high": 60,
  "action_low": "pass",
  "action_mid": "slider",
  "action_high": "deny",
  "rate_limit_rpm": 30,
  "ip_whitelist": ["10.0.0.0/8"],
  "ip_blacklist": ["192.0.2.0/24"]
}
```

## Attack Monitoring

Real-time threat detection dashboard with 5 attack types:

| Threat Type | Description |
|-------------|-------------|
| `credential_stuffing` | Distributed login attempts from Tor/proxy |
| `bot_wave` | Registration bots from datacenter IPs |
| `rate_abuse` | Single IP exceeding rate limits |
| `scraping` | Systematic crawling with rotating fingerprints |
| `brute_force` | Password brute force targeting specific accounts |

Threat actions: view details, mitigate, resolve.

## Internationalization

10 languages out of the box:

| | | | | |
|---|---|---|---|---|
| English `en` | 中文 `zh` | 日本語 `ja` | 한국어 `ko` | Español `es` |
| Português `pt` | Indonesia `id` | ไทย `th` | Tiếng Việt `vi` | العربية `ar` |

## A/B Testing

Built-in experiment framework for challenge optimization:
- Deterministic variant assignment (consistent per fingerprint via SHA-256)
- Weighted traffic splitting with configurable experiment percentage
- Real-time metrics: impressions, completions, failures, avg duration, conversion rate

## Webhooks

8 event types with HMAC-signed payloads and 3x exponential backoff retry:

| Event | Description |
|-------|-------------|
| `challenge.created` | New challenge session started |
| `challenge.passed` | User passed challenge |
| `challenge.failed` | User failed challenge |
| `risk.high_detected` | High risk score detected |
| `risk.deny_triggered` | Request denied |
| `bot.detected` | Bot pattern identified |
| `feedback.received` | Customer feedback submitted |
| `rate_limit.hit` | Rate limit exceeded |

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go, Gin, bcrypt |
| Database | PostgreSQL (auto-fallback to in-memory) |
| Cache | Redis (auto-fallback to in-memory) |
| Console | Next.js, React, TypeScript, Tailwind, Recharts |
| JS SDK | TypeScript (Canvas/WebGL fingerprint, behavior tracking) |
| iOS SDK | Swift (WKWebView, touch entropy, UIDevice fingerprint) |
| Android SDK | Kotlin (WebView, MotionEvent tracking, emulator detection) |
| Auth | API Key + Site Key + bcrypt + HMAC token signing |
| Monitoring | Prometheus metrics |
| Containerization | Docker, docker-compose |

## Project Structure

```
engagelab-captcha/
├── cmd/server/                    # Application entrypoint
├── sdk/
│   ├── js/                        # Browser JavaScript SDK
│   ├── ios/                       # iOS Swift Package
│   └── android/                   # Android Kotlin library
├── internal/
│   ├── service/
│   │   ├── risk/                  # Risk engine + proxy/VPN/Tor + bot pattern detectors
│   │   ├── challenge/             # Challenge engine + i18n (10 langs) + A/B testing
│   │   ├── policy/                # Policy evaluation (thresholds, whitelist, rate limits)
│   │   ├── verify/                # Server-side HMAC token verification + replay prevention
│   │   └── webhook/               # Event delivery with HMAC signing + retry
│   ├── handler/                   # 12 HTTP handlers (30 endpoints)
│   ├── metrics/                   # Prometheus metrics collector
│   ├── middleware/                # API key + site key auth, CORS
│   ├── model/                     # Data models (7 files)
│   ├── repository/                # In-memory store with seed data
│   └── router/                    # Route wiring + dependency injection
├── migrations/                    # 2 PostgreSQL migrations
├── web/console/                   # Next.js dashboard (5 pages)
├── Dockerfile
├── docker-compose.yaml
└── Makefile
```

## Use Cases

| Scenario | Protection |
|----------|-----------|
| **Registration** | Block registration bots, fake accounts |
| **Login** | Prevent credential stuffing, brute force |
| **Promotions** | Stop coupon/voucher abuse, scalper bots |
| **Comments** | Filter spam posts and reviews |
| **API Calls** | Rate limit automated API abuse |
| **Ticketing** | Prevent scalper bots from bulk-buying |

## License

MIT

---

# EngageLab CAPTCHA 中文

[English](#engagelab-captcha)

**智能 Bot 防护与人机验证平台**

EngageLab CAPTCHA 是一个风险优先的自适应验证平台，保护 Web 应用免受机器人、撞库、账户欺诈和滥用的侵害 —— 同时最大限度降低对真实用户的干扰。

与传统验证码对每个请求都弹挑战不同，EngageLab 先评估风险，再决定响应：静默放行、无感验证、交互式挑战或直接拦截。

## 工作原理

```
请求 → 风险引擎 (6 层信号) → 策略引擎 → 决策
                                          │
                                ┌─────────┼─────────┐
                                │         │         │
                              放行      挑战       拦截
                           (score<15) (15-70)    (>70)
                                │         │
                           返回 token  显示验证码
                                       (滑块/点选)
                                          │
                                      用户完成
                                          │
                                    服务端校验
```

## 风险引擎 — 6 层信号

| 层级 | 信号 | 分值范围 |
|------|------|---------|
| **IP 信誉** | 黑名单 IP、数据中心 CIDR (AWS/GCP/Azure) | 0-30 |
| **频率限制** | 按 IP 和指纹的滑动窗口计数器 | 0-30 |
| **UA 分析** | Bot 特征 (curl, python-requests, headless, selenium) | 0-15 |
| **行为分析** | 鼠标熵值、页面停留时间、按键、滚动事件 | 0-25 |
| **代理/VPN/Tor** | Tor 出口节点、VPN 范围、CGNAT、扩展数据中心 | 0-25 |
| **Bot 模式** | HeadlessChrome、WebDriver、自动化框架、不可能鼠标速率 | 0-25 |

分值 0-100。风险等级：低 (0-15)、中 (16-40)、高 (41-70)、极高 (71+)。

## API 端点 (30 个)

### SDK 端点 (site_key 认证)
| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/v1/risk/precheck` | 风险评估 + 挑战决策 |
| `POST` | `/v1/challenge/render` | 获取挑战配置 |
| `POST` | `/v1/challenge/verify` | 提交挑战答案 |

### 服务端校验
| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/v1/siteverify` | 校验挑战 token (secret key) |

### 认证（无需鉴权）
| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/v1/auth/register` | 注册用户 + 创建组织 + 默认应用 |
| `POST` | `/v1/auth/login` | 邮箱密码登录 |

### 控制台管理 (API key 认证)
| 方法 | 路径 | 说明 |
|------|------|------|
| `POST/GET` | `/v1/apps` | 创建 / 列出应用 |
| `GET/DELETE` | `/v1/apps/:id` | 获取 / 删除应用 |
| `POST/GET` | `/v1/scenes` | 创建 / 列出场景 |
| `GET` | `/v1/policies` | 列出策略 |
| `POST/GET/DELETE` | `/v1/webhooks` | 管理 Webhook |
| `POST` | `/v1/events/feedback` | 提交误杀/漏报反馈 |
| `POST/GET` | `/v1/account/api-keys` | 生成 / 列出 API Key |

### 分析与监控
| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/v1/stats/dashboard` | 仪表盘统计 |
| `GET` | `/v1/analytics/countries` | 国家分布分析 |
| `GET` | `/v1/analytics/devices` | 浏览器/平台/OS 分析 |
| `GET` | `/v1/analytics/challenges` | 挑战类型效果 |
| `GET` | `/v1/analytics/risk-trends` | 7 天风险趋势 |

### 攻击监控
| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/v1/threats` | 威胁事件列表 |
| `GET` | `/v1/threats/dashboard` | 威胁仪表盘（时间线、Top IP、严重度）|
| `POST` | `/v1/threats/:id/mitigate` | 标记威胁已缓解 |

### 基础设施
| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/v1/i18n/:lang` | 挑战 UI 翻译（10 种语言）|
| `GET` | `/health` | 健康检查 |
| `GET` | `/metrics` | Prometheus 指标 |

## 快速开始

### 环境要求
- Go 1.23+
- Node.js 20+ (控制台)
- PostgreSQL 16 & Redis 7 (可选 — 自动降级内存)

### 启动后端
```bash
cd "engagelab captcha"
go mod tidy
make dev
```

### 启动控制台
```bash
cd web/console && npm install && npm run dev
```

### Docker 启动
```bash
docker-compose up -d
```

### 测试 API
```bash
# 注册账号
curl -X POST http://localhost:8080/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"dev@example.com","password":"SecureP@ss1","name":"Dev","company":"MyCo"}'

# 正常用户 → 放行 (score=5)
curl -X POST http://localhost:8080/v1/risk/precheck \
  -H "X-Site-Key: sk_demo_site_key_abc" -H "Content-Type: application/json" \
  -d '{"app_id":"app-001","ip":"8.8.8.8","ua":"Mozilla/5.0 Chrome/120","fingerprint":"fp_user","behavior_data":{"mouse_moves":80,"mouse_entropy":0.7,"duration_ms":6000}}'

# Headless Bot → 拦截 (score=87)
curl -X POST http://localhost:8080/v1/risk/precheck \
  -H "X-Site-Key: sk_demo_site_key_abc" -H "Content-Type: application/json" \
  -d '{"app_id":"app-001","ip":"159.89.1.1","ua":"HeadlessChrome","fingerprint":"fp_bot","behavior_data":{"mouse_moves":0,"duration_ms":50,"webdriver":true}}'

# 查看威胁仪表盘
curl http://localhost:8080/v1/threats/dashboard -H "X-API-Key: ak_demo_key_123456"
```

### 运行测试
```bash
make test   # 59 个测试
```

## SDK

### JavaScript (浏览器)
```html
<script src="https://cdn.engagelab.com/captcha/v1/engagelab-captcha.js"></script>
<script>
  EngagelabCaptcha.init({
    siteKey: 'your_site_key', scene: 'register', container: '#captcha',
    onSuccess: (token) => { /* 发送 token 到后端校验 */ },
  });
  form.addEventListener('submit', (e) => { e.preventDefault(); EngagelabCaptcha.execute(); });
</script>
```

### iOS (Swift)
```swift
let captcha = EngagelabCaptcha(config: CaptchaConfig(siteKey: "sk_...", scene: .register))
let result = try await captcha.execute()
switch result {
case .passed(let token): /* 使用 token */
case .challengeRequired: /* 自动显示 WKWebView 挑战 */
case .denied: /* 已拦截 */
}
```

### Android (Kotlin)
```kotlin
val captcha = EngagelabCaptcha(CaptchaConfig(siteKey = "sk_...", scene = SceneType.REGISTER))
when (val result = captcha.execute(context)) {
    is CaptchaResult.Passed -> { /* 使用 result.token */ }
    is CaptchaResult.ChallengeRequired -> { /* 自动 WebView 挑战 */ }
    is CaptchaResult.Denied -> { /* 已拦截 */ }
}
```

## 攻击监控

实时威胁检测，5 种攻击类型：

| 类型 | 说明 |
|------|------|
| `credential_stuffing` | Tor/代理分布式撞库 |
| `bot_wave` | 数据中心 IP 注册机波次 |
| `rate_abuse` | 单 IP 超频率限制 |
| `scraping` | 轮换指纹系统化爬取 |
| `brute_force` | 密码暴力破解 |

## 挑战类型

| 类型 | 触发条件 | 用户体验 |
|------|---------|---------|
| **无感** | 低风险 (< 15) | 零摩擦，静默签发 token |
| **滑块** | 中风险 (15-50) | 拖动滑块到目标位置 |
| **点选** | 高风险 (50-70) | 按正确顺序点击目标 |
| **拦截** | 极高风险 (> 70) | 直接拒绝请求 |

## 国际化

开箱支持 10 种语言：en / zh / ja / ko / es / pt / id / th / vi / ar

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go, Gin, bcrypt |
| 数据库 | PostgreSQL（自动降级内存）|
| 缓存 | Redis（自动降级内存）|
| 控制台 | Next.js, React, TypeScript, Tailwind, Recharts |
| JS SDK | TypeScript (Canvas/WebGL 指纹, 行为追踪) |
| iOS SDK | Swift (WKWebView, 触控熵值, UIDevice 指纹) |
| Android SDK | Kotlin (WebView, MotionEvent 追踪, 模拟器检测) |
| 认证 | API Key + Site Key + bcrypt + HMAC 签名 |
| 监控 | Prometheus 指标 |
| 容器化 | Docker, docker-compose |

## 目标客户

- 电商 / 金融科技 / 游戏 / 社交
- 在线教育 / SaaS / 票务平台
- 出海 App / 开发者平台

## 许可证

MIT
