import AppKit
import FlockKit
import SwiftUI

// MARK: - Color Palette

extension Color {
    /// Modern accent coral matching brand mark
    static let flare = Color.flockMark
    /// Deep coral / red
    static let flareDeep = Color(red: 0.93, green: 0.22, blue: 0.33)
    /// Modern status green
    static let live = Color(red: 0.19, green: 0.80, blue: 0.47)
    /// Modern warning yellow / amber
    static let pending = Color(red: 0.98, green: 0.69, blue: 0.18)
    /// Modern danger / error red
    static let fault = Color(red: 0.96, green: 0.30, blue: 0.30)
}

// MARK: - Modern Surface

/// Quiet raised surface modifier that looks sleek and natural in both macOS light and dark appearances.
struct Surface: ViewModifier {
    var radius: CGFloat = 14
    var highlighted: Bool = false

    func body(content: Content) -> some View {
        content
            .background {
                RoundedRectangle(cornerRadius: radius, style: .continuous)
                    .fill(Color.primary.opacity(highlighted ? 0.065 : 0.04))
            }
            .overlay {
                RoundedRectangle(cornerRadius: radius, style: .continuous)
                    .strokeBorder(Color.primary.opacity(0.07), lineWidth: 1)
            }
    }
}

extension View {
    func surface(radius: CGFloat = 14, highlighted: Bool = false) -> some View {
        modifier(Surface(radius: radius, highlighted: highlighted))
    }
}

// MARK: - Button & Interaction Styles

/// ButtonStyle with subtle spring dip on press and smooth hover highlight.
struct PressableStyle: ButtonStyle {
    var hoverFill = true

    func makeBody(configuration: Configuration) -> some View {
        HoverHost(hoverFill: hoverFill) {
            configuration.label
                .scaleEffect(configuration.isPressed ? 0.94 : 1)
                .animation(.spring(response: 0.28, dampingFraction: 0.85), value: configuration.isPressed)
        }
    }

    private struct HoverHost<Label: View>: View {
        let hoverFill: Bool
        @ViewBuilder let label: Label
        @State private var hovering = false

        var body: some View {
            label
                .background {
                    if hoverFill {
                        RoundedRectangle(cornerRadius: 6, style: .continuous)
                            .fill(Color.primary.opacity(hovering ? 0.08 : 0))
                    }
                }
                .onHover { hovering = $0 }
                .animation(.easeOut(duration: 0.12), value: hovering)
        }
    }
}

/// Copy button that animates to a green checkmark upon copying.
struct CopyButton: View {
    let copied: Bool
    let action: () -> Void

    var body: some View {
        Button(action: action) {
            ZStack {
                Image(systemName: copied ? "checkmark" : "doc.on.doc")
                    .id(copied)
                    .transition(.scale(scale: 0.6).combined(with: .opacity))
            }
            .font(.system(size: 11, weight: .semibold))
            .frame(width: 22, height: 22)
            .foregroundStyle(copied ? Color.live : .secondary)
            .contentShape(Rectangle())
        }
        .buttonStyle(PressableStyle())
        .animation(.spring(response: 0.3, dampingFraction: 0.86), value: copied)
        .help(copied ? "Copied" : "Copy to clipboard")
    }
}

// MARK: - Brand & Header Components

/// High-res squircle brand logo mark.
struct DevEzBrandMark: View {
    var size: CGFloat = 28

    var body: some View {
        if let icon = NSApp.applicationIconImage, icon.isValid {
            Image(nsImage: icon)
                .resizable()
                .frame(width: size, height: size)
                .clipShape(RoundedRectangle(cornerRadius: size * 0.24, style: .continuous))
                .shadow(color: Color.flockMark.opacity(0.35), radius: size * 0.12, y: size * 0.06)
        } else {
            RoundedRectangle(cornerRadius: size * 0.24, style: .continuous)
                .fill(LinearGradient(colors: [Color.flockMark, Color(red: 0.90, green: 0.18, blue: 0.12)], startPoint: .topLeading, endPoint: .bottomTrailing))
                .overlay {
                    HStack(spacing: size * 0.08) {
                        Image(systemName: "chevron.right")
                            .font(.system(size: size * 0.38, weight: .bold))
                        RoundedRectangle(cornerRadius: size * 0.04)
                            .frame(width: size * 0.28, height: size * 0.10)
                    }
                    .foregroundStyle(.white)
                }
                .frame(width: size, height: size)
                .shadow(color: Color.flockMark.opacity(0.35), radius: size * 0.12, y: size * 0.06)
        }
    }
}

