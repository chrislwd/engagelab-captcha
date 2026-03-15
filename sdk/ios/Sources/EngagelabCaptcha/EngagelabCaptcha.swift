import Foundation
#if canImport(UIKit)
import UIKit
import WebKit
#endif

/// Main entry point for the EngageLab CAPTCHA SDK.
///
/// Usage:
/// ```swift
/// let config = CaptchaConfig(siteKey: "your-site-key", scene: .login)
/// let captcha = EngagelabCaptcha(config: config)
/// let result = await captcha.execute()
/// switch result {
/// case .passed(let token):
///     // send token to your backend
/// case .challengeRequired(let challengeID, let type):
///     // present challenge UI
/// case .denied:
///     // block the action
/// case .error(let err):
///     // handle error
/// }
/// ```
public final class EngagelabCaptcha: @unchecked Sendable {

    // MARK: - Errors

    public enum CaptchaError: Error, LocalizedError {
        case invalidConfig(String)
        case networkError(Error)
        case serverError(statusCode: Int, message: String)
        case decodingError(Error)
        case challengeTimeout
        case challengeFailed(String)
        case webViewUnavailable
        case cancelled

        public var errorDescription: String? {
            switch self {
            case .invalidConfig(let msg): return "Invalid config: \(msg)"
            case .networkError(let err): return "Network error: \(err.localizedDescription)"
            case .serverError(let code, let msg): return "Server error \(code): \(msg)"
            case .decodingError(let err): return "Decoding error: \(err.localizedDescription)"
            case .challengeTimeout: return "Challenge timed out"
            case .challengeFailed(let msg): return "Challenge failed: \(msg)"
            case .webViewUnavailable: return "WebView is not available"
            case .cancelled: return "CAPTCHA was cancelled"
            }
        }
    }

    // MARK: - Properties

    public let config: CaptchaConfig
    public let behaviorTracker: BehaviorTracker
    private let deviceCollector: DeviceInfoCollector
    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    // MARK: - Init

    public init(config: CaptchaConfig) {
        self.config = config
        self.behaviorTracker = BehaviorTracker()
        self.deviceCollector = DeviceInfoCollector.shared

        let sessionConfig = URLSessionConfiguration.default
        sessionConfig.timeoutIntervalForRequest = 15
        sessionConfig.timeoutIntervalForResource = 30
        self.session = URLSession(configuration: sessionConfig)

        self.decoder = JSONDecoder()
        self.encoder = JSONEncoder()
    }

    // MARK: - Execute

    /// Run the full CAPTCHA flow: precheck -> optional challenge -> result.
    @available(iOS 14.0, *)
    public func execute() async -> CaptchaResult {
        do {
            // Validate config
            guard !config.siteKey.isEmpty else {
                return .error(CaptchaError.invalidConfig("siteKey must not be empty"))
            }

            // Step 1: Collect signals
            let deviceInfo = deviceCollector.collect()
            let behavior = behaviorTracker.behaviorData()

            // Step 2: Precheck
            let precheckResult = try await precheck(deviceInfo: deviceInfo, behaviorData: behavior)

            // Step 3: Handle action
            switch precheckResult.action {
            case .pass:
                return .passed(token: precheckResult.token)

            case .invisible:
                // Invisible verification passed server-side
                return .passed(token: precheckResult.token)

            case .challenge:
                guard let challengeID = precheckResult.challengeID,
                      let challengeType = precheckResult.challengeType else {
                    return .error(CaptchaError.serverError(
                        statusCode: 0,
                        message: "Challenge required but no challenge_id or challenge_type returned"
                    ))
                }
                return .challengeRequired(challengeID: challengeID, type: challengeType)

            case .deny:
                return .denied
            }
        } catch {
            return .error(error)
        }
    }

    // MARK: - Verify

    /// Submit the user's challenge answer and get a verification result.
    @available(iOS 14.0, *)
    public func verify(challengeID: String, answer: String) async -> CaptchaResult {
        do {
            let verifyResponse = try await submitChallengeAnswer(
                challengeID: challengeID,
                answer: answer
            )

            if verifyResponse.success, let token = verifyResponse.token {
                return .passed(token: token)
            } else {
                let message = verifyResponse.errorMessage ?? "Verification failed"
                return .error(CaptchaError.challengeFailed(message))
            }
        } catch {
            return .error(error)
        }
    }

    // MARK: - Challenge WebView

