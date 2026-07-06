import SwiftUI

/// 起動アニメーションビュー（読み込み中インジケーター）
/// みかんが上から落ちて転がり、棒とロゴにブラーをかけた後に完成形を表示する。
/// 親ビューは初期読み込みが完了し次第このビューを取り除く（アニメーションは途中でもスキップされる）。
struct SplashView: View {
    // MARK: - アニメーション状態

    /// みかんのY軸オフセット（落下位置）
    @State private var mikanOffsetY: CGFloat = -300
    /// みかんのX軸オフセット（転がり移動）
    @State private var mikanOffsetX: CGFloat = 0
    /// みかんの回転角度（転がり演出）
    @State private var mikanRotation: Double = -180
    /// みかんの縦方向スケール（着地の潰れ演出）
    @State private var mikanScaleY: CGFloat = 1.0
    /// みかんのブラー量（転がり中のモーションブラー）
    @State private var mikanBlur: CGFloat = 0
    /// 棒のブラー量
    @State private var penBlur: CGFloat = 0
    /// 背景とロゴ全体のフェードイン
    @State private var logoOpacity: Double = 0
    /// 完成形ロゴのスケール
    @State private var logoScale: CGFloat = 0.7
    /// 転がりフェーズかどうか（初期状態から転がりフェーズにして最初のフレームから絵を出す）
    @State private var isRolling: Bool = true
    /// 棒の上を転がる開始位置（penWidthに対する比率）
    @State private var rollProgress: CGFloat = -0.6

    // MARK: - レイアウト定数

    private let logoSize: CGFloat = 200
    private let mikanSize: CGFloat = 120
    private let penWidth: CGFloat = 160
    private let penHeight: CGFloat = 28

    var body: some View {
        ZStack {
            // 背景色（ロゴの背景と同じ青緑）
            Color(red: 0.169, green: 0.584, blue: 0.647)
                .ignoresSafeArea()

            if isRolling {
                rollingPhaseView
            } else {
                finalLogoView
            }
        }
        .onAppear {
            // コールドローンチ直後は初回描画が確定する前に withAnimation を呼んでも
            // アニメーションされず最終状態へジャンプするため、
            // 描画確定後（次のランループ）まで遅らせてから開始する
            DispatchQueue.main.async {
                startAnimation()
            }
        }
    }

    // MARK: - 転がりフェーズ

    /// 転がりフェーズ：みかんが落下してロゴの棒の上を転がる
    private var rollingPhaseView: some View {
        ZStack {
            rollingPenView
            rollingMikanView
        }
    }

    /// 転がり中の棒（ペン）
    private var rollingPenView: some View {
        RoundedRectangle(cornerRadius: 4)
            .fill(penGradient)
            .frame(width: penWidth, height: penHeight)
            .blur(radius: penBlur)
            .offset(y: 60)
    }

    /// 転がり中のみかん（円）
    private var rollingMikanView: some View {
        Circle()
            .fill(mikanGradient(radius: mikanSize / 2))
            .frame(width: mikanSize, height: mikanSize)
            .scaleEffect(x: 1.0, y: mikanScaleY)
            .rotationEffect(.degrees(mikanRotation))
            .blur(radius: mikanBlur)
            // X軸は棒上の転がり位置、Y軸は棒の直上
            .offset(x: mikanOffsetX, y: mikanOffsetY)
    }

    // MARK: - 完成形ロゴ

    /// 完成形：ロゴ全体をフェードインで表示
    private var finalLogoView: some View {
        ZStack {
            // みかん（楕円）
            Ellipse()
                .fill(mikanGradient(radius: logoSize * 0.38))
                .frame(width: logoSize * 0.61, height: logoSize * 0.60)
                .offset(y: -logoSize * 0.07)

            // 棒（ペン）
            RoundedRectangle(cornerRadius: 3)
                .fill(penGradient)
                .frame(width: logoSize * 0.76, height: logoSize * 0.145)
                .offset(y: logoSize * 0.235)
        }
        .frame(width: logoSize, height: logoSize)
        .opacity(logoOpacity)
        .scaleEffect(logoScale)
    }

    // MARK: - グラデーション

    private var penGradient: LinearGradient {
        LinearGradient(
            colors: [
                Color(red: 0.914, green: 0.878, blue: 0.714),
                Color(red: 0.914, green: 0.878, blue: 0.714).opacity(0)
            ],
            startPoint: .leading,
            endPoint: .trailing
        )
    }

    private func mikanGradient(radius: CGFloat) -> RadialGradient {
        RadialGradient(
            colors: [
                Color(red: 0.925, green: 0.412, blue: 0.188),
                Color(red: 0.925, green: 0.412, blue: 0.188).opacity(0)
            ],
            center: .center,
            startRadius: 0,
            endRadius: radius
        )
    }

    // MARK: - アニメーション制御

    /// 起動アニメーション全体のシーケンス
    private func startAnimation() {
        dropMikan()
    }

    /// フェーズ1: みかんが棒の左端に落下
    private func dropMikan() {
        withAnimation(.interpolatingSpring(stiffness: 100, damping: 10).delay(0.1)) {
            // 棒Y=60、みかん半径=60、棒高さ半分=14 → 棒の直上に乗る
            mikanOffsetY = 15
            mikanOffsetX = rollProgress * penWidth
        }
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) { squashMikan() }
    }

    /// フェーズ2: 着地の潰れ演出
    private func squashMikan() {
        withAnimation(.easeOut(duration: 0.12)) {
            mikanScaleY = 0.78
            mikanBlur = 2
            penBlur = 3
        }
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.12) { unsquashMikan() }
    }

    /// 潰れからの戻り
    private func unsquashMikan() {
        withAnimation(.easeIn(duration: 0.1)) {
            mikanScaleY = 1.0
            mikanBlur = 0
        }
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.18) { rollMikan() }
    }

    /// フェーズ3: 棒の上を右端に転がる（モーションブラー付き）
    private func rollMikan() {
        withAnimation(.easeInOut(duration: 0.6)) {
            mikanOffsetX = 0.55 * penWidth
            mikanRotation = 180
            mikanBlur = 6
            penBlur = 5
        }
        DispatchQueue.main.asyncAfter(deadline: .now() + 0.6) { showFinalLogo() }
    }

    /// フェーズ4: 完成ロゴをフェードインして表示
    private func showFinalLogo() {
        isRolling = false
        withAnimation(.easeOut(duration: 0.5)) {
            logoOpacity = 1.0
            logoScale = 1.0
        }
    }
}
