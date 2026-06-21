import SwiftUI

/// ログイン画面 - Liquid Glassデザイン対応
struct LoginView: View {
    @State private var email = ""
    @State private var password = ""
    @State private var showRegister = false
    @Bindable var viewModel: AuthViewModel
    
    var body: some View {
        NavigationStack {
            ZStack {
                // 背景のグラデーション
                LinearGradient(
                    colors: [.blue.opacity(0.3), .purple.opacity(0.3)],
                    startPoint: .topLeading,
                    endPoint: .bottomTrailing
                )
                .ignoresSafeArea()
                
                ScrollView {
                    VStack(spacing: 32) {
                        Spacer()
                            .frame(height: 60)
                        
                        // アプリタイトル - Liquid Glassエフェクト
                        Text("umi.mikan")
                            .font(.system(size: 48, weight: .bold, design: .rounded))
                            .foregroundStyle(.white)
                            .padding(.horizontal, 32)
                            .padding(.vertical, 20)
                            .glassEffect(.regular.tint(.blue).interactive())
                        
                        // ログインフォーム - GlassEffectContainer使用
                        GlassEffectContainer(spacing: 20.0) {
                            VStack(spacing: 20) {
                                // メールアドレス入力フィールド
                                VStack(alignment: .leading, spacing: 8) {
                                    Text("メールアドレス")
                                        .font(.caption)
                                        .foregroundStyle(.white.opacity(0.8))
                                        .padding(.leading, 4)
                                    
                                    TextField("mail@example.com", text: $email)
                                        .textContentType(.emailAddress)
                                        .keyboardType(.emailAddress)
                                        .autocapitalization(.none)
                                        .textInputAutocapitalization(.never)
                                        .padding()
                                        .background(.white.opacity(0.15))
                                        .clipShape(RoundedRectangle(cornerRadius: 12))
                                        .glassEffect(.regular.interactive(), in: .rect(cornerRadius: 12))
                                }
                                
                                // パスワード入力フィールド
                                VStack(alignment: .leading, spacing: 8) {
                                    Text("パスワード")
                                        .font(.caption)
                                        .foregroundStyle(.white.opacity(0.8))
                                        .padding(.leading, 4)
                                    
                                    SecureField("••••••••", text: $password)
                                        .textContentType(.password)
                                        .padding()
                                        .background(.white.opacity(0.15))
                                        .clipShape(RoundedRectangle(cornerRadius: 12))
                                        .glassEffect(.regular.interactive(), in: .rect(cornerRadius: 12))
                                }
                            }
                            .padding(.horizontal, 24)
                        }
                        
                        // エラーメッセージ
                        if let error = viewModel.errorMessage {
                            Text(error)
                                .foregroundStyle(.red)
                                .font(.callout)
                                .padding(.horizontal, 24)
                                .padding(.vertical, 12)
                                .background(.red.opacity(0.15))
                                .clipShape(RoundedRectangle(cornerRadius: 12))
                                .glassEffect(.regular.tint(.red), in: .rect(cornerRadius: 12))
                                .padding(.horizontal, 24)
                        }
                        
                        // ボタン群
                        VStack(spacing: 16) {
                            // ログインボタン
                            Button {
                                Task { 
                                    await viewModel.login(email: email, password: password) 
                                }
                            } label: {
                                Group {
                                    if viewModel.isLoading {
                                        ProgressView()
                                            .tint(.white)
                                    } else {
                                        Text("ログイン")
                                            .fontWeight(.semibold)
                                    }
                                }
                                .frame(maxWidth: .infinity)
                                .frame(height: 52)
                            }
                            .buttonStyle(.glassProminent)
                            .disabled(viewModel.isLoading || email.isEmpty || password.isEmpty)
                            
                            // 新規登録ボタン
                            Button {
                                showRegister = true
                            } label: {
                                Text("アカウントを作成")
                                    .fontWeight(.medium)
                                    .frame(maxWidth: .infinity)
                                    .frame(height: 48)
                            }
                            .buttonStyle(.glass)
                        }
                        .padding(.horizontal, 24)
                        
                        Spacer()
                            .frame(height: 60)
                    }
                }
            }
            .navigationDestination(isPresented: $showRegister) {
                RegisterView(viewModel: viewModel)
            }
        }
    }
}
#Preview {
    LoginView(viewModel: AuthViewModel())
}

