# Stage 1: Build frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /frontend
COPY frontend/ .
RUN npm ci
RUN npm run build

# Stage 2: Build backend
FROM golang:1.21-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache git gcc musl-dev
COPY . .
RUN go mod download && \
    go mod tidy && \
    go mod vendor && \
    go mod verify
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -ldflags="-w -s" -o k-monitor .

# Stage 3: Final image
FROM node:18-alpine
WORKDIR /app
RUN apk add --no-cache nginx supervisor
COPY --from=backend-builder /app/k-monitor .
COPY --from=frontend-builder /frontend/.next /app/frontend/.next
COPY --from=frontend-builder /frontend/public /app/frontend/public
COPY --from=frontend-builder /frontend/package*.json /app/frontend/
COPY nginx.conf /etc/nginx/nginx.conf
COPY supervisord.conf /etc/supervisord.conf
WORKDIR /app/frontend
RUN npm install --production
WORKDIR /app
RUN mkdir -p /run/nginx && \
    chown -R node:node /app /run/nginx /var/log/nginx /var/lib/nginx /etc/nginx
EXPOSE 80 3000 8081
USER node
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]