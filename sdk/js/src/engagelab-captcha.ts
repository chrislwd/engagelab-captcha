/**
 * EngageLab CAPTCHA JavaScript SDK
 *
 * Usage:
 * ```html
 * <script src="https://cdn.engagelab.com/captcha/v1/engagelab-captcha.js"></script>
 * <script>
 *   EngagelabCaptcha.init({
 *     siteKey: 'your_site_key',
 *     scene: 'register',
 *     container: '#captcha-container',
 *     onSuccess: (token) => { /* submit token to your server *\/ },
 *     onError: (err) => { console.error(err); },
 *   });
 * </script>
 * ```
 */

interface CaptchaConfig {
  siteKey: string;
  scene: string;
  container?: string | HTMLElement;
  apiBase?: string;
  lang?: string;
  theme?: 'light' | 'dark';
  onSuccess?: (token: string) => void;
  onError?: (error: Error) => void;
  onReady?: () => void;
  onChallengeShow?: (type: string) => void;
}

interface PrecheckResponse {
  action: 'pass' | 'invisible' | 'challenge' | 'deny';
  risk_score: number;
  challenge_type: string;
  challenge_id: string;
  token: string;
}

interface ChallengeRenderResponse {
  challenge_id: string;
  type: string;
  config: {
    bg_image?: string;
    slider_image?: string;
    target_x?: number;
    targets?: Array<{ id: string; label: string; x: number; y: number }>;
    image_url?: string;
  };
  expires_at: string;
}

interface ChallengeVerifyResponse {
  success: boolean;
  token: string;
}

class EngagelabCaptchaSDK {
  private config: CaptchaConfig | null = null;
  private apiBase: string = 'https://api.engagelab.com';
  private containerEl: HTMLElement | null = null;
  private sessionId: string = '';
  private fingerprint: string = '';

  /**
   * Initialize the CAPTCHA SDK.
   */
  async init(config: CaptchaConfig): Promise<void> {
    this.config = config;
    this.apiBase = config.apiBase || this.apiBase;
    this.sessionId = this.generateId();
    this.fingerprint = await this.collectFingerprint();

    // Resolve container
    if (config.container) {
      if (typeof config.container === 'string') {
        this.containerEl = document.querySelector(config.container);
      } else {
        this.containerEl = config.container;
      }
    }

    config.onReady?.();

    // Start behavior tracking
    this.startBehaviorTracking();
  }

  /**
   * Execute CAPTCHA verification.
   * Call this when the user submits a form.
   */
  async execute(): Promise<string> {
    if (!this.config) throw new Error('SDK not initialized');

    const behaviorData = this.getBehaviorData();

    // Step 1: Risk precheck
    const precheck = await this.precheck(behaviorData);

    // Step 2: Handle action
    switch (precheck.action) {
      case 'pass':
      case 'invisible':
        this.config.onSuccess?.(precheck.token);
        return precheck.token;

      case 'deny':
        const err = new Error('Request denied by risk engine');
        this.config.onError?.(err);
        throw err;

      case 'challenge':
        return await this.showChallenge(precheck.challenge_id, precheck.challenge_type);
    }
  }

  /**
   * Reset the CAPTCHA state.
   */
  reset(): void {
    this.sessionId = this.generateId();
    this.behaviorEvents = [];
    if (this.containerEl) {
      this.containerEl.innerHTML = '';
    }
  }

  // --- Private Methods ---

