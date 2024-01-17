FROM volvlabs/volvlabs-ci-base:latest

WORKDIR /home/app

COPY ./bin/run.sh /home/app/run.sh

COPY ./* /home/app/

RUN go install github.com/cespare/reflex@latest
