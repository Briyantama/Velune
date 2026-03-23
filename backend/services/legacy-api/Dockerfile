# Build from repository root: docker build -f services/legacy-api/Dockerfile .
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY shared /src/shared
COPY services/legacy-api /src/services/legacy-api
WORKDIR /src/services/legacy-api
RUN go mod download && CGO_ENABLED=0 go build -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=build /src/services/legacy-api/migrations /app/migrations
ENV MIGRATIONS_PATH=file:///app/migrations
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/api"]
