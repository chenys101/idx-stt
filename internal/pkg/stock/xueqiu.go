package stock

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// 发送 HTTP 请求获取页面内容
func getPageContent(url string) ([]byte, error) {
	// 创建一个新的 HTTP 请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头，模拟浏览器行为
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	// 创建一个 HTTP 客户端
	client := &http.Client{}
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 读取响应体内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// 解析 HTML 内容，提取帖子信息
func parsePosts(htmlContent []byte) {
	// 使用 goquery 加载 HTML 内容
	doc, err := goquery.NewDocumentFromReader(ioutil.NopCloser(bytes.NewReader(htmlContent)))
	if err != nil {
		fmt.Println("Error parsing HTML:", err)
		return
	}

	// 这里需要根据雪球网实际的 HTML 结构来选择合适的选择器
	// 以下选择器仅为示例，需要根据实际情况调整
	doc.Find(".post-item").Each(func(i int, s *goquery.Selection) {
		// 提取帖子标题
		title := s.Find(".post-title").Text()
		// 提取帖子作者
		author := s.Find(".post-author").Text()

		fmt.Printf("Title: %s\nAuthor: %s\n\n", title, author)
	})
}
