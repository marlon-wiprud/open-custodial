FROM golang:1.16-alpine as builder
WORKDIR "$GOPATH/src/open_custodial"

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN apk --update --no-cache add g++
# RUN go build -o /run/app ./cmd/hot_wallet

FROM alpine:latest

RUN apk update
RUN apk add bash 
RUN apk add build-base
RUN apk add openssl-dev
RUN apk add opensc

RUN wget https://dist.opendnssec.org/source/softhsm-2.3.0.tar.gz
RUN tar -xzf softhsm-2.3.0.tar.gz
WORKDIR /softhsm-2.3.0
RUN ./configure --disable-gost
RUN make install

WORKDIR /
COPY --from=builder /run/app /
COPY ./abi/social_money.json /abi/social_money.json


# initialize softhsm token
RUN softhsm2-util --init-token --slot 0 --label "hot_wallet" --pin "test12345" --so-pin "test12345"
# generate key
RUN pkcs11-tool --module /usr/local/lib/softhsm/libsofthsm2.so  --token-label "hot_wallet" --login --pin "test12345" --keypairgen --id 1 --key-type EC:secp256k1

ENV GRPC_SERVER_ADDRESS=:3001
ENV HSM_LIB_PATH=/usr/local/lib/softhsm/libsofthsm2.so
ENV CU_USERNAME=test12345
ENV CU_PASSWORD=test12345
ENV CERT_NAME=test
ENV KEY_NAME=test
ENV CERT_AUTHORITY_NAME=test
ENV AWS_ACCESS_KEY=test
ENV AWS_SECRET_KEY=test
ENV HOT_WALLET_LABEL=hot_wallet
ENV ELASTIC_URL=test
ENV ELASTIC_USERNAME=test
ENV ELASTIC_PASSWORD=test
ENV GO_ENV=local


EXPOSE 3001:3001

ENTRYPOINT [ "/app" ]
