FROM golang:alpine

ENV GOARCH $TARGETARCH

ARG USER=extractor
ARG HOME_DIR=/home/$USER

RUN apk add --no-cache --update \
	git \
	exiftool

RUN adduser \
	--disabled-password \
	--uid 1000 \
	$USER

USER $USER
WORKDIR $HOME_DIR

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN go build -buildvcs=false

ENTRYPOINT ["sh", "entrypoint.sh"]
