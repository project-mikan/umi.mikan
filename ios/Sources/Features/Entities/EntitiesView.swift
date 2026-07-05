import SwiftUI

/// よびな管理ページ - 日記に登場するひとや場所のよびなを一覧・作成する
struct EntitiesView: View {
    @State private var viewModel: EntitiesViewModel
    @State private var showCreateSheet = false

    // swiftlint:disable:next type_contents_order
    init(authViewModel: AuthViewModel) {
        _viewModel = State(initialValue: EntitiesViewModel(authViewModel: authViewModel))
    }

    var body: some View {
        ScrollView {
            LazyVStack(alignment: .leading, spacing: 12) {
                filterButtons
                if viewModel.isLoading {
                    loadingView
                } else if viewModel.filteredEntities.isEmpty {
                    emptyView
                } else {
                    entityList
                }
            }
            .padding(16)
        }
        .task {
            await viewModel.fetch()
        }
        .refreshable {
            await viewModel.fetch()
        }
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Button {
                    showCreateSheet = true
                } label: {
                    Label("新規作成", systemImage: "plus")
                }
                .buttonStyle(.glass)
            }
        }
        .sheet(isPresented: $showCreateSheet) {
            EntityFormView(viewModel: viewModel, mode: .create)
        }
        .overlay(alignment: .bottom) {
            if let error = viewModel.errorMessage {
                errorBanner(message: error)
            }
        }
    }

    // MARK: - コンポーネント

    private var filterButtons: some View {
        HStack(spacing: 8) {
            ForEach(EntitiesViewModel.CategoryFilter.allCases, id: \.self) { filter in
                Button {
                    viewModel.filter = filter
                } label: {
                    Text(filter.label)
                        .font(.caption)
                        .fontWeight(.medium)
                        .padding(.horizontal, 14)
                        .padding(.vertical, 8)
                        .background(viewModel.filter == filter ? Color.twBlue : Color.gray.opacity(0.15))
                        .foregroundStyle(viewModel.filter == filter ? .white : Color.twBody)
                        .clipShape(Capsule())
                }
                .buttonStyle(.plain)
            }
            Spacer()
        }
    }

    private var entityList: some View {
        ForEach(viewModel.filteredEntities, id: \.id) { entity in
            NavigationLink {
                EntityDetailView(viewModel: viewModel, entity: entity)
            } label: {
                entityRow(entity)
            }
            .buttonStyle(.plain)
        }
    }

    private var loadingView: some View {
        VStack(spacing: 16) {
            ProgressView()
                .controlSize(.large)
            Text("読み込み中...")
                .font(.subheadline)
                .foregroundStyle(Color.twSecondary)
        }
        .frame(maxWidth: .infinity)
        .padding(.vertical, 60)
    }

    private var emptyView: some View {
        Text("よびなが登録されていません")
            .font(.subheadline)
            .foregroundStyle(Color.twSecondary)
            .frame(maxWidth: .infinity)
            .padding(.vertical, 40)
    }

    private func entityRow(_ entity: Entity_Entity) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            entityRowHeader(entity)
            entityAliasChips(entity)
            if !entity.memo.isEmpty {
                Text(entity.memo)
                    .font(.caption)
                    .foregroundStyle(Color.twSecondary)
                    .lineLimit(2)
            }
        }
        .padding(14)
        .frame(maxWidth: .infinity, alignment: .leading)
        .clipShape(RoundedRectangle(cornerRadius: 14))
        .glassEffect(.regular, in: .rect(cornerRadius: 14))
    }

    /// よびな行のヘッダー（名前・カテゴリバッジ）
    private func entityRowHeader(_ entity: Entity_Entity) -> some View {
        HStack {
            Text(entity.name)
                .font(.headline)
                .foregroundStyle(Color.twHeading)
            Text(viewModel.categoryLabel(entity.category))
                .font(.caption2)
                .padding(.horizontal, 8)
                .padding(.vertical, 3)
                .background(entity.category == .people ? Color.twBlue.opacity(0.15) : Color.gray.opacity(0.15))
                .foregroundStyle(entity.category == .people ? Color.twBlue : Color.twSecondary)
                .clipShape(Capsule())
            Spacer()
            Image(systemName: "chevron.right")
                .font(.caption)
                .foregroundStyle(Color.twSecondary)
        }
    }

    /// 他のよびかた（エイリアス）のチップ表示
    @ViewBuilder
    private func entityAliasChips(_ entity: Entity_Entity) -> some View {
        if !entity.aliases.isEmpty {
            HStack(spacing: 6) {
                ForEach(entity.aliases, id: \.id) { alias in
                    Text(alias.alias)
                        .font(.caption2)
                        .padding(.horizontal, 8)
                        .padding(.vertical, 3)
                        .background(Color.twIndigo.opacity(0.12))
                        .foregroundStyle(Color.twIndigo)
                        .clipShape(Capsule())
                }
            }
        }
    }

    private func errorBanner(message: String) -> some View {
        HStack(spacing: 8) {
            Image(systemName: "exclamationmark.circle.fill")
            Text(message)
                .font(.subheadline)
            Spacer()
            Button {
                viewModel.errorMessage = nil
            } label: {
                Image(systemName: "xmark")
                    .font(.caption)
            }
        }
        .padding(16)
        .background(.red.opacity(0.15))
        .foregroundStyle(Color.twRed)
        .clipShape(RoundedRectangle(cornerRadius: 12))
        .glassEffect(.regular.tint(.red), in: .rect(cornerRadius: 12))
        .padding(16)
    }
}
