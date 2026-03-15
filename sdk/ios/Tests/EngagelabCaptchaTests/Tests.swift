import XCTest
@testable import EngagelabCaptcha

final class EngagelabCaptchaTests: XCTestCase {

    // MARK: - Config Tests

    func testConfigDefaults() {
        let config = CaptchaConfig(siteKey: "test-key-123")
        XCTAssertEqual(config.siteKey, "test-key-123")
        XCTAssertEqual(config.scene, .login)
        XCTAssertEqual(config.apiBase, "https://captcha.engagelab.com/api/v1")
        XCTAssertNil(config.lang)
        XCTAssertEqual(config.theme, .light)
    }

    func testConfigCustomValues() {
        let config = CaptchaConfig(
            siteKey: "custom-key",
            scene: .register,
            apiBase: "https://custom.api.com/v2",
            lang: "zh-CN",
            theme: .dark
        )
        XCTAssertEqual(config.siteKey, "custom-key")
        XCTAssertEqual(config.scene, .register)
        XCTAssertEqual(config.apiBase, "https://custom.api.com/v2")
        XCTAssertEqual(config.lang, "zh-CN")
        XCTAssertEqual(config.theme, .dark)
    }

    // MARK: - Scene Type Tests

    func testAllSceneTypes() {
        let scenes: [SceneType] = [.register, .login, .activity, .comment, .api]
        let rawValues = ["register", "login", "activity", "comment", "api"]

        for (scene, raw) in zip(scenes, rawValues) {
            XCTAssertEqual(scene.rawValue, raw)
            XCTAssertEqual(SceneType(rawValue: raw), scene)
        }
    }

    func testInvalidSceneType() {
        XCTAssertNil(SceneType(rawValue: "invalid"))
    }

    // MARK: - ChallengeAction Tests

    func testAllChallengeActions() {
        let actions: [ChallengeAction] = [.pass, .invisible, .challenge, .deny]
        let rawValues = ["pass", "invisible", "challenge", "deny"]

        for (action, raw) in zip(actions, rawValues) {
            XCTAssertEqual(action.rawValue, raw)
            XCTAssertEqual(ChallengeAction(rawValue: raw), action)
        }
    }

    // MARK: - Theme Tests

    func testThemeRawValues() {
        XCTAssertEqual(CaptchaTheme.light.rawValue, "light")
        XCTAssertEqual(CaptchaTheme.dark.rawValue, "dark")
        XCTAssertEqual(CaptchaTheme.auto.rawValue, "auto")
    }

    // MARK: - Model Tests

    func testPrecheckResponseDecoding() throws {
        let json = """
        {
            "action": "pass",
            "risk_score": 0.1,
            "challenge_type": null,
            "challenge_id": null,
            "token": "abc-token-123"
        }
        """.data(using: .utf8)!

        let response = try JSONDecoder().decode(PrecheckResponse.self, from: json)
        XCTAssertEqual(response.action, .pass)
        XCTAssertEqual(response.riskScore, 0.1, accuracy: 0.001)
        XCTAssertNil(response.challengeType)
        XCTAssertNil(response.challengeID)
        XCTAssertEqual(response.token, "abc-token-123")
    }

    func testPrecheckResponseChallengeDecoding() throws {
        let json = """
        {
            "action": "challenge",
            "risk_score": 0.65,
            "challenge_type": "image_select",
            "challenge_id": "ch-456",
            "token": "session-token"
        }
        """.data(using: .utf8)!

        let response = try JSONDecoder().decode(PrecheckResponse.self, from: json)
        XCTAssertEqual(response.action, .challenge)
        XCTAssertEqual(response.riskScore, 0.65, accuracy: 0.001)
        XCTAssertEqual(response.challengeType, "image_select")
        XCTAssertEqual(response.challengeID, "ch-456")
    }

    func testChallengeRenderResponseDecoding() throws {
        let json = """
        {
            "challenge_id": "ch-789",
            "challenge_type": "slider",
            "render_url": "https://captcha.engagelab.com/render/ch-789",
            "expires_at": "2026-01-01T00:00:00Z"
        }
        """.data(using: .utf8)!

        let response = try JSONDecoder().decode(ChallengeRenderResponse.self, from: json)
        XCTAssertEqual(response.challengeID, "ch-789")
        XCTAssertEqual(response.challengeType, "slider")
        XCTAssertEqual(response.renderURL, "https://captcha.engagelab.com/render/ch-789")
        XCTAssertEqual(response.expiresAt, "2026-01-01T00:00:00Z")
    }

    func testChallengeVerifyResponseDecoding() throws {
        let json = """
        {
            "success": true,
            "token": "verified-token-abc",
            "error_message": null,
            "retry_allowed": false
        }
        """.data(using: .utf8)!

        let response = try JSONDecoder().decode(ChallengeVerifyResponse.self, from: json)
        XCTAssertTrue(response.success)
        XCTAssertEqual(response.token, "verified-token-abc")
        XCTAssertNil(response.errorMessage)
        XCTAssertFalse(response.retryAllowed)
    }

