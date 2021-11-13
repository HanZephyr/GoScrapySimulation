# GoScrapySimulation

[![Release latest](https://img.shields.io/badge/Release-latest-blue.svg?style=flat-square)](https://github.com/allwaysLove/GoScrapySimulation/releases) [![MIT License](https://img.shields.io/badge/LICENSE-MIT-yellow.svg?style=flat-square)](https://github.com/allwaysLove/GoScrapySimulation/main/LICENSE)


## 概述

**[GoScrapySimulation](https://github.com/allwaysLove/GoScrapySimulation)** 基于 Go 语言编写的类似于 Python 的 Scrapy 框架的开源网络爬虫框架


## 安装

```shell
go get github.com/allwaysLove/GoScrapySimulation
```

## 使用示例

```go
package main

import (
   "GoScrapySimulation"
   "GoScrapySimulation/engine"
   "github.com/antchfx/htmlquery"
   "strings"
)

// GithubProjectInfo 声明数据的结构
type GithubProjectInfo struct {
   star string
   fork string
}

// ParseFunc 编写请求到网页内容后的解析方法，通过 ItemDataChannel 将 数据传输给 Pipeline 进行输出
func ParseFunc(content []byte, UrlItemChannel chan GoScrapySimulation.RequestItem, ItemDataChannel chan interface{}) {
   docParser, err := htmlquery.Parse(strings.NewReader(string(content)))
   if err != nil {
      return
   }
   staredCount := htmlquery.FindOne(docParser, `//a[contains(@aria-label, "starred")]`)
   forkedCount := htmlquery.FindOne(docParser, `//a[contains(@aria-label, "forked")]`)
   ItemDataChannel <- GithubProjectInfo{
      star: strings.TrimSpace(htmlquery.InnerText(staredCount)),
      fork: strings.TrimSpace(htmlquery.InnerText(forkedCount)),
   }
}

func main() {
   // 将需要请求的 URL 以及其他参数 构造成 RequestItem，并存储到 RequestItem 类型的切片中
   requestItemList := []GoScrapySimulation.RequestItem{{
      Url:    "https://github.com/allwaysLove/GoScrapySimulation",
      Parser: ParseFunc,
   }, {
      Url:    "https://github.com/allwaysLove/SchoolAssignmentManageSystem",
      Method: "Get",
      Parser: ParseFunc,
   }}
   // 创建一个爬虫引擎，并传入配置
   MyEngine := engine.NewEngine(
      // 起始 RequestItem 列表
      engine.SetStartRequestItemList(requestItemList),
      // 设定最大的并发数量
      engine.SetConcurrentCount(200),
      // 设定超时时间，当超过设定时间，仍没有任何新任务出现时，便会结束爬虫
      engine.SetQuiteSpiderTimeout(5),
   )
   // 启动爬虫引擎
   MyEngine.StartCrawler()
}
```
输出结果为
```shell
2021/11/14 01:35:40 Engine Config {
                        ConcurrentCount: 200
                        Pipeline: *GoScrapySimulation.DefaultPipeline, Memory Address: 0xc000050cd0
                        requestItemList: [
                            1. Url: https://github.com/allwaysLove/GoScrapySimulation, ParserFunc: main.ParseFunc
                            2. Url: https://github.com/allwaysLove/SchoolAssignmentManageSystem, ParserFunc: main.ParseFunc
                        ]
                    }
2021/11/14 01:35:40 Spider Running ...
2021/11/14 01:35:41 Fetch Request Url: https://github.com/allwaysLove/GoScrapySimulation
2021/11/14 01:35:41 ProcessItem: {1 0}
2021/11/14 01:35:42 Fetch Request Url: https://github.com/allwaysLove/SchoolAssignmentManageSystem
2021/11/14 01:35:42 ProcessItem: {8 2}
2021/11/14 01:35:47 Info Of Engine {
                        TotalRequestCount: 2
                        TotalRequestErrorCount: 0
                        StartTime: 2021-11-14 01:35:40
                        EndTime: 2021-11-14 01:35:47
                        SpiderTimeConsuming: 1.997866s
                        TotalTimeConsuming: 6.997866s
                    }
2021/11/14 01:35:47 Timeout 5s, exit the crawler
```
这个示例中，使用的是框架默认的 Pipeline（DefaultPipeline），这个 Pipeline 在接收到 Item 数据时，就会将数据输出的控制台中，即上述输出结果的 `ProcessItem`。但这只是简单的将结构体输出出来，如果希望获取到更详细的信息，有两种方法：
1. 实现自定义数据结构体的 String 方法
    ```go
    func (info GithubProjectInfo) String() string {
        return fmt.Sprintf("star: %s , fork: %s", info.star, info.fork)
    }
    ```
2. 定义自己的 Pipeline
   ```go
   // 自定义 Pipeline 结构体
   type MyPipeline struct {
       GoScrapySimulation.DefaultPipeline
   }
   // 编写对每条数据的处理方法
   func (defaultPipeline *MyPipeline) ProcessItemFunc(parameter interface{}, dataItem interface{}) interface{}  {
        info := dataItem.(GithubProjectInfo)
        fmt.Printf("name: %s, star: %s, fork: %s\n", info.name, info.star, info.fork)
        return nil
   }
   ```

此时对每条数据的输出格式就变成了
```shell
2021/11/14 01:53:33 ProcessItem: name: GoScrapySimulation , star: 1 , fork: 0
2021/11/14 01:53:34 ProcessItem: name: SchoolAssignmentManageSystem , star: 8 , fork: 22021/11/14 01:53:33 Fetch Request Url: https://github.com/allwaysLove/GoScrapySimulation
```

需要注意的是，自定义 Pipeline 需要实现 `GoScrapySimulation。Pipeline` 接口.该接口包含三个方法：`StartPipeline`, `ProcessItemFunc`, `ClosePipeline`，具体写法详见文件 `pipeline.go` 中 DefaultPipeline 的实现

而本例中因不需要在 Pipeline 启动和关闭时进行操作，故为了简化操作，直接继承了默认的 `DefaultPipeline`，而在实际项目中，可以独自实现接口而无需继承。

示例代码详见 example 目录