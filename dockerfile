FROM alpine:latest AS build

WORKDIR /build
RUN apk add --no-cache --update go gcc g++

COPY ./go.mod .
COPY ./go.sum .

RUN go mod download
RUN --mount=type=cache,target=/go/pkg/mod 
RUN CGO_ENABLED=1 GOOS=linux

COPY . .
RUN go build -o ./src/main ./src/main.go

FROM alpine AS main

WORKDIR /app

COPY --from=build /build .

WORKDIR /app/src
# RUN --mount=type=secret,OAUTH2_CREDENTIALS=/run/secrets/oauth2
ENTRYPOINT ["./main"]
