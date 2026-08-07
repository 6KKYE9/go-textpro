// go-textpro 一个纯文本处理工具箱，把 awk/sed/cut 那些常用操作收进一个二进制。
// 子命令：
//
//	grep     按正则/子串过滤行（-v 取反，-i 忽略大小写）
//	replace  按正则/子串替换
//	column   按分隔符把文本转成对齐的表格
//	cut      取指定列（类似 cut -f）
//	sort     按行排序（-n 数字序，-r 反转）
//	unique   去重（-c 统计次数）
//	align    把多行按某列左对齐
//
// 输入输出走标准输入/标准输出，方便接管道。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printTextUsage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	var err error
	switch cmd {
	case "grep":
		err = tpGrep(args)
	case "replace":
		err = tpReplace(args)
	case "column":
		err = tpColumn(args)
	case "cut":
		err = tpCut(args)
	case "sort":
		err = tpSort(args)
	case "unique":
		err = tpUnique(args)
	case "align":
		err = tpAlign(args)
	case "help", "-h", "--help":
		printTextUsage()
	default:
		fmt.Fprintf(os.Stderr, "未知子命令: %s\n", cmd)
		printTextUsage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "错误:", err)
		os.Exit(1)
	}
}

func printTextUsage() {
	fmt.Println(`go-textpro 文本处理工具箱

用法:
  go-textpro <子命令> [参数]

子命令:
  grep    [-i] [-v] -p <模式>          过滤匹配行
  replace -p <模式> -r <替换串>        正则替换
  column  [-d 分隔符]                  对齐成表格
  cut     -f <列号> [-d 分隔符]        取列（从 1 计）
  sort    [-n] [-r]                    排序
  unique  [-c]                         去重（可计数）
  align   -f <列号> [-d 分隔符]        按列对齐`)
}

func readLines() ([]string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	s := strings.TrimRight(string(data), "\n")
	if s == "" {
		return []string{}, nil
	}
	return strings.Split(s, "\n"), nil
}

func writeLines(lines []string) {
	for _, l := range lines {
		fmt.Println(l)
	}
}

// ----- grep -----
func tpGrep(args []string) error {
	fs := flag.NewFlagSet("grep", flag.ContinueOnError)
	pat := fs.String("p", "", "匹配模式（正则）")
	insensitive := fs.Bool("i", false, "忽略大小写")
	invert := fs.Bool("v", false, "取反（不匹配的行）")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *pat == "" {
		return fmt.Errorf("需要 -p 模式")
	}
	re, err := regexp.Compile(*pat)
	if err != nil {
		return err
	}
	if *insensitive {
		re = regexp.MustCompile("(?i)" + *pat)
	}
	lines, err := readLines()
	if err != nil {
		return err
	}
	for _, l := range lines {
		match := re.MatchString(l)
		if match != *invert {
			fmt.Println(l)
		}
	}
	return nil
}

// ----- replace -----
func tpReplace(args []string) error {
	fs := flag.NewFlagSet("replace", flag.ContinueOnError)
	pat := fs.String("p", "", "匹配模式（正则）")
	rep := fs.String("r", "", "替换串")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *pat == "" {
		return fmt.Errorf("需要 -p 模式")
	}
	re, err := regexp.Compile(*pat)
	if err != nil {
		return err
	}
	lines, err := readLines()
	if err != nil {
		return err
	}
	for _, l := range lines {
		fmt.Println(re.ReplaceAllString(l, *rep))
	}
	return nil
}

// ----- column -----
func tpColumn(args []string) error {
	fs := flag.NewFlagSet("column", flag.ContinueOnError)
	delim := fs.String("d", "\t", "分隔符")
	if err := fs.Parse(args); err != nil {
		return err
	}
	lines, err := readLines()
	if err != nil {
		return err
	}
	cols := columnize(lines, *delim)
	for _, l := range cols {
		fmt.Println(l)
	}
	return nil
}