    func testChallengeVerifyFailureDecoding() throws {
        let json = """
        {
            "success": false,
            "token": null,
            "error_message": "Incorrect answer",
            "retry_allowed": true
        }
        """.data(using: .utf8)!

        let response = try JSONDecoder().decode(ChallengeVerifyResponse.self, from: json)
        XCTAssertFalse(response.success)
        XCTAssertNil(response.token)
        XCTAssertEqual(response.errorMessage, "Incorrect answer")
        XCTAssertTrue(response.retryAllowed)
    }

    func testSiteVerifyResponseDecoding() throws {
        let json = """
        {
            "success": true,
            "challenge_ts": "2026-01-15T12:00:00Z",
            "hostname": "example.com",
            "score": 0.9,
            "action": "login",
            "error_codes": []
        }
        """.data(using: .utf8)!

        let response = try JSONDecoder().decode(SiteVerifyResponse.self, from: json)
        XCTAssertTrue(response.success)
        XCTAssertEqual(response.challengeTimestamp, "2026-01-15T12:00:00Z")
        XCTAssertEqual(response.hostname, "example.com")
        XCTAssertEqual(response.score, 0.9, accuracy: 0.001)
        XCTAssertEqual(response.action, "login")
        XCTAssertEqual(response.errorCodes, [])
    }

    // MARK: - CaptchaResult Tests

    func testCaptchaResultPassedToken() {
        let result = CaptchaResult.passed(token: "my-token")
        XCTAssertEqual(result.token, "my-token")
        XCTAssertNil(result.challengeID)
    }

    func testCaptchaResultChallengeRequired() {
        let result = CaptchaResult.challengeRequired(challengeID: "ch-1", type: "slider")
        XCTAssertNil(result.token)
        XCTAssertEqual(result.challengeID, "ch-1")
    }

    func testCaptchaResultDenied() {
        let result = CaptchaResult.denied
        XCTAssertNil(result.token)
        XCTAssertNil(result.challengeID)
    }

    // MARK: - Device Info Tests

    func testDeviceInfoCollect() {
        let info = DeviceInfoCollector.shared.collect()
        XCTAssertFalse(info.isEmpty)
        XCTAssertNotNil(info["lang"])
        XCTAssertNotNil(info["timezone"])
        XCTAssertNotNil(info["timezone_offset"])
        XCTAssertNotNil(info["user_agent"])
        XCTAssertNotNil(info["fingerprint"])
    }

    func testFingerprintDeterministic() {
        let info: [String: String] = [
            "model": "iPhone",
            "os": "iOS 17",
            "lang": "en"
        ]
        let fp1 = DeviceInfoCollector.shared.generateFingerprint(from: info)
        let fp2 = DeviceInfoCollector.shared.generateFingerprint(from: info)
        XCTAssertEqual(fp1, fp2)
        XCTAssertEqual(fp1.count, 64) // SHA-256 hex is 64 chars
    }

    func testFingerprintChangesWithInput() {
        let info1 = ["model": "iPhone"]
        let info2 = ["model": "iPad"]
        let fp1 = DeviceInfoCollector.shared.generateFingerprint(from: info1)
        let fp2 = DeviceInfoCollector.shared.generateFingerprint(from: info2)
        XCTAssertNotEqual(fp1, fp2)
    }

    // MARK: - Behavior Tracker Tests

    func testBehaviorTrackerEmptySession() {
        let tracker = BehaviorTracker()
        let data = tracker.behaviorData()
        XCTAssertEqual(data["touch_count"], "0")
        XCTAssertEqual(data["touch_entropy"], "0.0000")
        XCTAssertNotNil(data["session_duration_ms"])
    }

    func testBehaviorTrackerWithTouches() {
        let tracker = BehaviorTracker()
        tracker.recordTouch(x: 100, y: 200, phase: "began")
        tracker.recordTouch(x: 110, y: 210, phase: "moved")
        tracker.recordTouch(x: 120, y: 220, phase: "moved")
        tracker.recordTouch(x: 120, y: 220, phase: "ended")

        let data = tracker.behaviorData()
        XCTAssertEqual(data["touch_count"], "4")
        XCTAssertEqual(data["touch_began_count"], "1")
        XCTAssertEqual(data["touch_moved_count"], "2")
        XCTAssertEqual(data["touch_ended_count"], "1")
        XCTAssertNotNil(data["touch_samples"])
    }

    func testBehaviorTrackerReset() {
        let tracker = BehaviorTracker()
        tracker.recordTouch(x: 50, y: 50, phase: "began")
        tracker.reset()
        let data = tracker.behaviorData()
        XCTAssertEqual(data["touch_count"], "0")
    }

    // MARK: - PrecheckRequest Encoding

    func testPrecheckRequestEncoding() throws {
        let request = PrecheckRequest(
            siteKey: "key-123",
            scene: "login",
            deviceInfo: ["model": "iPhone"],
            behaviorData: ["touch_count": "5"],
            lang: "en",
            theme: "light"
        )
        let data = try JSONEncoder().encode(request)
        let dict = try JSONSerialization.jsonObject(with: data) as? [String: Any]
        XCTAssertEqual(dict?["site_key"] as? String, "key-123")
        XCTAssertEqual(dict?["scene"] as? String, "login")
        XCTAssertNotNil(dict?["device_info"])
        XCTAssertNotNil(dict?["behavior_data"])
    }
}
