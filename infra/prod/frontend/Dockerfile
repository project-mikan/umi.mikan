# Build stage
FROM node:22-alpine AS builder

WORKDIR /frontend

# pnpm 11の対話プロンプト（node_modules purge確認等）をTTYなし環境で無効化
ENV CI=true

# Install pnpm
RUN npm install -g pnpm@latest

# Copy package files first for better caching
# pnpm 11はallowBuilds等の設定をpnpm-workspace.yamlから読むため、COPYしないと
# ビルドスクリプト許可が効かずERR_PNPM_IGNORED_BUILDSでinstallが失敗する
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./

# Install dependencies (cached layer when package.json unchanged)
RUN --mount=type=cache,target=/root/.pnpm-store \
    pnpm install --frozen-lockfile

# Copy source code after dependency installation
COPY . .

# Build application with cache mounts
RUN --mount=type=cache,target=/root/.pnpm-store \
    --mount=type=cache,target=/frontend/.svelte-kit \
    --mount=type=cache,target=/tmp \
    cp svelte.config.prod.js svelte.config.js && \
    pnpm prepare && \
    pnpm build

# deps stage: package.jsonのdependenciesをnpmでインストール（シンボリックリンクなし）
# pnpmのnode_modulesはシンボリックリンク構造のためCOPYで解決できないため、npmを使用
FROM node:22-alpine AS deps

WORKDIR /deps

COPY package.json ./
RUN npm install --omit=dev --ignore-scripts

# Production stage
FROM node:22-alpine AS production

WORKDIR /frontend

RUN addgroup -g 1001 -S nodejs && adduser -S sveltekit -u 1001

# --chownでCOPY時に所有権を設定し、chown -R レイヤーのコピーコストを回避
COPY --from=builder --chown=sveltekit:nodejs /frontend/build ./build
COPY --from=deps --chown=sveltekit:nodejs /deps/node_modules ./node_modules

USER sveltekit

EXPOSE 3000

CMD ["node", "build"]
