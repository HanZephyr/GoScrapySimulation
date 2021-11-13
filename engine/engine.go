package engine

import (
	"GoScrapySimulation"
	"fmt"
	"log"
	"reflect"
	"time"
)

type InfoOfEngine struct {
	TotalRequestCount      int
	TotalRequestErrorCount int
	StartTime              time.Time
	EndTime            time.Time
	SpiderTimeConsuming time.Duration
	TotalTimeConsuming time.Duration
}

func (infoOfEngine InfoOfEngine) String() string {
	return "Info Of Engine {\n" +
		fmt.Sprintf("\t\t\t\t\t\tTotalRequestCount: %d\n", infoOfEngine.TotalRequestCount) +
		fmt.Sprintf("\t\t\t\t\t\tTotalRequestErrorCount: %d\n", infoOfEngine.TotalRequestErrorCount) +
		fmt.Sprintf("\t\t\t\t\t\tStartTime: %s\n", infoOfEngine.StartTime.Format("2006-01-02 15:04:05")) +
		fmt.Sprintf("\t\t\t\t\t\tEndTime: %s\n", infoOfEngine.EndTime.Format("2006-01-02 15:04:05")) +
		fmt.Sprintf("\t\t\t\t\t\tSpiderTimeConsuming: %s\n", infoOfEngine.SpiderTimeConsuming) +
		fmt.Sprintf("\t\t\t\t\t\tTotalTimeConsuming: %s\n", infoOfEngine.TotalTimeConsuming) +
		"\t\t\t\t\t}"
}

var infoOfEngine InfoOfEngine

// ConfigOfEngine 引擎配置
type ConfigOfEngine struct {
	requestItemList         []GoScrapySimulation.RequestItem
	ConcurrentCount         int
	Pipeline                GoScrapySimulation.Pipeline
	ErrorRequestItemHandler func(ErrorRequestItemList []GoScrapySimulation.RequestItem) interface{}
	QuiteSpiderTimeout      int
}

// ConfigOfEngine 信息
func (configOfEngine ConfigOfEngine) String() string {
	var requestItemString = "[\n"
	for index, requestItem := range configOfEngine.requestItemList {
		requestItemString += fmt.Sprintf("\t\t\t\t\t\t\t%d. %s\n", index+1, requestItem)
	}
	requestItemString += "\t\t\t\t\t\t]"
	return "Engine Config {\n" +
		fmt.Sprintf("\t\t\t\t\t\tConcurrentCount: %d\n", configOfEngine.ConcurrentCount) +
		fmt.Sprintf("\t\t\t\t\t\tPipeline: %s, Memory Address: %p\n", reflect.TypeOf(configOfEngine.Pipeline).String(), configOfEngine.Pipeline) +
		fmt.Sprintf("\t\t\t\t\t\trequestItemList: %s\n", requestItemString) +
		"\t\t\t\t\t}"
}

// OptionEngine 定义配置选项函数（关键）
type OptionEngine func(*ConfigOfEngine)

// DefaultRequestItemErrorHandler 默认的失败请求处理函数
func DefaultRequestItemErrorHandler(ErrorRequestItemList []GoScrapySimulation.RequestItem) interface{} {
	var resultString string
	if len(ErrorRequestItemList) > 0 {
		resultString += fmt.Sprintf("A total of %d URL requests failed, the details are as follows: {\n", len(ErrorRequestItemList))
		for index, ErrorItem := range ErrorRequestItemList {
			resultString += fmt.Sprintf("\t\t\t\t\t\t%d. %s\n", index+1, ErrorItem)
		}
		log.Printf("%s\t\t\t\t\t}\n", resultString)
	}
	return nil
}

// SetStartRequestItemList 设置起始 RequestItem 数组
// 返回一个 OptionEngine 类型的函数（闭包）：接收 ConfigOfEngine 类型指针参数并修改之
func SetStartRequestItemList(StartRequestItemList []GoScrapySimulation.RequestItem) OptionEngine {
	return func(this *ConfigOfEngine) {
		if StartRequestItemList != nil {
			this.requestItemList = StartRequestItemList
		}
	}
}

// SetConcurrentCount 设置最大并发数量
// 返回一个 OptionEngine 类型的函数（闭包）：接收 ConfigOfEngine 类型指针参数并修改之
func SetConcurrentCount(ConcurrentCount int) OptionEngine {
	return func(this *ConfigOfEngine) {
		if ConcurrentCount >= 0 {
			this.ConcurrentCount = ConcurrentCount
		}
	}
}

// SetPipeline 设定数据输出管道处理类
// 返回一个 OptionEngine 类型的函数（闭包）：接收 ConfigOfEngine 类型指针参数并修改之
func SetPipeline(Pipeline GoScrapySimulation.Pipeline) OptionEngine {
	return func(this *ConfigOfEngine) {
		if Pipeline != nil {
			this.Pipeline = Pipeline
		}
	}
}

