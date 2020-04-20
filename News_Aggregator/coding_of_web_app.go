package main

import ("fmt"
		"net/http"
		"html/template"
		"io/ioutil"
		"encoding/xml"
		"sync")

var wg sync.WaitGroup

func home_handler(w http.ResponseWriter,r *http.Request){
	fmt.Fprintf(w, "<h1>Hi Jim</h1>")
}

var washPostXML = []byte(`
<sitemapindex>
   <sitemap>
      <loc>http://www.washingtonpost.com/news-politics-sitemap.xml</loc>
   </sitemap>
   <sitemap>
      <loc>http://www.washingtonpost.com/news-blogs-politics-sitemap.xml</loc>
   </sitemap>
   <sitemap>
      <loc>http://www.washingtonpost.com/news-opinions-sitemap.xml</loc>
   </sitemap>
</sitemapindex>
`)

type Sitemapindex struct {
	Locations []string `xml:"sitemap>loc"`
}

type News struct {
	Titles []string `xml:"url>news>title"`
	Keywords []string `xml:"url>news>keywords"`
	Locations []string `xml:"url>loc"`
}

type NewsMap struct {
	Keyword string
	Location string
}

type NewsAggPage struct {
	Title string
	News map[string]NewsMap
}

func newsRoutine(c chan News,Location string) {
	defer wg.Done()
	var n News
	resp, _ := http.Get(Location)
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &n)
	resp.Body.Close()

	c <- n
}

func newsAggHandler(w http.ResponseWriter,r *http.Request){
	var s Sitemapindex

	bytes := washPostXML
	xml.Unmarshal(bytes, &s)
	news_map := make(map[string]NewsMap)
	queue := make(chan News,30)

	for _, Location := range s.Locations {
		wg.Add(1)
		go newsRoutine(queue,Location)
	}

	wg.Wait()
	close(queue)

	for elem := range queue {
		for idx, _ := range elem.Keywords {
			news_map[elem.Titles[idx]] = NewsMap{elem.Keywords[idx], elem.Locations[idx]}
		}
	}
	p := NewsAggPage{Title: "News Aggregator", News : news_map}
	t,_ := template.ParseFiles("basictemplating.html")
	t.Execute(w,p)
}

func main() {
	http.HandleFunc("/",home_handler)
	http.HandleFunc("/agg/",newsAggHandler)
	http.ListenAndServe(":8000",nil)
}
