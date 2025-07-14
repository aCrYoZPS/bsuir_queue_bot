FROM alpine:latest AS build

WORKDIR /build
RUN apk add --no-cache --update go gcc g++

COPY ./go.mod .
COPY ./go.sum .

RUN go mod download
RUN --mount=type=cache,target=/go/pkg/mod 
RUN CGO_ENABLED=1 GOOS=linux

COPY . .
RUN go build -o ./bin/main ./src/main.go

FROM alpine AS main

WORKDIR /app

COPY --from=build /build .

WORKDIR /app
RUN --mount=type=secret,id=credentials.json
RUN --mount=type=secret,id=token.json
ENTRYPOINT ["./bin/main"]
