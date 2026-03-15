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

## API Endpoints (21)

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

### Console Management (API key auth)
| Method | Path | Description |
|--------|------|-------------|
| `POST/GET` | `/v1/apps` | Create / list applications |
| `GET/DELETE` | `/v1/apps/:id` | Get / delete application |
| `POST/GET` | `/v1/scenes` | Create / list scenes |
| `POST/GET/DELETE` | `/v1/webhooks` | Manage webhooks |
| `POST` | `/v1/events/feedback` | Submit false positive/negative feedback |
| `GET` | `/v1/stats/dashboard` | Dashboard statistics |
| `GET` | `/v1/analytics/countries` | Per-country breakdown |
| `GET` | `/v1/analytics/devices` | Browser/platform/OS breakdown |
| `GET` | `/v1/analytics/challenges` | Challenge type performance |
| `GET` | `/v1/analytics/risk-trends` | 7-day risk trends |

### Other
| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/v1/i18n/:lang` | Challenge UI translations |
| `GET` | `/health` | Health check |

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
# Risk precheck (normal user → pass)
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

# Risk precheck (headless bot → deny)
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
```

### Run Tests
```bash
make test   # 59 tests
```

## JavaScript SDK

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

SDK features:
- Browser fingerprinting (Canvas, WebGL, screen, timezone, plugins)
- Behavior tracking (mouse moves, clicks, keystrokes, scroll, entropy calculation)
- Slider and click-select challenge renderers (touch support)
- Automatic precheck → challenge → verify flow

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
  "ip_blacklist": ["192.0.2.0/24"]
}
```

## Internationalization

10 languages supported out of the box:

| Language | Code |
|----------|------|
| English | `en` |
| 中文 | `zh` |
| 日本語 | `ja` |
| 한국어 | `ko` |
| Español | `es` |
| Português | `pt` |
| Bahasa Indonesia | `id` |
| ภาษาไทย | `th` |
| Tiếng Việt | `vi` |
| العربية | `ar` |

## A/B Testing

Built-in experiment framework for optimizing challenge UX:
- Deterministic variant assignment (consistent per fingerprint)
- Weighted traffic splitting
- Real-time metrics: impressions, completions, failures, avg duration, conversion rate

## Webhooks

Real-time event notifications with HMAC-signed payloads:

| Event | Description |
|-------|-------------|
| `challenge.created` | New challenge session started |
| `challenge.passed` | User passed challenge |
| `challenge.failed` | User failed challenge |
| `risk.high_detected` | High risk score detected |
| `risk.deny_triggered` | Request denied |
| `bot.detected` | Bot pattern identified |
| `rate_limit.hit` | Rate limit exceeded |

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go, Gin |
| Database | PostgreSQL (auto-fallback to in-memory) |
| Cache | Redis (auto-fallback to in-memory) |
| Console | Next.js, React, TypeScript, Tailwind, Recharts |
| SDK | TypeScript (browser) |
| Auth | API Key (console), Site Key (SDK), HMAC token signing |
| Containerization | Docker, docker-compose |

## Project Structure

```
engagelab-captcha/
├── cmd/server/                    # Application entrypoint
├── sdk/js/                        # Browser JavaScript SDK
├── internal/
│   ├── service/
│   │   ├── risk/                  # Risk engine + proxy/bot detectors
│   │   ├── challenge/             # Challenge engine + i18n + A/B testing
│   │   ├── policy/                # Policy evaluation engine
│   │   ├── verify/                # Server-side token verification
│   │   └── webhook/               # Webhook delivery service
│   ├── handler/                   # 9 HTTP handlers + analytics
│   ├── middleware/                # API key + site key auth, CORS
│   ├── model/                     # Data models (6 files)
│   ├── repository/                # In-memory store with seed data
│   └── router/                    # Route wiring + dependency injection
├── migrations/                    # PostgreSQL schema
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

