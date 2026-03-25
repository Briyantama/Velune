FROM golang:1.25-alpine AS build
WORKDIR /src
COPY shared /src/shared
COPY services/admin-service /src/services/admin-service
WORKDIR /src/services/admin-service
RUN go mod download && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/admin ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/admin /app/admin
USER nonroot:nonroot
EXPOSE 8099
ENTRYPOINT ["/app/admin"]
