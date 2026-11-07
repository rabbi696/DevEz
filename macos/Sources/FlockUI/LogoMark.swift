import AppKit
import SwiftUI

/// Flock's logomark, from the template images that build-app.sh puts in the app bundle
/// (rendered by design/icongen). Template images take the foreground style, and in the menu bar
/// macOS colors them for its appearance. Without the bundle (`swift run`), an SF Symbol bird.
struct LogoMark: View {
    enum Mode {
        case running, stopped, warning

        /// The menu bar's 18-point image, or a 128-pixel one to scale down inside windows.
        func resource(large: Bool) -> String {
            let suffix = switch self {
            case .running: ""
            case .stopped: "Stopped"
            case .warning: "Warning"
            }
            return (large ? "LogoMark" : "MenuBar") + suffix
        }

        var fallback: String {
            switch self {
            case .running: "server.rack"
            case .stopped: "server.rack"
            case .warning: "exclamationmark.triangle.fill"
            }
        }
    }

    var mode: Mode = .running
    /// Points; nil keeps the image's own 18-point size (the menu bar's).
    var size: CGFloat? = nil

    var body: some View {
        if let size, let image = Self.image(mode.resource(large: true)) {
            Image(nsImage: image).resizable().frame(width: size, height: size)
        } else if size == nil, let image = Self.image(mode.resource(large: false)) {
            Image(nsImage: image)
        } else {
            Image(systemName: mode.fallback)
                .font(size.map { .system(size: $0 * 0.8) } ?? .body)
        }
    }

    @MainActor private static var cache: [String: NSImage] = [:]

    @MainActor private static func image(_ resource: String) -> NSImage? {
        if let image = cache[resource] {
            return image
        }
        guard let image = Bundle.main.image(forResource: resource) else {
            return nil
        }
        image.isTemplate = true
        cache[resource] = image
        return image
    }
}

extension Color {
    /// DevEz signature brand color (Laravel Herd-inspired vibrant coral/red).
    static let flockMark = Color(nsColor: NSColor(name: nil) { appearance in
        appearance.bestMatch(from: [.aqua, .darkAqua]) == .darkAqua
            ? NSColor(srgbRed: 0xFF / 255, green: 0x4F / 255, blue: 0x43 / 255, alpha: 1)
            : NSColor(srgbRed: 0xFF / 255, green: 0x2D / 255, blue: 0x20 / 255, alpha: 1)
    })

    /// Herd amber / orange for warnings or degraded states.
    static let flockEmber = Color(.sRGB, red: 0xFB / 255, green: 0x92 / 255, blue: 0x3C / 255)
}

/// The app icon, as Finder shows it (the tile is 80% of it: Apple's icon grid leaves room for the shadow).
struct AppIconView: View {
    var size: CGFloat

    var body: some View {
        Image(nsImage: NSApp.applicationIconImage)
            .resizable()
            .frame(width: size, height: size)
    }
}
