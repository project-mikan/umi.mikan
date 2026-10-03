#!/usr/bin/env bash
# iOS スクリーンショットを ios-screenshots ブランチへ push し、PR に画像一覧をコメントする（既存コメントは上書き）。
# GitHub API ではコメントへ直接画像を添付できないため、公開リポジトリの専用ブランチに置いて raw URL で参照する。
# 使い方: .github/scripts/ios-screenshot-comment.sh <スクリーンショットのディレクトリ>
# 必要な環境変数: GH_TOKEN, GITHUB_REPOSITORY, PR_NUMBER, HEAD_SHA
set -euo pipefail

SCREENSHOT_DIR="$(cd "${1:?スクリーンショットのディレクトリを指定してください}" && pwd)"
BRANCH="ios-screenshots"
DEST="pr-${PR_NUMBER}/${HEAD_SHA}"
MARKER="<!-- ios-screenshots -->"

# 列（端末）と行（画面）の定義。ファイル名は <端末>_<画面>.png（ios/Tests/ScreenshotTests.swift）
DEVICES=("iphone|iPhone" "iphone-duo|iPhone Duo" "ipad|iPad")
# CI の Xcode に iPhone Duo のシミュレータがない場合は撮影されないため、列ごと出さない
NOTE=""
if ! ls "$SCREENSHOT_DIR"/iphone-duo_*.png >/dev/null 2>&1; then
  DEVICES=("iphone|iPhone" "ipad|iPad")
  NOTE="※ CI の Xcode に iPhone Duo のシミュレータがないため、iPhone Duo は撮影していません。"
fi
SCREENS=("home|ホーム" "monthly|月ごと" "search|検索" "detail|詳細モーダル" "settings|設定")

# --- 画像を専用ブランチへ push する ---
WORKTREE="$(mktemp -d)"
git config --global user.name "github-actions[bot]"
git config --global user.email "41898282+github-actions[bot]@users.noreply.github.com"
if git fetch --depth=1 origin "$BRANCH" 2>/dev/null; then
  git worktree add -B "$BRANCH" "$WORKTREE" FETCH_HEAD
else
  # 初回のみ、履歴を持たない専用ブランチを作る
  git worktree add --detach "$WORKTREE"
  git -C "$WORKTREE" checkout --orphan "$BRANCH"
  git -C "$WORKTREE" rm -rf --quiet .
fi

mkdir -p "$WORKTREE/$DEST"
cp "$SCREENSHOT_DIR"/*.png "$WORKTREE/$DEST/"
git -C "$WORKTREE" add "$DEST"
git -C "$WORKTREE" commit --quiet -m "iOS screenshots for #${PR_NUMBER} (${HEAD_SHA})"

# 他の PR のジョブと同時に push すると拒否されるため、取り込み直して数回リトライする（パスが重ならないので競合しない）
for attempt in 1 2 3 4 5; do
  if git -C "$WORKTREE" push origin "$BRANCH"; then
    break
  fi
  if [ "$attempt" -eq 5 ]; then
    echo "::error::${BRANCH} ブランチへの push に失敗しました" >&2
    exit 1
  fi
  sleep $((attempt * 2))
  git -C "$WORKTREE" fetch --depth=1 origin "$BRANCH"
  git -C "$WORKTREE" rebase FETCH_HEAD
done

# --- コメント本文を組み立てる ---
BASE_URL="https://raw.githubusercontent.com/${GITHUB_REPOSITORY}/${BRANCH}/${DEST}"
{
  echo "$MARKER"
  echo "## 📱 iOS スクリーンショット"
  echo
  echo "コミット ${HEAD_SHA:0:7} 時点の画面です（ダミーデータで描画しています）。"
  if [ -n "$NOTE" ]; then
    echo
    echo "$NOTE"
  fi
  echo
  header="| 画面 |"
  divider="| --- |"
  for device in "${DEVICES[@]}"; do
    header+=" ${device#*|} |"
    divider+=" --- |"
  done
  echo "$header"
  echo "$divider"
  for screen in "${SCREENS[@]}"; do
    row="| ${screen#*|} |"
    for device in "${DEVICES[@]}"; do
      file="${device%%|*}_${screen%%|*}.png"
      if [ -f "$SCREENSHOT_DIR/$file" ]; then
        row+=" <img src=\"${BASE_URL}/${file}\" width=\"240\"> |"
      else
        row+=" （撮影失敗） |"
      fi
    done
    echo "$row"
  done
} > "$RUNNER_TEMP/ios-screenshot-comment.md"

# --- 既存のコメントがあれば上書き、なければ新規作成する ---
comment_id="$(gh api --paginate "repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments" \
  --jq ".[] | select(.user.login == \"github-actions[bot]\" and (.body | startswith(\"${MARKER}\"))) | .id" | head -n1)"
if [ -n "$comment_id" ]; then
  gh api --method PATCH "repos/${GITHUB_REPOSITORY}/issues/comments/${comment_id}" -F body=@"$RUNNER_TEMP/ios-screenshot-comment.md" >/dev/null
else
  gh api --method POST "repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments" -F body=@"$RUNNER_TEMP/ios-screenshot-comment.md" >/dev/null
fi
echo "✅ PR #${PR_NUMBER} にスクリーンショットをコメントしました"
