# Build stage
FROM node:22-alpine AS builder

WORKDIR /frontend

# Install pnpm
RUN npm install -g pnpm@latest

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install dependencies
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

# Copy built application and node_modules from builder stage
COPY --from=builder /frontend/build ./build
COPY --from=builder /frontend/package.json ./package.json

# Cannot find package '@bufbuild/protobuf'になるので必要↓
RUN mkdir -p /frontend/node_modules/@bufbuild/protobuf
RUN mkdir -p /frontend/node_modules/@connectrpc/connect
RUN mkdir -p /frontend/node_modules/@connectrpc/connect-node
COPY --from=builder /frontend/node_modules/@bufbuild/protobuf/ ./node_modules/@bufbuild/protobuf/
COPY --from=builder /frontend/node_modules/@connectrpc/connect/ ./node_modules/@connectrpc/connect/
COPY --from=builder /frontend/node_modules/@connectrpc/connect-node/ ./node_modules/@connectrpc/connect-node/



# Change ownership of the app directory
RUN chown -R sveltekit:nodejs /frontend

USER sveltekit

EXPOSE 3000

CMD ["node", "build"]
