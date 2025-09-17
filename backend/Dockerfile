# 1. Dùng image Deno chính thức
FROM denoland/deno:2.5.0

WORKDIR /app

COPY . .

EXPOSE 5000

CMD ["run", "--allow-net", "--allow-env", "--allow-read", "--allow-write", "--allow-sys", "app.ts"]

