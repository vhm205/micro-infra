// cpu_worker.ts
self.onmessage = (evt) => {
  const { duration } = evt.data;
  const start = Date.now();
  console.log("CPU worker started");
  while (Date.now() - start < duration) {
    // công việc nặng — ví dụ tính toán
    Math.log(Math.random() + Math.random());
  }
  console.log("CPU worker done");
  postMessage("done");
};
