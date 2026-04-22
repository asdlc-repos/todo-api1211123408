# --- Build stage ---------------------------------------------------------
FROM node:20-alpine AS builder

WORKDIR /app

# Install dependencies first for better layer caching
COPY package.json package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

# Copy source and build
COPY tsconfig.json tsconfig.node.json vite.config.ts index.html ./
COPY src ./src
COPY public ./public

# VITE_API_BASE is read at build time; default to /api/v1 which the ingress
# should route to the todo-api service. Override with --build-arg VITE_API_BASE=...
ARG VITE_API_BASE=/api/v1
ENV VITE_API_BASE=${VITE_API_BASE}

RUN npm run build

# --- Runtime stage -------------------------------------------------------
FROM nginx:alpine

# Replace default server config
RUN rm -f /etc/nginx/conf.d/default.conf /etc/nginx/nginx.conf
COPY nginx.conf /etc/nginx/nginx.conf

# Copy built static assets
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 8080

# Sanity-check the nginx config at build time
RUN nginx -t

CMD ["nginx", "-g", "daemon off;"]
