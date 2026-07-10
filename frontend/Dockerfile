# ---- Build stage ----
FROM node:20-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY package.json package-lock.json* ./
RUN npm ci

# Copy source
COPY . .

# Allow the API base URL to be set at build time.
ARG VITE_API_BASE_URL=/api/v1
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

# Build the production bundle
RUN npm run build

# ---- Runtime stage ----
FROM nginx:alpine

# Copy the built static files to nginx's serve directory
COPY --from=builder /app/dist /usr/share/nginx/html

# Custom nginx config:
#  - Serves the SPA with fallback to index.html for client-side routing
#  - Proxies /api requests to the backend container so the frontend works
#    inside Docker Compose without CORS configuration.
RUN echo $'server { \n\
    listen 80; \n\
    \n\
    location / { \n\
        root /usr/share/nginx/html; \n\
        index index.html; \n\
        try_files $uri $uri/ /index.html; \n\
    } \n\
    \n\
    location /api/ { \n\
        proxy_pass http://backend:8080; \n\
        proxy_set_header Host $host; \n\
        proxy_set_header X-Real-IP $remote_addr; \n\
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; \n\
        proxy_set_header X-Forwarded-Proto $scheme; \n\
    } \n\
    \n\
    location /docs/ { \n\
        proxy_pass http://backend:8080; \n\
        proxy_set_header Host $host; \n\
    } \n\
}' > /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
