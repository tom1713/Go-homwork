FROM golang:1.18.4-alpine
WORKDIR /go/src/api
ADD . /go/src/api
RUN  apk add git \
     && cd /go/src/api \
     && mkdir uploaded \
     && go get github.com/gorilla/mux  \
     && go get go.mongodb.org/mongo-driver/mongo \
     && go get go.mongodb.org/mongo-driver/mongo/options \
     && go get github.com/aws/aws-sdk-go/aws \
     && go build
EXPOSE 3000
ENTRYPOINT ./api