    #if canImport(UIKit)
    /// Present a WKWebView to render the challenge and return the answer.
    /// Call this when `execute()` returns `.challengeRequired`.
    @MainActor
    @available(iOS 14.0, *)
    public func presentChallenge(
        challengeID: String,
        challengeType: String,
        from viewController: UIViewController
    ) async -> CaptchaResult {
        do {
            // Fetch render data
            let renderResponse = try await fetchChallengeRender(challengeID: challengeID)

            // Create and present WebView
            return await withCheckedContinuation { continuation in
                let webVC = ChallengeWebViewController(
                    renderURL: renderResponse.renderURL,
                    challengeID: challengeID,
                    config: config
                ) { [weak self] result in
                    guard let self = self else {
                        continuation.resume(returning: .error(CaptchaError.cancelled))
                        return
                    }
                    switch result {
                    case .answer(let answer):
                        Task {
                            let verifyResult = await self.verify(
                                challengeID: challengeID,
                                answer: answer
                            )
                            continuation.resume(returning: verifyResult)
                        }
                    case .cancelled:
                        continuation.resume(returning: .error(CaptchaError.cancelled))
                    case .error(let error):
                        continuation.resume(returning: .error(error))
                    }
                }
                viewController.present(webVC, animated: true)
            }
        } catch {
            return .error(error)
        }
    }
    #endif

    // MARK: - Network

    private func precheck(
        deviceInfo: [String: String],
        behaviorData: [String: String]
    ) async throws -> PrecheckResponse {
        let body = PrecheckRequest(
            siteKey: config.siteKey,
            scene: config.scene.rawValue,
            deviceInfo: deviceInfo,
            behaviorData: behaviorData,
            lang: config.lang,
            theme: config.theme.rawValue
        )
        return try await post(path: "/precheck", body: body)
    }

    private func fetchChallengeRender(challengeID: String) async throws -> ChallengeRenderResponse {
        let url = try buildURL(path: "/challenge/render", query: [
            "challenge_id": challengeID,
            "site_key": config.siteKey
        ])
        return try await get(url: url)
    }

    private func submitChallengeAnswer(
        challengeID: String,
        answer: String
    ) async throws -> ChallengeVerifyResponse {
        // Collect fresh device info from the current session
        let deviceInfo = deviceCollector.collect()
        let token = deviceInfo["fingerprint"] ?? ""

        let body = ChallengeVerifyRequest(
            challengeID: challengeID,
            answer: answer,
            token: token
        )
        return try await post(path: "/challenge/verify", body: body)
    }

    // MARK: - HTTP Helpers

    private func buildURL(path: String, query: [String: String]? = nil) throws -> URL {
        guard var components = URLComponents(string: config.apiBase + path) else {
            throw CaptchaError.invalidConfig("Invalid API base URL: \(config.apiBase)")
        }
        if let query = query {
            components.queryItems = query.map { URLQueryItem(name: $0.key, value: $0.value) }
        }
        guard let url = components.url else {
            throw CaptchaError.invalidConfig("Could not construct URL for path: \(path)")
        }
        return url
    }

    private func post<T: Encodable, R: Decodable>(path: String, body: T) async throws -> R {
        let url = try buildURL(path: path)
        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue(DeviceInfoCollector.shared.collect()["user_agent"], forHTTPHeaderField: "User-Agent")
        request.httpBody = try encoder.encode(body)
        return try await perform(request: request)
    }

    private func get<R: Decodable>(url: URL) async throws -> R {
        var request = URLRequest(url: url)
        request.httpMethod = "GET"
        request.setValue("application/json", forHTTPHeaderField: "Accept")
        request.setValue(DeviceInfoCollector.shared.collect()["user_agent"], forHTTPHeaderField: "User-Agent")
        return try await perform(request: request)
    }

    private func perform<R: Decodable>(request: URLRequest) async throws -> R {
        let (data, response): (Data, URLResponse)
        do {
            (data, response) = try await session.data(for: request)
        } catch {
            throw CaptchaError.networkError(error)
        }

        guard let http = response as? HTTPURLResponse else {
            throw CaptchaError.serverError(statusCode: 0, message: "Non-HTTP response")
        }

        guard (200..<300).contains(http.statusCode) else {
            let body = String(data: data, encoding: .utf8) ?? "no body"
            throw CaptchaError.serverError(statusCode: http.statusCode, message: body)
        }

        do {
            return try decoder.decode(R.self, from: data)
        } catch {
            throw CaptchaError.decodingError(error)
        }
    }
}

// MARK: - Challenge WebView Controller

#if canImport(UIKit)
/// Internal view controller that hosts a WKWebView to render CAPTCHA challenges.
@available(iOS 14.0, *)
internal final class ChallengeWebViewController: UIViewController, WKNavigationDelegate, WKScriptMessageHandler {

    enum ChallengeResult {
        case answer(String)
        case cancelled
        case error(Error)
    }

