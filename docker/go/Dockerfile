FROM golang:1.26

WORKDIR /app

# ホットリロード用ツール
RUN go install github.com/air-verse/air@latest
ENV PATH="/root/go/bin:${PATH}"

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
