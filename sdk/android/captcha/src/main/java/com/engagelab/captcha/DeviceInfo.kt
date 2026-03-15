package com.engagelab.captcha

import android.content.Context
import android.content.res.Resources
import android.os.Build
import android.util.DisplayMetrics
import java.security.MessageDigest
import java.util.Locale
import java.util.TimeZone

/**
 * Collects device-level signals for risk analysis.
 */
object DeviceInfoCollector {

    /**
     * Gather all device info into a string-keyed map.
     *
     * @param context Android context, used to read display metrics and configuration.
     */
    fun collect(context: Context): Map<String, String> {
        val info = mutableMapOf<String, String>()

        // Hardware
        info["model"] = Build.MODEL
        info["manufacturer"] = Build.MANUFACTURER
        info["brand"] = Build.BRAND
        info["device"] = Build.DEVICE
        info["hardware"] = Build.HARDWARE
        info["product"] = Build.PRODUCT

        // OS
        info["os_version"] = Build.VERSION.RELEASE
        info["sdk_int"] = Build.VERSION.SDK_INT.toString()

        // Screen
        val dm: DisplayMetrics = Resources.getSystem().displayMetrics
        info["screen_width"] = dm.widthPixels.toString()
        info["screen_height"] = dm.heightPixels.toString()
        info["screen_density"] = dm.density.toString()
        info["screen_dpi"] = dm.densityDpi.toString()

        // Locale / Language
        val locale = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
            context.resources.configuration.locales[0]
        } else {
            @Suppress("DEPRECATION")
            context.resources.configuration.locale
        }
        info["lang"] = locale.language
        info["lang_tag"] = locale.toLanguageTag()
        info["country"] = locale.country

        // Timezone
        val tz = TimeZone.getDefault()
        info["timezone"] = tz.id
        info["timezone_offset"] = (tz.rawOffset / 1000).toString()

        // User agent
        info["user_agent"] = buildUserAgent()

        // Emulator detection
        info["is_emulator"] = if (isEmulator()) "true" else "false"

        // Fingerprint
        info["fingerprint"] = generateFingerprint(info)

        return info
    }

    /**
     * Build a user-agent string describing this SDK and the device.
     */
    private fun buildUserAgent(): String {
        val sdkVersion = "1.0.0"
        return "EngagelabCaptcha-Android/$sdkVersion (${Build.MODEL}; Android ${Build.VERSION.RELEASE})"
    }

    /**
     * Detect common emulator signatures.
     */
    fun isEmulator(): Boolean {
        return (Build.FINGERPRINT.startsWith("generic")
                || Build.FINGERPRINT.startsWith("unknown")
                || Build.MODEL.contains("google_sdk")
                || Build.MODEL.contains("Emulator")
                || Build.MODEL.contains("Android SDK built for x86")
                || Build.MANUFACTURER.contains("Genymotion")
                || (Build.BRAND.startsWith("generic") && Build.DEVICE.startsWith("generic"))
                || Build.PRODUCT == "google_sdk"
                || Build.PRODUCT == "sdk"
                || Build.PRODUCT == "sdk_x86"
                || Build.PRODUCT == "sdk_gphone_x86"
                || Build.PRODUCT == "sdk_gphone64_arm64"
                || Build.PRODUCT == "vbox86p"
                || Build.HARDWARE.contains("goldfish")
                || Build.HARDWARE.contains("ranchu"))
    }

    /**
     * Generate a SHA-256 fingerprint hash from the collected info values.
     */
    fun generateFingerprint(info: Map<String, String>): String {
        val sortedKeys = info.keys.sorted()
        val combined = sortedKeys.joinToString("&") { key -> "$key=${info[key] ?: ""}" }
        return sha256(combined)
    }

    private fun sha256(input: String): String {
        val bytes = MessageDigest.getInstance("SHA-256").digest(input.toByteArray(Charsets.UTF_8))
        return bytes.joinToString("") { "%02x".format(it) }
    }
}
