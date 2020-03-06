FROM golang:latest
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .
EXPOSE 5000
RUN useradd -r -u 1000 -U -d /app app
RUN chown -R app:app /app
USER 1000:1000
CMD ["./main"]
