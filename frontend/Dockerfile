#
# Frontend Dockerfile
# Builds the production-ready Vite bundle and serves it via Nginx.
#

# 1) Build the React/Vite application
FROM node:24-alpine AS builder

WORKDIR /app

# Install dependencies first (leveraging Docker layer caching)
COPY package.json package-lock.json ./
RUN npm ci

# Copy the remaining source files and build the app
COPY . .
RUN npm run build

# 2) Serve the compiled assets with Nginx
FROM nginx:alpine AS runner

WORKDIR /usr/share/nginx/html

# Copy custom Nginx config that supports client-side routing
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built assets from the builder stage
COPY --from=builder /app/dist ./

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
