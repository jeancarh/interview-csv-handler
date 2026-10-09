import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

const backend = "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/api": backend,
      "/docs": backend,
      "/healthz": backend,
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: ["./src/shared/test/setup.ts"],
  },
});
