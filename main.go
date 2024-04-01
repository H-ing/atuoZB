package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"flag"
	"regexp"
)

var (
	author  string
	since   string
	until   string
	items   string
	pattern string
	name    string
	process string
	logArg  string = `--pretty=format:"%s" `

	zbTpl string = `
【${projectName}】
一、本周项目实施情况汇总：
${projectName}进度${projectProcess}%。${totalMsg}。

二、下周计划开展内容：
${nextPlain}

三、需要协调处理的问题：
${needHelp}`

	detailTpl string = `
【${projectName}】
【完成功能清单】
${addList}

【修复bug清单】
${fixList}

【计划工作清单】
${todoList}
`

	tplParams map[string]string = make(map[string]string)
)

func main() {
	parseArg()
	readyPattern()
	logMsg := getGitLog()
	analyzeLog(logMsg)
	printZb()
}

func parseArg() {
	// 使用flag包定义命令行参数
	flag.StringVar(&author, "author", "", "")
	flag.StringVar(&since, "since", "", "")
	flag.StringVar(&until, "until", "", "")
	flag.StringVar(&name, "name", "", "")
	flag.StringVar(&process, "process", "", "")
	flag.StringVar(&items, "items", "add|fix|todo", "")

	// 解析命令行参数
	flag.Parse()

	if len(author) > 0 {
		logArg += "--author=" + author
	}

	if len(since) > 0 {
		logArg += "--since=" + since
	}

	if len(until) > 0 {
		logArg += "--until=" + until
	}

	tplParams["projectName"] = name
	tplParams["projectProcess"] = process
}

func readyPattern() {
	pattern = fmt.Sprintf(`(?m)(%s)(:|：).*?(\s|$)`, items)
}

func getGitLog() string {
	cmd := exec.Command("git", "log", logArg)
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing git status:", err)
		return ""
	}
	// fmt.Println(string(output))

	logMsg := string(output)
	logMsg = strings.ReplaceAll(logMsg, "\"", "")
	logMsg = strings.ReplaceAll(logMsg, "：", ":")

	return logMsg
}

func analyzeLog(logMsg string) {
	re, _ := regexp.Compile(pattern)

	// 使用FindAllString查找所有匹配的字符串
	matches := re.FindAllString(logMsg, -1)

	addTotal := 0
	addList := make([]string, 0)
	fixTotal := 0
	fixList := make([]string, 0)
	todoTotal := 0
	todoList := make([]string, 0)
	// 打印匹配结果
	for _, match := range matches {
		split := strings.Split(match, ":")
		dataType := split[0]
		msg := split[1]

		if dataType == "add" {
			addTotal++
			addList = append(addList, msg)
		}

		if dataType == "fix" {
			fixTotal++
			fixList = append(fixList, msg)
		}

		if dataType == "todo" {
			todoTotal++
			todoList = append(todoList, msg)
		}
	}

	tplParams["totalMsg"] = fmt.Sprintf(`本周共完成%d个功能开发，修复了%d个bug`, addTotal, fixTotal)
	tplParams["addTotal"] = strconv.Itoa(addTotal)
	tplParams["fixTotal"] = strconv.Itoa(fixTotal)
	tplParams["todoTotal"] = strconv.Itoa(todoTotal)
	tplParams["addList"] = strings.Join(addList, "\n")
	tplParams["fixList"] = strings.Join(fixList, "\n")
	tplParams["todoList"] = strings.Join(todoList, "\n")
}

func printZb() {
	printMsg(zbTpl)
	printMsg(detailTpl)
}

func printMsg(tpl string) {
	re := regexp.MustCompile(`\${([^}]+)}`)
	matches := re.FindAllStringSubmatch(tpl, -1)
	// 打印匹配结果
	for _, match := range matches {
		replaceTarget := match[0]
		key := match[1]
		val := tplParams[key]
		tpl = strings.ReplaceAll(tpl, replaceTarget, val)
	}

	fmt.Println(tpl)
}
