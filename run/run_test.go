package run

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/xhd2015/go-best-practice/skill"
)

func TestListTopics(t *testing.T) {
	topics, err := listTopics()
	if err != nil {
		t.Fatalf("listTopics: %v", err)
	}
	topicSet := make(map[string]bool, len(topics))
	for _, tp := range topics {
		topicSet[tp] = true
	}
	expectedTopics := []string{
		"cli",
		"cli/color",
		"cli/dry-run",
		"cli/inline-tui-mouse",
		"cli/output-alignment",
		"cli/project-layout",
		"cli/skill-cli",
		"cli/streaming",
		"cmd-exec",
		"flags-parsing",
		"flags-parsing/collect",
		"flags-parsing/cut",
		"flags-parsing/subcommand",
		"flags-parsing/types",
		"go-embed-assets",
		"go-embed-version",
		"kool-create",
	}
	for _, expected := range expectedTopics {
		if !topicSet[expected] {
			t.Errorf("missing topic: %q", expected)
		}
	}
}

func TestReadTopicExistingTopLevel(t *testing.T) {
	content, ok, err := readTopic("kool-create")
	if err != nil {
		t.Fatalf("readTopic(kool-create): %v", err)
	}
	if !ok {
		t.Fatal("expected ok for kool-create")
	}
	if !strings.Contains(content, "kool create") {
		t.Errorf("unexpected content for kool-create: %s", content)
	}
}

func TestReadTopicGoEmbedAssets(t *testing.T) {
	content, ok, err := readTopic("go-embed-assets")
	if err != nil {
		t.Fatalf("readTopic(go-embed-assets): %v", err)
	}
	if !ok {
		t.Fatal("expected ok for go-embed-assets")
	}
	if !strings.Contains(content, "go-best-practice/go-embed-assets") {
		t.Errorf("go-embed-assets missing nested frontmatter name: %s", content)
	}
	if !strings.Contains(content, "placeholder") {
		t.Errorf("go-embed-assets missing placeholder guidance: %s", content)
	}
	if !strings.Contains(content, "hydrate") {
		t.Errorf("go-embed-assets missing hydrate guidance: %s", content)
	}
}

func TestReadTopicGoEmbedVersion(t *testing.T) {
	content, ok, err := readTopic("go-embed-version")
	if err != nil {
		t.Fatalf("readTopic(go-embed-version): %v", err)
	}
	if !ok {
		t.Fatal("expected ok for go-embed-version")
	}
	if !strings.Contains(content, "go-best-practice/go-embed-version") {
		t.Errorf("go-embed-version missing nested frontmatter name: %s", content)
	}
	if !strings.Contains(content, "restore") {
		t.Errorf("go-embed-version missing restore guidance: %s", content)
	}
	if !strings.Contains(content, "//go:embed VERSION.txt") {
		t.Errorf("go-embed-version missing embed directive: %s", content)
	}
}

func TestReadTopicSkillCLI(t *testing.T) {
	content, ok, err := readTopic("cli/skill-cli")
	if err != nil {
		t.Fatalf("readTopic(cli/skill-cli): %v", err)
	}
	if !ok {
		t.Fatal("expected ok for cli/skill-cli")
	}
	if !strings.Contains(content, "skill --install") && !strings.Contains(content, "--install") {
		t.Errorf("unexpected content for cli/skill-cli: %s", content)
	}
	if !strings.Contains(content, "go-best-practice/cli/skill-cli") {
		t.Errorf("cli/skill-cli missing nested frontmatter name: %s", content)
	}
}

func TestReadTopicCLIStreaming(t *testing.T) {
	content, ok, err := readTopic("cli/streaming")
	if err != nil {
		t.Fatalf("readTopic(cli/streaming): %v", err)
	}
	if !ok {
		t.Fatal("expected ok for cli/streaming")
	}
	if !strings.Contains(content, "go-best-practice/cli/streaming") {
		t.Errorf("cli/streaming missing nested frontmatter name: %s", content)
	}
	if !strings.Contains(content, "stream") {
		t.Errorf("cli/streaming missing stream guidance: %s", content)
	}
}

func TestReadTopicCLIColor(t *testing.T) {
	content, ok, err := readTopic("cli/color")
	if err != nil {
		t.Fatalf("readTopic(cli/color): %v", err)
	}
	if !ok {
		t.Fatal("expected ok for cli/color")
	}
	if !strings.Contains(content, "go-best-practice/cli/color") {
		t.Errorf("cli/color missing nested frontmatter name: %s", content)
	}
}

