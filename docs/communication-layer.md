# Communication Layer

Documentation for the pluggable communication architecture that allows swapping between different protocols (REST, gRPC, WebSocket, Mock).

## Overview

The communication layer is designed to be **completely pluggable**, allowing you to swap between different protocols without changing business logic. This is ideal for a toy project where you want to experiment with different approaches.

```
Frontend UI
     ↓
Frontend Service Layer
     ↓
APIClient Protocol (Abstract Interface)
     ↓
┌────┴─────┬──────────┬─────────┐
│          │          │         │
REST     gRPC    WebSocket    Mock
```

---

## Design Principles

### 1. Protocol Independence

**Business logic never knows about the protocol:**

```
✅ Service Layer → APIClient Protocol → REST/gRPC/WebSocket
❌ Service Layer → HTTP directly
```

### 2. Interface-Based Abstraction

All implementations conform to the same interface:

```swift
protocol APIClient {
    func getTasks() async throws -> [Task]
    func createTask(_ task: TaskCreateRequest) async throws -> Task
    // ...
}
```

### 3. Configuration-Based Selection

Choose protocol via configuration, not code changes:

```swift
// Switch protocols by changing one line
let client = APIClientFactory.create(type: .rest)  // or .grpc, .mock
```

---

## Frontend Architecture

### Directory Structure

```
frontend/QuestTodo/
├── Services/
│   ├── APIClient.swift           # Protocol definition
│   ├── APIClientFactory.swift    # Factory for creating clients
│   ├── Models/
│   │   ├── APIResponse.swift     # Common response wrapper
│   │   ├── APIError.swift        # Error types
│   │   └── Requests.swift        # Request DTOs
│   ├── REST/
│   │   └── RESTClient.swift      # REST implementation
│   ├── gRPC/
│   │   ├── GRPCClient.swift      # gRPC implementation
│   │   └── Protos/               # Generated proto files
│   ├── WebSocket/
│   │   └── WebSocketClient.swift # WebSocket implementation
│   └── Mock/
│       └── MockClient.swift      # Mock for testing
└── Config/
    └── AppConfig.swift            # Configuration
```

---

## APIClient Protocol

### Complete Interface Definition

```swift
// Services/APIClient.swift

import Foundation

/// Abstract protocol for all communication implementations
protocol APIClient {
    // MARK: - Tasks

    /// List all tasks with optional filtering
    func getTasks(filter: TaskFilter?) async throws -> [Task]

    /// Get a single task by ID
    func getTask(id: String) async throws -> Task

    /// Create a new task
    func createTask(_ task: TaskCreateRequest) async throws -> Task

    /// Update an existing task
    func updateTask(id: String, _ task: TaskUpdateRequest) async throws -> Task

    /// Delete a task
    func deleteTask(id: String) async throws

    /// Update task status
    func updateTaskStatus(id: String, status: TaskStatus) async throws -> Task

    /// Bulk operations
    func createTasksBulk(_ tasks: [TaskCreateRequest]) async throws -> [Task]
    func deleteTasksBulk(ids: [String]) async throws

    /// Search tasks
    func searchTasks(query: String) async throws -> [Task]

    // MARK: - Objectives

    func addObjective(taskID: String, objective: ObjectiveCreateRequest) async throws -> Objective
    func updateObjective(taskID: String, objectiveID: String, completed: Bool) async throws -> Objective
    func deleteObjective(taskID: String, objectiveID: String) async throws

    // MARK: - Categories

    func getCategories() async throws -> [Category]
    func getCategory(id: String) async throws -> Category
    func createCategory(_ category: CategoryCreateRequest) async throws -> Category
    func updateCategory(id: String, _ category: CategoryUpdateRequest) async throws -> Category
    func deleteCategory(id: String) async throws

    // MARK: - Stats

    func getStats() async throws -> Stats
    func getDailyStats(from: Date, to: Date) async throws -> [DailyStat]
    func getCategoryStats() async throws -> [String: CategoryStat]

    // MARK: - History

    func getHistory(from: Date?, to: Date?) async throws -> [Task]

    // MARK: - Health

    func ping() async throws -> Bool
    func version() async throws -> String
}
```

### Common Models

