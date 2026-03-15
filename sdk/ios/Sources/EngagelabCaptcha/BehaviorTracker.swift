import Foundation
#if canImport(UIKit)
import UIKit
#endif

/// Tracks user touch behavior and computes behavioral signals for risk analysis.
public final class BehaviorTracker: @unchecked Sendable {

    // MARK: - Types

    /// A single recorded touch event.
    public struct TouchEvent: Sendable {
        public let x: Double
        public let y: Double
        public let timestamp: TimeInterval
        public let phase: String // began, moved, ended

        public init(x: Double, y: Double, timestamp: TimeInterval, phase: String) {
            self.x = x
            self.y = y
            self.timestamp = timestamp
            self.phase = phase
        }
    }

    // MARK: - Properties

    private let lock = NSLock()
    private var touchEvents: [TouchEvent] = []
    private var sessionStartTime: TimeInterval
    private var accelerometerAvailable: Bool = false

    // MARK: - Init

    public init() {
        self.sessionStartTime = ProcessInfo.processInfo.systemUptime
        self.accelerometerAvailable = checkAccelerometerAvailability()
    }

    // MARK: - Public API

    /// Record a touch event. Call from your UIView's touch handlers.
    public func recordTouch(x: Double, y: Double, phase: String) {
        let event = TouchEvent(
            x: x,
            y: y,
            timestamp: ProcessInfo.processInfo.systemUptime,
            phase: phase
        )
        lock.lock()
        touchEvents.append(event)
        lock.unlock()
    }

    #if canImport(UIKit)
    /// Convenience to record a UITouch directly.
    public func recordUITouch(_ touch: UITouch, in view: UIView) {
        let location = touch.location(in: view)
        let phase: String
        switch touch.phase {
        case .began: phase = "began"
        case .moved: phase = "moved"
        case .ended: phase = "ended"
        case .cancelled: phase = "cancelled"
        default: phase = "other"
        }
        recordTouch(x: Double(location.x), y: Double(location.y), phase: phase)
    }
    #endif

    /// Reset all collected data for a new session.
    public func reset() {
        lock.lock()
        touchEvents.removeAll()
        sessionStartTime = ProcessInfo.processInfo.systemUptime
        lock.unlock()
    }

    /// Return behavioral data as a dictionary suitable for sending to the API.
    public func behaviorData() -> [String: String] {
        lock.lock()
        let events = touchEvents
        let startTime = sessionStartTime
        lock.unlock()

        let now = ProcessInfo.processInfo.systemUptime
        let sessionDuration = now - startTime

        var data = [String: String]()
        data["session_duration_ms"] = "\(Int(sessionDuration * 1000))"
        data["touch_count"] = "\(events.count)"
        data["accelerometer_available"] = accelerometerAvailable ? "true" : "false"

        // Compute touch entropy
        let entropy = computeTouchEntropy(events: events)
        data["touch_entropy"] = String(format: "%.4f", entropy)

        // Average touch velocity
        let velocity = computeAverageVelocity(events: events)
        data["avg_touch_velocity"] = String(format: "%.2f", velocity)

        // Phase distribution
        let phases = events.map { $0.phase }
        let beganCount = phases.filter { $0 == "began" }.count
        let movedCount = phases.filter { $0 == "moved" }.count
        let endedCount = phases.filter { $0 == "ended" }.count
        data["touch_began_count"] = "\(beganCount)"
        data["touch_moved_count"] = "\(movedCount)"
        data["touch_ended_count"] = "\(endedCount)"

        // Serialize touch positions as compact JSON for server-side analysis
        if !events.isEmpty {
            let sample = Array(events.prefix(50))
            let touchArray = sample.map { "[\(Int($0.x)),\(Int($0.y)),\(Int($0.timestamp * 1000))]" }
            data["touch_samples"] = "[\(touchArray.joined(separator: ","))]"
        }

        return data
    }

    // MARK: - Analysis

    /// Compute Shannon entropy over a discretized grid of touch positions.
    /// Higher entropy indicates more natural, human-like touch patterns.
    private func computeTouchEntropy(events: [TouchEvent]) -> Double {
        guard events.count > 1 else { return 0.0 }

        // Discretize positions into a 10x10 grid
        let gridSize = 10.0
        var bins = [String: Int]()

        let xs = events.map { $0.x }
        let ys = events.map { $0.y }
        let minX = xs.min() ?? 0
        let maxX = xs.max() ?? 1
        let minY = ys.min() ?? 0
        let maxY = ys.max() ?? 1
        let rangeX = max(maxX - minX, 1.0)
        let rangeY = max(maxY - minY, 1.0)

        for event in events {
            let gx = Int(((event.x - minX) / rangeX) * gridSize)
            let gy = Int(((event.y - minY) / rangeY) * gridSize)
            let key = "\(min(gx, Int(gridSize) - 1)),\(min(gy, Int(gridSize) - 1))"
            bins[key, default: 0] += 1
        }

        let total = Double(events.count)
        var entropy = 0.0
        for (_, count) in bins {
            let p = Double(count) / total
            if p > 0 {
                entropy -= p * log2(p)
            }
        }

        return entropy
    }

    /// Compute average velocity (pixels/second) between consecutive move events.
    private func computeAverageVelocity(events: [TouchEvent]) -> Double {
        let moveEvents = events.filter { $0.phase == "moved" }
        guard moveEvents.count > 1 else { return 0.0 }

        var totalVelocity = 0.0
        var count = 0

        for i in 1..<moveEvents.count {
            let prev = moveEvents[i - 1]
            let curr = moveEvents[i]
            let dt = curr.timestamp - prev.timestamp
            guard dt > 0 else { continue }

            let dx = curr.x - prev.x
            let dy = curr.y - prev.y
            let distance = sqrt(dx * dx + dy * dy)
            totalVelocity += distance / dt
            count += 1
        }

        return count > 0 ? totalVelocity / Double(count) : 0.0
    }

    // MARK: - Accelerometer

    private func checkAccelerometerAvailability() -> Bool {
        #if canImport(CoreMotion)
        // CoreMotion is available on iOS; actual availability checked at runtime
        return true
        #else
        return false
        #endif
    }
}
