import Connect
import Foundation

/// よびな（エンティティ）管理のViewModel
@MainActor
@Observable
final class EntitiesViewModel {
    /// カテゴリフィルタ
    enum CategoryFilter: CaseIterable {
        case all
        case people
        case noCategory

        /// 表示名
        var label: String {
            switch self {
            case .all:
                return "すべて"

            case .people:
                return "人物"

            case .noCategory:
                return "未分類"
            }
        }
    }

    var entities: [Entity_Entity] = []
    var filter: CategoryFilter = .all
    var isLoading: Bool = false
    var errorMessage: String?

    /// フィルタ適用後のエンティティ一覧
    var filteredEntities: [Entity_Entity] {
        switch filter {
        case .all:
            return entities

        case .people:
            return entities.filter { $0.category == .people }

        case .noCategory:
            return entities.filter { $0.category == .noCategory }
        }
    }

    private let authViewModel: AuthViewModel

    init(authViewModel: AuthViewModel) {
        self.authViewModel = authViewModel
    }

    /// よびな一覧を取得する
    func fetch() async {
        isLoading = entities.isEmpty
        errorMessage = nil

        let client = Entity_EntityServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Entity_ListEntitiesRequest()
        request.allCategories = true

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.listEntities(request: request, headers: ConnectClient.shared.headers())
        }

        if let error = response.error {
            if !APIHelper.isNetworkError(error) {
                errorMessage = APIHelper.errorMessage(error)
            }
            isLoading = false
            return
        }
        entities = response.message?.entities ?? []
        isLoading = false
    }

    /// よびなを新規作成する
    func create(name: String, category: Entity_EntityCategory, memo: String) async -> Bool {
        let client = Entity_EntityServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Entity_CreateEntityRequest()
        request.name = name
        request.category = category
        request.memo = memo

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.createEntity(request: request, headers: ConnectClient.shared.headers())
        }
        if let error = response.error {
            errorMessage = APIHelper.errorMessage(error)
            return false
        }
        await fetch()
        return true
    }

    /// よびなを更新する
    func update(id: String, name: String, category: Entity_EntityCategory, memo: String) async -> Bool {
        let client = Entity_EntityServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Entity_UpdateEntityRequest()
        request.id = id
        request.name = name
        request.category = category
        request.memo = memo

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.updateEntity(request: request, headers: ConnectClient.shared.headers())
        }
        if let error = response.error {
            errorMessage = APIHelper.errorMessage(error)
            return false
        }
        await fetch()
        return true
    }

    /// よびなを削除する
    func delete(id: String) async -> Bool {
        let client = Entity_EntityServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Entity_DeleteEntityRequest()
        request.id = id

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.deleteEntity(request: request, headers: ConnectClient.shared.headers())
        }
        if let error = response.error {
            errorMessage = APIHelper.errorMessage(error)
            return false
        }
        await fetch()
        return true
    }

    /// 他のよびかた（エイリアス）を追加する
    func addAlias(entityID: String, alias: String) async -> Bool {
        let client = Entity_EntityServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Entity_CreateEntityAliasRequest()
        request.entityID = entityID
        request.alias = alias

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.createEntityAlias(request: request, headers: ConnectClient.shared.headers())
        }
        if let error = response.error {
            errorMessage = APIHelper.errorMessage(error)
            return false
        }
        await fetch()
        return true
    }

    /// 他のよびかた（エイリアス）を削除する
    func deleteAlias(id: String) async -> Bool {
        let client = Entity_EntityServiceClient(client: ConnectClient.shared.protocolClient)
        var request = Entity_DeleteEntityAliasRequest()
        request.id = id

        let response = await APIHelper.withTokenRefresh(authViewModel) {
            await client.deleteEntityAlias(request: request, headers: ConnectClient.shared.headers())
        }
        if let error = response.error {
            errorMessage = APIHelper.errorMessage(error)
            return false
        }
        await fetch()
        return true
    }

    /// カテゴリの表示名を返す
    func categoryLabel(_ category: Entity_EntityCategory) -> String {
        switch category {
        case .people:
            return "人物"

        default:
            return "未分類"
        }
    }
}
