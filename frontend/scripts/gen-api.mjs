import { readFile, writeFile } from "node:fs/promises";
import openapiTS, { astToString } from "openapi-typescript";
import ts from "typescript";

const spec = new URL("../../backend/api/openapi.yaml", import.meta.url);
const out = new URL("../src/shared/api/schema.d.ts", import.meta.url);

const generated = astToString(await openapiTS(await readFile(spec, "utf8")));
const source = ts.createSourceFile("schema.d.ts", generated, ts.ScriptTarget.Latest, false, ts.ScriptKind.TS);
const printer = ts.createPrinter({ removeComments: true, newLine: ts.NewLineKind.LineFeed });

await writeFile(out, printer.printFile(source));