// columnize 把每行按分隔符拆成多列，再按每列最大宽度补空格对齐。
func columnize(lines []string, delim string) []string {
	rows := make([][]string, 0, len(lines))
	maxCols := 0
	for _, l := range lines {
		f := strings.Split(l, delim)
		if len(f) > maxCols {
			maxCols = len(f)
		}
		rows = append(rows, f)
	}
	widths := make([]int, maxCols)
	for _, r := range rows {
		for i, c := range r {
			if len(c) > widths[i] {
				widths[i] = len(c)
			}
		}
	}
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		var parts []string
		for i, c := range r {
			if i < maxCols {
				parts = append(parts, padRight(c, widths[i]))
			}
		}
		out = append(out, strings.Join(parts, "  "))
	}
	return out
}

func padRight(s string, n int) string {
	if len(s) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(s))
}

// ----- cut -----
func tpCut(args []string) error {
	fs := flag.NewFlagSet("cut", flag.ContinueOnError)
	field := fs.Int("f", 1, "列号（从 1 计）")
	delim := fs.String("d", "\t", "分隔符")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *field < 1 {
		return fmt.Errorf("-f 必须 >= 1")
	}
	lines, err := readLines()
	if err != nil {
		return err
	}
	for _, l := range lines {
		f := strings.Split(l, *delim)
		if *field <= len(f) {
			fmt.Println(f[*field-1])
		}
	}
	return nil
}

// ----- sort -----
func tpSort(args []string) error {
	fs := flag.NewFlagSet("sort", flag.ContinueOnError)
	numeric := fs.Bool("n", false, "数字序")
	reverse := fs.Bool("r", false, "反转")
	if err := fs.Parse(args); err != nil {
		return err
	}
	lines, err := readLines()
	if err != nil {
		return err
	}
	if *numeric {
		sort.SliceStable(lines, func(i, j int) bool {
			vi, ej := atoiSafe(lines[i])
			vj, ei := atoiSafe(lines[j])
			if ej == nil && ei == nil {
				if *reverse {
					return vi > vj
				}
				return vi < vj
			}
			// 有一方不是数字就退化为字符串比较
			if *reverse {
				return lines[i] > lines[j]
			}
			return lines[i] < lines[j]
		})
	} else {
		sort.SliceStable(lines, func(i, j int) bool {
			if *reverse {
				return lines[i] > lines[j]
			}
			return lines[i] < lines[j]
		})
	}
	writeLines(lines)
	return nil
}

func atoiSafe(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

// ----- unique -----
func tpUnique(args []string) error {
	fs := flag.NewFlagSet("unique", flag.ContinueOnError)
	count := fs.Bool("c", false, "统计出现次数")
	if err := fs.Parse(args); err != nil {
		return err
	}
	lines, err := readLines()
	if err != nil {
		return err
	}
	counts := map[string]int{}
	var order []string
	for _, l := range lines {
		if _, ok := counts[l]; !ok {
			order = append(order, l)
		}
		counts[l]++
	}
	for _, l := range order {
		if *count {
			fmt.Printf("%5d %s\n", counts[l], l)
		} else {
			fmt.Println(l)
		}
	}
	return nil
}

// ----- align -----
func tpAlign(args []string) error {
	fs := flag.NewFlagSet("align", flag.ContinueOnError)
	field := fs.Int("f", 1, "按第几列对齐（从 1 计）")
	delim := fs.String("d", " ", "分隔符")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *field < 1 {
		return fmt.Errorf("-f 必须 >= 1")
	}
	lines, err := readLines()
	if err != nil {
		return err
	}
	maxW := 0
	split := make([][]string, len(lines))
	for i, l := range lines {
		f := strings.Split(l, *delim)
		split[i] = f
		if *field <= len(f) && len(f[*field-1]) > maxW {
			maxW = len(f[*field-1])
		}
	}
	for _, f := range split {
		if *field <= len(f) {
			f[*field-1] = padRight(f[*field-1], maxW)
		}
		fmt.Println(strings.Join(f, *delim))
	}
	return nil
}
