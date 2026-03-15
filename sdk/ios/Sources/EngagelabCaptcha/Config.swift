import Foundation

/// Configuration for the EngageLab CAPTCHA SDK.
public struct CaptchaConfig {
    /// The site key issued by the EngageLab dashboard.
    public let siteKey: String

    /// The scene in which the CAPTCHA is being triggered.
    public let scene: SceneType

    /// Base URL for the CAPTCHA API.
    public let apiBase: String

    /// Language code (e.g. "en", "zh-CN"). Nil uses the device language.
    public let lang: String?

    /// Visual theme for challenge UI.
    public let theme: CaptchaTheme

    public init(
        siteKey: String,
        scene: SceneType = .login,
        apiBase: String = "https://captcha.engagelab.com/api/v1",
        lang: String? = nil,
        theme: CaptchaTheme = .light
    ) {
        self.siteKey = siteKey
        self.scene = scene
        self.apiBase = apiBase
        self.lang = lang
        self.theme = theme
    }
}

/// Visual theme for the challenge rendering.
public enum CaptchaTheme: String, Codable, Sendable {
    case light
    case dark
    case auto
}
