FROM golang:1.22-alpine AS build
WORKDIR /src
COPY shared /src/shared
COPY services/transaction-service /src/services/transaction-service
WORKDIR /src/services/transaction-service
RUN go mod download && CGO_ENABLED=0 go build -o /out/srv ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/srv /app/server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/app/server"]
