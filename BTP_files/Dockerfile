FROM golang:1.23-alpine as builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o custom-scheduler .

FROM alpine:3.14
COPY --from=builder /app/custom-scheduler /usr/local/bin/custom-scheduler
CMD ["custom-scheduler"]

