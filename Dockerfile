FROM alpine:3.13 as builder
RUN apk add -U go git
ADD ./cmd /app/cmd
RUN cd /app && \
    go get "github.com/vmihailenco/msgpack" && \
    go build ./cmd/escape-csv && \
    go build ./cmd/expand

FROM alpine:3.13
COPY --from=builder /app/escape-csv /usr/local/bin/
COPY --from=builder /app/expand /usr/local/bin/
