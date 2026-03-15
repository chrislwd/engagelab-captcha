package com.engagelab.captcha

import android.annotation.SuppressLint
import android.app.Activity
import android.content.Context
import android.net.Uri
import android.os.Bundle
import android.view.ViewGroup
import android.webkit.JavascriptInterface
import android.webkit.WebChromeClient
import android.webkit.WebView
import android.webkit.WebViewClient
import android.widget.FrameLayout
import androidx.appcompat.app.AppCompatActivity
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.suspendCancellableCoroutine
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.io.IOException
import java.util.concurrent.TimeUnit
import kotlin.coroutines.resume

/**
 * Main entry point for the EngageLab CAPTCHA SDK.
 *
 * Usage:
 * ```kotlin
 * val config = CaptchaConfig(siteKey = "your-site-key", scene = SceneType.LOGIN)
 * val captcha = EngagelabCaptcha(config)
 * val result = captcha.execute(context)
 * when (result) {
 *     is CaptchaResult.Passed -> // send result.token to your backend
 *     is CaptchaResult.ChallengeRequired -> // present challenge UI
 *     is CaptchaResult.Denied -> // block the action
 *     is CaptchaResult.Error -> // handle result.exception
 * }
 * ```
 */
class EngagelabCaptcha(
    val config: CaptchaConfig
) {

    // ----------------------------------------------------------------
    // Errors
    // ----------------------------------------------------------------

    class CaptchaException(message: String, cause: Throwable? = null) : Exception(message, cause)

    // ----------------------------------------------------------------
    // Properties
    // ----------------------------------------------------------------

    val behaviorTracker = BehaviorTracker()

    private val json = Json {
        ignoreUnknownKeys = true
        encodeDefaults = true
    }

    private val httpClient = OkHttpClient.Builder()
        .connectTimeout(15, TimeUnit.SECONDS)
        .readTimeout(15, TimeUnit.SECONDS)
        .writeTimeout(15, TimeUnit.SECONDS)
        .build()

    private val jsonMediaType = "application/json; charset=utf-8".toMediaType()

    // ----------------------------------------------------------------
    // Execute
    // ----------------------------------------------------------------

    /**
     * Run the full CAPTCHA flow: precheck -> optional challenge -> result.
     *
     * @param context Android context used for device info collection.
     */
    suspend fun execute(context: Context): CaptchaResult {
        return try {
            // Validate config
            if (config.siteKey.isBlank()) {
                return CaptchaResult.Error(CaptchaException("siteKey must not be empty"))
            }

            // Step 1: Collect signals
            val deviceInfo = DeviceInfoCollector.collect(context)
            val behavior = behaviorTracker.behaviorData()

            // Step 2: Precheck
            val precheckResponse = precheck(deviceInfo, behavior)

            // Step 3: Handle action
            when (precheckResponse.action) {
                ChallengeAction.PASS -> CaptchaResult.Passed(precheckResponse.token)
                ChallengeAction.INVISIBLE -> CaptchaResult.Passed(precheckResponse.token)
                ChallengeAction.CHALLENGE -> {
                    val challengeId = precheckResponse.challengeId
                    val challengeType = precheckResponse.challengeType
                    if (challengeId != null && challengeType != null) {
                        CaptchaResult.ChallengeRequired(challengeId, challengeType)
                    } else {
                        CaptchaResult.Error(
                            CaptchaException("Challenge required but no challenge_id or challenge_type returned")
                        )
                    }
                }
                ChallengeAction.DENY -> CaptchaResult.Denied
            }
        } catch (e: Exception) {
            CaptchaResult.Error(e)
        }
    }

    // ----------------------------------------------------------------
    // Verify
    // ----------------------------------------------------------------

    /**
     * Submit the user's challenge answer and get a verification result.
     */
    suspend fun verify(challengeId: String, answer: String): CaptchaResult {
        return try {
            val response = submitChallengeAnswer(challengeId, answer)
            if (response.success && response.token != null) {
                CaptchaResult.Passed(response.token)
            } else {
                val message = response.errorMessage ?: "Verification failed"
                CaptchaResult.Error(CaptchaException(message))
            }
        } catch (e: Exception) {
            CaptchaResult.Error(e)
        }
    }

    // ----------------------------------------------------------------
    // Challenge WebView
    // ----------------------------------------------------------------

    /**
     * Present a WebView to render the challenge and return the answer.
     * Call this when [execute] returns [CaptchaResult.ChallengeRequired].
     *
     * @param activity The activity to host the WebView.
     * @param challengeId The challenge ID from the precheck response.
     * @param challengeType The challenge type from the precheck response.
     */
    suspend fun presentChallenge(
        activity: Activity,
        challengeId: String,
        challengeType: String
    ): CaptchaResult {
        return try {
            // Fetch render data
            val renderResponse = fetchChallengeRender(challengeId)

            // Show WebView and wait for answer
            val answer = showChallengeWebView(activity, renderResponse, challengeId)
                ?: return CaptchaResult.Error(CaptchaException("Challenge cancelled"))

            // Verify
            verify(challengeId, answer)
        } catch (e: Exception) {
            CaptchaResult.Error(e)
        }
    }

    @SuppressLint("SetJavaScriptEnabled")
    private suspend fun showChallengeWebView(
        activity: Activity,
        renderResponse: ChallengeRenderResponse,
        challengeId: String
    ): String? = suspendCancellableCoroutine { cont ->
        activity.runOnUiThread {
            val container = FrameLayout(activity).apply {
                layoutParams = ViewGroup.LayoutParams(
                    ViewGroup.LayoutParams.MATCH_PARENT,
                    ViewGroup.LayoutParams.MATCH_PARENT
                )
            }

            val webView = WebView(activity).apply {
                layoutParams = ViewGroup.LayoutParams(
                    ViewGroup.LayoutParams.MATCH_PARENT,
                    ViewGroup.LayoutParams.MATCH_PARENT
                )
                settings.javaScriptEnabled = true
                settings.domStorageEnabled = true
                webChromeClient = WebChromeClient()
                webViewClient = WebViewClient()
            }

            // Add JS bridge
            webView.addJavascriptInterface(object {
                @JavascriptInterface
                fun submit(answer: String) {
                    activity.runOnUiThread {
                        (container.parent as? ViewGroup)?.removeView(container)
                        if (cont.isActive) cont.resume(answer)
                    }
                }

                @JavascriptInterface
                fun cancel() {
                    activity.runOnUiThread {
                        (container.parent as? ViewGroup)?.removeView(container)
                        if (cont.isActive) cont.resume(null)
                    }
                }
            }, "EngagelabCaptcha")

            container.addView(webView)

            // Add to the activity's root view
            val rootView = activity.findViewById<ViewGroup>(android.R.id.content)
            rootView.addView(container)

            // Build URL
            val uri = Uri.parse(renderResponse.renderUrl).buildUpon()
                .appendQueryParameter("challenge_id", challengeId)
                .appendQueryParameter("theme", config.theme.value)
                .apply { config.lang?.let { appendQueryParameter("lang", it) } }
                .appendQueryParameter("platform", "android")
                .build()

            webView.loadUrl(uri.toString())

            cont.invokeOnCancellation {
                activity.runOnUiThread {
                    (container.parent as? ViewGroup)?.removeView(container)
                }
            }
        }
    }

    // ----------------------------------------------------------------
    // Network
    // ----------------------------------------------------------------

    private suspend fun precheck(
        deviceInfo: Map<String, String>,
        behaviorData: Map<String, String>
    ): PrecheckResponse {
        val body = PrecheckRequest(
            siteKey = config.siteKey,
            scene = config.scene.value,
            deviceInfo = deviceInfo,
            behaviorData = behaviorData,
            lang = config.lang,
            theme = config.theme.value
        )
        return post("/precheck", body)
    }

    private suspend fun fetchChallengeRender(challengeId: String): ChallengeRenderResponse {
        val url = buildUrl("/challenge/render", mapOf(
            "challenge_id" to challengeId,
            "site_key" to config.siteKey
        ))
        return get(url)
    }

    private suspend fun submitChallengeAnswer(
        challengeId: String,
        answer: String
    ): ChallengeVerifyResponse {
        val body = ChallengeVerifyRequest(
            challengeId = challengeId,
            answer = answer,
            token = "" // Token will be set by the server from the session
        )
        return post("/challenge/verify", body)
    }

    // ----------------------------------------------------------------
    // HTTP Helpers
    // ----------------------------------------------------------------

    private fun buildUrl(path: String, query: Map<String, String>? = null): String {
        val base = config.apiBase.trimEnd('/') + path
        if (query.isNullOrEmpty()) return base
        val uri = Uri.parse(base).buildUpon()
        query.forEach { (k, v) -> uri.appendQueryParameter(k, v) }
        return uri.build().toString()
    }

    private fun userAgent(): String {
        return "EngagelabCaptcha-Android/1.0.0"
    }

    private suspend inline fun <reified T : Any> post(path: String, body: Any): T {
        return withContext(Dispatchers.IO) {
            val url = buildUrl(path)
            val jsonBody = json.encodeToString(
                kotlinx.serialization.serializer(body::class.java),
                body
            )
            val request = Request.Builder()
                .url(url)
                .post(jsonBody.toRequestBody(jsonMediaType))
                .header("Content-Type", "application/json")
                .header("User-Agent", userAgent())
                .build()

            executeRequest(request)
        }
    }

    private suspend inline fun <reified T : Any> get(url: String): T {
        return withContext(Dispatchers.IO) {
            val request = Request.Builder()
                .url(url)
                .get()
                .header("Accept", "application/json")
                .header("User-Agent", userAgent())
                .build()

            executeRequest(request)
        }
    }

    private inline fun <reified T : Any> executeRequest(request: Request): T {
        val response = try {
            httpClient.newCall(request).execute()
        } catch (e: IOException) {
            throw CaptchaException("Network error: ${e.message}", e)
        }

        val responseBody = response.body?.string()
            ?: throw CaptchaException("Empty response body, status=${response.code}")

        if (!response.isSuccessful) {
            throw CaptchaException("Server error ${response.code}: $responseBody")
        }

        return try {
            json.decodeFromString<T>(responseBody)
        } catch (e: Exception) {
            throw CaptchaException("Decoding error: ${e.message}", e)
        }
    }
}