```swift
// Services/Models/APIResponse.swift

struct APIResponse<T: Codable>: Codable {
    let success: Bool
    let data: T?
    let error: APIErrorDetail?
    let meta: MetaData?
}

struct APIErrorDetail: Codable {
    let code: String
    let message: String
    let fields: [String: String]?
}

struct MetaData: Codable {
    let total: Int?
    let limit: Int?
    let offset: Int?
    let hasMore: Bool?
}

// Services/Models/APIError.swift

enum APIError: Error {
    case networkError(Error)
    case invalidResponse
    case serverError(code: String, message: String)
    case notFound
    case validationError(fields: [String: String])
    case unauthorized

    static func from(_ detail: APIErrorDetail) -> APIError {
        switch detail.code {
        case "NOT_FOUND":
            return .notFound
        case "VALIDATION_ERROR":
            return .validationError(fields: detail.fields ?? [:])
        case "UNAUTHORIZED":
            return .unauthorized
        default:
            return .serverError(code: detail.code, message: detail.message)
        }
    }
}

// Services/Models/Requests.swift

struct TaskCreateRequest: Codable {
    let title: String
    let description: String?
    let priority: Int
    let deadline: Deadline?
    let category: String?
    let objectives: [ObjectiveCreateRequest]?
    let notes: String?
    let reward: Int?
    let tags: [String]?
}

struct TaskUpdateRequest: Codable {
    let title: String?
    let description: String?
    let priority: Int?
    let deadline: Deadline?
    let category: String?
    let notes: String?
    let reward: Int?
    let tags: [String]?
}

struct TaskFilter: Codable {
    let status: [TaskStatus]?
    let priority: [Int]?
    let categories: [String]?
    let tags: [String]?
    let deadlineType: String?
    let dateFrom: Date?
    let dateTo: Date?
    let includeCompleted: Bool?
    let sortBy: String?
    let sortOrder: String?
    let limit: Int?
    let offset: Int?

    func toQueryItems() -> [URLQueryItem] {
        var items: [URLQueryItem] = []

        if let status = status {
            items.append(URLQueryItem(name: "status", value: status.map(\.rawValue).joined(separator: ",")))
        }
        if let priority = priority {
            items.append(URLQueryItem(name: "priority", value: priority.map(String.init).joined(separator: ",")))
        }
        // ... other filters

        return items
    }
}
```

---

## REST Implementation