func TestReadTopicCLIProjectLayout(t *testing.T) {
	content, ok, err := readTopic("cli/project-layout")
	if err != nil {
		t.Fatalf("readTopic(cli/project-layout): %v", err)
	}
	if !ok {
		t.Fatal("expected ok for cli/project-layout")
	}
	if !strings.Contains(content, "go-best-practice/cli/project-layout") {
		t.Errorf("cli/project-layout missing nested frontmatter name: %s", content)
	}
	if !strings.Contains(content, "kool create go-cli") && !strings.Contains(content, "go-cli") {
		t.Errorf("cli/project-layout missing go-cli guidance: %s", content)
	}
	if !strings.Contains(content, "run.Main") && !strings.Contains(content, "run/") {
		t.Errorf("cli/project-layout missing run package guidance: %s", content)
	}
}

func TestReadTopicCLIParent(t *testing.T) {
	content, ok, err := readTopic("cli")
	if err != nil {
		t.Fatalf("readTopic(cli): %v", err)
	}
	if !ok {
		t.Fatal("expected ok for cli")
	}
	if !strings.Contains(content, "go-best-practice/cli") {
		t.Errorf("cli missing frontmatter name: %s", content)
	}
}

func TestHandleSkillCLITopicViaShow(t *testing.T) {
	output, err := captureStdout(t, func() error {
		return handle([]string{"skill", "--show", "cli/skill-cli"})
	})
	if err != nil {
		t.Fatalf("handle(skill --show cli/skill-cli): %v", err)
	}
	if !strings.Contains(output, "skill-cli") {
		t.Errorf("cli/skill-cli output missing expected content: %s", output)
	}
	if !strings.Contains(output, "go-best-practice/cli/skill-cli") {
		t.Errorf("missing nested name: %s", output)
	}
}

