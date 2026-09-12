import assert from "node:assert/strict";
import test from "node:test";
import { readFileSync } from "node:fs";
import vm from "node:vm";
import ts from "typescript";

test("reference picker preserves draft selection during parent renders and snapshots on reopen", () => {
  const source = ts.createSourceFile("BottomBar.tsx", readFileSync(new URL("./BottomBar.tsx", import.meta.url), "utf8"), ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);
  let effect;
  function visit(node) {
    if (ts.isCallExpression(node) && node.expression.getText(source) === "useEffect" && node.arguments[0]?.getText(source).includes("setPickedRefs(")) effect = node.getText(source);
    ts.forEachChild(node, visit);
  }
  visit(source);
  assert.ok(effect, "missing picker initialization effect");
  let previous, picked;
  const context = {
    assetOpen: false, referencePickMode: true, referenceImages: [], referenceImagesRef: { current: [] },
    setAssetTab() {}, setGalleryQuery() {}, setAssetKind() {}, setAssetType() {},
    setPickedRefs(value) { picked = value; },
    useEffect(fn, deps) {
      if (!previous || deps.some((value, i) => value !== previous[i])) fn();
      previous = deps;
    },
  };
  const script = ts.transpileModule(effect, { compilerOptions: { target: ts.ScriptTarget.ES2022 } }).outputText;
  const render = () => vm.runInNewContext(script, context);
  render();
  context.assetOpen = true;
  render();
  picked = [{ url: "selected-in-dialog" }];
  // The carousel gives the child a fresh array on every parent render.
  context.referenceImages = [];
  context.referenceImagesRef.current = context.referenceImages;
  render();
  assert.equal(picked[0].url, "selected-in-dialog");
  context.assetOpen = false;
  render();
  context.referenceImagesRef.current = [{ url: "latest-confirmed" }];
  context.assetOpen = true;
  render();
  assert.equal(picked[0].url, "latest-confirmed");
});
