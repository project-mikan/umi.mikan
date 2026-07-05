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

    var body: some Scene {
        WindowGroup {
            ZStack {
                ContentView(launchState: launchState)

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
