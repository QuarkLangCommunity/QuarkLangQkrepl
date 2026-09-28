// qkrepl: QuarkLang 交互式解释器（REPL）——工具链成员
//
// 用法: qkrepl
//
//	qk> int x = 1;            // 语句：累积执行（每次重放，只打印新增输出）
//	qk> fn add(int a, int b) int { return a + b; }   // 顶层声明
//	qk> :decls / :reset / :q
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"

	"quarklang/internal/lang"
)

var declPrefixes = []string{"program ", "fn ", "type ", "impl", "space ", "library ", "import ", "pub ", "#", "const "}

func isDecl(line string) bool {
	t := strings.TrimSpace(line)
	for _, p := range declPrefixes {
		if strings.HasPrefix(t, p) {
			return true
		}
	}
	return false
}

func bracesBalanced(lines []string) bool {
	depth := 0
	inStr, inRaw, inLine, inBlock := false, false, false, false
	for _, ln := range lines {
		for i := 0; i < len(ln); i++ {
			c := ln[i]
			switch {
			case inLine:
				continue
			case inBlock:
				if c == '*' && i+1 < len(ln) && ln[i+1] == '/' {
					inBlock = false
					i++
				}
			case inStr:
				if c == 0x5c {
					i++
				} else if c == '"' {
					inStr = false
				}
			case inRaw:
				if c == 0x60 {
					inRaw = false
				}
			default:
				switch {
				case c == '/' && i+1 < len(ln) && ln[i+1] == '/':
					inLine = true
				case c == '/' && i+1 < len(ln) && ln[i+1] == '*':
					inBlock = true
					i++
				case c == '"':
					inStr = true
				case c == 0x60:
					inRaw = true
				case c == '{':
					depth++
				case c == '}':
					depth--
				}
			}
		}
		inLine = false
	}
	return depth <= 0
}

func build(decls, stmts []string) string {
	var b strings.Builder
	b.WriteString("program main;\n")
	for _, d := range decls {
		b.WriteString(d)
		b.WriteString("\n")
	}
	b.WriteString("fn main(IOStream io) {\n")
	for _, s := range stmts {
		b.WriteString(s)
		b.WriteString("\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func run(src string) (string, error) {
	prog, err := lang.Compile(src)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := lang.Run(prog, "repl", nil, os.Stdin, &buf); err != nil {
		return buf.String(), err
	}
	return buf.String(), nil
}

func main() {
	var decls, stmts []string
	printed := 0
	in := bufio.NewScanner(os.Stdin)
	in.Buffer(make([]byte, 0, 1<<20), 1<<20)
	fmt.Println("QuarkLang REPL（:q 退出，:reset 清空，:decls 查看声明）")
	for {
		if len(stmts) == 0 && len(decls) == 0 {
			fmt.Print("qk> ")
		} else {
			fmt.Print("..> ")
		}
		if !in.Scan() {
			fmt.Println()
			return
		}
		line := in.Text()
		t := strings.TrimSpace(line)
		switch t {
		case ":q", ":quit", ":exit":
			return
		case ":reset":
			decls, stmts, printed = nil, nil, 0
			fmt.Println("已清空")
			continue
		case ":decls":
			for _, d := range decls {
				fmt.Println(strings.SplitN(d, "\n", 2)[0])
			}
			continue
		case "":
			continue
		}
		// 多行块：括号未闭合则继续读
		block := []string{line}
		for !bracesBalanced(block) {
			fmt.Print("...> ")
			if !in.Scan() {
				break
			}
			block = append(block, in.Text())
		}
		chunk := strings.Join(block, "\n")
		var nd, ns []string
		if isDecl(chunk) {
			nd = append(append([]string{}, decls...), chunk)
			ns = stmts
		} else {
			nd = decls
			ns = append(append([]string{}, stmts...), chunk)
		}
		src := build(nd, ns)
		out, err := run(src)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			continue
		}
		if len(out) > printed {
			fmt.Print(out[printed:])
		}
		printed = len(out)
		decls, stmts = nd, ns
	}
}
