package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
)

// Structs for Open Library API response
type TOCEntry struct {
	Title string `json:"title"`
}

type BookData struct {
	Title           string     `json:"title"`
	TableOfContents []TOCEntry `json:"table_of_contents"`
}

// Structs for .mm mindmap format
type Node struct {
	XMLName xml.Name `xml:"node"`
	Text    string   `xml:"TEXT,attr"`
	Nodes   []Node   `xml:"node,omitempty"`
}

type Map struct {
	XMLName xml.Name `xml:"map"`
	Version string   `xml:"version,attr"`
	Root    Node     `xml:"node"`
}

func fetchBookData(isbn string) (BookData, error) {
	var book BookData
	url := fmt.Sprintf("https://openlibrary.org/api/books?bibkeys=ISBN:%s&format=json&jscmd=data", isbn)

	resp, err := http.Get(url)
	if err != nil {
		return book, err
	}
	defer resp.Body.Close()

	var result map[string]BookData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return book, err
	}

	key := "ISBN:" + isbn
	book, ok := result[key]
	if !ok {
		return book, fmt.Errorf("no book data found for ISBN %s", isbn)
	}

	return book, nil
}

func buildMindmap(book BookData, filename string) error {
	root := Node{Text: book.Title}

	for _, entry := range book.TableOfContents {
		root.Nodes = append(root.Nodes, Node{Text: entry.Title})
	}

	m := Map{
		Version: "1.0.1",
		Root:    root,
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	file.WriteString(xml.Header)
	enc := xml.NewEncoder(file)
	enc.Indent("", "  ")
	return enc.Encode(m)
}

func main() {
	var isbn string
	fmt.Print("Enter ISBN: ")
	fmt.Scanln(&isbn)

	book, err := fetchBookData(isbn)
	if err != nil {
		fmt.Println("Error fetching book data:", err)
		return
	}

	if len(book.TableOfContents) == 0 {
		fmt.Println("No table of contents found.")
		return
	}

	filename := "mindmap.mm"
	err = buildMindmap(book, filename)
	if err != nil {
		fmt.Println("Error writing mindmap:", err)
		return
	}

	fmt.Printf("Mindmap saved to %s\n", filename)
}
