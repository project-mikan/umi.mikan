import SwiftUI

/// よびな詳細ページ - 編集・削除・他のよびかたの管理を行う
struct EntityDetailView: View {
    let viewModel: EntitiesViewModel

    @State private var entity: Entity_Entity
    @State private var showEditSheet = false
    @State private var showDeleteConfirm = false
    @State private var newAlias = ""
    @State private var isAddingAlias = false

    @Environment(\.dismiss)
    private var dismiss

    // swiftlint:disable:next type_contents_order
    init(viewModel: EntitiesViewModel, entity: Entity_Entity) {
        self.viewModel = viewModel
        _entity = State(initialValue: entity)
    }

    var body: some View {
        ScrollView {
            VStack(alignment: .leading, spacing: 16) {
                infoCard
                aliasCard
                deleteButton
            }
            .padding(16)
        }
        .navigationTitle(entity.name)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Button("編集") { showEditSheet = true }
                    .buttonStyle(.glass)
            }
        }
        .sheet(isPresented: $showEditSheet, onDismiss: reloadEntity) {
            EntityFormView(viewModel: viewModel, mode: .edit(entity))
        }
        .confirmationDialog("このよびなを削除しますか？", isPresented: $showDeleteConfirm, titleVisibility: .visible) {
            Button("削除する", role: .destructive) {
                Task {
                    if await viewModel.delete(id: entity.id) {
                        dismiss()
                    }
                }
            }
        }
    }

    // MARK: - コンポーネント

    private var infoCard: some View {
        VStack(alignment: .leading, spacing: 10) {
            LabeledContent("名前") {
                Text(entity.name).foregroundStyle(Color.twBody)
            }
            LabeledContent("カテゴリ") {
                Text(viewModel.categoryLabel(entity.category)).foregroundStyle(Color.twBody)
            }
            if !entity.memo.isEmpty {
                VStack(alignment: .leading, spacing: 4) {
                    Text("メモ")
                        .foregroundStyle(Color.twSecondary)
                    Text(entity.memo)
                        .foregroundStyle(Color.twBody)
                }
            }
        }
        .font(.subheadline)
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .glassEffect(.regular, in: .rect(cornerRadius: 14))
    }

    private var aliasCard: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("他のよびかた")
                .font(.subheadline)
                .fontWeight(.semibold)
                .foregroundStyle(Color.twHeading)
            aliasList
            aliasAddForm
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .glassEffect(.regular, in: .rect(cornerRadius: 14))
    }

    /// 他のよびかたの一覧
    @ViewBuilder private var aliasList: some View {
        if entity.aliases.isEmpty {
            Text("他のよびかたは登録されていません")
                .font(.caption)
                .foregroundStyle(Color.twSecondary)
        } else {
            ForEach(entity.aliases, id: \.id) { alias in
                aliasRow(alias)
            }
        }
    }

    /// 他のよびかたの追加フォーム
    private var aliasAddForm: some View {
        HStack(spacing: 8) {
            TextField("よびかたを追加", text: $newAlias)
                .textFieldStyle(.roundedBorder)
            Button {
                Task { await addAlias() }
            } label: {
                if isAddingAlias {
                    ProgressView().controlSize(.small)
                } else {
                    Image(systemName: "plus")
                }
            }
            .buttonStyle(.glassProminent)
            .disabled(newAlias.trimmingCharacters(in: .whitespaces).isEmpty || isAddingAlias)
        }
    }

    private var deleteButton: some View {
        Button(role: .destructive) {
            showDeleteConfirm = true
        } label: {
            Label("このよびなを削除", systemImage: "trash")
                .frame(maxWidth: .infinity)
                .frame(height: 44)
        }
        .buttonStyle(.glass)
        .tint(.red)
    }

    // MARK: - 操作

    /// 他のよびかたの1行（削除ボタン付き）
    private func aliasRow(_ alias: Entity_EntityAlias) -> some View {
        HStack {
            Text(alias.alias)
                .font(.subheadline)
                .foregroundStyle(Color.twBody)
            Spacer()
            Button {
                Task {
                    if await viewModel.deleteAlias(id: alias.id) {
                        reloadEntity()
                    }
                }
            } label: {
                Image(systemName: "trash")
                    .font(.caption)
                    .foregroundStyle(Color.twRed)
            }
            .buttonStyle(.plain)
        }
        .padding(.vertical, 2)
    }

    /// 他のよびかたを追加する
    private func addAlias() async {
        isAddingAlias = true
        defer { isAddingAlias = false }
        let alias = newAlias.trimmingCharacters(in: .whitespaces)
        if await viewModel.addAlias(entityID: entity.id, alias: alias) {
            newAlias = ""
            reloadEntity()
        }
    }

    /// 一覧の最新データからこのエンティティを再読込する
    private func reloadEntity() {
        if let updated = viewModel.entities.first(where: { $0.id == entity.id }) {
            entity = updated
        }
    }
}
