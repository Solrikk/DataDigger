package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/xuri/excelize/v2"
)

type ResponseData struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

func scrapeHandler(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "URL параметр отсутствует", http.StatusBadRequest)
		return
	}

	response, err := http.Get(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при HTTP-запросе: %v", err), http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Не удалось получить доступ к странице, статус код %d", response.StatusCode), http.StatusInternalServerError)
		return
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при чтении HTML-документа: %v", err), http.StatusInternalServerError)
		return
	}

	var data []ResponseData
	doc.Find("body").Each(func(i int, body *goquery.Selection) {
		textData := strings.TrimSpace(body.Text())
		lines := strings.Split(textData, "\n")
		for _, line := range lines {
			cleanLine := strings.TrimSpace(line)
			if cleanLine != "" {
				data = append(data, ResponseData{Text: cleanLine, URL: ""})
			}
		}
		body.Find("a").Each(func(j int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if exists {
				data = append(data, ResponseData{Text: "", URL: href})
			}
		})
	})

	f := excelize.NewFile()
	index := f.NewSheet("Sheet1")

	for i, item := range data {
		if item.Text != "" {
			cell := fmt.Sprintf("A%d", i+1)
			f.SetCellValue("Sheet1", cell, item.Text)
		} else if item.URL != "" {
			cell := fmt.Sprintf("B%d", i+1)
			f.SetCellValue("Sheet1", cell, item.URL)
		}
	}

	f.SetActiveSheet(index)

	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment;filename=results.xlsx")
	w.Header().Set("File-Name", "results.xlsx")
	w.Header().Set("Content-Transfer-Encoding", "binary")

	if err := f.Write(w); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка при записи Excel файла: %v", err), http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/scrape", scrapeHandler)

	fs := http.FileServer(http.Dir("./"))
	http.Handle("/", fs)

	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