```swift
// Services/REST/RESTClient.swift

import Foundation

class RESTClient: APIClient {
    private let baseURL: URL
    private let session: URLSession
    private let decoder: JSONDecoder
    private let encoder: JSONEncoder

    init(baseURL: String) {
        self.baseURL = URL(string: baseURL)!

        let config = URLSessionConfiguration.default
        config.timeoutIntervalForRequest = 30
        config.timeoutIntervalForResource = 300
        self.session = URLSession(configuration: config)

        self.decoder = JSONDecoder()
        self.decoder.dateDecodingStrategy = .iso8601

        self.encoder = JSONEncoder()
        self.encoder.dateEncodingStrategy = .iso8601
    }

    // MARK: - Tasks

    func getTasks(filter: TaskFilter?) async throws -> [Task] {
        var components = URLComponents(
            url: baseURL.appendingPathComponent("/tasks"),
            resolvingAgainstBaseURL: true
        )!

        if let filter = filter {
            components.queryItems = filter.toQueryItems()
        }

        let response: APIResponse<[Task]> = try await request(url: components.url!)

        guard response.success, let tasks = response.data else {
            throw APIError.from(response.error!)
        }

        return tasks
    }

    func getTask(id: String) async throws -> Task {
        let url = baseURL.appendingPathComponent("/tasks/\(id)")
        let response: APIResponse<Task> = try await request(url: url)

        guard response.success, let task = response.data else {
            throw APIError.from(response.error!)
        }

        return task
    }

    func createTask(_ task: TaskCreateRequest) async throws -> Task {
        let url = baseURL.appendingPathComponent("/tasks")
        let response: APIResponse<Task> = try await request(
            url: url,
            method: "POST",
            body: task
        )

        guard response.success, let createdTask = response.data else {
            throw APIError.from(response.error!)
        }

        return createdTask
    }

    func updateTask(id: String, _ task: TaskUpdateRequest) async throws -> Task {
        let url = baseURL.appendingPathComponent("/tasks/\(id)")
        let response: APIResponse<Task> = try await request(
            url: url,
            method: "PUT",
            body: task
        )

        guard response.success, let updatedTask = response.data else {
            throw APIError.from(response.error!)
        }

        return updatedTask
    }

    func deleteTask(id: String) async throws {
        let url = baseURL.appendingPathComponent("/tasks/\(id)")
        let response: APIResponse<EmptyResponse> = try await request(
            url: url,
            method: "DELETE"
        )

        guard response.success else {
            throw APIError.from(response.error!)
        }
    }

    func updateTaskStatus(id: String, status: TaskStatus) async throws -> Task {
        let url = baseURL.appendingPathComponent("/tasks/\(id)/status")
        let response: APIResponse<Task> = try await request(
            url: url,
            method: "PATCH",
            body: ["status": status.rawValue]
        )

        guard response.success, let task = response.data else {
            throw APIError.from(response.error!)
        }

        return task
    }

    func searchTasks(query: String) async throws -> [Task] {
        var components = URLComponents(
            url: baseURL.appendingPathComponent("/tasks/search"),
            resolvingAgainstBaseURL: true
        )!
        components.queryItems = [URLQueryItem(name: "q", value: query)]

        let response: APIResponse<[Task]> = try await request(url: components.url!)

        guard response.success, let tasks = response.data else {
            throw APIError.from(response.error!)
        }

        return tasks
    }

    // MARK: - Categories

    func getCategories() async throws -> [Category] {
        let url = baseURL.appendingPathComponent("/categories")
        let response: APIResponse<[Category]> = try await request(url: url)

        guard response.success, let categories = response.data else {
            throw APIError.from(response.error!)
        }

        return categories
    }

    func createCategory(_ category: CategoryCreateRequest) async throws -> Category {
        let url = baseURL.appendingPathComponent("/categories")
        let response: APIResponse<Category> = try await request(
            url: url,
            method: "POST",
            body: category
        )

        guard response.success, let created = response.data else {
            throw APIError.from(response.error!)
        }

        return created
    }

    // MARK: - Stats

    func getStats() async throws -> Stats {
        let url = baseURL.appendingPathComponent("/stats")
        let response: APIResponse<Stats> = try await request(url: url)

        guard response.success, let stats = response.data else {
            throw APIError.from(response.error!)
        }

        return stats
    }

    func getDailyStats(from: Date, to: Date) async throws -> [DailyStat] {
        var components = URLComponents(
            url: baseURL.appendingPathComponent("/stats/daily"),
            resolvingAgainstBaseURL: true
        )!

        let formatter = ISO8601DateFormatter()
        components.queryItems = [
            URLQueryItem(name: "from", value: formatter.string(from: from)),
            URLQueryItem(name: "to", value: formatter.string(from: to))
        ]

        let response: APIResponse<[DailyStat]> = try await request(url: components.url!)

        guard response.success, let stats = response.data else {
            throw APIError.from(response.error!)
        }

        return stats
    }

    // MARK: - Health

    func ping() async throws -> Bool {
        let url = baseURL.appendingPathComponent("/health")
        do {
            let _: APIResponse<HealthResponse> = try await request(url: url)
            return true
        } catch {
            return false
        }
    }

    func version() async throws -> String {
        let url = baseURL.appendingPathComponent("/version")
        let response: APIResponse<VersionResponse> = try await request(url: url)

        guard response.success, let version = response.data else {
            throw APIError.from(response.error!)
        }

        return version.version
    }

    // MARK: - Generic Request Helper

    private func request<T: Decodable>(
        url: URL,
        method: String = "GET",
        body: (any Encodable)? = nil
    ) async throws -> T {
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("application/json", forHTTPHeaderField: "Accept")

        if let body = body {
            request.httpBody = try encoder.encode(body)
        }

        do {
            let (data, response) = try await session.data(for: request)

            guard let httpResponse = response as? HTTPURLResponse else {
                throw APIError.invalidResponse
            }

            guard (200...299).contains(httpResponse.statusCode) else {
                // Try to decode error response
                if let errorResponse = try? decoder.decode(APIResponse<EmptyResponse>.self, from: data),
                   let error = errorResponse.error {
                    throw APIError.from(error)
                }
                throw APIError.serverError(
                    code: "HTTP_\(httpResponse.statusCode)",
                    message: "Request failed with status \(httpResponse.statusCode)"
                )
            }

            return try decoder.decode(T.self, from: data)
        } catch let error as APIError {
            throw error
        } catch {
            throw APIError.networkError(error)
        }
    }
}

// Helper types
private struct EmptyResponse: Codable {}
private struct HealthResponse: Codable {
    let status: String
}
private struct VersionResponse: Codable {
    let version: String
}
```

