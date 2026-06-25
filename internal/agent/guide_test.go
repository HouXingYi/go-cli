package agent

import (
	"os"
	"strings"
	"testing"

	"ziniao/internal/variant"
)

func TestGuideIsNonEmpty(t *testing.T) {
	variant.SetCurrent(variant.Ent)
	got := Guide()
	if strings.TrimSpace(got) == "" {
		t.Fatal("Guide() is empty")
	}
}

func TestGuideHasFrontmatter(t *testing.T) {
	variant.SetCurrent(variant.Ent)
	got := Guide()
	if !strings.HasPrefix(got, "---\n") {
		t.Fatal("Guide() missing YAML frontmatter opening")
	}
	if !strings.Contains(got, "name: zn-ent") {
		t.Fatal("Guide() missing name: zn-ent in frontmatter")
	}
}

func TestGuideHasKeySections(t *testing.T) {
	variant.SetCurrent(variant.Ent)
	got := Guide()
	for _, section := range []string{
		"## 非交互策略",
		"## 推荐工作流",
		"## API 发现",
		"## HTTP 调用",
		"## 常见错误",
	} {
		if !strings.Contains(got, section) {
			t.Fatalf("Guide() missing section %q", section)
		}
	}
}

func TestGuideRendersPerVariant(t *testing.T) {
	variant.SetCurrent(variant.Eco)
	got := Guide()
	if strings.Contains(got, "zn-ent") {
		t.Fatal("eco guide should not contain zn-ent")
	}
	if !strings.Contains(got, "zn-eco") {
		t.Fatal("eco guide should contain zn-eco command examples")
	}
}

func TestGuideMatchesSourceTemplate(t *testing.T) {
	variant.SetCurrent(variant.Ent)
	source, err := os.ReadFile("../../skills/shared/SKILL.md")
	if err != nil {
		t.Fatalf("read source SKILL.md: %v", err)
	}
	want := strings.ReplaceAll(strings.ReplaceAll(string(source), "\r\n", "\n"), "{{AppName}}", "zn-ent")
	if Guide() != want {
		t.Fatal("embedded Guide() differs from rendered skills/shared/SKILL.md template")
	}
}
