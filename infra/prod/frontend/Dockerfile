# Build stage
FROM node:22-alpine AS builder

WORKDIR /frontend

# Install pnpm
RUN npm install -g pnpm@latest

# Copy package files first for better caching
COPY package.json pnpm-lock.yaml ./

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

# Production stage
FROM node:22-alpine AS production

WORKDIR /frontend

# Create non-root user
RUN addgroup -g 1001 -S nodejs
RUN adduser -S sveltekit -u 1001

# Copy built application and node_modules from builder stage
COPY --from=builder /frontend/build ./build
COPY --from=builder /frontend/package.json ./package.json

# 本番環境で必要な依存関係をpackage.jsonに基づいてインストール
# pnpmのキャッシュマウントを使用しているため、node_modulesは永続化されない
# package.jsonとpnpm-lock.yamlを使用して一貫性のある依存関係管理を行う
RUN npm install -g pnpm@latest

# pnpm-lock.yamlもコピーしてロックファイルを使用
COPY --from=builder /frontend/pnpm-lock.yaml ./pnpm-lock.yaml

# 本番依存関係のみをインストール（devDependenciesを除外）
RUN pnpm install --prod --frozen-lockfile



# Change ownership of the app directory
RUN chown -R sveltekit:nodejs /frontend

USER sveltekit

EXPOSE 3000

CMD ["node", "build"]
