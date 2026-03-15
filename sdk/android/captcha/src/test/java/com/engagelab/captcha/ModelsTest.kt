package com.engagelab.captcha

import kotlinx.serialization.json.Json
import org.junit.Assert.*
import org.junit.Test

class ModelsTest {

    private val json = Json {
        ignoreUnknownKeys = true
        encodeDefaults = true
    }

    // ----------------------------------------------------------------
    // Config Tests
    // ----------------------------------------------------------------

    @Test
    fun `config has correct defaults`() {
        val config = CaptchaConfig(siteKey = "test-key-123")
        assertEquals("test-key-123", config.siteKey)
        assertEquals(SceneType.LOGIN, config.scene)
        assertEquals("https://captcha.engagelab.com/api/v1", config.apiBase)
        assertNull(config.lang)
        assertEquals(CaptchaTheme.LIGHT, config.theme)
    }

    @Test
    fun `config with custom values`() {
        val config = CaptchaConfig(
            siteKey = "custom-key",
            scene = SceneType.REGISTER,
            apiBase = "https://custom.api.com/v2",
            lang = "zh-CN",
            theme = CaptchaTheme.DARK
        )
        assertEquals("custom-key", config.siteKey)
        assertEquals(SceneType.REGISTER, config.scene)
        assertEquals("https://custom.api.com/v2", config.apiBase)
        assertEquals("zh-CN", config.lang)
        assertEquals(CaptchaTheme.DARK, config.theme)
    }

    // ----------------------------------------------------------------
    // Scene Type Tests
    // ----------------------------------------------------------------

    @Test
    fun `all scene types have correct values`() {
        assertEquals("register", SceneType.REGISTER.value)
        assertEquals("login", SceneType.LOGIN.value)
        assertEquals("activity", SceneType.ACTIVITY.value)
        assertEquals("comment", SceneType.COMMENT.value)
        assertEquals("api", SceneType.API.value)
    }

    @Test
    fun `scene type fromValue works`() {
        assertEquals(SceneType.REGISTER, SceneType.fromValue("register"))
        assertEquals(SceneType.LOGIN, SceneType.fromValue("login"))
        assertEquals(SceneType.ACTIVITY, SceneType.fromValue("activity"))
        assertEquals(SceneType.COMMENT, SceneType.fromValue("comment"))
        assertEquals(SceneType.API, SceneType.fromValue("api"))
        assertNull(SceneType.fromValue("invalid"))
    }

    // ----------------------------------------------------------------
    // Challenge Action Tests
    // ----------------------------------------------------------------

    @Test
    fun `all challenge actions have correct values`() {
        assertEquals("pass", ChallengeAction.PASS.value)
        assertEquals("invisible", ChallengeAction.INVISIBLE.value)
        assertEquals("challenge", ChallengeAction.CHALLENGE.value)
        assertEquals("deny", ChallengeAction.DENY.value)
    }

    @Test
    fun `challenge action fromValue works`() {
        assertEquals(ChallengeAction.PASS, ChallengeAction.fromValue("pass"))
        assertEquals(ChallengeAction.INVISIBLE, ChallengeAction.fromValue("invisible"))
        assertEquals(ChallengeAction.CHALLENGE, ChallengeAction.fromValue("challenge"))
        assertEquals(ChallengeAction.DENY, ChallengeAction.fromValue("deny"))
        assertNull(ChallengeAction.fromValue("unknown"))
    }

    // ----------------------------------------------------------------
    // Theme Tests
    // ----------------------------------------------------------------

    @Test
    fun `theme values are correct`() {
        assertEquals("light", CaptchaTheme.LIGHT.value)
        assertEquals("dark", CaptchaTheme.DARK.value)
        assertEquals("auto", CaptchaTheme.AUTO.value)
    }

    @Test
    fun `theme fromValue works`() {
        assertEquals(CaptchaTheme.LIGHT, CaptchaTheme.fromValue("light"))
        assertEquals(CaptchaTheme.DARK, CaptchaTheme.fromValue("dark"))
        assertEquals(CaptchaTheme.AUTO, CaptchaTheme.fromValue("auto"))
        assertNull(CaptchaTheme.fromValue("invalid"))
    }

    // ----------------------------------------------------------------
    // Precheck Response Serialization
    // ----------------------------------------------------------------

    @Test
    fun `precheck response pass deserialization`() {
        val jsonStr = """
        {
            "action": "pass",
            "risk_score": 0.1,
            "challenge_type": null,
            "challenge_id": null,
            "token": "abc-token-123"
        }
        """.trimIndent()

        val response = json.decodeFromString<PrecheckResponse>(jsonStr)
        assertEquals(ChallengeAction.PASS, response.action)
        assertEquals(0.1, response.riskScore, 0.001)
        assertNull(response.challengeType)
        assertNull(response.challengeId)
        assertEquals("abc-token-123", response.token)
    }

