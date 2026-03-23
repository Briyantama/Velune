FROM golang:1.22-alpine AS build
WORKDIR /src
COPY shared /src/shared
COPY services/auth-service /src/services/auth-service
WORKDIR /src/services/auth-service
RUN go mod download && CGO_ENABLED=0 go build -o /out/srv ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/srv /app/srv
COPY --from=build /src/services/auth-service/migrations /app/migrations
USER nonroot:nonroot
EXPOSE 8081
ENTRYPOINT ["/app/srv"]
