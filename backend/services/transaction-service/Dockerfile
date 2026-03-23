FROM golang:1.22-alpine AS build
WORKDIR /src
COPY shared /src/shared
COPY services/transaction-service /src/services/transaction-service
WORKDIR /src/services/transaction-service
RUN go mod download && CGO_ENABLED=0 go build -o /out/srv ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/srv /app/srv
COPY --from=build /src/services/transaction-service/migrations /app/migrations
USER nonroot:nonroot
EXPOSE 8082
ENTRYPOINT ["/app/srv"]