---

## Mock Implementation

Perfect for UI development and testing without a backend.

```swift
// Services/Mock/MockClient.swift

import Foundation

class MockClient: APIClient {
    // In-memory storage
    private var tasks: [Task] = []
    private var categories: [Category] = []
    private var nextTaskID = 1
    private var nextCategoryID = 1

    // Simulate network delay
    private let simulateDelay: Bool
    private let delayDuration: UInt64 = 200_000_000 // 200ms

    init(simulateDelay: Bool = true, seedData: Bool = true) {
        self.simulateDelay = simulateDelay

        if seedData {
            loadSeedData()
        }
    }

    // MARK: - Tasks

    func getTasks(filter: TaskFilter?) async throws -> [Task] {
        await delay()

        var filtered = tasks

        if let filter = filter {
            // Apply status filter
            if let statuses = filter.status, !statuses.isEmpty {
                filtered = filtered.filter { statuses.contains($0.status) }
            }

            // Apply priority filter
            if let priorities = filter.priority, !priorities.isEmpty {
                filtered = filtered.filter { priorities.contains($0.priority) }
            }

            // Apply category filter
            if let categories = filter.categories, !categories.isEmpty {
                filtered = filtered.filter { task in
                    guard let category = task.category else { return false }
                    return categories.contains(category)
                }
            }

            // Apply sorting
            if let sortBy = filter.sortBy {
                filtered.sort { lhs, rhs in
                    let ascending = filter.sortOrder != "desc"

                    switch sortBy {
                    case "priority":
                        return ascending ? lhs.priority < rhs.priority : lhs.priority > rhs.priority
                    case "created_at":
                        return ascending ? lhs.createdAt < rhs.createdAt : lhs.createdAt > rhs.createdAt
                    case "deadline":
                        let lhsDate = lhs.deadline.date ?? Date.distantFuture
                        let rhsDate = rhs.deadline.date ?? Date.distantFuture
                        return ascending ? lhsDate < rhsDate : lhsDate > rhsDate
                    default:
                        return true
                    }
                }
            }

            // Apply pagination
            if let limit = filter.limit {
                let offset = filter.offset ?? 0
                let start = min(offset, filtered.count)
                let end = min(offset + limit, filtered.count)
                filtered = Array(filtered[start..<end])
            }
        }

        return filtered
    }

    func getTask(id: String) async throws -> Task {
        await delay()

        guard let task = tasks.first(where: { $0.id == id }) else {
            throw APIError.notFound
        }

        return task
    }

    func createTask(_ request: TaskCreateRequest) async throws -> Task {
        await delay()

        let task = Task(
            id: "task-\(nextTaskID)",
            title: request.title,
            description: request.description,
            priority: request.priority,
            deadline: request.deadline ?? Deadline(type: "none", date: nil),
            category: request.category,
            status: .active,
            objectives: request.objectives?.map { obj in
                Objective(
                    id: UUID().uuidString,
                    text: obj.text,
                    completed: false,
                    order: obj.order ?? 0,
                    createdAt: Date()
                )
            } ?? [],
            notes: request.notes,
            reward: request.reward ?? 0,
            tags: request.tags ?? [],
            order: tasks.count,
            createdAt: Date(),
            updatedAt: Date(),
            completedAt: nil,
            progress: 0,
            isOverdue: false,
            daysLeft: nil
        )

        nextTaskID += 1
        tasks.append(task)

        return task
    }

    func updateTask(id: String, _ request: TaskUpdateRequest) async throws -> Task {
        await delay()

        guard let index = tasks.firstIndex(where: { $0.id == id }) else {
            throw APIError.notFound
        }

        var task = tasks[index]

        if let title = request.title { task.title = title }
        if let description = request.description { task.description = description }
        if let priority = request.priority { task.priority = priority }
        if let deadline = request.deadline { task.deadline = deadline }
        if let category = request.category { task.category = category }
        if let notes = request.notes { task.notes = notes }
        if let reward = request.reward { task.reward = reward }
        if let tags = request.tags { task.tags = tags }

        task.updatedAt = Date()

        tasks[index] = task
        return task
    }

    func deleteTask(id: String) async throws {
        await delay()

        guard let index = tasks.firstIndex(where: { $0.id == id }) else {
            throw APIError.notFound
        }

        tasks.remove(at: index)
    }

    func updateTaskStatus(id: String, status: TaskStatus) async throws -> Task {
        await delay()

        guard let index = tasks.firstIndex(where: { $0.id == id }) else {
            throw APIError.notFound
        }

        var task = tasks[index]
        task.status = status
        task.updatedAt = Date()

        if status == .complete {
            task.completedAt = Date()
        }

        tasks[index] = task
        return task
    }

    func searchTasks(query: String) async throws -> [Task] {
        await delay()

        let lowercased = query.lowercased()
        return tasks.filter { task in
            task.title.lowercased().contains(lowercased) ||
            (task.description?.lowercased().contains(lowercased) ?? false)
        }
    }

    func createTasksBulk(_ requests: [TaskCreateRequest]) async throws -> [Task] {
        await delay()

        var created: [Task] = []
        for request in requests {
            let task = try await createTask(request)
            created.append(task)
        }
        return created
    }

    func deleteTasksBulk(ids: [String]) async throws {
        await delay()

        tasks.removeAll { ids.contains($0.id) }
    }

    // MARK: - Objectives

    func addObjective(taskID: String, objective: ObjectiveCreateRequest) async throws -> Objective {
        await delay()

        guard let index = tasks.firstIndex(where: { $0.id == taskID }) else {
            throw APIError.notFound
        }

        let obj = Objective(
            id: UUID().uuidString,
            text: objective.text,
            completed: false,
            order: objective.order ?? tasks[index].objectives.count,
            createdAt: Date()
        )

        tasks[index].objectives.append(obj)
        return obj
    }

    func updateObjective(taskID: String, objectiveID: String, completed: Bool) async throws -> Objective {
        await delay()

        guard let taskIndex = tasks.firstIndex(where: { $0.id == taskID }) else {
            throw APIError.notFound
        }

        guard let objIndex = tasks[taskIndex].objectives.firstIndex(where: { $0.id == objectiveID }) else {
            throw APIError.notFound
        }

        tasks[taskIndex].objectives[objIndex].completed = completed
        return tasks[taskIndex].objectives[objIndex]
    }

    func deleteObjective(taskID: String, objectiveID: String) async throws {
        await delay()

        guard let taskIndex = tasks.firstIndex(where: { $0.id == taskID }) else {
            throw APIError.notFound
        }

        tasks[taskIndex].objectives.removeAll { $0.id == objectiveID }
    }

    // MARK: - Categories

    func getCategories() async throws -> [Category] {
        await delay()
        return categories
    }

    func getCategory(id: String) async throws -> Category {
        await delay()

        guard let category = categories.first(where: { $0.id == id }) else {
            throw APIError.notFound
        }

        return category
    }

    func createCategory(_ request: CategoryCreateRequest) async throws -> Category {
        await delay()

        let category = Category(
            id: "cat-\(nextCategoryID)",
            name: request.name,
            color: request.color,
            icon: request.icon,
            type: request.type,
            order: categories.count,
            createdAt: Date()
        )

        nextCategoryID += 1
        categories.append(category)

        return category
    }

    func updateCategory(id: String, _ request: CategoryUpdateRequest) async throws -> Category {
        await delay()

        guard let index = categories.firstIndex(where: { $0.id == id }) else {
            throw APIError.notFound
        }

        var category = categories[index]

        if let name = request.name { category.name = name }
        if let color = request.color { category.color = color }
        if let icon = request.icon { category.icon = icon }

        categories[index] = category
        return category
    }

    func deleteCategory(id: String) async throws {
        await delay()

        guard let index = categories.firstIndex(where: { $0.id == id }) else {
            throw APIError.notFound
        }

        categories.remove(at: index)
    }

    // MARK: - Stats

    func getStats() async throws -> Stats {
        await delay()

        let total = tasks.count
        let active = tasks.filter { $0.status == .active || $0.status == .inProgress }.count
        let completed = tasks.filter { $0.status == .complete }.count
        let failed = tasks.filter { $0.status == .failed }.count

        return Stats(
            totalTasks: total,
            activeTasks: active,
            completedTasks: completed,
            failedTasks: failed,
            totalRewards: tasks.reduce(0) { $0 + $1.reward },
            completionRate: total > 0 ? Double(completed) / Double(total) : 0,
            averageTimeToComplete: 0, // TODO: Calculate
            streakDays: 0,
            categoryStats: [:],
            priorityStats: [:]
        )
    }

    func getDailyStats(from: Date, to: Date) async throws -> [DailyStat] {
        await delay()
        return []
    }

    func getCategoryStats() async throws -> [String : CategoryStat] {
        await delay()
        return [:]
    }

    // MARK: - History

    func getHistory(from: Date?, to: Date?) async throws -> [Task] {
        await delay()

        var completed = tasks.filter { $0.status == .complete }

        if let from = from {
            completed = completed.filter { $0.completedAt ?? Date.distantPast >= from }
        }

        if let to = to {
            completed = completed.filter { $0.completedAt ?? Date.distantFuture <= to }
        }

        return completed.sorted { ($0.completedAt ?? Date.distantPast) > ($1.completedAt ?? Date.distantPast) }
    }

    // MARK: - Health

    func ping() async throws -> Bool {
        await delay()
        return true
    }

    func version() async throws -> String {
        await delay()
        return "1.0.0-mock"
    }

    // MARK: - Helpers

    private func delay() async {
        if simulateDelay {
            try? await Task.sleep(nanoseconds: delayDuration)
        }
    }

    private func loadSeedData() {
        // Seed categories
        categories = [
            Category(id: "cat-1", name: "Work", color: "#3A7F8F", icon: "briefcase", type: "main", order: 0, createdAt: Date()),
            Category(id: "cat-2", name: "Personal", color: "#6BA573", icon: "person", type: "main", order: 1, createdAt: Date()),
            Category(id: "cat-3", name: "Side Projects", color: "#E8A958", icon: "lightbulb", type: "side", order: 2, createdAt: Date()),
        ]

        // Seed tasks
        tasks = [
            Task(
                id: "task-1",
                title: "Complete Project Proposal",
                description: "Draft and finalize Q1 project proposal for review",
                priority: 5,
                deadline: Deadline(type: "short", date: Calendar.current.date(byAdding: .day, value: 2, to: Date())),
                category: "cat-1",
                status: .active,
                objectives: [
                    Objective(id: "obj-1", text: "Research requirements", completed: true, order: 0, createdAt: Date()),
                    Objective(id: "obj-2", text: "Draft outline", completed: true, order: 1, createdAt: Date()),
                    Objective(id: "obj-3", text: "Write proposal", completed: false, order: 2, createdAt: Date()),
                ],
                notes: "Check with team lead before submitting",
                reward: 50,
                tags: ["important", "q1"],
                order: 0,
                createdAt: Date(),
                updatedAt: Date(),
                completedAt: nil,
                progress: 0.66,
                isOverdue: false,
                daysLeft: 2
            ),
            Task(
                id: "task-2",
                title: "Review Code Changes",
                description: "Review and merge pending pull requests",
                priority: 3,
                deadline: Deadline(type: "medium", date: Calendar.current.date(byAdding: .day, value: 5, to: Date())),
                category: "cat-1",
                status: .inProgress,
                objectives: [],
                notes: nil,
                reward: 20,
                tags: ["code-review"],
                order: 1,
                createdAt: Date(),
                updatedAt: Date(),
                completedAt: nil,
                progress: 0,
                isOverdue: false,
                daysLeft: 5
            ),
            Task(
                id: "task-3",
                title: "Learn SwiftUI Animations",
                description: "Study and practice SwiftUI animation techniques",
                priority: 2,
                deadline: Deadline(type: "long", date: Calendar.current.date(byAdding: .day, value: 14, to: Date())),
                category: "cat-3",
                status: .active,
                objectives: [],
                notes: "Focus on implicit and explicit animations",
                reward: 30,
                tags: ["learning", "swiftui"],
                order: 2,
                createdAt: Date(),
                updatedAt: Date(),
                completedAt: nil,
                progress: 0,
                isOverdue: false,
                daysLeft: 14
            ),
        ]

        nextTaskID = 4
        nextCategoryID = 4
    }
}

// Helper request models
struct ObjectiveCreateRequest: Codable {
    let text: String
    let order: Int?
}

struct CategoryCreateRequest: Codable {
    let name: String
    let color: String
    let icon: String?
    let type: String
}

struct CategoryUpdateRequest: Codable {
    let name: String?
    let color: String?
    let icon: String?
}
```

