import { defineConfig } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";

const backendTarget = process.env.VITE_BACKEND_URL ?? "http://localhost:8080";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    proxy: {
      "/api": {
        target: backendTarget,
        changeOrigin: true,
      },
      // MCP server and OAuth authorization server (KEG-12). /oauth/consent is
      // a frontend route, so only the backend endpoints are proxied.
      "/mcp": { target: backendTarget, changeOrigin: true },
      "/.well-known": { target: backendTarget, changeOrigin: true },
      "^/oauth/(authorize|token|register|revoke)": { target: backendTarget, changeOrigin: true },
    },
  },
});
