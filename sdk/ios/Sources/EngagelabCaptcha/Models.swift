import Foundation

// MARK: - Enums

/// Action returned by the precheck endpoint indicating how to proceed.
public enum ChallengeAction: String, Codable, Sendable {
    case pass
    case invisible
    case challenge
    case deny
}

/// Predefined scene types that describe where the CAPTCHA is invoked.
public enum SceneType: String, Codable, Sendable {
    case register
    case login
    case activity
    case comment
    case api
}

// MARK: - Request Models

/// Body sent to the precheck endpoint.
public struct PrecheckRequest: Codable, Sendable {
    public let siteKey: String
    public let scene: String
    public let deviceInfo: [String: String]
    public let behaviorData: [String: String]
    public let lang: String?
    public let theme: String?

    enum CodingKeys: String, CodingKey {
        case siteKey = "site_key"
        case scene
        case deviceInfo = "device_info"
        case behaviorData = "behavior_data"
        case lang
        case theme
    }
}

/// Body sent to the challenge verify endpoint.
public struct ChallengeVerifyRequest: Codable, Sendable {
    public let challengeID: String
    public let answer: String
    public let token: String

    enum CodingKeys: String, CodingKey {
        case challengeID = "challenge_id"
        case answer
        case token
    }
}

// MARK: - Response Models

/// Response from the precheck endpoint.
public struct PrecheckResponse: Codable, Sendable {
    public let action: ChallengeAction
    public let riskScore: Double
    public let challengeType: String?
    public let challengeID: String?
    public let token: String

    enum CodingKeys: String, CodingKey {
        case action
        case riskScore = "risk_score"
        case challengeType = "challenge_type"
        case challengeID = "challenge_id"
        case token
    }

    public init(
        action: ChallengeAction,
        riskScore: Double,
        challengeType: String?,
        challengeID: String?,
        token: String
    ) {
        self.action = action
        self.riskScore = riskScore
        self.challengeType = challengeType
        self.challengeID = challengeID
        self.token = token
    }
}

/// Response when requesting challenge render data (HTML/assets URL).
public struct ChallengeRenderResponse: Codable, Sendable {
    public let challengeID: String
    public let challengeType: String
    public let renderURL: String
    public let expiresAt: String

    enum CodingKeys: String, CodingKey {
        case challengeID = "challenge_id"
        case challengeType = "challenge_type"
        case renderURL = "render_url"
        case expiresAt = "expires_at"
    }

    public init(
        challengeID: String,
        challengeType: String,
        renderURL: String,
        expiresAt: String
    ) {
        self.challengeID = challengeID
        self.challengeType = challengeType
        self.renderURL = renderURL
        self.expiresAt = expiresAt
    }
}

/// Response after the user submits a challenge answer.
public struct ChallengeVerifyResponse: Codable, Sendable {
    public let success: Bool
    public let token: String?
    public let errorMessage: String?
    public let retryAllowed: Bool

    enum CodingKeys: String, CodingKey {
        case success
        case token
        case errorMessage = "error_message"
        case retryAllowed = "retry_allowed"
    }

    public init(
        success: Bool,
        token: String?,
        errorMessage: String?,
        retryAllowed: Bool
    ) {
        self.success = success
        self.token = token
        self.errorMessage = errorMessage
        self.retryAllowed = retryAllowed
    }
}

/// Response from the server-side site-verify endpoint.
public struct SiteVerifyResponse: Codable, Sendable {
    public let success: Bool
    public let challengeTimestamp: String?
    public let hostname: String?
    public let score: Double?
    public let action: String?
    public let errorCodes: [String]?

    enum CodingKeys: String, CodingKey {
        case success
        case challengeTimestamp = "challenge_ts"
        case hostname
        case score
        case action
        case errorCodes = "error_codes"
    }

    public init(
        success: Bool,
        challengeTimestamp: String?,
        hostname: String?,
        score: Double?,
        action: String?,
        errorCodes: [String]?
    ) {
        self.success = success
        self.challengeTimestamp = challengeTimestamp
        self.hostname = hostname
        self.score = score
        self.action = action
        self.errorCodes = errorCodes
    }
}

// MARK: - SDK Result

/// The result of executing the CAPTCHA flow.
public enum CaptchaResult: Sendable {
    /// The user passed verification; includes a verification token.
    case passed(token: String)
    /// A challenge is required; the caller should present the challenge UI.
    case challengeRequired(challengeID: String, type: String)
    /// The request was denied (high risk score).
    case denied
    /// An error occurred during the flow.
    case error(Error)

    /// Convenience check for the passed case.
    public var token: String? {
        if case .passed(let token) = self { return token }
        return nil
    }

    /// Convenience check for challenge required case.
    public var challengeID: String? {
        if case .challengeRequired(let id, _) = self { return id }
        return nil
    }
}