    @Test
    fun `precheck response challenge deserialization`() {
        val jsonStr = """
        {
            "action": "challenge",
            "risk_score": 0.65,
            "challenge_type": "image_select",
            "challenge_id": "ch-456",
            "token": "session-token"
        }
        """.trimIndent()

        val response = json.decodeFromString<PrecheckResponse>(jsonStr)
        assertEquals(ChallengeAction.CHALLENGE, response.action)
        assertEquals(0.65, response.riskScore, 0.001)
        assertEquals("image_select", response.challengeType)
        assertEquals("ch-456", response.challengeId)
        assertEquals("session-token", response.token)
    }

    @Test
    fun `precheck response deny deserialization`() {
        val jsonStr = """
        {
            "action": "deny",
            "risk_score": 0.95,
            "token": "deny-token"
        }
        """.trimIndent()

        val response = json.decodeFromString<PrecheckResponse>(jsonStr)
        assertEquals(ChallengeAction.DENY, response.action)
        assertEquals(0.95, response.riskScore, 0.001)
        assertNull(response.challengeType)
        assertNull(response.challengeId)
    }

    // ----------------------------------------------------------------
    // Challenge Render Response Serialization
    // ----------------------------------------------------------------

    @Test
    fun `challenge render response deserialization`() {
        val jsonStr = """
        {
            "challenge_id": "ch-789",
            "challenge_type": "slider",
            "render_url": "https://captcha.engagelab.com/render/ch-789",
            "expires_at": "2026-01-01T00:00:00Z"
        }
        """.trimIndent()

        val response = json.decodeFromString<ChallengeRenderResponse>(jsonStr)
        assertEquals("ch-789", response.challengeId)
        assertEquals("slider", response.challengeType)
        assertEquals("https://captcha.engagelab.com/render/ch-789", response.renderUrl)
        assertEquals("2026-01-01T00:00:00Z", response.expiresAt)
    }

    // ----------------------------------------------------------------
    // Challenge Verify Response Serialization
    // ----------------------------------------------------------------

    @Test
    fun `challenge verify success deserialization`() {
        val jsonStr = """
        {
            "success": true,
            "token": "verified-token-abc",
            "error_message": null,
            "retry_allowed": false
        }
        """.trimIndent()

        val response = json.decodeFromString<ChallengeVerifyResponse>(jsonStr)
        assertTrue(response.success)
        assertEquals("verified-token-abc", response.token)
        assertNull(response.errorMessage)
        assertFalse(response.retryAllowed)
    }

    @Test
    fun `challenge verify failure deserialization`() {
        val jsonStr = """
        {
            "success": false,
            "token": null,
            "error_message": "Incorrect answer",
            "retry_allowed": true
        }
        """.trimIndent()

        val response = json.decodeFromString<ChallengeVerifyResponse>(jsonStr)
        assertFalse(response.success)
        assertNull(response.token)
        assertEquals("Incorrect answer", response.errorMessage)
        assertTrue(response.retryAllowed)
    }

    // ----------------------------------------------------------------
    // Site Verify Response Serialization
    // ----------------------------------------------------------------

    @Test
    fun `site verify response deserialization`() {
        val jsonStr = """
        {
            "success": true,
            "challenge_ts": "2026-01-15T12:00:00Z",
            "hostname": "example.com",
            "score": 0.9,
            "action": "login",
            "error_codes": []
        }
        """.trimIndent()

        val response = json.decodeFromString<SiteVerifyResponse>(jsonStr)
        assertTrue(response.success)
        assertEquals("2026-01-15T12:00:00Z", response.challengeTimestamp)
        assertEquals("example.com", response.hostname)
        assertEquals(0.9, response.score!!, 0.001)
        assertEquals("login", response.action)
        assertTrue(response.errorCodes!!.isEmpty())
    }

    @Test
    fun `site verify response failure deserialization`() {
        val jsonStr = """
        {
            "success": false,
            "error_codes": ["missing-input-response", "timeout-or-duplicate"]
        }
        """.trimIndent()

        val response = json.decodeFromString<SiteVerifyResponse>(jsonStr)
        assertFalse(response.success)
        assertNull(response.challengeTimestamp)
        assertNull(response.hostname)
        assertNull(response.score)
        assertEquals(2, response.errorCodes!!.size)
        assertEquals("missing-input-response", response.errorCodes!![0])
    }

