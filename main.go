
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/xuri/excelize/v2"
)

type ResponseData struct {
	Text     string `json:"text"`
	URL      string `json:"url"`
	Type     string `json:"type"`
	Tag      string `json:"tag"`
	MetaData string `json:"metadata"`
	Date     string `json:"date"`
}

func scrapeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}

	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	response, err := http.Get(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error during HTTP request: %v", err), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Failed to access the page, status code %d", response.StatusCode), http.StatusInternalServerError)
		return
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error reading HTML document: %v", err), http.StatusInternalServerError)
		return
	}

	var data []ResponseData
	
	metaData := make(map[string]string)
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		property, _ := s.Attr("property")
		content, _ := s.Attr("content")
		
		key := name
		if key == "" {
			key = property
		}
		
		if key != "" && content != "" {
			metaData[key] = content
		}
	})
	
	title := doc.Find("title").Text()
	if title != "" {
		data = append(data, ResponseData{
			Text: title,
			URL:  "",
			Type: "title",
			Tag:  "title",
			MetaData: "",
			Date: time.Now().Format("2006-01-02"),
		})
	}

	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			headingType := s.Get(0).Data // h1, h2, etc.
			data = append(data, ResponseData{
				Text: text,
				URL:  "",
				Type: "heading",
				Tag:  headingType,
				MetaData: "",
				Date: time.Now().Format("2006-01-02"),
			})
		}
	})

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			data = append(data, ResponseData{
				Text: text,
				URL:  "",
				Type: "paragraph",
				Tag:  "p",
				MetaData: "",
				Date: time.Now().Format("2006-01-02"),
			})
		}
	})

	doc.Find("ul, ol").Each(func(i int, s *goquery.Selection) {
		listType := s.Get(0).Data // ul or ol
		s.Find("li").Each(func(j int, li *goquery.Selection) {
			text := strings.TrimSpace(li.Text())
			if text != "" {
				data = append(data, ResponseData{
					Text: text,
					URL:  "",
					Type: "list-item",
					Tag:  listType + "-li",
					MetaData: "",
					Date: time.Now().Format("2006-01-02"),
				})
			}
		})
	})

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		text := strings.TrimSpace(s.Text())
		
		if exists && href != "" {
			if strings.HasPrefix(href, "/") {
				baseURL := getBaseURL(url)
				href = baseURL + href
			}
			
			data = append(data, ResponseData{
				Text: text,
				URL:  href,
				Type: "link",
				Tag:  "a",
				MetaData: "",
				Date: time.Now().Format("2006-01-02"),
			})
		}
	})

	// Extract images
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		alt, _ := s.Attr("alt")
		
		if exists && src != "" {
			// Convert relative URLs to absolute
			if strings.HasPrefix(src, "/") {
				baseURL := getBaseURL(url)
				src = baseURL + src
			}
			
			data = append(data, ResponseData{
				Text: alt,
				URL:  src,
				Type: "image",
				Tag:  "img",
				MetaData: "",
				Date: time.Now().Format("2006-01-02"),
			})
		}
	})

	// Extract tables
	doc.Find("table").Each(func(tableIdx int, table *goquery.Selection) {
		tableData := ""
		
		table.Find("tr").Each(func(rowIdx int, row *goquery.Selection) {
			if rowIdx > 0 {
				tableData += "\n"
			}
			
			row.Find("th, td").Each(func(colIdx int, cell *goquery.Selection) {
				if colIdx > 0 {
					tableData += " | "
				}
				tableData += strings.TrimSpace(cell.Text())
			})
		})
		
		if tableData != "" {
			data = append(data, ResponseData{
				Text: tableData,
				URL:  "",
				Type: "table",
				Tag:  "table",
				MetaData: "",
				Date: time.Now().Format("2006-01-02"),
			})
		}
	})

	// Add metadata as separate entries
	for key, value := range metaData {
		data = append(data, ResponseData{
			Text: value,
			URL:  "",
			Type: "metadata",
			Tag:  key,
			MetaData: "",
			Date: time.Now().Format("2006-01-02"),
		})
	}

	acceptHeader := r.Header.Get("Accept")
	if strings.Contains(acceptHeader, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, fmt.Sprintf("Error encoding JSON: %v", err), http.StatusInternalServerError)
		}
		return
	}

	// Create Excel file
	f := excelize.NewFile()
	index := f.NewSheet("Sheet1")

	// Set headers
	headers := []string{"Content Type", "HTML Tag", "Text", "URL", "Metadata", "Date"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue("Sheet1", cell, header)
	}
	
	// Style headers
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#DDDDDD"}, Pattern: 1},
	})
	if err == nil {
		f.SetCellStyle("Sheet1", "A1", string(rune('A'+len(headers)-1))+"1", headerStyle)
	}

	// Populate data
	for i, item := range data {
		rowIdx := i + 2 // +2 because headers are at row 1
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", rowIdx), item.Type)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", rowIdx), item.Tag)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", rowIdx), item.Text)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", rowIdx), item.URL)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", rowIdx), item.MetaData)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", rowIdx), item.Date)
	}

	// Auto column width
	f.SetColWidth("Sheet1", "A", "A", 15)
	f.SetColWidth("Sheet1", "B", "B", 15)
	f.SetColWidth("Sheet1", "C", "C", 60)
	f.SetColWidth("Sheet1", "D", "D", 40)
	f.SetColWidth("Sheet1", "E", "E", 20)
	f.SetColWidth("Sheet1", "F", "F", 15)

	f.SetActiveSheet(index)

	// Set headers for download
	fileName := getCleanDomainName(url) + "_data.xlsx"
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")

	// Write Excel file
	if err := f.Write(w); err != nil {
		http.Error(w, fmt.Sprintf("Error writing Excel file: %v", err), http.StatusInternalServerError)
	}
}

func getBaseURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) >= 3 {
		return strings.Join(parts[:3], "/")
	}
	return url
}

func getCleanDomainName(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "www.")
	
	parts := strings.Split(url, "/")
	domain := parts[0]
	domain = strings.ReplaceAll(domain, ".", "_")
	
	return domain
}

func main() {
	http.HandleFunc("/scrape", scrapeHandler)

	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", fs)

	port := ":8080"
	log.Printf("Server started at http://0.0.0.0%s", port)
	log.Fatal(http.ListenAndServe("0.0.0.0"+port, nil))
}