---

## Factory Pattern

```swift
// Services/APIClientFactory.swift

import Foundation

enum APIClientType: String {
    case rest
    case grpc
    case websocket
    case mock
}

class APIClientFactory {
    static func create(type: APIClientType? = nil) -> APIClient {
        let clientType = type ?? AppConfig.apiClientType

        switch clientType {
        case .rest:
            return RESTClient(baseURL: AppConfig.backendURL)
        case .grpc:
            return GRPCClient(host: AppConfig.grpcHost, port: AppConfig.grpcPort)
        case .websocket:
            return WebSocketClient(url: URL(string: AppConfig.websocketURL)!)
        case .mock:
            return MockClient(simulateDelay: true, seedData: true)
        }
    }
}
```

---

## Configuration

```swift
// Config/AppConfig.swift

import Foundation

struct AppConfig {
    // MARK: - API Client Configuration

    static let apiClientType: APIClientType = {
        #if DEBUG
        // Use mock client in previews and debug builds
        if ProcessInfo.processInfo.environment["XCODE_RUNNING_FOR_PREVIEWS"] == "1" {
            return .mock
        }
        return .rest  // Or .mock for UI-only development
        #else
        return .rest
        #endif
    }()

    // MARK: - REST Configuration

    static let backendURL = "http://localhost:3000/api/v1"

    // MARK: - gRPC Configuration

    static let grpcHost = "localhost"
    static let grpcPort = 50051

    // MARK: - WebSocket Configuration

    static let websocketURL = "ws://localhost:3000/ws"

    // MARK: - App Settings

    static let enableLogging = true
    static let logLevel: LogLevel = .debug
}

enum LogLevel {
    case debug
    case info
    case warning
    case error
}
```

