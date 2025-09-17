import { Hono } from "hono";
import { connect } from "https://deno.land/x/amqp@v0.23.1/mod.ts";
import { S3Client, PutObjectCommand } from "npm:@aws-sdk/client-s3";

import {
  printCpu,
  printMemory,
  allocateMemory,
  cpuHeavyWork,
} from "./helper.ts";

const MINIO_ENDPOINT = Deno.env.get("MINIO_ENDPOINT") || "localhost:9000";
const MINIO_ACCESS_KEY = Deno.env.get("MINIO_ACCESS_KEY") || "minioadmin";
const MINIO_SECRET_KEY = Deno.env.get("MINIO_SECRET_KEY") || "minioadmin";
const MINIO_BUCKET = Deno.env.get("MINIO_BUCKET") || "test";
const USE_SSL = false; // or true if using TLS
console.log({
  MINIO_ENDPOINT,
  MINIO_ACCESS_KEY,
  MINIO_SECRET_KEY,
  MINIO_BUCKET,
  USE_SSL,
});

// initialize S3 client that points to MinIO
const s3Client = new S3Client({
  endpoint: `http${USE_SSL ? "s" : ""}://${MINIO_ENDPOINT}`,
  region: "us-east-1", // region string; MinIO ignores region in many cases
  credentials: {
    accessKeyId: MINIO_ACCESS_KEY,
    secretAccessKey: MINIO_SECRET_KEY,
  },
  forcePathStyle: true, // important for many S3 compatible storages including MinIO
});

const app = new Hono();

const channels: any[] = [];

async function connectToRabbitMQ() {
  const hostname = Deno.env.get("RABBITMQ_HOST") || "localhost";
  const port = +Deno.env.get("RABBITMQ_PORT") || 5672;
  const username = Deno.env.get("RABBITMQ_USERNAME") || "guest";
  const password = Deno.env.get("RABBITMQ_PASSWORD") || "guest";
  console.log({
    hostname,
    port,
    username,
    password,
  });

  const connection = await connect({
    hostname,
    port,
    username,
    password,
  });
  return connection;
}

const connection = await (async () => {
  try {
    return await connectToRabbitMQ();
  } catch (error) {
    console.error("RabbitMQ connection failed", error);
    return null;
  }
})();

const getChannel = async () => {
  if (channels.length > 0) {
    return channels[0];
  }

  if (!connection) return null;

  const channel = await connection.openChannel();
  channels.push(channel);

  await channel.declareQueue({
    queue: "product-service",
    durable: true,
  });
  await channel.bindQueue({
    queue: "product-service",
    exchange: "amq.direct",
    routingKey: "product.import",
  });

  return channel;
};

app.post("/upload", async (c) => {
  const body = await c.req.parseBody({ type: "form-data" });
  const file = body?.files;

  if (!file) {
    return c.text("No file provided", 400);
  }

  const arrayBuffer = await file.arrayBuffer();
  const uint8array = new Uint8Array(arrayBuffer);

  const key = `uploads/${file.name}`; // or make unique
  const putParams = {
    Bucket: MINIO_BUCKET,
    Key: key,
    Body: uint8array,
    ContentType: file.type,
  };

  try {
    const result = await s3Client.send(new PutObjectCommand(putParams));
    const channel = await getChannel();
    await channel.publish(
      { exchange: "amq.direct", routingKey: "product.import" },
      { contentType: "application/json" },
      new TextEncoder().encode(JSON.stringify({ key, result })),
    );
    console.log(`Message published with payload ${JSON.stringify(result)}`);

    return c.json({
      message: "Uploaded successfully",
      key,
      result,
    });
  } catch (err) {
    console.error("Upload error:", err);
    return c.text("Upload failed", 500);
  }
});

app.post("/enqueue", async (c) => {
  try {
    const query = c.req.query();
    const body = await c.req.json();

    const jobs = Array.from({ length: 100 });
    const channel = await getChannel();

    for (const job of jobs) {
      await channel.publish(
        { exchange: "amq.direct", routingKey: "product.import" },
        { contentType: "application/json" },
        new TextEncoder().encode(JSON.stringify({ ...body, job })),
      );
      console.log(`Message published with payload ${JSON.stringify(body)}`);
    }

    return c.json({ status: "queued", query, body });
  } catch (error) {
    console.log(error);
    return c.json({ status: "error", error });
  }
});

// Create a heavy task
app.get("/heavy", async (c) => {
  const query = c.req.query();

  try {
    const totalBuf = parseInt(query.totalBuf || "100");
    const duration = parseInt(query.duration || "1000");

    // for (let i = 0; i < totalLoop; i++) {
    //   const start = Date.now();
    //   await Promise.all(
    //     Array.from({ length: concurrency }, () => {
    //       return new Promise((resolve) => {
    //         setTimeout(() => {
    //           resolve(true);
    //         }, delay);
    //       });
    //     }),
    //   );
    //   const end = Date.now();
    //   console.log(`Loop ${i} took ${end - start}ms`);
    // }

    console.log("Start stress test");

    // Tăng CPU usage: chạy vòng lặp nặng trong một Worker
    const cpuWorker = new Worker(new URL("worker.ts", import.meta.url).href, {
      type: "module",
    });
    // gửi yêu cầu xét nặng
    cpuWorker.postMessage({ duration: 5000 }); // 5 giây công việc nặng

    // Tăng memory usage
    const bigBuffers = [];
    for (let i = 0; i < totalBuf; i++) {
      console.log(`Allocating buffer #${i}`);
      const buf = await allocateMemory(50); // mỗi buffer 50 MB → tổng 500 MB
      bigBuffers.push(buf);
      await new Promise((resolve) => setTimeout(resolve, 500));
    }

    cpuHeavyWork(duration); // tính toán nặng trong một Worker

    console.log("Stress test done");

    return c.json({ status: "ok" });
  } catch (error) {
    console.log(error);
    return c.json({ status: "error", error });
  }
});

app.get("/health", async (c) => {
  const appName = Deno.env.get("APP_NAME");

  console.clear();
  printMemory();
  printCpu();

  return c.json({ status: "ok", appName });
});

setInterval(() => {
  printMemory();
  printCpu();
}, 3000);

Deno.serve({ port: 5001 }, app.fetch);

Deno.addSignalListener("SIGINT", async () => {
  console.log("Caught SIGINT, shutting down server...");

  if (connection) {
    await connection.close();
  }

  // Add any other cleanup logic here
  Deno.exit(0);
});
