package GoScrapySimulation

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
)

type ParserFunc interface{}

type RequestHeader map[string]interface{}
type RequestQuery map[string]interface{}

type RequestItem struct {
	// 访问网址
	Url string
	// 请求方法
	Method string
	// 请求参数
	Header RequestHeader
	Query RequestQuery
	Cookie http.Cookie
	Body interface{} // 可以为 string， *io.Reader, bytes(文件), *os.File 文件(ioutil.ReadAll(fp))

	// 自定义参数（传递指针变量，便于在不同请求和解析函数间做数据传输）
	Meta *map[string]string
	// 对应的网页解析函数
	Parser func(content []byte, UrlItemChannel chan RequestItem, ItemDataChannel chan interface{})
}

func (urlItem RequestItem) String() string {
	return fmt.Sprintf("Url: %s, ParserFunc: %s", urlItem.Url, runtime.FuncForPC(reflect.ValueOf(urlItem.Parser).Pointer()).Name())
}
