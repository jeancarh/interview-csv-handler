import js from "@eslint/js";
import { defineConfig } from "eslint/config";
import reactHooks from "eslint-plugin-react-hooks";
import globals from "globals";
import tseslint from "typescript-eslint";

const outer = ["**/app/**", "**/shared/api/**", "**/shared/http/**"];

function layer(files, banned, message) {
  return {
    files,
    rules: {
      "no-restricted-imports": ["error", { patterns: [{ group: banned, message }] }],
    },
  };
}

export default defineConfig(
  { ignores: ["dist"] },
  js.configs.recommended,
  tseslint.configs.recommended,
  reactHooks.configs.flat.recommended,
  {
    files: ["**/*.{ts,tsx}"],
    languageOptions: { globals: globals.browser },
  },
  {
    files: ["scripts/**", "*.config.{js,ts}"],
    languageOptions: { globals: globals.node },
  },
  layer(
    ["src/features/*/domain/**"],
    ["react", "react-dom", "react-router", "@tanstack/*", "**/application/**", "**/infrastructure/**", "**/ui/**", ...outer],
    "domain code is plain TypeScript: no React, no HTTP, no outer layers",
  ),
  layer(
    ["src/features/*/application/**"],
    ["**/infrastructure/**", "**/ui/**", ...outer],
    "application code talks to repository ports, never to adapters or UI",
  ),
  layer(
    ["src/features/*/ui/**"],
    ["**/infrastructure/**", ...outer],
    "UI goes through application hooks, it never calls the API directly",
  ),
  layer(
    ["src/features/*/infrastructure/**"],
    ["**/application/**", "**/ui/**", "**/app/**"],
    "adapters implement domain ports and know nothing about React or the UI",
  ),
);