## API 端点 (21 个)

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

### 控制台管理 (API key 认证)
| 方法 | 路径 | 说明 |
|------|------|------|
| `POST/GET` | `/v1/apps` | 创建 / 列出应用 |
| `GET/DELETE` | `/v1/apps/:id` | 获取 / 删除应用 |
| `POST/GET` | `/v1/scenes` | 创建 / 列出场景 |
| `POST/GET/DELETE` | `/v1/webhooks` | 管理 Webhook |
| `POST` | `/v1/events/feedback` | 提交误杀/漏报反馈 |
| `GET` | `/v1/stats/dashboard` | 仪表盘统计 |
| `GET` | `/v1/analytics/countries` | 国家分布分析 |
| `GET` | `/v1/analytics/devices` | 浏览器/平台/OS 分析 |
| `GET` | `/v1/analytics/challenges` | 挑战类型效果 |
| `GET` | `/v1/analytics/risk-trends` | 7 天风险趋势 |

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
cd web/console
npm install
npm run dev
```

### Docker 启动
```bash
docker-compose up -d
```

### 测试 API
```bash
# 正常用户 → 放行
curl -X POST http://localhost:8080/v1/risk/precheck \
  -H "X-Site-Key: sk_demo_site_key_abc" \
  -H "Content-Type: application/json" \
  -d '{"app_id":"app-001","ip":"8.8.8.8","ua":"Mozilla/5.0 Chrome/120","fingerprint":"fp_user","behavior_data":{"mouse_moves":80,"mouse_entropy":0.7,"duration_ms":6000}}'

# Headless Bot → 拦截
curl -X POST http://localhost:8080/v1/risk/precheck \
  -H "X-Site-Key: sk_demo_site_key_abc" \
  -H "Content-Type: application/json" \
  -d '{"app_id":"app-001","ip":"159.89.1.1","ua":"HeadlessChrome","fingerprint":"fp_bot","behavior_data":{"mouse_moves":0,"duration_ms":50,"webdriver":true}}'
```

### 运行测试
```bash
make test   # 59 个测试
```

## JavaScript SDK

```html
<script src="https://cdn.engagelab.com/captcha/v1/engagelab-captcha.js"></script>
<script>
  EngagelabCaptcha.init({
    siteKey: 'your_site_key',
    scene: 'register',
    container: '#captcha',
    onSuccess: (token) => {
      // 将 token 发送到后端做服务端校验
    },
  });

  // 表单提交时触发验证
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    await EngagelabCaptcha.execute();
  });
</script>
```

## 挑战类型

| 类型 | 触发条件 | 用户体验 |
|------|---------|---------|
| **无感** | 低风险 (score < 15) | 零摩擦，静默签发 token |
| **滑块** | 中风险 (15-50) | 拖动滑块到目标位置 |
| **点选** | 高风险 (50-70) | 按正确顺序点击目标 |
| **拦截** | 极高风险 (> 70) | 直接拒绝请求 |

## 国际化

开箱即用支持 10 种语言：en / zh / ja / ko / es / pt / id / th / vi / ar

## 应用场景

| 场景 | 防护能力 |
|------|---------|
| **注册** | 防注册机、虚假账号 |
| **登录** | 防撞库、暴力破解 |
| **营销活动** | 防薅羊毛、优惠券滥用 |
| **评论发帖** | 防垃圾内容 |
| **API 调用** | 防脚本刷爆 |
| **票务/抢购** | 防黄牛机器人 |

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go, Gin |
| 数据库 | PostgreSQL（自动降级内存）|
| 缓存 | Redis（自动降级内存）|
| 控制台 | Next.js, React, TypeScript, Tailwind, Recharts |
| SDK | TypeScript (浏览器) |
| 认证 | API Key (控制台) + Site Key (SDK) + HMAC 签名 |
| 容器化 | Docker, docker-compose |

## 许可证

MIT
