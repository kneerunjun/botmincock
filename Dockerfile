# FROM kneerunjun/gogingonic:latest
FROM golang:1.19-alpine3.16
# from the vanilla image of go gin with mgo driver
# mapping for log files
ARG SRC
ARG LOG
ARG RUN
ARG ETC 
ARG BIN
RUN apk add git
RUN mkdir -p ${SRC} && mkdir -p ${LOG} && mkdir -p ${RUN} && mkdir -p ${ETC}
WORKDIR ${SRC}
# https://stackoverflow.com/questions/30215830/dockerfile-copy-keep-subdirectory-structure
# since we want the entire directory structure recursively to be copied onto the container
COPY . .
RUN go mod download 
RUN go build -o ${BIN}/botmincock .