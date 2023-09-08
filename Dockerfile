FROM golang:1.20-alpine3.18 as builder
WORKDIR /app
COPY . .
RUN go build -tags hook_1,hook_2,hook_3
RUN ls

FROM alpine:3.18
COPY --from=builder /app/scheduler /app
COPY app.env app.env
ENTRYPOINT ["/app"]
