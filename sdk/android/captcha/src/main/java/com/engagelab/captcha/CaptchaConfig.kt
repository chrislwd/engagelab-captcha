package com.engagelab.captcha

import kotlinx.serialization.Serializable

/**
 * Configuration for the EngageLab CAPTCHA SDK.
 *
 * @param siteKey The site key issued by the EngageLab dashboard.
 * @param scene The scene in which the CAPTCHA is being triggered.
 * @param apiBase Base URL for the CAPTCHA API.
 * @param lang Language code (e.g. "en", "zh-CN"). Null uses the device language.
 * @param theme Visual theme for challenge UI.
 */
data class CaptchaConfig(
    val siteKey: String,
    val scene: SceneType = SceneType.LOGIN,
    val apiBase: String = "https://captcha.engagelab.com/api/v1",
    val lang: String? = null,
    val theme: CaptchaTheme = CaptchaTheme.LIGHT
)

/**
 * Predefined scene types that describe where the CAPTCHA is invoked.
 */
@Serializable
enum class SceneType(val value: String) {
    REGISTER("register"),
    LOGIN("login"),
    ACTIVITY("activity"),
    COMMENT("comment"),
    API("api");

    companion object {
        fun fromValue(value: String): SceneType? =
            entries.firstOrNull { it.value == value }
    }
}

/**
 * Visual theme for the challenge rendering.
 */
@Serializable
enum class CaptchaTheme(val value: String) {
    LIGHT("light"),
    DARK("dark"),
    AUTO("auto");

    companion object {
        fun fromValue(value: String): CaptchaTheme? =
            entries.firstOrNull { it.value == value }
    }
}
