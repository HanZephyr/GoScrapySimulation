```go
package main

// 导入 GoScrapySimulation
import (
	"GoScrapySimulation"
	"GoScrapySimulation/engine"
	"fmt"
	"github.com/antchfx/htmlquery"
	"strings"
)

// GithubProjectInfo 声明数据的结构
type GithubProjectInfo struct {
	name string
	star string
	fork string
}

func (info GithubProjectInfo) String() string {
	return fmt.Sprintf("name: %s , star: %s , fork: %s", info.name, info.star, info.fork)
}

type MyPipeline struct {
	GoScrapySimulation.DefaultPipeline
}

func (defaultPipeline *MyPipeline) ProcessItemFunc(parameter interface{}, dataItem interface{}) interface{}  {
	info := dataItem.(GithubProjectInfo)
	fmt.Printf("name: %s, star: %s, fork: %s\n", info.name, info.star, info.fork)
	return nil
}

// ParseFunc 编写请求到网页内容后的解析方法，通过 ItemDataChannel 将 数据传输给 Pipeline 进行输出
func ParseFunc(content []byte, UrlItemChannel chan GoScrapySimulation.RequestItem, ItemDataChannel chan interface{}) {
	docParser, err := htmlquery.Parse(strings.NewReader(string(content)))
	if err != nil { return }
	name := htmlquery.FindOne(docParser, `//a[@data-pjax="#js-repo-pjax-container"]`)
	staredCount := htmlquery.FindOne(docParser, `//a[contains(@aria-label, "starred")]`)
	forkedCount := htmlquery.FindOne(docParser, `//a[contains(@aria-label, "forked")]`)
	ItemDataChannel <- GithubProjectInfo{
		name: strings.TrimSpace(htmlquery.InnerText(name)),
		star: strings.TrimSpace(htmlquery.InnerText(staredCount)),
		fork: strings.TrimSpace(htmlquery.InnerText(forkedCount)),
	}
}


func main() {
	// 将需要请求的 URL 以及其他参数 构造成 RequestItem，并存储到 RequestItem 类型的切片中
	requestItemList := []GoScrapySimulation.RequestItem{{
		Url:"https://github.com/allwaysLove/GoScrapySimulation",
		Parser: ParseFunc,
	}, {
		Url:"https://github.com/allwaysLove/SchoolAssignmentManageSystem",
		Method: "Get",
		Parser: ParseFunc,
	}}
	myPipeline := MyPipeline{}
	// 创建一个爬虫引擎，并传入配置
	MyEngine := engine.NewEngine(
		// 起始 RequestItem 列表
		engine.SetStartRequestItemList(requestItemList),
		engine.SetPipeline(&myPipeline),
		// 设定最大的并发数量
		engine.SetConcurrentCount(200),
		// 设定超时时间，当超过设定时间，仍没有任何新任务出现时，便会结束爬虫
		engine.SetQuiteSpiderTimeout(5),
	)
	// 启动爬虫引擎
	MyEngine.StartCrawler()
}
```