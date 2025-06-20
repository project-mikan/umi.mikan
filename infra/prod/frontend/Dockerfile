# Build stage
FROM node:22-alpine AS builder

WORKDIR /frontend

# Install pnpm
RUN npm install -g pnpm@latest

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install dependencies including @sveltejs/adapter-node
RUN pnpm add -D @sveltejs/adapter-node@^5.0.1
RUN pnpm install --frozen-lockfile

# Copy source code
COPY . .

# Use production config and build
RUN cp svelte.config.prod.js svelte.config.js
RUN pnpm build

# Production stage
FROM node:22-alpine AS production

WORKDIR /frontend

# Create non-root user
RUN addgroup -g 1001 -S nodejs
RUN adduser -S sveltekit -u 1001

# Copy built application and package.json from builder stage
COPY --from=builder /frontend/build ./build
COPY --from=builder /frontend/package.json ./package.json

# Change ownership of the app directory
RUN chown -R sveltekit:nodejs /frontend

USER sveltekit

EXPOSE 3000

ENV NODE_ENV=production
ENV BODY_SIZE_LIMIT=0

CMD ["node", "build"]
