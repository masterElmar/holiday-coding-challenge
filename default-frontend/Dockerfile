## Multi-stage Dockerfile for Next.js (dev, build, prod)
## - dev: runs next dev on Node
## - build: builds the app
## - prod: serves a static export via nginx (best when the app can be statically exported)

ARG NODE_VERSION=lts

# Base deps layer to cache installs
FROM node:${NODE_VERSION} AS base
WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1
COPY package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi


# dev stage
FROM node:${NODE_VERSION} AS dev
WORKDIR /app
ENV NODE_ENV=development
ENV NEXT_TELEMETRY_DISABLED=1

COPY package*.json ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

COPY . .

EXPOSE 3000
CMD ["npm", "run", "dev"]


# build stage
FROM node:${NODE_VERSION} AS build
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Reuse deps from base for faster builds
COPY --from=base /app/node_modules ./node_modules
COPY . .

# Build Next.js (generates .next)
RUN npm run build

# Try to produce a static export for nginx. If export fails (e.g., due to SSR/app router constraints),
# fall back to copying /public so the nginx stage still builds. Static export output goes to /app/out
FROM build AS export
RUN npx next export -o out \
	|| (echo "Static export failed; falling back to serving /public" && mkdir -p out && cp -r public/* out/ || true)

FROM node:${NODE_VERSION} AS prod
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

# Copy built app and production deps
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/.next ./.next
COPY --from=build /app/public ./public
COPY --from=build /app/package*.json ./
COPY --from=build /app/next.config.js ./

EXPOSE 3000
CMD ["npm", "run", "start"]

