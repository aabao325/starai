package main

import (
	"strings"
	"unicode"
)

// Visual references establish appearance, not performance or fabric composition.
// Keep uncertain sales claims only when the complete copy quotes user input.
func groundedDetailAnalysis(analysis, inputs map[string]interface{}) map[string]interface{} {
	out := copyMap(analysis)
	sections, ok := analysis["detail_sections"].([]interface{})
	if !ok {
		return out
	}
	userFacts := firstNonEmpty(stringAny(inputs["user_prompt"]), firstUserPrompt(inputs))
	grounded := func(text string) string {
		text = strings.TrimSpace(text)
		for _, r := range text {
			if unicode.IsDigit(r) && !strings.Contains(userFacts, text) {
				return ""
			}
		}
		for _, claim := range []string{"柔软", "舒适", "亲肤", "透气", "轻盈", "轻便", "无负担", "方便", "便捷", "实用", "环保", "安全", "无异味", "保暖", "弹性", "弹力", "不紧绷", "防风", "防水", "防雨", "防护", "保护", "活动自如", "一甩即干", "牢固", "耐用", "耐磨", "防滑", "抗菌", "防晒", "防皱", "不起球", "塑料", "纯棉", "羊毛", "羊绒", "新款", "销量", "认证", "售后", "春秋", "冬季"} {
			if strings.Contains(text, claim) && !strings.Contains(userFacts, text) {
				return ""
			}
		}
		return text
	}
	cleaned := make([]interface{}, 0, len(sections))
	for _, raw := range sections {
		section, ok := mapAny(raw)
		if !ok {
			continue
		}
		next := copyMap(section)
		next["copy_title"] = grounded(stringAny(section["copy_title"]))
		points := []string{}
		for _, point := range stringSlice(section["copy_points"]) {
			if text := grounded(point); text != "" {
				points = append(points, text)
			}
		}
		next["copy_points"] = points
		if stringAny(section["type"]) == "specification" && len(points) == 0 {
			next["type"], next["title"], next["copy_title"] = "closing", "外观收尾", "外观细节"
			next["objective"] = "以已观察到的外观细节收尾；用户未提供规格，不制作空参数表"
			next["image_prompt"] = "参考商品正面可见局部的单处近景，保留原有颜色和结构，与整页统一底色和光线；不制作尺寸图、参数表、多角度阵列或留白占位表格"
		}
		sectionPlan := stringAny(section["title"]) + stringAny(section["objective"]) + stringAny(section["image_prompt"])
		if strings.Contains(sectionPlan, "包装") && !strings.Contains(userFacts, "包装") {
			next["type"], next["title"], next["copy_title"] = "closing", "商品收尾", ""
			next["copy_points"] = []string{}
			next["objective"] = "回到参考商品本身完成详情页收尾"
			next["image_prompt"] = "参考商品完整或半身英雄式展示，保持原商品与人物身份，延续整页背景和光线；不生成包装盒、吊牌、赠品或品牌道具"
		}
		cleaned = append(cleaned, next)
	}
	out["detail_sections"] = cleaned
	return out
}
