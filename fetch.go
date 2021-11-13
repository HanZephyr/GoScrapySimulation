package GoScrapySimulation

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

type myBuffer struct {
	bytes.Buffer
}

func (buffer myBuffer)Close() error { return nil }

func Request(requestItem RequestItem, ErrorRequestItemList *[]RequestItem, RequestItemChannel chan RequestItem, ItemDataChannel chan interface{}) {
	// 对出现错误的请求保存到 ErrorRequestItemList 中
	OccurredError := false
	defer func() {
		if OccurredError {
			*ErrorRequestItemList = append(*ErrorRequestItemList, requestItem)
		}
	}()

	// 处理 Query Params
	queryParams := url.Values{}
	Url, err := url.Parse(requestItem.Url)
	if err != nil {
		OccurredError = true
		return
	}
	for key, value := range requestItem.Query {
		queryParams.Set(key, fmt.Sprintf("%v", value))
	}
	// 如果参数中有中文参数,这个方法会进行 URLEncode
	Url.RawQuery = queryParams.Encode()
	EncodedUrl := Url.String()

	// 创建 Request 对象
	request, _ := http.NewRequest(strings.ToUpper(requestItem.Method), EncodedUrl, nil)

	// 处理 Body Params
	switch requestItem.Body.(type) {
	case io.Reader:
		if !checkMapFieldIsDefined(requestItem.Header, "Content-Type") {
			request.Header.Set("Content-Type", "text/plain")
		}
		request.Body = requestItem.Body.(io.ReadCloser)
	// 普通字符串
	case string:
		request.Body = ioutil.NopCloser(strings.NewReader(requestItem.Body.(string)))
		if !checkMapFieldIsDefined(requestItem.Header, "Content-Type") {
			request.Header.Set("Content-Type", "text/plain")
		}
	// 键值对
	case map[string]interface{}:
		form := url.Values{}
		for key, value := range requestItem.Body.(map[string]interface{}) {
			form.Set(key, fmt.Sprintf("%v", value))
		}
		request.Body = ioutil.NopCloser(strings.NewReader(form.Encode()))
		if !checkMapFieldIsDefined(requestItem.Header, "Content-Type") {
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		}
	case []byte:
		request.Body = ioutil.NopCloser(bytes.NewReader(requestItem.Body.([]byte)))
		if !checkMapFieldIsDefined(requestItem.Header, "Content-Type") {
			request.Header.Set("Content-Type", "multipart/form-data")
		}
	case os.File:
		file := requestItem.Body.(os.File)
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {}
		}(&file)

		fileName := path.Base(file.Name())

		//创建一个模拟的form中的一个选项,这个form项现在是空的
		bodyBuf := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuf)

		//关键的一步操作, 设置文件的上传参数叫 file, 文件名是 fileName,
		//相当于现在还没选择文件, form项里选择文件的选项
		fileWriter, _ := bodyWriter.CreateFormFile("file", fileName)

		//io.Copy 这里相当于选择了文件,将文件放到 form 中
		_, err = io.Copy(fileWriter, &file)
		if err != nil {
			return
		}

		//这个很关键,必须这样写关闭,不能使用defer关闭,不然会导致错误
		_ = bodyWriter.Close()

		request.Body = ioutil.NopCloser(bodyBuf)

		if !checkMapFieldIsDefined(requestItem.Header, "Content-Type") {
			request.Header.Set("Content-Type", bodyWriter.FormDataContentType())
		}
	}

	// 处理 Header Params
	for key, value := range requestItem.Header {
		request.Header.Set(key, fmt.Sprintf("%v", value))
	}

	response, err := http.DefaultClient.Do(request)
	defer func() {
		log.Printf("Fetch Request Url: %s", requestItem.Url)
		err = response.Body.Close()
		if err != nil {}
	}()

	if err != nil || response.StatusCode != http.StatusOK {
		OccurredError = true
		return
	}

	result, _ := ioutil.ReadAll(response.Body)

	go func() {
		requestItem.Parser(result, RequestItemChannel, ItemDataChannel)
	}()
}

func checkMapFieldIsDefined(mapItem map[string]interface{}, fieldName string) bool {
	if mapItem[fieldName] == nil || mapItem[fieldName] == "" {
		return false
	}
	return true
}

