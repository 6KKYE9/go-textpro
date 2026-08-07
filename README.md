# go-textpro

纯文本处理工具箱，把 awk/sed/cut 那些常用操作收进一个二进制。输入输出走标准输入/标准输出，方便接管道。

## 安装

```bash
go build -o go-textpro.exe
```

## 子命令

### grep

```bash
cat log.txt | go-textpro grep -p "ERROR"        # 匹配行
go-textpro grep -p "err" -i -v access.log       # 忽略大小写、取反
```

### replace

```bash
cat a.txt | go-textpro replace -p "foo" -r "bar"
```

### column

```bash
printf "a\tb\nccc\td" | go-textpro column       # 按分隔符对齐成表格
```

### cut

```bash
cat tsv.txt | go-textpro cut -f 2 -d ","        # 取第 2 列
```

### sort

```bash
cat nums.txt | go-textpro sort -n -r            # 数字序、反转
```

### unique

```bash
cat list.txt | go-textpro unique                # 去重
go-textpro unique -c list.txt                   # 去重并计数
```

### align

```bash
cat data.txt | go-textpro align -f 1            # 按第 1 列左对齐
```

## 说明

零依赖纯 Go。分隔符默认是制表符，除 `grep/replace` 的模式是正则外，其余按字面处理。`column` 和 `align` 都能做列对齐，区别在前者把所有列都对齐、后者只对齐指定列。
