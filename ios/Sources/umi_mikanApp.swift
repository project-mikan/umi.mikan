//
//  umi_mikanApp.swift
//  umi.mikan
//
//  Created by usuyuki on 2026/06/20.
//

import SwiftUI

@main
struct umi_mikanApp: App {
    /// スプラッシュアニメーションの表示状態
    @State private var showSplash = true

    var body: some Scene {
        WindowGroup {
            if showSplash {
                SplashView {
                    withAnimation(.easeInOut(duration: 0.3)) {
                        showSplash = false
                    }
                }
            } else {
                ContentView()
            }
        }
    }
}
