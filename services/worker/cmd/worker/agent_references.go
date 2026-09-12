package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Read the actual pixels before analysis. A localhost URL in text is not a
// reference image, and external vision providers cannot fetch local storage.
func agentAnalysisReferenceImages(ctx context.Context, inputs map[string]interface{}) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	refs := referenceImageURLs(inputs)
	for i, ref := range refs {
		refs[i] = normalizeReferenceImage(ctx, ref)
		if !strings.HasPrefix(refs[i], "data:image/") {
			return nil, fmt.Errorf("第%d张参考图读取失败，请检查素材是否可访问；已停止无图分析", i+1)
		}
	}
	return refs, nil
}

func agentAnalysisModelAcceptsImages(model agentAnalysisModel) bool {
	caps := mapAnyOr(model.RuntimeRule["capabilities"], nil)
	return boolAny(caps["vision"]) || boolAny(caps["multimodal"]) || boolAny(caps["image_input"])
}

func detailSectionCount(inputs map[string]interface{}) int {
	if n := intAny(inputs["detail_section_count"]); n >= 4 && n <= 8 {
		return n
	}
	if n := intAny(inputs["count"]); n >= 4 && n <= 8 {
		return n
	}
	return 5
}

func applyAgentVisionContent(ctx context.Context, body map[string]interface{}, protocol, mode, system, user string, images []string) {
	switch normalizeWorkerLLMProtocol(protocol) {
	case "gemini":
		parts := []map[string]interface{}{{"text": user}}
		for _, ref := range collectBananaReferenceImages(ctx, images) {
			parts = append(parts, geminiImagePart(ref))
		}
		body["contents"] = []map[string]interface{}{{"role": "user", "parts": parts}}
	case "claude":
		parts := []map[string]interface{}{{"type": "text", "text": user}}
		for _, ref := range images {
			header, data, ok := strings.Cut(ref, ",")
			if !ok || !strings.HasPrefix(header, "data:image/") {
				continue
			}
			parts = append(parts, map[string]interface{}{"type": "image", "source": map[string]interface{}{"type": "base64", "media_type": strings.TrimSuffix(strings.TrimPrefix(header, "data:"), ";base64"), "data": data}})
		}
		body["messages"] = []map[string]interface{}{{"role": "user", "content": parts}}
	default:
		parts := []map[string]interface{}{{"type": "text", "text": user}}
		key := "messages"
		if mode == "responses" {
			key = "input"
			parts[0]["type"] = "input_text"
		}
		for _, ref := range images {
			part := map[string]interface{}{"type": "image_url", "image_url": map[string]string{"url": ref}}
			if mode == "responses" {
				part = map[string]interface{}{"type": "input_image", "image_url": ref}
			}
			parts = append(parts, part)
		}
		body[key] = []map[string]interface{}{{"role": "system", "content": system}, {"role": "user", "content": parts}}
	}
}
