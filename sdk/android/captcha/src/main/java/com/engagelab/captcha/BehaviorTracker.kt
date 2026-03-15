package com.engagelab.captcha

import android.os.SystemClock
import android.view.MotionEvent
import kotlin.math.ln
import kotlin.math.max
import kotlin.math.min
import kotlin.math.sqrt

/**
 * Tracks user touch behavior and computes behavioral signals for risk analysis.
 */
class BehaviorTracker {

    // ----------------------------------------------------------------
    // Types
    // ----------------------------------------------------------------

    /**
     * A single recorded touch event.
     */
    data class TouchEvent(
        val x: Float,
        val y: Float,
        val timestampMs: Long,
        val action: String
    )

    // ----------------------------------------------------------------
    // State
    // ----------------------------------------------------------------

    private val lock = Any()
    private val touchEvents = mutableListOf<TouchEvent>()
    private var sessionStartMs: Long = SystemClock.elapsedRealtime()

    // ----------------------------------------------------------------
    // Public API
    // ----------------------------------------------------------------

    /**
     * Record a raw touch event by coordinates and action name.
     */
    fun recordTouch(x: Float, y: Float, action: String) {
        val event = TouchEvent(
            x = x,
            y = y,
            timestampMs = SystemClock.elapsedRealtime(),
            action = action
        )
        synchronized(lock) {
            touchEvents.add(event)
        }
    }

    /**
     * Convenience: record from an Android [MotionEvent].
     */
    fun recordMotionEvent(event: MotionEvent) {
        val action = when (event.actionMasked) {
            MotionEvent.ACTION_DOWN -> "down"
            MotionEvent.ACTION_MOVE -> "move"
            MotionEvent.ACTION_UP -> "up"
            MotionEvent.ACTION_CANCEL -> "cancel"
            else -> "other"
        }
        recordTouch(event.x, event.y, action)
    }

    /**
     * Reset all collected data for a new session.
     */
    fun reset() {
        synchronized(lock) {
            touchEvents.clear()
            sessionStartMs = SystemClock.elapsedRealtime()
        }
    }

    /**
     * Return behavioral data as a map suitable for sending to the API.
     */
    fun behaviorData(): Map<String, String> {
        val events: List<TouchEvent>
        val startMs: Long
        synchronized(lock) {
            events = touchEvents.toList()
            startMs = sessionStartMs
        }

        val now = SystemClock.elapsedRealtime()
        val sessionDuration = now - startMs

        val data = mutableMapOf<String, String>()
        data["session_duration_ms"] = sessionDuration.toString()
        data["touch_count"] = events.size.toString()

        // Touch entropy
        val entropy = computeTouchEntropy(events)
        data["touch_entropy"] = "%.4f".format(entropy)

        // Average velocity
        val velocity = computeAverageVelocity(events)
        data["avg_touch_velocity"] = "%.2f".format(velocity)

        // Phase distribution
        val downCount = events.count { it.action == "down" }
        val moveCount = events.count { it.action == "move" }
        val upCount = events.count { it.action == "up" }
        data["touch_down_count"] = downCount.toString()
        data["touch_move_count"] = moveCount.toString()
        data["touch_up_count"] = upCount.toString()

        // Serialize touch positions as compact JSON for server-side analysis
        if (events.isNotEmpty()) {
            val sample = events.take(50)
            val touchArray = sample.joinToString(",") { "[${it.x.toInt()},${it.y.toInt()},${it.timestampMs}]" }
            data["touch_samples"] = "[$touchArray]"
        }

        return data
    }

    // ----------------------------------------------------------------
    // Analysis
    // ----------------------------------------------------------------

    /**
     * Compute Shannon entropy over a discretized grid of touch positions.
     * Higher entropy indicates more natural, human-like touch patterns.
     */
    private fun computeTouchEntropy(events: List<TouchEvent>): Double {
        if (events.size <= 1) return 0.0

        val gridSize = 10.0
        val xs = events.map { it.x.toDouble() }
        val ys = events.map { it.y.toDouble() }
        val minX = xs.min()
        val maxX = xs.max()
        val minY = ys.min()
        val maxY = ys.max()
        val rangeX = max(maxX - minX, 1.0)
        val rangeY = max(maxY - minY, 1.0)

        val bins = mutableMapOf<String, Int>()
        for (event in events) {
            val gx = min(((event.x - minX) / rangeX * gridSize).toInt(), gridSize.toInt() - 1)
            val gy = min(((event.y - minY) / rangeY * gridSize).toInt(), gridSize.toInt() - 1)
            val key = "$gx,$gy"
            bins[key] = (bins[key] ?: 0) + 1
        }

        val total = events.size.toDouble()
        var entropy = 0.0
        for ((_, count) in bins) {
            val p = count / total
            if (p > 0) {
                entropy -= p * (ln(p) / ln(2.0))
            }
        }

        return entropy
    }

    /**
     * Compute average velocity (pixels/second) between consecutive move events.
     */
    private fun computeAverageVelocity(events: List<TouchEvent>): Double {
        val moveEvents = events.filter { it.action == "move" }
        if (moveEvents.size <= 1) return 0.0

        var totalVelocity = 0.0
        var count = 0

        for (i in 1 until moveEvents.size) {
            val prev = moveEvents[i - 1]
            val curr = moveEvents[i]
            val dtMs = curr.timestampMs - prev.timestampMs
            if (dtMs <= 0) continue

            val dx = (curr.x - prev.x).toDouble()
            val dy = (curr.y - prev.y).toDouble()
            val distance = sqrt(dx * dx + dy * dy)
            val dtSec = dtMs / 1000.0
            totalVelocity += distance / dtSec
            count++
        }

        return if (count > 0) totalVelocity / count else 0.0
    }
}
