export type StreamEvent = { event: string; data: string };

// Keep raw boundaries until the next chunk arrives: CRLF and UTF-8 characters
// can both straddle network reads. Some providers omit the final blank line.
export async function* readEventStream(body: ReadableStream<Uint8Array>): AsyncGenerator<StreamEvent> {
  const reader = body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  try {
    while (true) {
      const { value, done } = await reader.read();
      buffer += decoder.decode(value, { stream: !done });
      const blocks = buffer.split(/\r\n\r\n|\n\n|\r\r/);
      buffer = blocks.pop() || "";
      if (done && buffer.trim()) blocks.push(buffer);
      for (const block of blocks) {
        let event = "";
        const data: string[] = [];
        for (const line of block.split(/\r\n|\n|\r/)) {
          const colon = line.indexOf(":");
          const field = colon < 0 ? line : line.slice(0, colon);
          const value = colon < 0 ? "" : line.slice(colon + 1).replace(/^ /, "");
          if (field === "event") event = value;
          if (field === "data") data.push(value);
        }
        if (data.length) yield { event, data: data.join("\n") };
      }
      if (done) break;
    }
  } finally {
    await reader.cancel().catch(() => {});
    reader.releaseLock();
  }
}
