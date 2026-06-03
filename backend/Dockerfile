FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/regear-api ./cmd/api

FROM alpine:3.22
WORKDIR /app
COPY --from=build /out/regear-api /app/regear-api
EXPOSE 8080
CMD ["/app/regear-api"]