    // ----------------------------------------------------------------
    // Precheck Request Serialization
    // ----------------------------------------------------------------

    @Test
    fun `precheck request serialization`() {
        val request = PrecheckRequest(
            siteKey = "key-123",
            scene = "login",
            deviceInfo = mapOf("model" to "Pixel 8"),
            behaviorData = mapOf("touch_count" to "5"),
            lang = "en",
            theme = "light"
        )
        val jsonStr = json.encodeToString(PrecheckRequest.serializer(), request)
        assertTrue(jsonStr.contains("\"site_key\":\"key-123\""))
        assertTrue(jsonStr.contains("\"scene\":\"login\""))
        assertTrue(jsonStr.contains("\"device_info\""))
        assertTrue(jsonStr.contains("\"behavior_data\""))
    }

    @Test
    fun `challenge verify request serialization`() {
        val request = ChallengeVerifyRequest(
            challengeId = "ch-100",
            answer = "42",
            token = "tok-abc"
        )
        val jsonStr = json.encodeToString(ChallengeVerifyRequest.serializer(), request)
        assertTrue(jsonStr.contains("\"challenge_id\":\"ch-100\""))
        assertTrue(jsonStr.contains("\"answer\":\"42\""))
        assertTrue(jsonStr.contains("\"token\":\"tok-abc\""))
    }

    // ----------------------------------------------------------------
    // CaptchaResult Tests
    // ----------------------------------------------------------------

    @Test
    fun `captcha result passed`() {
        val result = CaptchaResult.Passed("my-token")
        assertTrue(result is CaptchaResult.Passed)
        assertEquals("my-token", (result as CaptchaResult.Passed).token)
    }

    @Test
    fun `captcha result challenge required`() {
        val result = CaptchaResult.ChallengeRequired("ch-1", "slider")
        assertTrue(result is CaptchaResult.ChallengeRequired)
        assertEquals("ch-1", (result as CaptchaResult.ChallengeRequired).challengeId)
        assertEquals("slider", result.type)
    }

    @Test
    fun `captcha result denied`() {
        val result = CaptchaResult.Denied
        assertTrue(result is CaptchaResult.Denied)
    }

    @Test
    fun `captcha result error`() {
        val exception = RuntimeException("test error")
        val result = CaptchaResult.Error(exception)
        assertTrue(result is CaptchaResult.Error)
        assertEquals("test error", (result as CaptchaResult.Error).exception.message)
    }

    // ----------------------------------------------------------------
    // Fingerprint Tests
    // ----------------------------------------------------------------

    @Test
    fun `fingerprint is deterministic`() {
        val info = mapOf("model" to "Pixel", "os" to "Android 14")
        val fp1 = DeviceInfoCollector.generateFingerprint(info)
        val fp2 = DeviceInfoCollector.generateFingerprint(info)
        assertEquals(fp1, fp2)
        assertEquals(64, fp1.length) // SHA-256 hex is 64 chars
    }

    @Test
    fun `fingerprint changes with different input`() {
        val info1 = mapOf("model" to "Pixel")
        val info2 = mapOf("model" to "Samsung")
        val fp1 = DeviceInfoCollector.generateFingerprint(info1)
        val fp2 = DeviceInfoCollector.generateFingerprint(info2)
        assertNotEquals(fp1, fp2)
    }

    // ----------------------------------------------------------------
    // Behavior Tracker Tests
    // ----------------------------------------------------------------

    @Test
    fun `behavior tracker empty session`() {
        val tracker = BehaviorTracker()
        val data = tracker.behaviorData()
        assertEquals("0", data["touch_count"])
        assertEquals("0.0000", data["touch_entropy"])
        assertNotNull(data["session_duration_ms"])
    }

    @Test
    fun `behavior tracker with touches`() {
        val tracker = BehaviorTracker()
        tracker.recordTouch(100f, 200f, "down")
        tracker.recordTouch(110f, 210f, "move")
        tracker.recordTouch(120f, 220f, "move")
        tracker.recordTouch(120f, 220f, "up")

        val data = tracker.behaviorData()
        assertEquals("4", data["touch_count"])
        assertEquals("1", data["touch_down_count"])
        assertEquals("2", data["touch_move_count"])
        assertEquals("1", data["touch_up_count"])
        assertNotNull(data["touch_samples"])
    }

    @Test
    fun `behavior tracker reset`() {
        val tracker = BehaviorTracker()
        tracker.recordTouch(50f, 50f, "down")
        tracker.reset()
        val data = tracker.behaviorData()
        assertEquals("0", data["touch_count"])
    }
}
