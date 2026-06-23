package agent

import (
	"os"
	"strings"
	"testing"
)

func TestGuideIsNonEmpty(t *testing.T) {
	got := Guide()
	if strings.TrimSpace(got) == "" {
		t.Fatal("Guide() is empty")
	}
}

func TestGuideHasFrontmatter(t *testing.T) {
	got := Guide()
	if !strings.HasPrefix(got, "---\n") {
		t.Fatal("Guide() missing YAML frontmatter opening")
	}
	if !strings.Contains(got, "name: zn-cli") {
		t.Fatal("Guide() missing name: zn-cli in frontmatter")
	}
}

func TestGuideHasKeySections(t *testing.T) {
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

func TestGuideMatchesSourceFile(t *testing.T) {
	source, err := os.ReadFile("../../skills/zn-cli/SKILL.md")
	if err != nil {
		t.Fatalf("read source SKILL.md: %v", err)
	}
	want := strings.ReplaceAll(string(source), "\r\n", "\n")
	if Guide() != want {
		t.Fatal("embedded Guide() differs from skills/zn-cli/SKILL.md")
	}
}