/**
 * Optional: Standalone Activity that hosts the challenge WebView.
 * Can be launched via Intent when a full-screen challenge experience is desired.
 */
class ChallengeActivity : AppCompatActivity() {

    companion object {
        const val EXTRA_RENDER_URL = "render_url"
        const val EXTRA_CHALLENGE_ID = "challenge_id"
        const val EXTRA_THEME = "theme"
        const val EXTRA_LANG = "lang"
        const val RESULT_ANSWER = "answer"

        internal val pendingResult = CompletableDeferred<String?>()
    }

    @SuppressLint("SetJavaScriptEnabled")
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val renderUrl = intent.getStringExtra(EXTRA_RENDER_URL) ?: run {
            pendingResult.complete(null)
            finish()
            return
        }
        val challengeId = intent.getStringExtra(EXTRA_CHALLENGE_ID) ?: ""
        val theme = intent.getStringExtra(EXTRA_THEME) ?: "light"
        val lang = intent.getStringExtra(EXTRA_LANG)

        val webView = WebView(this).apply {
            layoutParams = ViewGroup.LayoutParams(
                ViewGroup.LayoutParams.MATCH_PARENT,
                ViewGroup.LayoutParams.MATCH_PARENT
            )
            settings.javaScriptEnabled = true
            settings.domStorageEnabled = true
            webChromeClient = WebChromeClient()
            webViewClient = WebViewClient()
        }

        webView.addJavascriptInterface(object {
            @JavascriptInterface
            fun submit(answer: String) {
                pendingResult.complete(answer)
                runOnUiThread { finish() }
            }

            @JavascriptInterface
            fun cancel() {
                pendingResult.complete(null)
                runOnUiThread { finish() }
            }
        }, "EngagelabCaptcha")

        setContentView(webView)

        val uri = Uri.parse(renderUrl).buildUpon()
            .appendQueryParameter("challenge_id", challengeId)
            .appendQueryParameter("theme", theme)
            .apply { lang?.let { appendQueryParameter("lang", it) } }
            .appendQueryParameter("platform", "android")
            .build()

        webView.loadUrl(uri.toString())
    }

    override fun onBackPressed() {
        super.onBackPressed()
        pendingResult.complete(null)
    }
}