    private let renderURL: String
    private let challengeID: String
    private let config: CaptchaConfig
    private let completion: (ChallengeResult) -> Void
    private var webView: WKWebView!

    init(
        renderURL: String,
        challengeID: String,
        config: CaptchaConfig,
        completion: @escaping (ChallengeResult) -> Void
    ) {
        self.renderURL = renderURL
        self.challengeID = challengeID
        self.config = config
        self.completion = completion
        super.init(nibName: nil, bundle: nil)
        self.modalPresentationStyle = .pageSheet
    }

    required init?(coder: NSCoder) {
        fatalError("init(coder:) has not been implemented")
    }

    override func viewDidLoad() {
        super.viewDidLoad()
        view.backgroundColor = .systemBackground
        setupWebView()
        loadChallenge()
    }

    private func setupWebView() {
        let contentController = WKUserContentController()
        contentController.add(self, name: "captchaCallback")

        // Inject JS bridge so the challenge page can communicate back
        let bridgeScript = WKUserScript(
            source: """
            window.EngagelabCaptcha = {
                submit: function(answer) {
                    window.webkit.messageHandlers.captchaCallback.postMessage({
                        type: 'submit',
                        answer: answer
                    });
                },
                cancel: function() {
                    window.webkit.messageHandlers.captchaCallback.postMessage({
                        type: 'cancel'
                    });
                }
            };
            """,
            injectionTime: .atDocumentStart,
            forMainFrameOnly: true
        )
        contentController.addUserScript(bridgeScript)

        let webConfig = WKWebViewConfiguration()
        webConfig.userContentController = contentController
        webConfig.allowsInlineMediaPlayback = true

        webView = WKWebView(frame: .zero, configuration: webConfig)
        webView.navigationDelegate = self
        webView.translatesAutoresizingMaskIntoConstraints = false
        view.addSubview(webView)

        NSLayoutConstraint.activate([
            webView.topAnchor.constraint(equalTo: view.safeAreaLayoutGuide.topAnchor),
            webView.leadingAnchor.constraint(equalTo: view.leadingAnchor),
            webView.trailingAnchor.constraint(equalTo: view.trailingAnchor),
            webView.bottomAnchor.constraint(equalTo: view.bottomAnchor)
        ])

        // Add close button
        let closeButton = UIButton(type: .close)
        closeButton.translatesAutoresizingMaskIntoConstraints = false
        closeButton.addTarget(self, action: #selector(closeTapped), for: .touchUpInside)
        view.addSubview(closeButton)

        NSLayoutConstraint.activate([
            closeButton.topAnchor.constraint(equalTo: view.safeAreaLayoutGuide.topAnchor, constant: 8),
            closeButton.trailingAnchor.constraint(equalTo: view.trailingAnchor, constant: -16)
        ])
    }

    private func loadChallenge() {
        // Build the render URL with query params
        var components = URLComponents(string: renderURL)
        var queryItems = components?.queryItems ?? []
        queryItems.append(URLQueryItem(name: "challenge_id", value: challengeID))
        queryItems.append(URLQueryItem(name: "theme", value: config.theme.rawValue))
        if let lang = config.lang {
            queryItems.append(URLQueryItem(name: "lang", value: lang))
        }
        queryItems.append(URLQueryItem(name: "platform", value: "ios"))
        components?.queryItems = queryItems

        if let url = components?.url {
            webView.load(URLRequest(url: url))
        } else {
            completion(.error(EngagelabCaptcha.CaptchaError.invalidConfig("Invalid render URL")))
            dismiss(animated: true)
        }
    }

    @objc private func closeTapped() {
        completion(.cancelled)
        dismiss(animated: true)
    }

    // MARK: - WKScriptMessageHandler

    func userContentController(
        _ userContentController: WKUserContentController,
        didReceive message: WKScriptMessage
    ) {
        guard let body = message.body as? [String: Any],
              let type = body["type"] as? String else { return }

        switch type {
        case "submit":
            if let answer = body["answer"] as? String {
                completion(.answer(answer))
                dismiss(animated: true)
            }
        case "cancel":
            completion(.cancelled)
            dismiss(animated: true)
        default:
            break
        }
    }

    // MARK: - WKNavigationDelegate

    func webView(
        _ webView: WKWebView,
        didFailProvisionalNavigation navigation: WKNavigation!,
        withError error: Error
    ) {
        completion(.error(error))
        dismiss(animated: true)
    }

    func webView(_ webView: WKWebView, didFail navigation: WKNavigation!, withError error: Error) {
        completion(.error(error))
        dismiss(animated: true)
    }
}
#endif
