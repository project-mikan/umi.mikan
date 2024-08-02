FROM node:22-alpine AS builder
WORKDIR /frontned
COPY package*.json .
RUN npm ci
COPY . .
RUN pnpm build
RUN pnpm prune --production

FROM node:22-alpine
WORKDIR /frontned
COPY --from=builder /frontned/build build/
COPY --from=builder /frontned/node_modules node_modules/
COPY package.json .
EXPOSE 3000
ENV NODE_ENV=production
CMD [ "node", "build" ]