---

## Usage in ViewModels

```swift
// ViewModels/TaskListViewModel.swift

import Foundation
import Combine

@MainActor
class TaskListViewModel: ObservableObject {
    @Published var tasks: [Task] = []
    @Published var isLoading = false
    @Published var error: APIError?

    private let client: APIClient
    private var cancellables = Set<AnyCancellable>()

    init(client: APIClient = APIClientFactory.create()) {
        self.client = client
    }

    func loadTasks(filter: TaskFilter? = nil) async {
        isLoading = true
        error = nil

        do {
            tasks = try await client.getTasks(filter: filter)
        } catch let error as APIError {
            self.error = error
        } catch {
            self.error = .networkError(error)
        }

        isLoading = false
    }

    func createTask(_ request: TaskCreateRequest) async {
        do {
            let task = try await client.createTask(request)
            tasks.append(task)
        } catch let error as APIError {
            self.error = error
        } catch {
            self.error = .networkError(error)
        }
    }

    func deleteTask(id: String) async {
        do {
            try await client.deleteTask(id: id)
            tasks.removeAll { $0.id == id }
        } catch let error as APIError {
            self.error = error
        } catch {
            self.error = .networkError(error)
        }
    }
}

// For SwiftUI Previews
extension TaskListViewModel {
    static var preview: TaskListViewModel {
        TaskListViewModel(client: MockClient(simulateDelay: false, seedData: true))
    }
}
```

