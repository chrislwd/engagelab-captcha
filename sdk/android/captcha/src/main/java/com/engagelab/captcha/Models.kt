package com.engagelab.captcha

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

// ============================================================
// Enums
// ============================================================

/**
 * Action returned by the precheck endpoint indicating how to proceed.
 */
@Serializable
enum class ChallengeAction(val value: String) {
    @SerialName("pass") PASS("pass"),
    @SerialName("invisible") INVISIBLE("invisible"),
    @SerialName("challenge") CHALLENGE("challenge"),
    @SerialName("deny") DENY("deny");

    companion object {
        fun fromValue(value: String): ChallengeAction? =
            entries.firstOrNull { it.value == value }
    }
}

// ============================================================
// Request Models
// ============================================================

/**
 * Body sent to the precheck endpoint.
 */
@Serializable
data class PrecheckRequest(
    @SerialName("site_key") val siteKey: String,
    @SerialName("scene") val scene: String,
    @SerialName("device_info") val deviceInfo: Map<String, String>,
    @SerialName("behavior_data") val behaviorData: Map<String, String>,
    @SerialName("lang") val lang: String? = null,
    @SerialName("theme") val theme: String? = null
)

/**
 * Body sent to the challenge verify endpoint.
 */
@Serializable
data class ChallengeVerifyRequest(
    @SerialName("challenge_id") val challengeId: String,
    @SerialName("answer") val answer: String,
    @SerialName("token") val token: String
)

// ============================================================
// Response Models
// ============================================================

/**
 * Response from the precheck endpoint.
 */
@Serializable
data class PrecheckResponse(
    @SerialName("action") val action: ChallengeAction,
    @SerialName("risk_score") val riskScore: Double,
    @SerialName("challenge_type") val challengeType: String? = null,
    @SerialName("challenge_id") val challengeId: String? = null,
    @SerialName("token") val token: String
)

/**
 * Response when requesting challenge render data (HTML/assets URL).
 */
@Serializable
data class ChallengeRenderResponse(
    @SerialName("challenge_id") val challengeId: String,
    @SerialName("challenge_type") val challengeType: String,
    @SerialName("render_url") val renderUrl: String,
    @SerialName("expires_at") val expiresAt: String
)

/**
 * Response after the user submits a challenge answer.
 */
@Serializable
data class ChallengeVerifyResponse(
    @SerialName("success") val success: Boolean,
    @SerialName("token") val token: String? = null,
    @SerialName("error_message") val errorMessage: String? = null,
    @SerialName("retry_allowed") val retryAllowed: Boolean = false
)

/**
 * Response from the server-side site-verify endpoint.
 */
@Serializable
data class SiteVerifyResponse(
    @SerialName("success") val success: Boolean,
    @SerialName("challenge_ts") val challengeTimestamp: String? = null,
    @SerialName("hostname") val hostname: String? = null,
    @SerialName("score") val score: Double? = null,
    @SerialName("action") val action: String? = null,
    @SerialName("error_codes") val errorCodes: List<String>? = null
)

// ============================================================
// SDK Result
// ============================================================

/**
 * The result of executing the CAPTCHA flow.
 */
sealed class CaptchaResult {
    /**
     * The user passed verification; includes a verification token.
     */
    data class Passed(val token: String) : CaptchaResult()

    /**
     * A challenge is required; the caller should present the challenge UI.
     */
    data class ChallengeRequired(
        val challengeId: String,
        val type: String
    ) : CaptchaResult()

    /**
     * The request was denied (high risk score).
     */
    data object Denied : CaptchaResult()

    /**
     * An error occurred during the flow.
     */
    data class Error(val exception: Throwable) : CaptchaResult()
}
