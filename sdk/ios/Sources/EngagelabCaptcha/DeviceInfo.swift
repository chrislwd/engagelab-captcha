import Foundation
#if canImport(UIKit)
import UIKit
#endif
import CommonCrypto

/// Collects device-level signals for risk analysis.
public final class DeviceInfoCollector: @unchecked Sendable {

    public static let shared = DeviceInfoCollector()

    private init() {}

    /// Gather all device info into a string-keyed dictionary.
    public func collect() -> [String: String] {
        var info = [String: String]()

        #if canImport(UIKit)
        let device = UIDevice.current
        info["model"] = device.model
        info["system_name"] = device.systemName
        info["system_version"] = device.systemVersion
        info["device_name"] = device.name
        info["identifier_for_vendor"] = device.identifierForVendor?.uuidString ?? "unknown"

        if let screen = screenSize() {
            info["screen_width"] = "\(Int(screen.width))"
            info["screen_height"] = "\(Int(screen.height))"
            info["screen_scale"] = "\(Int(screen.scale))"
        }
        #else
        info["model"] = "unknown"
        info["system_name"] = "unknown"
        info["system_version"] = "unknown"
        #endif

        info["lang"] = Locale.preferredLanguages.first ?? Locale.current.identifier
        info["timezone"] = TimeZone.current.identifier
        info["timezone_offset"] = "\(TimeZone.current.secondsFromGMT())"
        info["user_agent"] = userAgentString()
        info["fingerprint"] = generateFingerprint(from: info)

        return info
    }

    // MARK: - Screen

    #if canImport(UIKit)
    private struct ScreenInfo {
        let width: CGFloat
        let height: CGFloat
        let scale: CGFloat
    }

    private func screenSize() -> ScreenInfo? {
        let bounds = UIScreen.main.bounds
        let scale = UIScreen.main.scale
        return ScreenInfo(width: bounds.width, height: bounds.height, scale: scale)
    }
    #endif

    // MARK: - User Agent

    /// Build a simple user-agent string describing this SDK and the device.
    private func userAgentString() -> String {
        let sdkVersion = "1.0.0"
        #if canImport(UIKit)
        let device = UIDevice.current
        return "EngagelabCaptcha-iOS/\(sdkVersion) (\(device.model); \(device.systemName) \(device.systemVersion))"
        #else
        return "EngagelabCaptcha-iOS/\(sdkVersion)"
        #endif
    }

    // MARK: - Fingerprint

    /// Generate a simple SHA-256 fingerprint hash from the collected info values.
    public func generateFingerprint(from info: [String: String]) -> String {
        let sortedKeys = info.keys.sorted()
        let combined = sortedKeys.map { "\($0)=\(info[$0] ?? "")" }.joined(separator: "&")
        return sha256(combined)
    }

    private func sha256(_ input: String) -> String {
        guard let data = input.data(using: .utf8) else { return "" }
        var hash = [UInt8](repeating: 0, count: Int(CC_SHA256_DIGEST_LENGTH))
        data.withUnsafeBytes { buffer in
            _ = CC_SHA256(buffer.baseAddress, CC_LONG(data.count), &hash)
        }
        return hash.map { String(format: "%02x", $0) }.joined()
    }
}
