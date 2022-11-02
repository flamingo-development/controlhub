FROM golang:1.19.3 AS builder

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o control ./app/control/control.go

# ---

FROM golang

WORKDIR /app

COPY --from=builder /app/control .
COPY --from=builder /app/start.sh .
COPY --from=builder /app/config.json .
COPY --from=builder /app/plugins ./plugins
COPY --from=builder /app/formatters ./formatters

RUN chmod +x /app/start.sh
RUN chmod +x /app/plugins/install.sh

ENTRYPOINT [ "./start.sh" ]
CMD ["./control"]
