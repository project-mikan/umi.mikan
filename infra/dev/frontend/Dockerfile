FROM node:22.22.0-bookworm-slim
RUN npm install -g corepack@latest
RUN corepack enable
RUN pnpm config set store-dir ${PNPM_HOME}/store
# RUN npx playwright install-deps
WORKDIR /frontend

# Create user with UID 1000 to match host user
RUN useradd -u 1000 -g node -m nodeuser || true
RUN chown -R 1000:1000 /frontend

USER 1000:1000

# pnpm-workspace.yamlと.npmrcもCOPYしないと、ビルド時と実行時(バインドマウント)で
# インストール設定が不一致になりnode_modules再構築の対話プロンプトで起動が止まる
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml .npmrc ./
RUN pnpm install --frozen-lockfile

EXPOSE 5173

# Always run pnpm install on startup to handle missing node_modules
# CI=1でpnpmの対話プロンプトを無効化（tty:trueでも起動が止まらないように）
CMD ["sh", "-c", "CI=1 pnpm install --frozen-lockfile && exec pnpm dev --host 0.0.0.0"]
