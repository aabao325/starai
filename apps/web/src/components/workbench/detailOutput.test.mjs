import assert from "node:assert/strict";
import test from "node:test";
import { readFileSync } from "node:fs";
import vm from "node:vm";
import ts from "typescript";

test("detail previews keep typeset output when API hydrates original photo tasks", () => {
  const source = ts.createSourceFile("agent.tsx", readFileSync(new URL("./AgentWorkspace.tsx", import.meta.url), "utf8"), ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);
  const fn = source.statements.find(node => ts.isFunctionDeclaration(node) && node.name?.text === "resolvedAgentMediaTasks");
  assert.ok(fn);
  const ctx = {};
  vm.runInNewContext(ts.transpileModule(fn.getText(source), { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText, ctx);
  const task = { task_no: "photo", status: "succeeded", progress: 100, output: { image_url: "remote-photo" } };
  const finished = { ...task, output: { source_image_url: "remote-photo", image_url: "local-typeset" } };
  const project = { inputs: { creative_scene: "detail_image" }, media_tasks: [task], outputs: { media_tasks: [finished] } };
  assert.equal(ctx.resolvedAgentMediaTasks(project)[0].output.image_url, "local-typeset");
  assert.equal(task.output.image_url, "remote-photo");
  task.status = "failed";
  assert.equal(ctx.resolvedAgentMediaTasks(project)[0].status, "failed");
  project.inputs.creative_scene = "main_image";
  assert.equal(ctx.resolvedAgentMediaTasks(project)[0].output.image_url, "remote-photo");
});
