FROM golang:1.19.3 AS builder

USER root

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o control ./app/control/control.go

RUN chmod +x /app/start.sh
RUN chmod +x /app/plugins/install.sh

ENTRYPOINT [ "./start.sh" ]
CMD ["./control"]
