package main

import (
	"log"
	"fmt"
	"sync"
	"sync/atomic"
//	"math/rand"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

type URL struct {
	Url   string
	Depth int
}


type FetchWalker struct {
	mapFetched map[string]int
	lock       sync.Mutex
	maxDepth   int
}

var work_counter int32
//var random *math.Rand = rand.New(time.Now().UnixNano())

func (fetchWalker *FetchWalker) nonBlockingCrawler(urlChan chan *URL, waitGroup *sync.WaitGroup, fetcher Fetcher) {
	routineId := atomic.AddInt32(&work_counter, 1)
	log.Printf("Hey I'm routine %v\n", routineId)
	for {
		//		rand.Seed(time.Now().UnixNano())
		//		randInt := rand.Intn(500)
		//		fmt.Printf(".............. randint %v\n", randInt)
		//		time.Sleep(time.Duration(randInt) * time.Millisecond)
		url, more := <-urlChan
		if more {
			fetchWalker.lock.Lock()
			fetchWalker.mapFetched[url.Url] = 1
			fetchWalker.lock.Unlock()
			//do something with url
			body, urls, err := fetcher.Fetch(url.Url)
			if err != nil {
				log.Printf("Routine %v err: %v\n", routineId, err)
			}else {
				log.Printf("Routine %v crawled body: %v\n", routineId, body)
				if url.Depth < fetchWalker.maxDepth {
					urlChanChildren := make(chan *URL, 10)
					nbUrlAdded := fetchWalker.crawlChildUrls(urls, urlChanChildren, url.Depth + 1)
					close(urlChanChildren)

					//doneChanChildren := make(chan bool, 10)
					if nbUrlAdded > 0 {
						var waitGroupChildren sync.WaitGroup
						for i := 0; i < nbUrlAdded; i++ {
							waitGroupChildren.Add(1)
							log.Printf("Routine %v spawn a child\n", routineId)
							go fetchWalker.nonBlockingCrawler(urlChanChildren, &waitGroupChildren, fetcher)
						}
						waitGroupChildren.Wait()
					}
				}
			}
		}else {
			log.Printf("Routine %v done\n", routineId)
			waitGroup.Done()
			return
		}
	}
}


func (fetchWalker *FetchWalker) crawlChildUrls(urls []string, urlChan chan *URL, depth int) (nbUrlAdded int) {
	nbUrlAdded = 0
	for _, url := range urls {
		fetchWalker.lock.Lock()
		_, alreadyFetched := fetchWalker.mapFetched[url]
		if !alreadyFetched {
			fetchWalker.mapFetched[url] = 1
		}
		fetchWalker.lock.Unlock()

		if !alreadyFetched {
			urlChan <- &URL{url, depth}
			nbUrlAdded++
		}
	}
	return nbUrlAdded
}


func Crawl3(url string, depth int, fetcher Fetcher) {
	urlToCrawlChan := make(chan *URL, 10)
	//urlCrawled := map[string]int
	startUrl := URL{url, 0}

	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	var fetchWalker *FetchWalker = &FetchWalker{make(map[string]int), sync.Mutex{}, depth}
	go fetchWalker.nonBlockingCrawler(urlToCrawlChan, &waitGroup, fetcher)
	urlToCrawlChan <- &startUrl
	close(urlToCrawlChan)

	waitGroup.Wait()

}


//func crawler(urlChan chan *URL, maxDepth int, fetcher Fetcher) {
//	url, chanOpened := <-urlChan
//	for chanOpened {
//		fmt.Printf("url:%v, opened:%v\n", url, chanOpened)
//		if url.Depth <= maxDepth {
//			body, urls, err := fetcher.Fetch(url.Url)
//			if err != nil {
//				fmt.Println(err)
//			} else {
//				fmt.Printf("Yeah! Crawled : %v\n", body)
//				if url.Depth != maxDepth {
//					for _, u := range urls {
//						newUrl := URL{u, url.Depth + 1}
//						fmt.Printf("Pushing url to channel %v\n", newUrl)
//						urlChan <- &newUrl
//					}
//				}
//			}
//		}
//		url, chanOpened = <-urlChan
//	}
//}

//
//var crawlerDone chan bool = make(chan bool)
//func Crawl2(url string, depth int, fetcher Fetcher) {
//	// TODO: Fetch URLs in parallel.
//	// TODO: Don't fetch the same URL twice.
//	// This implementation doesn't do either:
//	urlToCrawlChan := make(chan *URL, 10)
//
//	//urlCrawled := map[string]int
//	startUrl := URL{url, 0}
//	for i := 2; i > 0; i-- {
//		go crawler(urlToCrawlChan, depth, fetcher)
//	}
//	urlToCrawlChan <- &startUrl
//
//}
//
//// Crawl uses fetcher to recursively crawl
//// pages starting with url, to a maximum of depth.
//func Crawl(url string, depth int, fetcher Fetcher) {
//	// TODO: Fetch URLs in parallel.
//	// TODO: Don't fetch the same URL twice.
//	// This implementation doesn't do either:
//	if depth <= 0 {
//		return
//	}
//	body, urls, err := fetcher.Fetch(url)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//	fmt.Printf("found: %s %q\n", url, body)
//	for _, u := range urls {
//		Crawl(u, depth-1, fetcher)
//	}
//	return
//}

func main() {
	//	ch := make(chan int, 2)
	//	done := make(chan bool)
	//	go func() {
	//		keepOn := true
	//		for keepOn {
	//			select {
	//			case v, more := <-ch:
	//				fmt.Printf("v %v more %v\n", v, more)
	//				keepOn = more
	//			default:
	//				fmt.Println("default")
	//			}
	//		}
	//		done <- true
	//	}()
	//	ch <- 6
	//	close(ch)
	//	<-done
	Crawl3("http://golang.org/", 4, &fetcher)
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}

