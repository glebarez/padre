FROM golang:1.14

WORKDIR /padre


# Install Root CA cert if any
COPY *.crt /usr/local/share/ca-certificates/extra/
RUN update-ca-certificates

# Deps
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN go build -o padre .

# Runn
CMD ["./padre"]