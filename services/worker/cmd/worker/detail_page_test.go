package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/starai/worker/internal/storage"
)

func TestAgentDetailSectionsBuildsCompletePagePlan(t *testing.T) {
	analysis := map[string]interface{}{
		"detail_sections": []interface{}{
			map[string]interface{}{"id": "custom", "type": "hero", "title": "定制首屏", "image_prompt": "定制商品首屏"},
		},
	}
	sections := agentDetailSections(analysis, map[string]interface{}{"count": float64(1)}, "基础商品方案")
	if len(sections) != 5 {
		t.Fatalf("sections=%d, want 5", len(sections))
	}
	if sections[0]["title"] != "定制首屏" {
		t.Fatalf("first section=%#v", sections[0])
	}
	if sections[4]["type"] != "closing" {
		t.Fatalf("last default section=%#v", sections[4])
	}
}

func TestDetailSectionPromptEnforcesSingleDesignedModule(t *testing.T) {
	prompt := detailSectionGenerationPrompt(
		"玻尿酸精华液，30ml，三重保湿",
		map[string]interface{}{"type": "material", "title": "材质细节", "objective": "展示瓶身和滴管", "image_prompt": "微距商品摄影", "copy_title": "瓶身细节", "copy_points": []string{"30ml"}},
		2,
		6,
		map[string]interface{}{"creative_scene": "detail_image", "creative_scene_label": "商品详情图"},
	)
	for _, expected := range []string{"DETAIL PAGE MODULE 3/6", "视觉真值", "不绘制任何新增文字", "模块镜头硬约束", "2–3个有层级的局部近景", "不得创造新颜色", "包装盒", "数量=1", "商品详情图"} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("prompt missing %q: %s", expected, prompt)
		}
	}
}

func TestAgentContentImageCardsKeepsPlanAndFillsRequestedCount(t *testing.T) {
	post := map[string]interface{}{
		"title": "秋季护肤指南",
		"cards": []interface{}{
			map[string]interface{}{"id": "cover", "role": "cover", "headline": "换季别焦虑", "copy": "三步稳住皮肤", "image_prompt": "秋日暖色护肤静物"},
		},
	}
	cards := agentContentImageCards(post, map[string]interface{}{"image_count": 4}, "秋季护肤")
	if len(cards) != 4 || cards[0]["id"] != "cover" || cards[3]["role"] != "cta" {
		t.Fatalf("unexpected content cards: %#v", cards)
	}
	for _, card := range cards {
		if strings.TrimSpace(stringAny(card["image_prompt"])) == "" {
			t.Fatalf("card has no image prompt: %#v", card)
		}
	}
}

func TestContentImageAnalysisPromptRequiresPublishableStructure(t *testing.T) {
	prompt := buildAgentAnalysisSystemPrompt("image", "content_image_post", 1, "content_image_post")
	for _, expected := range []string{"content_task", "objective", "deliverables", "content_post", "title", "body", "hashtags", "cards 数量必须严格等于", "不得把标题、正文、标签直接画进图片"} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("content image prompt missing %q: %s", expected, prompt)
		}
	}
}

func TestComposeDetailPageLongImage(t *testing.T) {
	root := t.TempDir()
	store, err := storage.NewLocal(root, "http://localhost:8080/uploads-local")
	if err != nil {
		t.Fatal(err)
	}
	previous := objectStore
	objectStore = store
	t.Cleanup(func() { objectStore = previous })

	urls := []string{
		testPNGDataURL(t, 120, 80, color.RGBA{R: 255, A: 255}),
		testPNGDataURL(t, 60, 30, color.RGBA{G: 255, A: 255}),
	}
	got, err := composeDetailPageLongImage(context.Background(), "wfp_test", urls)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "/workflows/wfp_test/detail-page-") {
		t.Fatalf("url=%q", got)
	}
	files, err := filepath.Glob(filepath.Join(root, "workflows", "wfp_test", "detail-page-*.jpg"))
	if err != nil || len(files) != 1 {
		t.Fatalf("files=%v err=%v", files, err)
	}
	f, err := os.Open(files[0])
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != 120 || img.Bounds().Dy() != 140 {
		t.Fatalf("bounds=%v", img.Bounds())
	}
}

func testPNGDataURL(t *testing.T, width, height int, fill color.Color) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, fill)
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(out.Bytes())
}

func TestDetailFallbackHasUniqueEvidenceBasedModules(t *testing.T) {
	for count := 4; count <= 8; count++ {
		sections := agentDetailSections(nil, map[string]interface{}{"detail_section_count": count}, "商品")
		if len(sections) != count {
			t.Fatalf("count %d: %#v", count, sections)
		}
		if count == 5 && sections[count-1]["type"] != "closing" {
			t.Fatalf("default five should close on the product: %#v", sections)
		}
		seen := map[string]bool{}
		for _, section := range sections {
			kind := stringAny(section["type"])
			if seen[kind] {
				t.Fatalf("duplicate %s", kind)
			}
			seen[kind] = true
		}
	}
}

func TestCommerceAnalysisRequiresGroundedClaimsAndDetailCopy(t *testing.T) {
	prompt := buildAgentAnalysisSystemPrompt("image", "ecommerce_image", 3, "detail_image")
	for _, want := range []string{"missing_information", "不能改变商品事实", "detail_section_count", "文字分别交付", "2–3个局部近景", "不制作空参数表", "无依据时留空"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("missing %s", want)
		}
	}
}

func TestCommerceReferencesPreserveAllViews(t *testing.T) {
	inputs := map[string]interface{}{"image_url": "https://example.com/front.jpg", "reference_images": []string{"https://example.com/front.jpg", "https://example.com/back.jpg", "https://example.com/detail.jpg"}}
	refs := referenceImageURLs(inputs)
	if len(refs) != 3 || refs[0] != "https://example.com/front.jpg" {
		t.Fatalf("references: %#v", refs)
	}
	task := agentMediaTaskInput(inputs, "商品详情", "qa")
	if len(referenceImageURLs(task)) != 3 {
		t.Fatalf("task lost reference views: %#v", task)
	}
}
