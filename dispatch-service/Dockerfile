FROM golang:1.26.3

WORKDIR /app

RUN mkdir contracts
COPY contracts/ ./contracts

RUN mkdir lib
COPY lib/ ./lib

RUN mkdir integrations
COPY integrations/ ./integrations

RUN mkdir dispatch-service
COPY dispatch-service/ ./dispatch-service

RUN cd dispatch-service/cmd && CGO_ENABLED=0 GOOS=linux go build -o run-dispatch-service

WORKDIR /app/dispatch-service
CMD ["tail", "-f", "/dev/null"]