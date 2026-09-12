package main

import (
	"strings"
	"testing"
)

func TestDetailCopyDropsUnsupportedClaimsAndEmptySpecifications(t *testing.T) {
	analysis := map[string]interface{}{"detail_sections": []interface{}{
		map[string]interface{}{"type": "material", "copy_title": "柔软亲肤", "copy_points": []interface{}{"米白到灰色渐变", "保暖透气"}},
		map[string]interface{}{"type": "specification", "copy_title": "规格参数"},
	}}
	clean := groundedDetailAnalysis(analysis, map[string]interface{}{"user_prompt": "请根据参考图生成衣服详情"})
	sections := clean["detail_sections"].([]interface{})
	material := sections[0].(map[string]interface{})
	if material["copy_title"] != "" || len(stringSlice(material["copy_points"])) != 1 || stringSlice(material["copy_points"])[0] != "米白到灰色渐变" {
		t.Fatalf("unsupported copy retained: %#v", material)
	}
	if sections[1].(map[string]interface{})["type"] != "closing" {
		t.Fatal("empty specifications should not become a blank module")
	}
	if analysis["detail_sections"].([]interface{})[0].(map[string]interface{})["copy_title"] != "柔软亲肤" {
		t.Fatal("mutated stored analysis")
	}
	clean = groundedDetailAnalysis(analysis, map[string]interface{}{"user_prompt": "已确认卖点：保暖透气"})
	if len(stringSlice(clean["detail_sections"].([]interface{})[0].(map[string]interface{})["copy_points"])) != 2 {
		t.Fatal("user-confirmed copy removed")
	}
}

func TestDetailCopyRejectsInventedMeasurementsAndPackaging(t *testing.T) {
	analysis := map[string]interface{}{"detail_sections": []interface{}{
		map[string]interface{}{"type": "specification", "copy_title": "尺寸参考", "copy_points": []interface{}{"衣长120cm"}, "image_prompt": "尺寸箭头展示"},
		map[string]interface{}{"type": "closing", "copy_title": "品牌收尾", "image_prompt": "商品与包装盒组合展示"},
	}}
	clean := groundedDetailAnalysis(analysis, map[string]interface{}{"user_prompt": "根据参考图生成详情"})
	sections := clean["detail_sections"].([]interface{})
	if sections[0].(map[string]interface{})["type"] != "closing" || len(stringSlice(sections[0].(map[string]interface{})["copy_points"])) != 0 {
		t.Fatalf("invented measurement retained: %#v", sections[0])
	}
	packaging := sections[1].(map[string]interface{})
	if strings.Contains(stringAny(packaging["image_prompt"]), "包装盒组合") || stringAny(packaging["title"]) != "商品收尾" {
		t.Fatalf("invented packaging retained: %#v", packaging)
	}
}

func TestDetailCopyKeepsAppearanceAndDropsInferredPerformance(t *testing.T) {
	analysis := map[string]interface{}{"detail_sections": []interface{}{
		map[string]interface{}{"type": "hero", "copy_title": "儿童长款透明雨衣", "copy_points": []interface{}{"蓝色透明设计，轻便防雨", "白色按扣清晰可见", "连帽按扣，穿脱便捷", "按扣闭合，牢固耐用"}},
	}}
	clean := groundedDetailAnalysis(analysis, map[string]interface{}{"user_prompt": "根据参考图生成详情"})
	section := clean["detail_sections"].([]interface{})[0].(map[string]interface{})
	points := stringSlice(section["copy_points"])
	if stringAny(section["copy_title"]) != "儿童长款透明雨衣" || len(points) != 1 || points[0] != "白色按扣清晰可见" {
		t.Fatalf("visual facts and inferred claims were not separated: %#v", section)
	}
}
