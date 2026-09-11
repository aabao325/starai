import assert from "node:assert/strict";
import test from "node:test";
import { readEventStream } from "./eventStream.ts";

test("chat events survive every byte boundary, CRLF, optional spaces and an unterminated final event", async () => {
  const bytes = new TextEncoder().encode('data:{"choices":[{"delta":{"content":"你好"}}]}\r\n\r\nevent: done\r\ndata: {"conversation_id":\r\ndata: "c1"}\r\n\r\ndata: [DONE]');
  for (let split = 0; split <= bytes.length; split++) {
    const stream = new ReadableStream({ start(c) { c.enqueue(bytes.slice(0, split)); c.enqueue(bytes.slice(split)); c.close(); } });
    const events = [];
    for await (const event of readEventStream(stream)) events.push(event);
    assert.equal(JSON.parse(events[0].data).choices[0].delta.content, "你好");
    assert.equal(events[1].event, "done");
    assert.equal(JSON.parse(events[1].data).conversation_id, "c1");
    assert.equal(events[2].data, "[DONE]");
  }
});
