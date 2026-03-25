FROM golang:1.22-alpine AS build
WORKDIR /src
COPY shared /src/shared
COPY services/notification-service /src/services/notification-service
WORKDIR /src/services/notification-service
RUN go mod download && CGO_ENABLED=0 go build -o /out/srv ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/srv /app/srv
USER nonroot:nonroot
EXPOSE 8086
ENTRYPOINT ["/app/srv"]
