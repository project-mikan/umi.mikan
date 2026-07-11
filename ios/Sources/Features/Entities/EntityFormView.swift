import SwiftUI

/// よびなの作成・編集フォーム（シート表示）
struct EntityFormView: View {
    /// フォームのモード
    enum Mode {
        case create
        case edit(Entity_Entity)
    }

    let viewModel: EntitiesViewModel
    let mode: Mode

    @State private var name: String = ""
    @State private var category: Entity_EntityCategory = .noCategory
    @State private var memo: String = ""
    @State private var isSubmitting = false

    @Environment(\.dismiss)
    private var dismiss

    // swiftlint:disable:next type_contents_order
    init(viewModel: EntitiesViewModel, mode: Mode) {
        self.viewModel = viewModel
        self.mode = mode
        if case let .edit(entity) = mode {
            _name = State(initialValue: entity.name)
            _category = State(initialValue: entity.category)
            _memo = State(initialValue: entity.memo)
        }
    }

    var body: some View {
        NavigationStack {
            formContent
                .navigationTitle(isCreate ? "よびなを作成" : "よびなを編集")
                .navigationBarTitleDisplayMode(.inline)
                .toolbar {
                    ToolbarItem(placement: .cancellationAction) {
                        Button("キャンセル") { dismiss() }
                    }
                    ToolbarItem(placement: .confirmationAction) {
                        submitButton
                    }
                }
        }
    }

    private var formContent: some View {
        Form {
            Section("名前") {
                TextField("よびなを入力", text: $name)
            }

            Section("カテゴリ") {
                Picker("カテゴリ", selection: $category) {
                    Text("未分類").tag(Entity_EntityCategory.noCategory)
                    Text("人物").tag(Entity_EntityCategory.people)
                }
                .pickerStyle(.segmented)
            }

            Section("メモ") {
                TextField("メモ（任意）", text: $memo, axis: .vertical)
                    .lineLimit(3 ... 6)
            }
        }
    }

    private var submitButton: some View {
        Button {
            Task { await submit() }
        } label: {
            if isSubmitting {
                ProgressView().controlSize(.small)
            } else {
                Text("保存")
            }
        }
        .disabled(name.trimmingCharacters(in: .whitespaces).isEmpty || isSubmitting)
    }

    private var isCreate: Bool {
        if case .create = mode {
            return true
        }
        return false
    }

    /// フォーム内容を送信する
    private func submit() async {
        isSubmitting = true
        defer { isSubmitting = false }

        let trimmedName = name.trimmingCharacters(in: .whitespaces)
        let success: Bool = switch mode {
        case .create:
            await viewModel.create(name: trimmedName, category: category, memo: memo)

        case let .edit(entity):
            await viewModel.update(id: entity.id, name: trimmedName, category: category, memo: memo)
        }
        if success {
            dismiss()
        }
    }
}
