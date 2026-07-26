# ==========================================
# Stage 1: Build static assets
# ==========================================
FROM node:20-alpine AS builder
WORKDIR /app

# Copy dependency definition and lock files
COPY package*.json ./

# Install project dependencies
RUN npm ci

# Copy full source and build static application
COPY . .
RUN npm run build:prod

# ==========================================
# Stage 2: Serve using Nginx Alpine
# ==========================================
FROM nginx:alpine

# Copy built application from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Replace standard Nginx configurations
COPY infra/nginx.conf /etc/nginx/conf.d/default.conf

# Setup startup script for dynamic environment injection
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 80
ENTRYPOINT ["/entrypoint.sh"]
