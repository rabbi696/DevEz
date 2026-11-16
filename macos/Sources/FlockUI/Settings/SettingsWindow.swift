import FlockKit
import SwiftUI

/// The panes of the Settings window, in sidebar order.
enum SettingsPane: String, CaseIterable, Identifiable {
    case sites, php, node, services, mail, dumps, debugger, ai, general, about

    var id: String { rawValue }

    var title: String {
        switch self {
        case .sites: "Sites"
        case .php: "PHP"
        case .node: "Node"
        case .services: "Services"
        case .mail: "Mail"
        case .dumps: "Dumps"
        case .debugger: "Debugger"
        case .ai: "AI"
        case .general: "General"
        case .about: "About"
        }
    }

    var symbol: String {
        switch self {
        case .sites: "globe"
        case .php: "chevron.left.forwardslash.chevron.right"
        case .node: "hexagon.fill"
        case .services: "server.rack"
        case .mail: "envelope.fill"
        case .dumps: "ladybug.fill"
        case .debugger: "ant.fill"
        case .ai: "sparkles"
        case .general: "gearshape.fill"
        case .about: "info"
        }
    }

    var color: Color {
        switch self {
        case .sites: Color.flockMark
        case .php: .indigo
        case .node: .green
        case .services: .orange
        case .mail: .cyan
        case .dumps: .pink
        case .debugger: .red
        case .ai: .purple
        case .general: .gray
        case .about: .secondary
        }
    }
}

/// Settings: a sidebar of panes and the selected pane on the right.
/// Modern macOS styling with brand header, categorized groups, and bottom privacy card.
struct SettingsWindow: View {
    @Environment(AppModel.self) private var model
    @Environment(\.openWindow) private var openWindow
    private var pane: SettingsPane {
        get { model.settingsPane }
        nonmutating set { model.settingsPane = newValue }
    }

    var body: some View {
        let store = model.settings

        HStack(spacing: 0) {
            VStack(spacing: 0) {
                // Top brand header
                SidebarBrandHeader()
                    .padding(.top, 40) // clear traffic lights
                    .padding(.horizontal, 16)
                    .padding(.bottom, 10)

                // Grouped navigation sidebar
                List(selection: Binding<SettingsPane?>(get: { pane }, set: { if let new = $0 { pane = new } })) {
                    Section {
                        sidebarRow(.sites, badge: (model.status?.sites.count ?? 0) + (model.status?.proxies.count ?? 0))
                    }
                    Section("Environment") {
                        sidebarRow(.php)
                        sidebarRow(.node)
                        sidebarRow(.services)
                        sidebarRow(.mail)
                        sidebarRow(.dumps)
                        sidebarRow(.debugger)
                        sidebarRow(.ai)
                    }
                    Section("Settings") {
                        sidebarRow(.general)
                        sidebarRow(.about)
                    }
                }
                .listStyle(.sidebar)
                .scrollContentBackground(.hidden)

                Spacer(minLength: 0)

                // Bottom privacy & tunnel status card
                SidebarStatusCard()
                    .padding(10)
            }
            .frame(width: 220)
            .background(Color(nsColor: .windowBackgroundColor))

            Divider()

            VStack(spacing: 0) {
                content
                    .frame(maxWidth: .infinity, maxHeight: .infinity, alignment: .topLeading)
                // The daemon's buttons here (Start, Restart) report through the model; "not running"
                // is already what the pane shows.
                SettingsError(message: (store.error == model.error ? nil : store.error) ?? model.actionError) {
                    store.error = nil
                    model.actionError = nil
                }
            }
            .background(Color(nsColor: .windowBackgroundColor))
        }
        .tint(Color.flockMark)
        .ignoresSafeArea(.container, edges: .top)
        .frame(minWidth: 850, minHeight: 520)
        .task {
            store.isVisible = true
            await store.refresh()
        }
        .onDisappear { store.isVisible = false }
    }

    private func sidebarRow(_ item: SettingsPane, badge: Int = 0) -> some View {
        Label {
            Text(item.title)
        } icon: {
            SettingsIcon(symbol: item.symbol, color: item.color)
        }
        .badge(badge > 0 ? "\(badge)" : "")
        .tag(item)
        .padding(.vertical, 2)
    }

    @ViewBuilder
    private var content: some View {
        if model.status == nil, pane != .about {
            if model.needsSetup {
                ContentUnavailableView {
                    Label { Text("DevEz isn't set up yet") } icon: { LogoMark(size: 44) }
                } description: {
                    Text("Set DevEz up to run your sites and change its settings.")
                } actions: {
                    Button("Set Up DevEz…") { openWindow.activating(WindowID.setup) }
                }
            } else {
                ContentUnavailableView {
                    Label { Text("DevEz is not running") } icon: { LogoMark(size: 44) }
                } description: {
                    Text(model.error ?? "Start the daemon to change its settings.")
                } actions: {
                    Button("Start DevEz") { Task { await model.restart() } }
                        .disabled(model.isRestarting)
                }
            }
        } else {
            switch pane {
            case .general: GeneralPane()
            case .sites: SitesPane()
            case .php: PHPPane()
            case .node: NodePane()
            case .services: ServicesPane()
            case .mail: MailPane()
            case .dumps: DumpsPane()
            case .debugger: DebuggerPane()
            case .ai: AIPane()
            case .about: AboutPane()
            }
        }
    }
}
