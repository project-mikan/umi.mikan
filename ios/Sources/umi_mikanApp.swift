//
//  umi_mikanApp.swift
//  umi.mikan
//
//  Created by usuyuki on 2026/06/20.
//

import SwiftUI
import UserNotifications

@main
struct umi_mikanApp: App {
    /// 初期読み込み状態（スプラッシュ表示の制御）
    @State private var launchState = AppLaunchState()
    /// 認証状態（アプリライフサイクル全体で共有する）
    @State private var authViewModel = AuthViewModel()
    /// 同期マネージャー（アプリ起動直後から存在させ、Live Activity Intent を確実に受け取る）
    @State private var syncManager: SyncManager

    // swiftlint:disable:next type_contents_order
    init() {
        let auth = AuthViewModel()
        _authViewModel = State(initialValue: auth)
        _syncManager = State(initialValue: SyncManager(authViewModel: auth))

        // フォアグラウンド中も通知バナーを表示させる
        UNUserNotificationCenter.current().delegate = NotificationDelegate.shared
        // リネーム前（OnThisDayNotificationManager）が残した孤児通知を一度だけ削除する。
        // 放置すると新しい「おもいで」通知と二重に発火し続けるため起動時に必ず実行する。
        MemoryNotificationManager.shared.migrateLegacyNotification()
    }

    var body: some Scene {
        WindowGroup {
            ZStack {
                ContentView(authViewModel: authViewModel, syncManager: syncManager, launchState: launchState)

                // スプラッシュは読み込み中インジケーターとして扱い、初期読み込みが終わり次第すぐ消す。
                // ローカルデータの読み込みは通常一瞬で終わるため、アニメーションは途中でもスキップされる
                if launchState.isInitialLoading {
                    SplashView()
                        // コンテナの animation(value:) は起動直後の変更で効かないことがあるため、
                        // transition 自体にアニメーションを付けて確実にフェードアウトさせる
                        .transition(.opacity.animation(.easeInOut(duration: 0.3)))
                }
            }
        }
    }
}