---

## Usage in SwiftUI Views

```swift
// QuestTodoApp.swift

@main
struct QuestTodoApp: App {
    @StateObject private var taskViewModel: TaskListViewModel

    init() {
        // Create client based on configuration
        let client = APIClientFactory.create()
        _taskViewModel = StateObject(wrappedValue: TaskListViewModel(client: client))
    }

    var body: some Scene {
        WindowGroup {
            MainView()
                .environmentObject(taskViewModel)
                .task {
                    await taskViewModel.loadTasks()
                }
        }
    }
}

// Views/TaskListView.swift

struct TaskListView: View {
    @EnvironmentObject var viewModel: TaskListViewModel

    var body: some View {
        List(viewModel.tasks) { task in
            TaskRowView(task: task)
        }
        .overlay {
            if viewModel.isLoading {
                ProgressView()
            }
        }
        .alert(item: $viewModel.error) { error in
            Alert(
                title: Text("Error"),
                message: Text(error.localizedDescription)
            )
        }
    }
}

// SwiftUI Preview with Mock
struct TaskListView_Previews: PreviewProvider {
    static var previews: some View {
        TaskListView()
            .environmentObject(TaskListViewModel.preview)
            .frame(width: 400, height: 600)
    }
}
```

---

## Backend: Protocol-Agnostic Service

