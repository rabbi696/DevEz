import FlockKit
import SwiftUI

struct GeneralPane: View {
    @Environment(AppModel.self) private var model
    private let loginErrorState = State<String?>(initialValue: nil)
    private var loginError: String? {
        get { loginErrorState.wrappedValue }
        nonmutating set { loginErrorState.wrappedValue = newValue }
    }
    private let selectedParkedState = State<String?>(initialValue: nil)
    private var selectedParked: String? {
        get { selectedParkedState.wrappedValue }
        nonmutating set { selectedParkedState.wrappedValue = newValue }
    }
    /// The project folder waiting for the removal to be confirmed.
    private let confirmUnparkState = State<String?>(initialValue: nil)
    private var confirmUnpark: String? {
        get { confirmUnparkState.wrappedValue }
        nonmutating set { confirmUnparkState.wrappedValue = newValue }
    }
    private let launchAtLoginState = State<Bool>(initialValue: LoginItem.isEnabled)
    private var launchAtLogin: Bool {
        get { launchAtLoginState.wrappedValue }
        nonmutating set { launchAtLoginState.wrappedValue = newValue }
    }

    var body: some View {
        let store = model.settings

        ScrollView {
            VStack(alignment: .leading, spacing: 20) {
                // Large modern title & subtitle matching modern design pattern
                VStack(alignment: .leading, spacing: 4) {
                    Text("General")
                        .font(.system(size: 28, weight: .bold))
                    Text("How DevEz behaves on your Mac.")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }
                .padding(.top, 14)

                // MARK: Card 1 - Behavior & Startup
                VStack(spacing: 0) {
                    toggleRow(
                        title: "Open at login",
                        subtitle: "Starts quietly in the menu bar. Sites and services come online by themselves.",
                        isOn: Binding(get: { launchAtLogin }, set: { setLaunchAtLogin($0) })
                    )
                    if let loginError {
                        Text(loginError).font(.caption).foregroundStyle(.orange)
                            .padding(.horizontal, 16).padding(.bottom, 8)
                    }

                    Divider().padding(.horizontal, 14)

                    // Updates row
                    HStack(alignment: .center, spacing: 14) {
                        VStack(alignment: .leading, spacing: 2) {
                            Text("Check for updates")
                                .font(.system(size: 13, weight: .medium))
                            Text(updateState(model.updater))
                                .font(.system(size: 11.5))
                                .foregroundStyle(.secondary)
                        }
                        Spacer()
                        if model.updater.isEnabled {
                            Button(model.updater.availableVersion.map { "Update to \($0)" } ?? "Check Now") {
                                model.updater.checkForUpdates()
                            }
                            .buttonStyle(.bordered)
                            .controlSize(.small)
                            .disabled(model.updater.isChecking)
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)

                    if let config = store.config {
                        Divider().padding(.horizontal, 14)

                        toggleRow(
                            title: "Beta versions",
                            subtitle: "Get new versions of DevEz, PHP and services earlier, including pre-releases.",
                            isOn: Binding(
                                get: { config.channel == "beta" },
                                set: { on in Task { await store.perform { _ = try await $0.updateSettings(SettingsChange(channel: on ? "beta" : "stable")) } } }
                            ),
                            disabled: config.channel == "experimental"
                        )
                        if config.channel == "experimental" {
                            Label("Experimental channel: packages that aren't released yet are offered. `devez channel stable` switches back.", systemImage: "flask.fill")
                                .font(.caption).foregroundStyle(.orange)
                                .padding(.horizontal, 16).padding(.bottom, 8)
                        }
                    }
                }
                .surface(radius: 14)

                // MARK: Card 2 - Running Tunnels
                VStack(spacing: 0) {
                    HStack(spacing: 12) {
                        StatusDot(color: model.liveShareCount > 0 ? Color.live : Color.secondary.opacity(0.4), pulsing: model.liveShareCount > 0, size: 8)
                        VStack(alignment: .leading, spacing: 2) {
                            Text("Running tunnels")
                                .font(.system(size: 13, weight: .medium))
                            Text(model.liveShareCount == 0 ? "No active Cloudflare tunnels" : "\(model.liveShareCount) tunnel(s) live via Cloudflare")
                                .font(.system(size: 11.5))
                                .foregroundStyle(.secondary)
                        }
                        Spacer()
                        Button {
                            Task { await model.stopAllShares() }
                        } label: {
                            if model.isStoppingShares {
                                ProgressView().controlSize(.mini)
                            } else {
                                Text("Stop All")
                                    .font(.system(size: 12, weight: .semibold))
                                    .foregroundStyle(model.liveShareCount > 0 ? Color.flareDeep : Color.secondary)
                            }
                        }
                        .buttonStyle(PressableStyle(hoverFill: false))
                        .disabled(model.liveShareCount == 0 || model.isStoppingShares)
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)
                }
                .surface(radius: 14)

                // MARK: Card 3 - Daemon & Environment
                VStack(spacing: 0) {
                    if let status = model.status {
                        HStack(spacing: 12) {
                            VStack(alignment: .leading, spacing: 2) {
                                HStack(spacing: 6) {
                                    StatusDot(color: status.isDegraded ? .orange : Color.live)
                                    Text(status.isDegraded ? "Daemon needs attention" : "Daemon is running")
                                        .font(.system(size: 13, weight: .medium))
                                }
                                Text("Version \(status.version) · PID \(status.pid)")
                                    .font(.system(size: 11.5))
                                    .foregroundStyle(.secondary)
                            }
                            Spacer()
                            Button {
                                Task { await model.restart() }
                            } label: {
                                BusyLabel(title: "Restart Daemon", busyTitle: "Restarting…", isBusy: model.isRestarting)
                            }
                            .buttonStyle(.bordered)
                            .controlSize(.small)
                            .disabled(model.isRestarting)
                        }
                        .padding(.horizontal, 16)
                        .padding(.vertical, 12)
                    }

                    if let config = store.config {
                        Divider().padding(.horizontal, 14)

                        HStack(alignment: .top, spacing: 14) {
                            VStack(alignment: .leading, spacing: 2) {
                                Text("Domain & Network")
                                    .font(.system(size: 13, weight: .medium))
                                Text("*.\(config.tld)")
                                    .font(.system(size: 12, design: .monospaced))
                                    .foregroundStyle(.secondary)
                                Text(domainExplanation(config.tld))
                                    .font(.system(size: 11))
                                    .foregroundStyle(.secondary)
                                    .fixedSize(horizontal: false, vertical: true)
                            }
                            Spacer()
                        }
                        .padding(.horizontal, 16)
                        .padding(.vertical, 12)

                        Divider().padding(.horizontal, 14)

                        HStack(spacing: 12) {
                            VStack(alignment: .leading, spacing: 2) {
                                Text("Data Folder")
                                    .font(.system(size: 13, weight: .medium))
                                Text(config.root)
                                    .font(.system(size: 11.5, design: .monospaced))
                                    .foregroundStyle(.secondary)
                                    .lineLimit(1)
                                    .truncationMode(.middle)
                            }
                            Spacer()
                            Button("Show in Finder") { reveal(config.root) }
                                .buttonStyle(.bordered)
                                .controlSize(.small)
                        }
                        .padding(.horizontal, 16)
                        .padding(.vertical, 12)
                    }
                }
                .surface(radius: 14)

                // MARK: Card 4 - Project Folders (Parked)
                if let config = store.config {
                    VStack(alignment: .leading, spacing: 10) {
                        HStack {
                            VStack(alignment: .leading, spacing: 2) {
                                Text("Project folders")
                                    .font(.system(size: 13, weight: .medium))
                                Text("Every folder inside a project folder is served as <folder>.\(config.tld).")
                                    .font(.system(size: 11.5))
                                    .foregroundStyle(.secondary)
                            }
                            Spacer()
                            HStack(spacing: 4) {
                                Button {
                                    for url in SettingsStore.chooseDirectories(prompt: "Add", multiple: true) {
                                        Task { await store.perform { _ = try await $0.park(url.path) } }
                                    }
                                } label: { Image(systemName: "plus").frame(width: 20, height: 16) }
                                .buttonStyle(.bordered)
                                .controlSize(.small)

                                Button {
                                    confirmUnpark = selectedParked
                                } label: { Image(systemName: "minus").frame(width: 20, height: 16) }
                                .buttonStyle(.bordered)
                                .controlSize(.small)
                                .disabled(selectedParked == nil)
                            }
                        }

                        if config.parked.isEmpty {
                            Text("No project folders added yet")
                                .font(.system(size: 12))
                                .foregroundStyle(.secondary)
                                .frame(maxWidth: .infinity, alignment: .center)
                                .padding(.vertical, 14)
                        } else {
                            VStack(spacing: 4) {
                                ForEach(config.parked, id: \.self) { path in
                                    HStack {
                                        Image(systemName: "folder")
                                            .foregroundStyle(Color.flockMark)
                                        Text(path)
                                            .font(.system(size: 12, design: .monospaced))
                                            .lineLimit(1)
                                            .truncationMode(.middle)
                                        Spacer()
                                    }
                                    .padding(.horizontal, 10)
                                    .padding(.vertical, 6)
                                    .background(selectedParked == path ? Color.accentColor.opacity(0.12) : Color.clear)
                                    .clipShape(RoundedRectangle(cornerRadius: 6))
                                    .contentShape(Rectangle())
                                    .onTapGesture {
                                        selectedParked = (selectedParked == path ? nil : path)
                                    }
                                }
                            }
                        }
                    }
                    .padding(16)
                    .surface(radius: 14)
                }
            }
            .padding(.horizontal, 28)
            .padding(.top, 10)
            .padding(.bottom, 28)
            .frame(maxWidth: 720, alignment: .leading)
            .frame(maxWidth: .infinity)
        }
        .confirmationDialog("Remove \(confirmUnpark.map { URL(fileURLWithPath: $0).lastPathComponent } ?? "") from project folders?",
                            isPresented: Binding(get: { confirmUnpark != nil }, set: { if !$0 { confirmUnpark = nil } }),
                            presenting: confirmUnpark) { path in
            Button("Remove", role: .destructive) {
                if selectedParked == path {
                    selectedParked = nil
                }
                Task { await store.perform { _ = try await $0.unpark(path) } }
            }
            Button("Cancel", role: .cancel) {}
        } message: { path in
            Text("The sites in \(path) are no longer served. The folder and its projects are not touched.")
        }
    }

    private func toggleRow(title: String, subtitle: String, isOn: Binding<Bool>, disabled: Bool = false) -> some View {
        HStack(alignment: .center, spacing: 14) {
            VStack(alignment: .leading, spacing: 2) {
                Text(title)
                    .font(.system(size: 13, weight: .medium))
                Text(subtitle)
                    .font(.system(size: 11.5))
                    .foregroundStyle(.secondary)
            }
            Spacer()
            Toggle(title, isOn: isOn)
                .toggleStyle(.switch)
                .labelsHidden()
                .tint(Color.flockMark)
                .disabled(disabled)
        }
        .padding(.horizontal, 16)
        .padding(.vertical, 12)
    }

    private func updateState(_ updater: Updater) -> String {
        if !updater.isEnabled {
            return "Updates are off in development builds"
        }
        if updater.isChecking {
            return "Checking for updates…"
        }
        if let version = updater.availableVersion {
            return "\(version) is available"
        }
        return updater.upToDate ? "DevEz is up to date (v\(Updater.appVersion))" : "DevEz v\(Updater.appVersion)"
    }

    private func domainExplanation(_ tld: String) -> String {
        guard let dns = model.status?.dns else { return "" }
        if dns.wildcard {
            return "Every name, any subdomain included, is resolved by DevEz's DNS server (/etc/resolver/\(tld))."
        }
        let problem = dns.wildcardProblem.isEmpty ? "DevEz's DNS server isn't running" : dns.wildcardProblem
        return "*.\(tld) doesn't resolve: \(problem). macOS has no hosts-file fallback, so sites don't open until it does; `devez install` sets /etc/resolver/\(tld) up again."
    }

    private func setLaunchAtLogin(_ on: Bool) {
        loginError = LoginItem.set(on)
        launchAtLogin = LoginItem.isEnabled
    }
}
