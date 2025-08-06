package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// ---------- Data Types ----------

type TOCEntry struct {
	Title string `json:"title"`
}

type BookData struct {
	Title           string     `json:"title"`
	TableOfContents []TOCEntry `json:"table_of_contents"`
}

type VolumeInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GoogleBookItem struct {
	VolumeInfo VolumeInfo `json:"volumeInfo"`
}

type GoogleBooksResponse struct {
	Items []GoogleBookItem `json:"items"`
}

// ---------- XML Structs for .mm ----------

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

// ---------- Fetch from Open Library ----------

func fetchFromOpenLibrary(isbn string) (BookData, error) {
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

// ---------- Fetch from Google Books ----------

func fetchFromGoogleBooks(isbn string) (BookData, error) {
	url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=isbn:%s", isbn)

	resp, err := http.Get(url)
	if err != nil {
		return BookData{}, err
	}
	defer resp.Body.Close()

	var result GoogleBooksResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return BookData{}, err
	}

	if len(result.Items) == 0 {
		return BookData{}, fmt.Errorf("no book found for ISBN %s", isbn)
	}

	info := result.Items[0].VolumeInfo
	var toc []TOCEntry

	if info.Description != "" {
		lines := strings.Split(info.Description, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && len(line) > 5 {
				toc = append(toc, TOCEntry{Title: line})
			}
		}
	}

	return BookData{
		Title:           info.Title,
		TableOfContents: toc,
	}, nil
}

// ---------- Fallback Wrapper ----------

func fetchBookData(isbn string) (BookData, error) {
	book, err := fetchFromOpenLibrary(isbn)
	if err == nil && len(book.TableOfContents) > 0 {
		fmt.Println("✓ Got ToC from Open Library")
		return book, nil
	}

	fmt.Println("⚠️  Open Library failed or had no ToC. Trying Google Books...")

	book, err = fetchFromGoogleBooks(isbn)
	if err != nil {
		return book, fmt.Errorf("both Open Library and Google Books failed: %v", err)
	}

	fmt.Println("✓ Fallback to Google Books successful")
	return book, nil
}

// ---------- Build .mm Mindmap File ----------

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

// ---------- Main ----------

func main() {
	var isbn string
	fmt.Print("Enter ISBN (no hyphens): ")
	fmt.Scanln(&isbn)

	book, err := fetchBookData(isbn)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if len(book.TableOfContents) == 0 {
		fmt.Println("No Table of Contents found.")
		return
	}

	filename := isbn + ".mm"
	if err := buildMindmap(book, filename); err != nil {
		fmt.Println("Error writing mindmap:", err)
		return
	}

	fmt.Printf("✅ Mindmap saved to %s\n", filename)
}
