FROM golang

WORKDIR /app

COPY ./config /app/config 
COPY ./cmd /app/cmd 
COPY ./internal /app/internal
COPY ./go* /app

ENV GOPRIVATE=github.com/IlianBuh
ENV CONFIG_PATH="/app/config/config.json"

EXPOSE 5051

RUN go env -w GOPRIVATE="github.com/IlianBuh"
RUN go mod tidy
RUN go install /app/cmd/main/main.go

CMD main