#!/usr/bin/env bash
# API ではコメントに画像を添付できないため、専用ブランチに置いて raw URL で参照する（要: GH_TOKEN, PR_NUMBER, HEAD_SHA）
set -euo pipefail

SCREENSHOT_DIR="$(cd "${1:?スクリーンショットのディレクトリを指定してください}" && pwd)"
BRANCH="ios-screenshots"
DEST="pr-${PR_NUMBER}/${HEAD_SHA}"
MARKER="<!-- ios-screenshots -->"

# ファイル名 <端末>_<画面>.png は ios/Tests/ScreenshotTests.swift で決めている
DEVICES=("iphone|iPhone" "iphone-duo|iPhone Duo" "ipad|iPad")
SCREENS=("home|ホーム" "monthly|月ごと" "search|検索" "detail|詳細モーダル" "settings|設定")

WORKTREE="$(mktemp -d)"
git config --global user.name "github-actions[bot]"
git config --global user.email "41898282+github-actions[bot]@users.noreply.github.com"
if git fetch --depth=1 origin "$BRANCH" 2>/dev/null; then
  git worktree add -B "$BRANCH" "$WORKTREE" FETCH_HEAD
else
  git worktree add --detach "$WORKTREE"
  git -C "$WORKTREE" checkout --orphan "$BRANCH"
  git -C "$WORKTREE" rm -rf --quiet .
fi

mkdir -p "$WORKTREE/$DEST"
shopt -s nullglob
for file in "$SCREENSHOT_DIR"/*.png "$SCREENSHOT_DIR"/*.unavailable; do
  cp "$file" "$WORKTREE/$DEST/"
done
git -C "$WORKTREE" add "$DEST"
# 同じ SHA の再実行で差分がないと commit が失敗するため
git -C "$WORKTREE" diff --cached --quiet || git -C "$WORKTREE" commit --quiet -m "iOS screenshots for #${PR_NUMBER} (${HEAD_SHA})"

# 他の PR のジョブと同時に push すると拒否されるためリトライする（パスが重ならないので競合しない）
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

# 端末ごとのジョブが並列に push するため、他のジョブの画像も含めた最新の状態で表を作る
git -C "$WORKTREE" fetch --depth=1 origin "$BRANCH"
git -C "$WORKTREE" reset --quiet --hard FETCH_HEAD
SHOTS="$WORKTREE/$DEST"

BASE_URL="https://raw.githubusercontent.com/${GITHUB_REPOSITORY}/${BRANCH}/${DEST}"
{
  echo "$MARKER"
  echo "## 📱 iOS スクリーンショット"
  echo
  echo "コミット ${HEAD_SHA:0:7} 時点の画面です（ダミーデータで描画しています）。"
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
      if [ -f "$SHOTS/$file" ]; then
        row+=" <img src=\"${BASE_URL}/${file}\" width=\"240\"> |"
      elif [ -f "$SHOTS/${device%%|*}.unavailable" ]; then
        row+=" シミュレータなし |"
      else
        row+=" 撮影中または失敗 |"
      fi
    done
    echo "$row"
  done
} > "$RUNNER_TEMP/ios-screenshot-comment.md"

comment_id="$(gh api --paginate "repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments" \
  --jq ".[] | select(.user.login == \"github-actions[bot]\" and (.body | startswith(\"${MARKER}\"))) | .id" | head -n1)"
if [ -n "$comment_id" ]; then
  gh api --method PATCH "repos/${GITHUB_REPOSITORY}/issues/comments/${comment_id}" -F body=@"$RUNNER_TEMP/ios-screenshot-comment.md" >/dev/null
else
  gh api --method POST "repos/${GITHUB_REPOSITORY}/issues/${PR_NUMBER}/comments" -F body=@"$RUNNER_TEMP/ios-screenshot-comment.md" >/dev/null
fi
echo "✅ PR #${PR_NUMBER} にスクリーンショットをコメントしました"