/// Brand header at the top of the sidebar.
struct SidebarBrandHeader: View {
    var body: some View {
        HStack(spacing: 10) {
            DevEzBrandMark(size: 32)
            VStack(alignment: .leading, spacing: 1.5) {
                Text("DevEz")
                    .font(.system(size: 14.5, weight: .bold))
                Text("Local Laravel & Web Suite")
                    .font(.system(size: 11))
                    .foregroundStyle(.secondary)
            }
            Spacer(minLength: 0)
        }
    }
}

/// Floating privacy & tunnel card at the bottom of the sidebar.
struct SidebarStatusCard: View {
    @Environment(AppModel.self) private var model

    var body: some View {
        let count = model.liveShareCount
        HStack(spacing: 10) {
            StatusDot(color: count > 0 ? Color.live : Color.secondary.opacity(0.5), pulsing: count > 0, size: 8)
                .frame(width: 14)
            VStack(alignment: .leading, spacing: 1) {
                Text(count == 0 ? "Nothing exposed" : count == 1 ? "1 tunnel live" : "\(count) tunnels live")
                    .font(.system(size: 12, weight: .semibold))
                Text(count == 0 ? "Your sites are private" : "Live via Cloudflare")
                    .font(.system(size: 11))
                    .foregroundStyle(.secondary)
            }
            Spacer(minLength: 0)
            if count > 0 {
                Button {
                    Task { await model.stopAllShares() }
                } label: {
                    if model.isStoppingShares {
                        ProgressView().controlSize(.mini)
                    } else {
                        Image(systemName: "stop.fill")
                            .font(.system(size: 9, weight: .bold))
                            .frame(width: 22, height: 22)
                    }
                }
                .buttonStyle(PressableStyle())
                .foregroundStyle(.secondary)
                .disabled(model.isStoppingShares)
                .help("Stop all Cloudflare tunnels")
            }
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
        .surface(radius: 11)
        .animation(.easeOut(duration: 0.2), value: count)
    }
}

// MARK: - Status & Indicator Components

/// Status dot with optional pulsing outer ring for active states.
struct StatusDot: View {
    var color: Color = .green
    var pulsing: Bool = false
    var size: CGFloat = 8

    @Environment(\.accessibilityReduceMotion) private var reduceMotion
    @State private var expanded = false

    init(color: Color = .green, pulsing: Bool = false, size: CGFloat = 8) {
        self.color = color
        self.pulsing = pulsing
        self.size = size
    }

    init(on: Bool, pulsing: Bool = false, size: CGFloat = 8) {
        self.color = on ? Color.live : Color.secondary.opacity(0.45)
        self.pulsing = pulsing
        self.size = size
    }

    var body: some View {
        Circle()
            .fill(color)
            .frame(width: size, height: size)
            .background {
                if pulsing && !reduceMotion {
                    Circle()
                        .stroke(color.opacity(expanded ? 0 : 0.6), lineWidth: 1.5)
                        .scaleEffect(expanded ? 2.6 : 1)
                        .onAppear {
                            withAnimation(.easeOut(duration: 1.6).repeatForever(autoreverses: false)) { expanded = true }
                        }
                        .onDisappear { expanded = false }
                }
            }
    }
}

/// Compact rounded capsule badge, e.g. "1 live".
struct Pill: View {
    let text: String
    var dot: Color? = nil
    var tint: Color = .secondary

    var body: some View {
        HStack(spacing: 5) {
            if let dot { StatusDot(color: dot, size: 6) }
            Text(text)
        }
        .font(.system(size: 11.5, weight: .semibold))
        .foregroundStyle(tint)
        .padding(.horizontal, 9)
        .padding(.vertical, 4)
        .background(Capsule().fill((dot ?? tint).opacity(0.13)))
    }
}

/// Four-bar latency meter for Cloudflare / relay tunnels.
struct SignalBars: View {
    let latency: TimeInterval?
    var reachable: Bool?

