package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gorilla/mux"
)

// Book - книга
type Book struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Author    string `json:"author"`
	Year      string `json:"year"`
	Publisher string `json:"publisher"`
	Image     string `json:"image"`
	Store     string `json:"store"`
}

var books []Book

func main() {
	file, err := excelize.OpenFile("./books.xlsx")
	if err != nil {
		fmt.Printf("Не удалось открыть Excel файл: %s", err)
		os.Exit(0)
	}
	rows, _ := file.GetRows(file.GetSheetList()[file.GetActiveSheetIndex()])
	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		if rows[rowIndex][0] == "" {
			break
		}
		books = append(books, Book{
			ID:        rowIndex - 1,
			Author:    rows[rowIndex][0],
			Name:      rows[rowIndex][1],
			Year:      rows[rowIndex][2],
			Publisher: rows[rowIndex][3],
			Store:     rows[rowIndex][4],
			Image:     rows[rowIndex][5],
		})
	}

	// booksJSON, _ := json.Marshal(books)
	// fmt.Println(string(booksJSON))

	router := mux.NewRouter()
	router.HandleFunc("/api/everything", getBooks).Methods("GET")
	router.HandleFunc("/api/search", searchBooks).Methods("GET")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	http.ListenAndServe(":80", router)
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func searchBooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var foundBooks []Book
	keys, ok := r.URL.Query()["q"]
	if !ok || len(keys[0]) < 1 {
		json.NewEncoder(w).Encode(unique(foundBooks))
		return
	}
	q := strings.TrimSpace(keys[0])
	q = strings.ReplaceAll(q, "\\Q", "")
	q = strings.ReplaceAll(q, "\\E", "")
	q = strings.ReplaceAll(q, "?", "\\E.\\Q")
	q = strings.ReplaceAll(q, "*", "\\E.*\\Q")
	words := strings.Split(q, " ")
	for _, word := range words {
		subwords := strings.Split(word, "&&")
		fmt.Println(len(subwords))
		if len(subwords) == 0 {
			pattern := `(?i)(^|\s)\Q` + word + `\E[а-яёa-z0-9!?.,-]{0,3}?`
			for _, book := range books {
				bookString := fmt.Sprintf("%s %s %s", book.Author, book.Name, book.Publisher)
				if matched, _ := regexp.Match(pattern, []byte(bookString)); matched || word == book.Year {
					foundBooks = append(foundBooks, book)
				}
			}
		} else {
			foundBooks1 := books
			var foundBooks2 []Book
			for _, subword := range subwords {
				fmt.Println(subword)
				pattern := `(?i)(^|\s)\Q` + subword + `\E[а-яёa-z0-9!?.,-]{0,3}?`
				for _, book := range foundBooks1 {
					bookString := fmt.Sprintf("%s %s %s", book.Author, book.Name, book.Publisher)
					if matched, _ := regexp.Match(pattern, []byte(bookString)); matched || subword == book.Year {
						foundBooks2 = append(foundBooks2, book)
					}
				}
				foundBooks1 = foundBooks2
				foundBooks2 = []Book{}
				fmt.Println(foundBooks1)
			}
			for _, book := range foundBooks1 {
				foundBooks = append(foundBooks, book)
			}
		}
	}
	json.NewEncoder(w).Encode(unique(foundBooks))
}

func unique(books []Book) []Book {
	keys := make(map[int]bool)
	list := []Book{}
	for _, entry := range books {
		if _, value := keys[entry.ID]; !value {
			keys[entry.ID] = true
			list = append(list, entry)
		}
	}
	return list
}
