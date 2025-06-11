FROM node:22-bullseye-slim
RUN npm install -g corepack@latest
RUN corepack enable
RUN pnpm config set store-dir ${PNPM_HOME}/store
# RUN npx playwright install-deps
WORKDIR /frontend
