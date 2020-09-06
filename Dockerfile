# Build the go application into a binary
FROM golang:alpine as builder
WORKDIR /app
ADD . ./
RUN CGO_ENABLED=0 GOOS=linux go build -mod vendor -a -installsuffix cgo -o torrent-ip-tracker .

# Run the binary on an empty container
FROM scratch
COPY --from=builder /app/torrent-ip-tracker .
ENV PORT=12345
EXPOSE ${PORT}
ENTRYPOINT ["/torrent-ip-tracker"]