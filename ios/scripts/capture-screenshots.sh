#!/usr/bin/env bash
# 使い方: ios/scripts/capture-screenshots.sh <出力ディレクトリ> [iphone|ipad|iphone-duo ...]（任意: DERIVED_DATA_PATH, SPM_CACHE_DIR）
set -euo pipefail

cd "$(dirname "$0")/../.."

OUTPUT_DIR="$(mkdir -p "${1:?出力ディレクトリを指定してください}" && cd "$1" && pwd)"
shift
DEVICES=("${@:-iphone ipad iphone-duo}")
# 引数なしの場合に1要素の文字列を単語分割する
read -r -a DEVICES <<<"${DEVICES[*]}"
DERIVED_DATA="${DERIVED_DATA_PATH:-ios/DerivedData}"

SECONDS=0
log() { echo "[$(printf '%3d' "$SECONDS")s] $*"; }

# 同名のシミュレータが複数ランタイムにある場合は最新のランタイムのものを使う
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

find_duo_device_type() {
  xcrun simctl list devicetypes -j | python3 -c "
import json, sys
for t in json.load(sys.stdin)['devicetypes']:
    if t['name'].startswith('iPhone') and 'duo' in t['name'].lower():
        print(t['name'] + '|' + t['identifier']); sys.exit()
"
}

SIMULATORS=()
DUO_TYPE=""
DOWNLOAD_PID=""
for device in "${DEVICES[@]}"; do
  case "$device" in
    iphone) SIMULATORS+=("iPhone 17") ;;
    ipad) SIMULATORS+=("iPad Pro 11-inch (M5)") ;;
    iphone-duo)
      DUO_TYPE="$(find_duo_device_type)"
      if [ -z "$DUO_TYPE" ]; then
        echo "⚠️ この Xcode には iPhone Duo のシミュレータがないため撮影しません"
        touch "$OUTPUT_DIR/iphone-duo.unavailable"
      elif [ -z "$(find_udid "${DUO_TYPE%%|*}")" ] && ! xcrun simctl create "${DUO_TYPE%%|*}" "${DUO_TYPE#*|}" >/dev/null 2>&1; then
        # 対応ランタイムが未導入で作成できない。数GBあるのでビルドと並行してダウンロードする
        xcodebuild -downloadPlatform iOS >"$OUTPUT_DIR/.runtime-download.log" 2>&1 &
        DOWNLOAD_PID=$!
      fi
      ;;
    *) echo "::error::未知の端末 ${device}" >&2; exit 1 ;;
  esac
done

BOOT_PIDS=()
BOOT_NAMES=()
boot_in_background() {
  xcrun simctl bootstatus "$1" -b >/dev/null &
  BOOT_PIDS+=("$!")
  BOOT_NAMES+=("$2")
}

DESTINATIONS=()
for name in "${SIMULATORS[@]}"; do
  udid="$(find_udid "$name")"
  if [ -z "$udid" ]; then
    echo "::error::シミュレータ ${name} が見つかりません" >&2
    exit 1
  fi
  DESTINATIONS+=(-destination "platform=iOS Simulator,id=${udid}")
  # 起動に数十秒かかるのでビルドと並行させる
  boot_in_background "$udid" "$name"
done

BUILD_FLAGS=(
  -project ios/umi.mikan.xcodeproj
  -scheme umi.mikan
  -derivedDataPath "$DERIVED_DATA"
  -quiet
)
if [ -n "${SPM_CACHE_DIR:-}" ]; then
  BUILD_FLAGS+=(-clonedSourcePackagesDirPath "$SPM_CACHE_DIR" -onlyUsePackageVersionsFromResolvedFile)
fi

# 起動するシミュレータに依存しないよう generic な宛先でビルドする
xcodebuild build-for-testing "${BUILD_FLAGS[@]}" \
  -destination "generic/platform=iOS Simulator" \
  -jobs "$(sysctl -n hw.ncpu)" \
  ONLY_ACTIVE_ARCH=YES \
  ARCHS=arm64 \
  CODE_SIGNING_ALLOWED=NO \
  COMPILER_INDEX_STORE_ENABLE=NO \
  SWIFT_COMPILATION_MODE=wholemodule
log "ビルドが完了しました"

if [ -n "$DUO_TYPE" ]; then
  duo_name="${DUO_TYPE%%|*}"
  if [ -n "$DOWNLOAD_PID" ] && ! wait "$DOWNLOAD_PID"; then
    cat "$OUTPUT_DIR/.runtime-download.log" >&2
  fi
  rm -f "$OUTPUT_DIR/.runtime-download.log"
  log "iPhone Duo のランタイムの準備が完了しました"
  if [ -z "$(find_udid "$duo_name")" ] && ! xcrun simctl create "$duo_name" "${DUO_TYPE#*|}" >/dev/null; then
    echo "⚠️ ${duo_name} のシミュレータを作成できなかったため撮影しません"
    touch "$OUTPUT_DIR/iphone-duo.unavailable"
  else
    duo_udid="$(find_udid "$duo_name")"
    DESTINATIONS+=(-destination "platform=iOS Simulator,id=${duo_udid}")
    boot_in_background "$duo_udid" "$duo_name"
  fi
fi

if [ "${#DESTINATIONS[@]}" -eq 0 ]; then
  log "撮影する端末がありません"
  exit 0
fi

for i in "${!BOOT_PIDS[@]}"; do
  if ! wait "${BOOT_PIDS[$i]}"; then
    echo "::error::シミュレータ ${BOOT_NAMES[$i]} の起動に失敗しました" >&2
    exit 1
  fi
done
log "シミュレータの起動が完了しました"

# 複数台を同時に走らせると非力なマシンで準備中に kill されたため1台ずつ撮る
TEST_RUNNER_SCREENSHOT_OUTPUT_DIR="$OUTPUT_DIR" \
  xcodebuild test-without-building "${BUILD_FLAGS[@]}" \
    "${DESTINATIONS[@]}" \
    -maximum-concurrent-test-simulator-destinations 1 \
    -only-testing:umi.mikanTests/ScreenshotTests \
    -parallel-testing-enabled NO
log "撮影が完了しました: ${OUTPUT_DIR}"
ls -1 "$OUTPUT_DIR"
