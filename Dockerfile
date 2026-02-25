FROM golang:1.25-alpine AS build

RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN templ generate
RUN CGO_ENABLED=0 go build -o /recipebox .

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /recipebox /usr/local/bin/recipebox
COPY --from=build /src/static /app/static

WORKDIR /app
RUN mkdir -p /data/uploads
ENV RECIPEBOX_STORAGE_UPLOADS_DIR=/data/uploads
EXPOSE 8080
ENTRYPOINT ["recipebox"]
CMD ["serve"]
