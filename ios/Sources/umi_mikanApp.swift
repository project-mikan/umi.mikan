//
//  umi_mikanApp.swift
//  umi.mikan
//
//  Created by usuyuki on 2026/06/20.
//

import SwiftUI

@main
struct umi_mikanApp: App {
    /// 初期読み込み状態（スプラッシュ表示の制御）
    @State private var launchState = AppLaunchState()
    /// 認証状態（アプリライフサイクル全体で共有する）
    @State private var authViewModel = AuthViewModel()
    /// 同期マネージャー（アプリ起動直後から存在させ、Live Activity Intent を確実に受け取る）
    @State private var syncManager: SyncManager?

    var body: some Scene {
        WindowGroup {
            ZStack {
                Group {
                    if let syncManager {
                        ContentView(authViewModel: authViewModel, syncManager: syncManager, launchState: launchState)
                    }
                }
                .task {
                    // SyncManager はアプリ起動直後に作成し、Live Activity Intent の通知を受け取れるようにする
                    if syncManager == nil {
                        syncManager = SyncManager(authViewModel: authViewModel)
                    }
                }

                // スプラッシュは読み込み中インジケーター。読み込みが終わったら出さない
                if launchState.isInitialLoading {
                    SplashView()
                        .transition(.opacity)
                }
            }
            .animation(.easeInOut(duration: 0.3), value: launchState.isInitialLoading)
        }
    }
}
