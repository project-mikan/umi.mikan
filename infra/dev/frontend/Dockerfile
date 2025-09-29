FROM node:22-bullseye-slim
RUN npm install -g corepack@latest
RUN corepack enable
RUN pnpm config set store-dir ${PNPM_HOME}/store
# RUN npx playwright install-deps
WORKDIR /frontend

# Create user with UID 1000 to match host user
RUN useradd -u 1000 -g node -m nodeuser || true
RUN chown -R 1000:1000 /frontend

USER 1000:1000

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

EXPOSE 5173

# Always run pnpm install on startup to handle missing node_modules
CMD ["sh", "-c", "pnpm install --frozen-lockfile && pnpm dev --host 0.0.0.0"]