The backend service layer remains completely protocol-agnostic:

```go
// internal/service/task_service.go

type TaskService struct {
    store storage.Storage
}

// This doesn't know about HTTP, gRPC, or any protocol
func (s *TaskService) CreateTask(ctx context.Context, req *CreateTaskRequest) (*Task, error) {
    // Validation
    if err := req.Validate(); err != nil {
        return nil, err
    }

    // Business logic
    task := &Task{
        ID:       uuid.New().String(),
        Title:    req.Title,
        Priority: req.Priority,
        Status:   StatusActive,
        // ...
    }

    // Persist
    if err := s.store.CreateTask(ctx, task); err != nil {
        return nil, err
    }

    return task, nil
}
```

Different protocol handlers call the same service:

```go
// REST Handler
func (h *RESTHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
    var req service.CreateTaskRequest
    json.NewDecoder(r.Body).Decode(&req)

    task, err := h.taskService.CreateTask(r.Context(), &req)
    // Return JSON response
}

// gRPC Handler
func (h *GRPCHandler) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.Task, error) {
    serviceReq := protoToServiceRequest(req)

    task, err := h.taskService.CreateTask(ctx, serviceReq)
    // Return proto response
}

// WebSocket Handler
func (h *WSHandler) HandleCreateTask(conn *websocket.Conn, msg Message) {
    var req service.CreateTaskRequest
    json.Unmarshal(msg.Data, &req)

    task, err := h.taskService.CreateTask(context.Background(), &req)
    // Send WebSocket response
}
```

---

## Benefits of Pluggable Communication

✅ **Experiment easily** - Try REST vs gRPC vs WebSocket
✅ **Fast UI development** - Use mock client without backend
✅ **Easy testing** - Unit test with mock implementation
✅ **Learn multiple protocols** - Compare approaches in real project
✅ **Flexible deployment** - Run multiple protocols simultaneously
✅ **Type safety** - Compiler ensures all implementations match interface
✅ **Future-proof** - Add new protocols without changing business logic

---

## Implementation Priority

### Phase 1: REST + Mock (MVP)
1. ✅ Define `APIClient` protocol
2. ✅ Implement `RESTClient`
3. ✅ Implement `MockClient`
4. ✅ Create `APIClientFactory`
5. ✅ Use in ViewModels

### Phase 2: Backend REST API
1. Implement Go REST handlers
2. Test with frontend
3. Full integration

### Phase 3: Additional Protocols (Optional)
1. Add gRPC support
2. Add WebSocket support
3. Benchmark and compare

---

## Summary

The pluggable communication layer provides:

- **Clean abstraction** via `APIClient` protocol
- **Multiple implementations**: REST, gRPC, WebSocket, Mock
- **Easy swapping** via factory and configuration
- **Protocol-agnostic business logic** on both frontend and backend
- **Perfect for experimentation** in a toy project

Start with REST + Mock, then add other protocols as you learn and experiment!
