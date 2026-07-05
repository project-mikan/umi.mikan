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
    /// スプラッシュの起動アニメーションが最後まで再生されたかどうか
    @State private var isSplashAnimationFinished = false
    /// 認証状態（アプリライフサイクル全体で共有する）
    @State private var authViewModel = AuthViewModel()
    /// 同期マネージャー（アプリ起動直後から存在させ、Live Activity Intent を確実に受け取る）
    @State private var syncManager: SyncManager

    // swiftlint:disable:next type_contents_order
    init() {
        let auth = AuthViewModel()
        _authViewModel = State(initialValue: auth)
        _syncManager = State(initialValue: SyncManager(authViewModel: auth))
    }

    var body: some Scene {
        WindowGroup {
            ZStack {
                ContentView(authViewModel: authViewModel, syncManager: syncManager, launchState: launchState)

                // スプラッシュは起動アニメーションを最後まで再生しつつ、初期読み込み完了まで表示する。
                // 読み込みが一瞬で終わってもアニメーションが途中で消えないよう両方の完了を待つ
                if launchState.isInitialLoading || !isSplashAnimationFinished {
                    SplashView { isSplashAnimationFinished = true }
                        // コンテナの animation(value:) は起動直後の変更で効かないことがあるため、
                        // transition 自体にアニメーションを付けて確実にフェードアウトさせる
                        .transition(.opacity.animation(.easeInOut(duration: 0.3)))
                }
            }
        }
    }
}
