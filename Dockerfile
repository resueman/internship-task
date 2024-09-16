FROM golang:1.22.2

WORKDIR /app

ADD tender-management-api/ .

RUN go mod download && go mod verify

RUN go build -o main

EXPOSE 8080

CMD [ "./main" ]
