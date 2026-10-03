#!/usr/bin/env bash
# iPhone / iPhone Duo（近似） / iPad の各画面スクリーンショットを撮影して PNG を出力する。
# 使い方: ios/scripts/capture-screenshots.sh <出力ディレクトリ> [--skip-build]
#   --skip-build : build-for-testing 済み（CI で単体テストと成果物を共有する場合）ならビルドを省く
# 撮影本体は ios/Tests/ScreenshotTests.swift。環境変数 TEST_RUNNER_* はテストプロセスへ接頭辞を外して渡される。
set -euo pipefail

cd "$(dirname "$0")/../.."

OUTPUT_DIR="$(mkdir -p "${1:?出力ディレクトリを指定してください}" && cd "$1" && pwd)"
SKIP_BUILD="${2:-}"
PROJECT="ios/umi.mikan.xcodeproj"
SCHEME="umi.mikan"
DERIVED_DATA="${DERIVED_DATA_PATH:-ios/DerivedData}"

# 端末ラベルとシミュレータ名の対応。
# 折りたたみiPhone（iPhone Duo）はシミュレータが存在しないため、開いた状態に近い iPad mini で代用する
DEVICES=(
  "iphone|iPhone 17"
  "iphone-duo|iPad mini (A17 Pro)"
  "ipad|iPad Pro 11-inch (M5)"
)

COMMON_FLAGS=(
  -project "$PROJECT"
  -scheme "$SCHEME"
  -derivedDataPath "$DERIVED_DATA"
  -quiet
)

if [ "$SKIP_BUILD" != "--skip-build" ]; then
  xcodebuild build-for-testing "${COMMON_FLAGS[@]}" \
    -destination "platform=iOS Simulator,name=iPhone 17,OS=latest" \
    ONLY_ACTIVE_ARCH=YES CODE_SIGNING_ALLOWED=NO COMPILER_INDEX_STORE_ENABLE=NO
fi

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

# 起動には1台あたり数十秒かかるため、撮影前に全台をまとめてバックグラウンドで起動しておく
UDIDS=()
for device in "${DEVICES[@]}"; do
  name="${device#*|}"
  udid="$(find_udid "$name")"
  if [ -z "$udid" ]; then
    echo "::error::シミュレータ ${name} が見つかりません" >&2
    exit 1
  fi
  UDIDS+=("$udid")
  xcrun simctl boot "$udid" >/dev/null 2>&1 &
done

for i in "${!DEVICES[@]}"; do
  label="${DEVICES[$i]%%|*}"
  udid="${UDIDS[$i]}"
  echo "📸 ${label} (${DEVICES[$i]#*|})"
  # 起動済みでもエラーにならないよう、起動完了を待ってから撮影する
  xcrun simctl bootstatus "$udid" -b >/dev/null

  TEST_RUNNER_SCREENSHOT_OUTPUT_DIR="$OUTPUT_DIR" \
  TEST_RUNNER_SCREENSHOT_DEVICE="$label" \
    xcodebuild test-without-building "${COMMON_FLAGS[@]}" \
      -destination "platform=iOS Simulator,id=${udid}" \
      -only-testing:umi.mikanTests/ScreenshotTests \
      -parallel-testing-enabled NO
done

echo "✅ ${OUTPUT_DIR} に保存しました"
ls -1 "$OUTPUT_DIR"