    private var level: Int {
        guard let latency else { return 0 }
        switch latency {
        case ..<0.06: return 4
        case ..<0.12: return 3
        case ..<0.25: return 2
        default: return 1
        }
    }

    private var tint: Color {
        reachable == false ? Color.fault : level >= 3 ? Color.live : Color.pending
    }

    var body: some View {
        HStack(alignment: .bottom, spacing: 1.5) {
            ForEach(1...4, id: \.self) { bar in
                RoundedRectangle(cornerRadius: 1)
                    .fill(bar <= level ? tint : Color.primary.opacity(0.15))
                    .frame(width: 3, height: 3 + CGFloat(bar) * 2.2)
            }
        }
        .help(latency.map { "Relay latency: \(Int($0 * 1000)) ms" } ?? "Relay offline")
    }
}

// MARK: - Legacy Components & Helpers

/// A System Settings–style sidebar icon: white symbol on a colored rounded square.
struct SettingsIcon: View {
    let symbol: String
    let color: Color
    var size: CGFloat = 20

    var body: some View {
        Image(systemName: symbol)
            .font(.system(size: size * 0.55, weight: .semibold))
            .foregroundStyle(.white)
            .frame(width: size, height: size)
            .background(color.gradient, in: RoundedRectangle(cornerRadius: size * 0.24, style: .continuous))
    }
}

/// Title and short explanation at the top of a pane.
struct PaneHeader: View {
    let title: String
    let subtitle: String

    var body: some View {
        VStack(alignment: .leading, spacing: 3) {
            Text(title).font(.title2.weight(.semibold))
            Text(subtitle).font(.callout).foregroundStyle(.secondary)
                .fixedSize(horizontal: false, vertical: true)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.horizontal, 20)
        .padding(.top, 34)
        .padding(.bottom, 4)
    }
}

/// A monospaced, selectable block of text with a copy button (environment snippets).
struct CodeBlock: View {
    let text: String

    var body: some View {
        ZStack(alignment: .topTrailing) {
            Text(text.isEmpty ? "—" : text)
                .font(.system(.callout, design: .monospaced))
                .textSelection(.enabled)
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(10)
                .background(Color(nsColor: .textBackgroundColor).opacity(0.6), in: RoundedRectangle(cornerRadius: 6))
                .overlay(RoundedRectangle(cornerRadius: 6).strokeBorder(.separator))
            if !text.isEmpty {
                Button {
                    NSPasteboard.general.clearContents()
                    NSPasteboard.general.setString(text, forType: .string)
                } label: {
                    Image(systemName: "doc.on.doc")
                }
                .buttonStyle(.borderless)
                .help("Copy")
                .padding(8)
            }
        }
    }
}

/// A capsule label such as "Default".
struct Badge: View {
    let text: String
    var color: Color = .accentColor

    var body: some View {
        Text(text)
            .font(.caption2.weight(.semibold))
            .padding(.horizontal, 6)
            .padding(.vertical, 2)
            .foregroundStyle(color)
            .background(color.opacity(0.15), in: Capsule())
    }
}

/// Shows the last error of the Settings store under a pane.
struct SettingsError: View {
    let message: String?
    let dismiss: () -> Void

    var body: some View {
        if let message {
            HStack(alignment: .top) {
                Image(systemName: "exclamationmark.triangle.fill").foregroundStyle(.orange)
                Text(message).font(.callout).textSelection(.enabled)
                    .fixedSize(horizontal: false, vertical: true)
                Spacer()
                Button("Dismiss", action: dismiss).buttonStyle(.borderless)
            }
            .padding(10)
            .background(.orange.opacity(0.12), in: RoundedRectangle(cornerRadius: 8))
            .padding(.horizontal, 20)
            .padding(.bottom, 12)
        }
    }
}

/// Opens a file or folder with its default app.
@MainActor
func reveal(_ path: String) {
    NSWorkspace.shared.activateFileViewerSelecting([URL(fileURLWithPath: path)])
}
