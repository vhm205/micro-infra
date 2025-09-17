import { cpus } from "node:os";

export function cpuHeavyWork(durationMs: number) {
  const start = Date.now();
  while (Date.now() - start < durationMs) {
    // một công việc nặng, tính toán
    Math.sqrt(Math.random() * Math.random());
  }
}

export async function allocateMemory(mb: number) {
  const size = mb * 1024 * 1024;
  const buffer = new ArrayBuffer(size);
  const view = new Uint8Array(buffer);
  // Ghi vào buffer để tránh bị GC ngay
  for (let i = 0; i < view.length; i += 4096) {
    view[i] = 0;
  }
  return buffer;
}

// Hàm lấy memory sử dụng
export function printMemory() {
  const mu = Deno.memoryUsage();
  console.log("===================================");
  console.log("Memory Usage:");
  console.log(`  RSS: ${(mu.rss / 1024 / 1024).toFixed(2)} MB`);
  console.log(`  Heap Total: ${(mu.heapTotal / 1024 / 1024).toFixed(2)} MB`);
  console.log(`  Heap Used: ${(mu.heapUsed / 1024 / 1024).toFixed(2)} MB`);
  console.log(`  External: ${(mu.external / 1024 / 1024).toFixed(2)} MB`);
  console.log("===================================");
}

// Hàm lấy CPU load
export function printCpu() {
  const load = Deno.loadavg(); // [1-min,5-min,15-min] trung bình load hệ thống :contentReference[oaicite:0]{index=0}
  const ncpu = cpus().length;
  console.log("===================================");
  console.log("CPU Load:");
  console.log(`  1-min Load Average: ${load[0].toFixed(2)} (of ${ncpu} cores)`);
  console.log(`  Utilization ~ ${((load[0] / ncpu) * 100).toFixed(2)}%`);
  console.log("===================================");
}