// SetErrorRequestItemHandler 设定数据输出管道处理类
// 返回一个 OptionEngine 类型的函数（闭包）：接收 ConfigOfEngine 类型指针参数并修改之
func SetErrorRequestItemHandler(function func([]GoScrapySimulation.RequestItem) interface{}) OptionEngine {
	return func(this *ConfigOfEngine) {
		if function != nil {
			this.ErrorRequestItemHandler = function
		}
	}
}

// SetQuiteSpiderTimeout 设定 Channel 超时多长时间未变化后，退出爬虫
// 返回一个 OptionEngine 类型的函数（闭包）：接收 ConfigOfEngine 类型指针参数并修改之
func SetQuiteSpiderTimeout(timeout int) OptionEngine {
	return func(this *ConfigOfEngine) {
		if timeout > 0 {
			this.QuiteSpiderTimeout = timeout
		}
	}
}

// NewEngine 创建一个爬虫引擎实例
func NewEngine(opts ...OptionEngine) ConfigOfEngine {
	// 初始化默认值
	defaultEngine := ConfigOfEngine{
		requestItemList:         []GoScrapySimulation.RequestItem{},
		ErrorRequestItemHandler: DefaultRequestItemErrorHandler,
		Pipeline:                &GoScrapySimulation.DefaultPipeline{},
		QuiteSpiderTimeout:      5,
	}
	// 依次调用 opts 函数列表中的函数，为结构体成员赋值
	for _, o := range opts {
		o(&defaultEngine)
	}
	return defaultEngine
}

func (configOfEngine ConfigOfEngine) StartCrawler() {
	infoOfEngine.StartTime = time.Now()

	var QuitType = "Unknown"
	defer func() {
		switch QuitType {
		case "QuitSignal":
			log.Print("QuitChannel command is received, exit the crawler")
		case "Timeout":
			log.Printf("Timeout %ds, exit the crawler", configOfEngine.QuiteSpiderTimeout)
		case "Unknown":
			log.Print("Unknown circumstances cause the crawler to exit")
		}
	}()

	// 输出爬虫初始配置信息
	log.Print(configOfEngine)
	// 创建 存储 RequestItem 的 channel
	RequestItemChannel := make(chan GoScrapySimulation.RequestItem, configOfEngine.ConcurrentCount)
	// 创建存储请求失败的 RequestItem 的切片
	var ErrorRequestItemList []GoScrapySimulation.RequestItem

	// 爬虫关闭时执行的操作
	defer func() {
		infoOfEngine.TotalRequestErrorCount = len(ErrorRequestItemList)
		infoOfEngine.EndTime = time.Now()
		infoOfEngine.SpiderTimeConsuming = infoOfEngine.EndTime.Sub(infoOfEngine.StartTime) - time.Second * time.Duration(configOfEngine.QuiteSpiderTimeout)
		infoOfEngine.TotalTimeConsuming = infoOfEngine.EndTime.Sub(infoOfEngine.StartTime)
		log.Print(infoOfEngine)
	}()

	// 程序退出时，将所有错误信息输出到文件
	defer func() {
		configOfEngine.ErrorRequestItemHandler(ErrorRequestItemList)
	}()

	// 创建 存储结果数据 Channel
	ItemDataChannel := make(chan interface{}, configOfEngine.ConcurrentCount)
	// 创建退出 Channel
	QuitChannel := make(chan interface{})

	// 爬虫开始前，创建 Pipeline 管道对象
	configOfEngine.Pipeline.StartPipeline("")

	// 爬虫结束后，退出时执行 CloseSpider 方法，做收尾工作
	defer configOfEngine.Pipeline.ClosePipeline("")

	// 将起始 requestItemList 中的所有 RequestItem 推入 requestItemListChannel
	for _, requestItem := range configOfEngine.requestItemList {
		RequestItemChannel <- requestItem
	}

	log.Print("Spider Running ...")
	// 启动爬虫
	for {
		select {
		// 读取新 RequestItem 调用 网络下载器进行下载
		case requestItem := <-RequestItemChannel:
			infoOfEngine.TotalRequestCount += 1
			go GoScrapySimulation.Request(requestItem, &ErrorRequestItemList, RequestItemChannel, ItemDataChannel)
		case itemData := <-ItemDataChannel:
			go configOfEngine.Pipeline.ProcessItemFunc("", itemData)
		// 当 QuitChannel 传入任意数据时，结束爬虫
		case <-QuitChannel:
			QuitType = "QuitSignal"
			return
		// 当无新数据超过 设定的 QuiteSpiderTimeout 时，结束爬虫
		case <-time.After(time.Second * time.Duration(configOfEngine.QuiteSpiderTimeout)):
			QuitType = "Timeout"
			return
		}
	}
}
