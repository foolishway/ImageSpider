package main

import (
	"fmt"
	"io"
	"strings"
	"net/url"
	"os"
	"io/ioutil"
	"log"
	"net/http"
	"flag"
	"encoding/json"
	"sync"
	"strconv"
	"time"
)
type SearchItem struct {
	ImgUrl string `json:"middleURL"`
}
type SearchResult struct {
	QueryWord string `json:"queryExt"`
	Data [] SearchItem `json:"data"`
}
var dirName string
var origin string = "http://image.baidu.com/search/acjson?tn=resultjson_com&ipn=rj&ct=201326592&is=&fp=result&cl=2&lm=-1&ie=utf-8&oe=utf-8&adpicid=&st=&z=&ic=&hd=&latest=&copyright=&s=&se=&tab=&width=&height=&face=&istype=&qc=&nc=1&fr=&expermode=&force=&rn=30&gsm=&1575860348327="
func main() {
	queryWords := flag.String("queryWord","", "the keyword to search")
	flag.Parse()
	if queryWords == nil {
		flag.Usage()
		return
	}

	qwArray := strings.Split(*queryWords,",")
	//fmt.Printf("%v", qwArray)
	wg := &sync.WaitGroup{}
	chn := make(chan struct{})
	//loading := []byte{'\\', '|', '/'}
	go spinner(1500, chn)

	for j := 0; j < len(qwArray); j++ {
		wg.Add(1)
		//create dir
		dirName = fmt.Sprint("./", qwArray[j])
		os.Mkdir(dirName, 0x777)
		os.Chmod(dirName, 0777)
		go func(n int) {
			//search 5 pages data
			for i := 1; i <= 5; i++ {
				getJsonData(qwArray[n], i * 30)
			}
			wg.Done()
		}(j)
	}
	//fmt.Println("close"))
	wg.Wait()
	chn <- struct{}{}

}

//loading
func spinner(delay time.Duration, chn chan struct{}) {
	loop:
		for {

			select {
				case <- chn:
					break loop
			default:

			}

			for _, r := range `-|/` {
				fmt.Printf("\r%c", r)
				time.Sleep(delay)
			}
		}
	fmt.Printf("\r%v", "")
}
//get JSON data from origin
func getJsonData(queryWord string, pageNum int) {
	//sURl := origin + "&queryword=" + queryWord + "&word=" + queryWord + "&pn=" + string(pageNum)
	sURL, _ := url.Parse(origin)
	q, _ := url.ParseQuery(sURL.RawQuery)
	q.Add("queryword", queryWord)
	q.Add("word", queryWord)
	q.Add("pn", strconv.Itoa(pageNum))
	sURL.RawQuery = q.Encode()

	//fmt.Print("origin-url:", fmt.Sprintf("%s", sURL))
	//fmt.Println()

	jsonRes, err := http.Get(fmt.Sprintf("%s", sURL))
	if err != nil {
		log.Println("Get JSON data error from :", origin, " error :", err)
		return
	}
	defer jsonRes.Body.Close()

	var searchResult SearchResult
	bs, err := ioutil.ReadAll(jsonRes.Body)
	//fmt.Println(string(bs))
	if err != nil {
		log.Println("Convert body to bytes error :", err)
		return
	}
	err = json.Unmarshal(bs, &searchResult)

	//fmt.Printf("%+v", &searchResult)
	if err != nil {
		log.Println("Unmarshal error :", err)
		return
	}

	var wg = &sync.WaitGroup{}
	for i := 0; i < len(searchResult.Data); i++ {
		wg.Add(1)
		//get image and save
		go getImage(searchResult.Data[i].ImgUrl, wg)
	}
	wg.Wait()
}

func getImage(imgUrl string, wg *sync.WaitGroup) {
	//fmt.Printf("imgUrl:%v", imgUrl)
	//
	//fmt.Println()
	if imgUrl == "" {
		wg.Done()
		return
	}

	//imgRes, err := http.Get(imgUrl)
	imgReq, err := http.NewRequest("GET", imgUrl, nil)
	imgReq.Header.Add("origin", origin)

	client := &http.Client{}
	imgRes, err := client.Do(imgReq)

	if err != nil {
		log.Println("Get image error :", err)
		return
	}
	defer imgRes.Body.Close()

	//save image to local
	u, err := url.Parse(imgUrl)
	if err != nil {
		log.Println("Parse image url error:", err)
		return
	}
	//imgURL, _ := u.Parse(imgUrl)
	path := u.Path

	fileName := strings.Split(strings.Split(path, "/")[2], "&")[0]
	//create image
	f, err := os.Create(dirName + "/" + fileName + ".jpg")
	if err != nil {
		log.Println("Create image error:", err)
		return
	}

	io.Copy(f, imgRes.Body)

	//wg.Done
	wg.Done()
}