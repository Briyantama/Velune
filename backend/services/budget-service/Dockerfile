FROM golang:1.22-alpine AS build
WORKDIR /src
COPY shared /src/shared
COPY services/budget-service /src/services/budget-service
WORKDIR /src/services/budget-service
RUN go mod download && CGO_ENABLED=0 go build -o /out/srv ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/srv /app/srv
COPY --from=build /src/services/budget-service/migrations /app/migrations
USER nonroot:nonroot
EXPOSE 8084
ENTRYPOINT ["/app/srv"]
