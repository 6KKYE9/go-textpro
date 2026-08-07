package main

import (
	"os"
	"strings"
	"testing"
)

func TestColumnize(t *testing.T) {
	lines := []string{"a\tbb\tc", "ddd\te\tff"}
	out := columnize(lines, "\t")
	if len(out) != 2 {
		t.Fatalf("应返回 2 行，实际 %d", len(out))
	}
	// 第一列 "a" 应对齐到宽度 3
	if !strings.HasPrefix(out[0], "a  ") {
		t.Fatalf("首行对齐异常: %q", out[0])
	}
}

func TestPadRight(t *testing.T) {
	if padRight("ab", 4) != "ab  " {
		t.Fatalf("padRight 异常: %q", padRight("ab", 4))
	}
	if padRight("abcd", 2) != "abcd" {
		t.Fatal("已超长不应再补空格")
	}
}

func TestAtoiSafe(t *testing.T) {
	v, err := atoiSafe(" 42 ")
	if err != nil || v != 42 {
		t.Fatalf("atoiSafe 异常: %d %v", v, err)
	}
	if _, err := atoiSafe("x"); err == nil {
		t.Fatal("非数字应报错")
	}
}

func TestReadLines(t *testing.T) {
	// 用临时改写：验证空输入返回空切片
	old := os.Stdin
	tmp, _ := os.CreateTemp("", "tp*.txt")
	tmp.WriteString("x\ny\n")
	tmp.Seek(0, 0)
	os.Stdin = tmp
	defer func() { os.Stdin = old; tmp.Close(); os.Remove(tmp.Name()) }()
	lines, err := readLines()
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 2 || lines[0] != "x" || lines[1] != "y" {
		t.Fatalf("readLines 异常: %v", lines)
	}
}
