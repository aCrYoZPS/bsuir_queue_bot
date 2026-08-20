FROM alpine:latest AS build

RUN apk add --no-cache --update go gcc g++

COPY ./go.mod .
COPY ./go.sum .

ARG TARGETARCH

RUN GOBIN=/bin go install github.com/go-delve/delve/cmd/dlv@latest
RUN --mount=type=cache,target=/go/pkg/mod 
RUN go mod download

RUN CGO_ENABLED=1 
RUN GOOS=linux GOARCH=${TARGETARCH}

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o /bin/main ./src/main.go

FROM alpine AS main

RUN apk add --no-cache --update musl-dev
WORKDIR /app

COPY run.sh /app/bin/run.sh
COPY --from=build /bin/dlv /bin/dlv
COPY --from=build /bin/main ./bin/main

WORKDIR /app/bin
RUN --mount=type=secret,id=credentials.json
RUN --mount=type=secret,id=token.json
ENTRYPOINT ["/app/bin/run.sh"]
