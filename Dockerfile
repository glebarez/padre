FROM golang:1.17

WORKDIR /padre

# Build
COPY . .
RUN go mod download
RUN go build -o padre .

# Runn
CMD ["./padre"]