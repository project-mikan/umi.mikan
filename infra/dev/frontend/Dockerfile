FROM node:22-bullseye-slim
RUN npm install -g corepack@latest
RUN corepack enable
RUN pnpm config set store-dir ${PNPM_HOME}/store
# RUN npx playwright install-deps
WORKDIR /frontend

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

EXPOSE 5173

CMD ["pnpm", "dev", "--host"]
