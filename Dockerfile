FROM golang:1.22.1-bullseye

LABEL version="3.0.0" description="Api to control whatsapp features through http requests." 
LABEL maintainer="jrCleber" git="https://github.com/jrCleber"
LABEL contact="suporte@codechat.dev"

ENV DOCKER_ENV=true

WORKDIR /home/codechat

COPY go.mod .

RUN go mod tidy

COPY ./dist/main .

RUN mkdir -p ./data/sqlite
RUN touch codechat.db

ENV DOCKER_ENV=true

CMD [ "./main" ]