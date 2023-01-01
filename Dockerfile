# syntax=docker/dockerfile:1

FROM golang:1.18-alpine as build-mmt

WORKDIR /app
COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /mmt .

FROM adiii717/ffmpeg as main
COPY --from=build-mmt /mmt .
CMD [ "./mmt" ]