func TestHandleBareTopicPathRejected(t *testing.T) {
	err := handle([]string{"cli/skill-cli"})
	if err == nil {
		t.Fatal("expected error for bare topic path")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestReadTopicExistingSubTopic(t *testing.T) {
	content, ok, err := readTopic("flags-parsing/types")
	if err != nil {
		t.Fatalf("readTopic(flags-parsing/types): %v", err)
	}
	if !ok {
		t.Fatal("expected ok for flags-parsing/types")
	}
	if !strings.Contains(content, "**T") {
		t.Errorf("unexpected content for flags-parsing/types: %s", content)
	}
}

func TestReadTopicNonExistent(t *testing.T) {
	_, ok, err := readTopic("nonexistent")
	if err != nil {
		t.Fatalf("readTopic(nonexistent): %v", err)
	}
	if ok {
		t.Error("expected not ok for nonexistent topic")
	}
}

func TestReadTopicEmpty(t *testing.T) {
	_, ok, err := readTopic("")
	if err != nil {
		t.Fatalf("readTopic(''): %v", err)
	}
	if ok {
		t.Error("expected not ok for empty topic")
	}
}

func TestValidateSegmentsOK(t *testing.T) {
	tests := []struct {
		segments []string
	}{
		{[]string{"kool-create"}},
		{[]string{"flags-parsing", "types"}},
		{[]string{"a"}},
	}
	for _, tt := range tests {
		if err := validateSegments(tt.segments); err != nil {
			t.Errorf("validateSegments(%v): unexpected error: %v", tt.segments, err)
		}
	}
}

func TestValidateSegmentsReject(t *testing.T) {
	tests := []struct {
		segments []string
	}{
		{[]string{""}},
		{[]string{"."}},
		{[]string{".."}},
		{[]string{"a", ""}},
		{[]string{"a", ".."}},
	}
	for _, tt := range tests {
		if err := validateSegments(tt.segments); err == nil {
			t.Errorf("validateSegments(%v): expected error, got nil", tt.segments)
		}
	}
}

func TestPrintTopicIndex(t *testing.T) {
	output, err := captureStdout(t, func() error {
		return printTopicIndex()
	})
	if err != nil {
		t.Fatalf("printTopicIndex: %v", err)
	}
	if !strings.Contains(output, "Available topics:") {
		t.Errorf("missing header in output: %s", output)
	}
	if !strings.Contains(output, "kool-create") {
		t.Errorf("topics output missing kool-create: %s", output)
	}
	if !strings.Contains(output, "cli") {
		t.Errorf("topics output missing cli: %s", output)
	}
	if !strings.Contains(output, "skill-cli") {
		t.Errorf("topics output missing skill-cli: %s", output)
	}
}

func TestHelpText(t *testing.T) {
	if !strings.Contains(help, "go-best-practice") {
		t.Error("help text missing go-best-practice reference")
	}
	if !strings.Contains(help, "install") {
		t.Error("help text missing install command")
	}
	if !strings.Contains(help, "topics") {
		t.Error("help text missing topics command")
	}
	if !strings.Contains(help, "skill --show") {
		t.Error("help text missing skill --show")
	}
}

func TestCollectTopicFiles(t *testing.T) {
	files, err := collectTopicFiles()
	if err != nil {
		t.Fatalf("collectTopicFiles: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("expected at least one topic file")
	}
	hasCmdExec := false
	hasSkillCLI := false
	for _, f := range files {
		if f.Path == "SKILL.md" || f.Path == "TOPIC.md" {
			t.Error("topic files should not include root skill index")
		}
		if f.Path == "cmd-exec/TOPIC.md" {
			hasCmdExec = true
		}
		if f.Path == "cli/skill-cli/TOPIC.md" {
			hasSkillCLI = true
		}
		if len(f.Content) == 0 {
			t.Errorf("empty content for %s", f.Path)
		}
	}
	if !hasCmdExec {
		t.Error("missing cmd-exec/TOPIC.md in collected files")
	}
	if !hasSkillCLI {
		t.Error("missing cli/skill-cli/TOPIC.md in collected files")
	}
}

func TestHandleHelp(t *testing.T) {
	output, err := captureStdout(t, func() error {
		return handle([]string{"-h"})
	})
	if err != nil {
		t.Fatalf("handle(-h): %v", err)
	}
	if !strings.Contains(output, "Usage:") {
		t.Errorf("help output missing Usage: %s", output)
	}
}

func TestHandleTopics(t *testing.T) {
	output, err := captureStdout(t, func() error {
		return handle([]string{"topics"})
	})
	if err != nil {
		t.Fatalf("handle(topics): %v", err)
	}
	if !strings.Contains(output, "Available topics:") {
		t.Errorf("topics output missing header: %s", output)
	}
}

func TestHandleUnknownCommand(t *testing.T) {
	err := handle([]string{"unknown-cmd"})
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestHandleNoArgs(t *testing.T) {
	output, err := captureStdout(t, func() error {
		return handle(nil)
	})
	if err != nil {
		t.Fatalf("handle(nil): %v", err)
	}
	if !strings.Contains(output, "Usage:") {
		t.Errorf("no-args output missing Usage: %s", output)
	}
	if !strings.Contains(output, "Available topics:") {
		t.Errorf("no-args output missing topic index: %s", output)
	}
}

func TestEmbeddedSkillMDNoInstallGuidelines(t *testing.T) {
	forbidden := []string{
		"skill install",
		"skill show",
		"install --cursor",
		"install --global",
	}
	lower := strings.ToLower(skill.Root)
	for _, phrase := range forbidden {
		if strings.Contains(lower, phrase) {
			t.Errorf("SKILL.md must not document CLI install/show plumbing (%q found); use --help and README instead", phrase)
		}
	}
}

func TestHandleSkillShow(t *testing.T) {
	output, err := captureStdout(t, func() error {
		return handle([]string{"skill", "--show"})
	})
	if err != nil {
		t.Fatalf("handle(skill --show): %v", err)
	}
	if !strings.Contains(output, "go-best-practice") {
		t.Errorf("skill --show output missing skill name: %s", output)
	}
}

func TestHandleSkillShowNestedBothOrders(t *testing.T) {
	for _, args := range [][]string{
		{"skill", "--show", "cli/skill-cli"},
		{"skill", "cli/skill-cli", "--show"},
	} {
		output, err := captureStdout(t, func() error {
			return handle(args)
		})
		if err != nil {
			t.Fatalf("handle(%v): %v", args, err)
		}
		if !strings.Contains(output, "go-best-practice/cli/skill-cli") {
			t.Errorf("handle(%v) missing nested name: %s", args, output)
		}
	}
}

func TestHandleSkillMissingAction(t *testing.T) {
	err := handle([]string{"skill"})
	if err == nil {
		t.Fatal("expected error for skill without action flags")
	}
	if !strings.Contains(err.Error(), "--show") && !strings.Contains(err.Error(), "--install") {
		t.Errorf("error should mention action flags: %v", err)
	}
}

func TestHandleLegacySkillShow(t *testing.T) {
	err := handle([]string{"skill", "show"})
	if err == nil {
		t.Fatal("expected error for legacy skill show")
	}
}

func TestHandleTopLevelInstallAlias(t *testing.T) {
	output, err := captureStdout(t, func() error {
		return handle([]string{"install", "--dry-run"})
	})
	if err != nil {
		t.Fatalf("handle(install --dry-run): %v", err)
	}
	if !strings.Contains(output, "[dry-run]") {
		t.Errorf("missing dry-run output: %s", output)
	}
	if !strings.Contains(output, ".agents/skills/go-best-practice") {
		t.Errorf("missing default target: %s", output)
	}
}

func captureStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	oldStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}

	stdoutCh := make(chan []byte, 1)
	readErrCh := make(chan error, 1)
	go func() {
		data, readErr := io.ReadAll(reader)
		stdoutCh <- data
		readErrCh <- readErr
	}()

	os.Stdout = writer
	runErr := fn()
	os.Stdout = oldStdout
	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	data := <-stdoutCh
	if err := <-readErrCh; err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return string(data), runErr
}
