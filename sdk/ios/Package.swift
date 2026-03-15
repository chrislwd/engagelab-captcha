// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "EngagelabCaptcha",
    platforms: [
        .iOS(.v14)
    ],
    products: [
        .library(
            name: "EngagelabCaptcha",
            targets: ["EngagelabCaptcha"]
        )
    ],
    targets: [
        .target(
            name: "EngagelabCaptcha",
            dependencies: [],
            path: "Sources/EngagelabCaptcha"
        ),
        .testTarget(
            name: "EngagelabCaptchaTests",
            dependencies: ["EngagelabCaptcha"],
            path: "Tests/EngagelabCaptchaTests"
        )
    ]
)
