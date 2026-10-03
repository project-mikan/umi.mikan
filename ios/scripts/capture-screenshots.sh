#!/usr/bin/env bash
# iPhone / iPhone Duo / iPad の各画面スクリーンショットを撮影して PNG を出力する。
# 使い方: ios/scripts/capture-screenshots.sh <出力ディレクトリ>
# 撮影本体は ios/Tests/ScreenshotTests.swift。環境変数 TEST_RUNNER_* はテストプロセスへ接頭辞を外して渡される。
#
# 高速化のため:
#   - シミュレータの起動はビルドと並行してバックグラウンドで行う
#   - 全端末への撮影を1回の xcodebuild（ビルド成果物は共有）で行う
# iPhone Duo のデバイスタイプがある Xcode ではそのシミュレータで撮影する（なければ作成する）。
# safe area などが実機と変わってしまうため近似はせず、デバイスタイプがない Xcode では iPhone Duo を撮らない。
# 任意の環境変数:
#   DERIVED_DATA_PATH : ビルド成果物の出力先（デフォルト: ios/DerivedData）
#   SPM_CACHE_DIR     : SPM 依存のクローン先（CI でキャッシュする場合に指定）
set -euo pipefail

cd "$(dirname "$0")/../.."

OUTPUT_DIR="$(mkdir -p "${1:?出力ディレクトリを指定してください}" && cd "$1" && pwd)"
DERIVED_DATA="${DERIVED_DATA_PATH:-ios/DerivedData}"
SIMULATORS=("iPhone 17" "iPad Pro 11-inch (M5)")

# 各フェーズの所要時間を出す（遅い箇所の切り分け用）
SECONDS=0
log() { echo "[$(printf '%3d' "$SECONDS")s] $*"; }

# シミュレータ名から UDID を引く（同名が複数ランタイムにある場合は最新のランタイムを使う）
find_udid() {
  xcrun simctl list devices available -j | python3 -c "
import json, sys
runtimes = json.load(sys.stdin)['devices']
for runtime in sorted(runtimes, reverse=True):
    for d in runtimes[runtime]:
        if d['name'] == sys.argv[1]:
            print(d['udid']); sys.exit()
" "$1"
}

# iPhone Duo のデバイスタイプ（名前に Duo を含む iPhone）の識別子を引く。なければ空
find_duo_device_type() {
  xcrun simctl list devicetypes -j | python3 -c "
import json, sys
for t in json.load(sys.stdin)['devicetypes']:
    if t['name'].startswith('iPhone') and 'duo' in t['name'].lower():
        print(t['name'] + '|' + t['identifier']); sys.exit()
"
}

DUO_TYPE="$(find_duo_device_type)"
if [ -n "$DUO_TYPE" ]; then
  duo_name="${DUO_TYPE%%|*}"
  # ランナーのイメージにシミュレータ本体が作られていない場合に備え、なければ最新ランタイムで作成する
  if [ -z "$(find_udid "$duo_name")" ]; then
    xcrun simctl create "$duo_name" "${DUO_TYPE#*|}" >/dev/null
  fi
  SIMULATORS+=("$duo_name")
else
  echo "⚠️ この Xcode には iPhone Duo のシミュレータがないため、iPhone Duo は撮影しません"
fi

UDIDS=()
DESTINATIONS=()
for name in "${SIMULATORS[@]}"; do
  udid="$(find_udid "$name")"
  if [ -z "$udid" ]; then
    echo "::error::シミュレータ ${name} が見つかりません" >&2
    exit 1
  fi
  UDIDS+=("$udid")
  DESTINATIONS+=(-destination "platform=iOS Simulator,id=${udid}")
  # 起動には1台あたり数十秒かかるため、ビルドと並行してバックグラウンドで起動しておく
  xcrun simctl boot "$udid" >/dev/null 2>&1 &
done
log "シミュレータの起動を開始しました"

BUILD_FLAGS=(
  -project ios/umi.mikan.xcodeproj
  -scheme umi.mikan
  -derivedDataPath "$DERIVED_DATA"
  -quiet
)
if [ -n "${SPM_CACHE_DIR:-}" ]; then
  BUILD_FLAGS+=(-clonedSourcePackagesDirPath "$SPM_CACHE_DIR" -onlyUsePackageVersionsFromResolvedFile)
fi

# 単体テストの workflow（301_ios_test.yml）と同じくビルド自体を軽くするフラグを付ける
xcodebuild build-for-testing "${BUILD_FLAGS[@]}" \
  -destination "platform=iOS Simulator,id=${UDIDS[0]}" \
  -jobs "$(sysctl -n hw.ncpu)" \
  ONLY_ACTIVE_ARCH=YES \
  CODE_SIGNING_ALLOWED=NO \
  COMPILER_INDEX_STORE_ENABLE=NO
log "ビルドが完了しました"

for udid in "${UDIDS[@]}"; do
  xcrun simctl bootstatus "$udid" >/dev/null
done
log "シミュレータの起動が完了しました"

# 複数のシミュレータで同時にテストを走らせると、非力なマシンでは準備中に kill されることがあったため
# 撮影は1台ずつ順番に行う（ビルドとシミュレータ起動は済んでいるので、1台あたりの追加コストは小さい）
TEST_RUNNER_SCREENSHOT_OUTPUT_DIR="$OUTPUT_DIR" \
  xcodebuild test-without-building "${BUILD_FLAGS[@]}" \
    "${DESTINATIONS[@]}" \
    -maximum-concurrent-test-simulator-destinations 1 \
    -only-testing:umi.mikanTests/ScreenshotTests \
    -parallel-testing-enabled NO
log "撮影が完了しました: ${OUTPUT_DIR}"
ls -1 "$OUTPUT_DIR"
