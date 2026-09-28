# QuarkLangQkrepl

QuarkLang 交互式解释器（REPL）——工具链成员。

## 用法

```sh
go build -o qkrepl .      # go.mod 里 replace 指向 ../QuarkLang
qkrepl
```

```
QuarkLang REPL（:q 退出，:reset 清空，:decls 查看声明）
qk> int x = 1;
qk> x = x + 41;
qk> io.println(x);
42
qk> fn add(int a, int b) int { return a + b; }
qk> io.println(add(2, 3));
5
qk> :q
```

## 工作方式

- **顶层声明**（`fn`/`type`/`impl`/`space`/`library`/`import`/`pub`/`#macro`/`program`）累积到程序头部；
  其余输入作为 `main` 里的语句累积。
- 每次回车把「声明 + 全部语句」重新编译执行一遍，**只打印新增的输出**（所以状态是持久的，且不会重复刷屏）。
- 大括号未闭合时进入多行模式（提示符 `...>`），支持字符串/注释里的括号干扰。
- 命令：`:q` 退出、`:reset` 清空、`:decls` 列出已累积的顶层声明。

## 说明

REPL 直接链接解释器（`internal/lang`），所以语法/类型检查与 `quark` 完全一致；
错误按解释器原文打印，不中断会话。