  private async precheck(behaviorData: object): Promise<PrecheckResponse> {
    const resp = await fetch(`${this.apiBase}/v1/risk/precheck`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        site_key: this.config!.siteKey,
        scene_id: this.config!.scene,
        ip: '', // Server will extract from request
        ua: navigator.userAgent,
        fingerprint: this.fingerprint,
        behavior_data: behaviorData,
        session_id: this.sessionId,
      }),
    });

    if (!resp.ok) throw new Error(`Precheck failed: ${resp.status}`);
    return resp.json();
  }

  private async showChallenge(challengeId: string, challengeType: string): Promise<string> {
    this.config?.onChallengeShow?.(challengeType);

    // Get challenge config
    const renderResp = await fetch(`${this.apiBase}/v1/challenge/render`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ challenge_id: challengeId }),
    });

    if (!renderResp.ok) throw new Error('Failed to render challenge');
    const challenge: ChallengeRenderResponse = await renderResp.json();

    // Render challenge UI
    return new Promise((resolve, reject) => {
      if (!this.containerEl) {
        this.containerEl = document.createElement('div');
        document.body.appendChild(this.containerEl);
      }

      switch (challenge.type) {
        case 'slider':
          this.renderSlider(challenge, resolve, reject);
          break;
        case 'click':
          this.renderClickSelect(challenge, resolve, reject);
          break;
        default:
          reject(new Error(`Unknown challenge type: ${challenge.type}`));
      }
    });
  }

  private renderSlider(
    challenge: ChallengeRenderResponse,
    resolve: (token: string) => void,
    reject: (err: Error) => void,
  ): void {
    const container = this.containerEl!;
    container.innerHTML = `
      <div class="ec-captcha-modal" style="
        position:fixed;top:0;left:0;width:100%;height:100%;
        background:rgba(0,0,0,0.5);display:flex;align-items:center;justify-content:center;z-index:99999;
      ">
        <div class="ec-captcha-box" style="
          background:#fff;border-radius:12px;padding:24px;width:340px;box-shadow:0 20px 60px rgba(0,0,0,0.2);
        ">
          <div style="font-size:14px;color:#333;margin-bottom:12px;font-weight:600;">
            Drag the slider to verify
          </div>
          <div class="ec-slider-track" style="
            position:relative;height:44px;background:#f0f0f0;border-radius:22px;
            border:1px solid #ddd;overflow:hidden;cursor:pointer;
          ">
            <div class="ec-slider-thumb" style="
              position:absolute;left:0;top:0;width:44px;height:44px;
              background:#4A90D9;border-radius:22px;cursor:grab;
              display:flex;align-items:center;justify-content:center;color:#fff;font-size:18px;
            ">→</div>
          </div>
          <div style="font-size:11px;color:#999;margin-top:8px;text-align:center;">
            Protected by EngageLab CAPTCHA
          </div>
        </div>
      </div>
    `;

    const thumb = container.querySelector('.ec-slider-thumb') as HTMLElement;
    const track = container.querySelector('.ec-slider-track') as HTMLElement;
    let dragging = false;
    let startX = 0;

    const onStart = (e: MouseEvent | TouchEvent) => {
      dragging = true;
      startX = 'touches' in e ? e.touches[0].clientX : e.clientX;
      thumb.style.cursor = 'grabbing';
    };

    const onMove = (e: MouseEvent | TouchEvent) => {
      if (!dragging) return;
      const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX;
      const dx = clientX - startX;
      const maxX = track.offsetWidth - thumb.offsetWidth;
      const x = Math.max(0, Math.min(dx, maxX));
      thumb.style.left = `${x}px`;
    };

    const onEnd = async () => {
      if (!dragging) return;
      dragging = false;
      thumb.style.cursor = 'grab';

      const x = parseInt(thumb.style.left || '0');
      const maxX = track.offsetWidth - thumb.offsetWidth;
      const position = x / maxX;

      try {
        const token = await this.submitChallenge(challenge.challenge_id, { position });
        container.innerHTML = '';
        this.config?.onSuccess?.(token);
        resolve(token);
      } catch (err) {
        thumb.style.left = '0px';
        reject(err as Error);
      }
    };

    thumb.addEventListener('mousedown', onStart);
    thumb.addEventListener('touchstart', onStart);
    document.addEventListener('mousemove', onMove);
    document.addEventListener('touchmove', onMove);
    document.addEventListener('mouseup', onEnd);
    document.addEventListener('touchend', onEnd);
  }

  private renderClickSelect(
    challenge: ChallengeRenderResponse,
    resolve: (token: string) => void,
    reject: (err: Error) => void,
  ): void {
    const targets = challenge.config.targets || [];
    const container = this.containerEl!;

    container.innerHTML = `
      <div class="ec-captcha-modal" style="
        position:fixed;top:0;left:0;width:100%;height:100%;
        background:rgba(0,0,0,0.5);display:flex;align-items:center;justify-content:center;z-index:99999;
      ">
        <div class="ec-captcha-box" style="
          background:#fff;border-radius:12px;padding:24px;width:360px;box-shadow:0 20px 60px rgba(0,0,0,0.2);
        ">
          <div style="font-size:14px;color:#333;margin-bottom:12px;font-weight:600;">
            Click the items in order: ${targets.map(t => t.label).join(' → ')}
          </div>
          <div class="ec-click-area" style="
            position:relative;width:300px;height:200px;background:#f8f8f8;
            border-radius:8px;border:1px solid #eee;margin:0 auto;
          ">
            ${targets.map(t => `
              <div class="ec-target" data-id="${t.id}" style="
                position:absolute;left:${t.x}px;top:${t.y}px;
                width:40px;height:40px;background:#4A90D9;border-radius:50%;
                cursor:pointer;display:flex;align-items:center;justify-content:center;
                color:#fff;font-size:12px;font-weight:bold;
              ">${t.label}</div>
            `).join('')}
          </div>
          <div style="font-size:11px;color:#999;margin-top:8px;text-align:center;">
            Protected by EngageLab CAPTCHA
          </div>
        </div>
      </div>
    `;

    const clicked: string[] = [];
    container.querySelectorAll('.ec-target').forEach((el) => {
      el.addEventListener('click', async () => {
        const id = (el as HTMLElement).dataset.id!;
        clicked.push(id);
        (el as HTMLElement).style.background = '#2ecc71';

        if (clicked.length === targets.length) {
          try {
            const token = await this.submitChallenge(challenge.challenge_id, { clicked_order: clicked });
            container.innerHTML = '';
            this.config?.onSuccess?.(token);
            resolve(token);
          } catch (err) {
            reject(err as Error);
          }
        }
      });
    });
  }

  private async submitChallenge(challengeId: string, answer: object): Promise<string> {
    const resp = await fetch(`${this.apiBase}/v1/challenge/verify`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        challenge_id: challengeId,
        answer,
        behavior_data: this.getBehaviorData(),
      }),
    });

    if (!resp.ok) throw new Error('Challenge verification failed');
    const result: ChallengeVerifyResponse = await resp.json();

    if (!result.success) throw new Error('Challenge failed');
    return result.token;
  }

  // --- Behavior Tracking ---

  private behaviorEvents: Array<{ type: string; ts: number; data?: object }> = [];
  private trackingStarted = false;

  private startBehaviorTracking(): void {
    if (this.trackingStarted) return;
    this.trackingStarted = true;

    const startTs = Date.now();

    document.addEventListener('mousemove', (e) => {
      if (this.behaviorEvents.length < 200) {
        this.behaviorEvents.push({
          type: 'mm',
          ts: Date.now() - startTs,
          data: { x: e.clientX, y: e.clientY },
        });
      }
    }, { passive: true });

    document.addEventListener('click', (e) => {
      this.behaviorEvents.push({
        type: 'click',
        ts: Date.now() - startTs,
        data: { x: e.clientX, y: e.clientY },
      });
    });

    document.addEventListener('keydown', () => {
      this.behaviorEvents.push({ type: 'key', ts: Date.now() - startTs });
    });

    document.addEventListener('scroll', () => {
      if (this.behaviorEvents.length < 200) {
        this.behaviorEvents.push({
          type: 'scroll',
          ts: Date.now() - startTs,
          data: { y: window.scrollY },
        });
      }
    }, { passive: true });
  }

  private getBehaviorData(): object {
    const events = this.behaviorEvents;
    const mouseEvents = events.filter(e => e.type === 'mm');
    const clicks = events.filter(e => e.type === 'click');
    const keys = events.filter(e => e.type === 'key');

    return {
      total_events: events.length,
      mouse_moves: mouseEvents.length,
      clicks: clicks.length,
      keystrokes: keys.length,
      duration_ms: events.length > 0 ? events[events.length - 1].ts : 0,
      mouse_speed_avg: this.computeMouseSpeed(mouseEvents),
      mouse_entropy: this.computeEntropy(mouseEvents),
      has_touch: 'ontouchstart' in window,
      page_visible_time: Math.round(performance.now()),
    };
  }

  private computeMouseSpeed(events: Array<{ data?: object; ts: number }>): number {
    if (events.length < 2) return 0;
    let totalDist = 0;
    for (let i = 1; i < events.length; i++) {
      const prev = events[i - 1].data as { x: number; y: number };
      const curr = events[i].data as { x: number; y: number };
      totalDist += Math.sqrt((curr.x - prev.x) ** 2 + (curr.y - prev.y) ** 2);
    }
    const totalTime = events[events.length - 1].ts - events[0].ts;
    return totalTime > 0 ? totalDist / totalTime : 0;
  }

  private computeEntropy(events: Array<{ data?: object }>): number {
    if (events.length < 3) return 0;
    // Compute direction changes as entropy proxy
    let changes = 0;
    for (let i = 2; i < events.length; i++) {
      const p0 = events[i - 2].data as { x: number; y: number };
      const p1 = events[i - 1].data as { x: number; y: number };
      const p2 = events[i].data as { x: number; y: number };
      const d1x = p1.x - p0.x, d1y = p1.y - p0.y;
      const d2x = p2.x - p1.x, d2y = p2.y - p1.y;
      if (Math.sign(d1x) !== Math.sign(d2x) || Math.sign(d1y) !== Math.sign(d2y)) {
        changes++;
      }
    }
    return changes / (events.length - 2);
  }

  // --- Fingerprint ---

  private async collectFingerprint(): Promise<string> {
    const components = [
      navigator.userAgent,
      navigator.language,
      `${screen.width}x${screen.height}`,
      `${screen.colorDepth}`,
      Intl.DateTimeFormat().resolvedOptions().timeZone,
      navigator.hardwareConcurrency?.toString() || '',
      navigator.platform || '',
      String(navigator.maxTouchPoints || 0),
    ];

    // Canvas fingerprint
    try {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      if (ctx) {
        ctx.textBaseline = 'top';
        ctx.font = '14px Arial';
        ctx.fillText('EngageLab CAPTCHA fp', 2, 2);
        components.push(canvas.toDataURL().slice(-50));
      }
    } catch {}

    // WebGL
    try {
      const canvas = document.createElement('canvas');
      const gl = canvas.getContext('webgl');
      if (gl) {
        const ext = gl.getExtension('WEBGL_debug_renderer_info');
        if (ext) {
          components.push(gl.getParameter(ext.UNMASKED_RENDERER_WEBGL) || '');
        }
      }
    } catch {}

    const str = components.join('|');
    // Simple hash
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const c = str.charCodeAt(i);
      hash = ((hash << 5) - hash + c) | 0;
    }
    return 'fp_' + Math.abs(hash).toString(36);
  }

  private generateId(): string {
    return 'sess_' + Math.random().toString(36).substring(2, 15) + Date.now().toString(36);
  }
}

// Export as global
const EngagelabCaptcha = new EngagelabCaptchaSDK();
if (typeof window !== 'undefined') {
  (window as any).EngagelabCaptcha = EngagelabCaptcha;
}

export default EngagelabCaptcha;
export { EngagelabCaptchaSDK, CaptchaConfig };
