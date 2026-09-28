FROM golang:1.26-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/api

FROM alpine:3.22
RUN apk add --no-cache ca-certificates
COPY --from=build /out/server /server
USER 10001:10001
EXPOSE 8081
ENTRYPOINT ["/server"]
