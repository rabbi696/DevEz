import Foundation
import Network

/// Probes local TCP ports and measures network latency to endpoints.
public enum PortProbe: Sendable {
    /// Common local development ports to discover running servers (e.g. Vite, Next.js, Nuxt, Rails, Express, Docker).
    public static let commonPorts: [UInt16] = [
        3000, 3001, 5173, 8000, 8080, 8081, 4321, 3333, 4000, 5000, 8888, 9000
    ]

    /// Whether something accepts TCP connections on `host:port` (IPv4 or IPv6).
    public static func isListening(_ port: UInt16, host: String = "127.0.0.1", timeout: TimeInterval = 0.6) async -> Bool {
        await measure(host: host, port: port, timeout: timeout) != nil
    }

    /// TCP connect time to `host:port`, or nil if unreachable within `timeout`.
    public static func measure(host: String, port: UInt16, timeout: TimeInterval = 1.0) async -> TimeInterval? {
        guard let nwPort = NWEndpoint.Port(rawValue: port) else { return nil }
        let endpoint = NWEndpoint.hostPort(host: NWEndpoint.Host(host), port: nwPort)
        let connection = NWConnection(to: endpoint, using: .tcp)
        let queue = DispatchQueue(label: "dev.devez.probe")
        let started = Date()
        return await withCheckedContinuation { continuation in
            let once = Once(continuation)
            connection.stateUpdateHandler = { state in
                switch state {
                case .ready:
                    once.resume(Date().timeIntervalSince(started))
                case .waiting, .failed, .cancelled:
                    once.resume(nil)
                default:
                    return
                }
                connection.stateUpdateHandler = nil
                connection.cancel()
            }
            connection.start(queue: queue)
            queue.asyncAfter(deadline: .now() + timeout) {
                once.resume(nil)
                connection.cancel()
            }
        }
    }

    /// Resumes a continuation at most once; callers are serialized on one queue.
    private final class Once: @unchecked Sendable {
        private var continuation: CheckedContinuation<TimeInterval?, Never>?
        init(_ continuation: CheckedContinuation<TimeInterval?, Never>) { self.continuation = continuation }
        func resume(_ value: TimeInterval?) {
            continuation?.resume(returning: value)
            continuation = nil
        }
    }
}
