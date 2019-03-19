package main

import (
	"io"
	"io/ioutil"
	"os"
	"encoding/json"
	"log"
	"net/http"
	"github.com/gorilla/mux"
)

// The Status type for status response
type Status struct {
	Status string `json:"status,omitempty"`
}

// The BookDetails type for details response
type BookDetails struct {
	ID			string	`json:"id,omitempty"`
	Author		string 	`json:"author,omitempty"`
	Year		string	`json:"year,omitempty"`
	Type		string	`json:"type,omitempty"`
	Pages		int		`json:"pages,omitempty"`
	Publisher	string	`json:"publisher,omitempty"`
	Language	string	`json:"language,omitempty"`
	ISBN10		string 	`json:"ISBN-10,omitempty"`
	ISBN13		string	`json:"ISBN-13,omitempty"`
}

// The ExternalBookDetails represent the answer from the Google API
type ExternalBookDetails struct {
	Kind       string `json:"kind"`
	TotalItems int    `json:"totalItems"`
	Items      []struct {
		Kind       string `json:"kind"`
		ID         string `json:"id"`
		Etag       string `json:"etag"`
		SelfLink   string `json:"selfLink"`
		VolumeInfo struct {
			Title               string   `json:"title"`
			Authors             []string `json:"authors"`
			Publisher           string   `json:"publisher"`
			PublishedDate       string   `json:"publishedDate"`
			Description         string   `json:"description"`
			IndustryIdentifiers []struct {
				Type       string `json:"type"`
				Identifier string `json:"identifier"`
			} `json:"industryIdentifiers"`
			ReadingModes struct {
				Text  bool `json:"text"`
				Image bool `json:"image"`
			} `json:"readingModes"`
			PageCount           int      `json:"pageCount"`
			PrintType           string   `json:"printType"`
			Categories          []string `json:"categories"`
			MaturityRating      string   `json:"maturityRating"`
			AllowAnonLogging    bool     `json:"allowAnonLogging"`
			ContentVersion      string   `json:"contentVersion"`
			PanelizationSummary struct {
				ContainsEpubBubbles  bool `json:"containsEpubBubbles"`
				ContainsImageBubbles bool `json:"containsImageBubbles"`
			} `json:"panelizationSummary"`
			ImageLinks struct {
				SmallThumbnail string `json:"smallThumbnail"`
				Thumbnail      string `json:"thumbnail"`
			} `json:"imageLinks"`
			Language            string `json:"language"`
			PreviewLink         string `json:"previewLink"`
			InfoLink            string `json:"infoLink"`
			CanonicalVolumeLink string `json:"canonicalVolumeLink"`
		} `json:"volumeInfo"`
		SaleInfo struct {
			Country     string `json:"country"`
			Saleability string `json:"saleability"`
			IsEbook     bool   `json:"isEbook"`
		} `json:"saleInfo"`
		AccessInfo struct {
			Country                string `json:"country"`
			Viewability            string `json:"viewability"`
			Embeddable             bool   `json:"embeddable"`
			PublicDomain           bool   `json:"publicDomain"`
			TextToSpeechPermission string `json:"textToSpeechPermission"`
			Epub                   struct {
				IsAvailable bool `json:"isAvailable"`
			} `json:"epub"`
			Pdf struct {
				IsAvailable bool `json:"isAvailable"`
			} `json:"pdf"`
			WebReaderLink       string `json:"webReaderLink"`
			AccessViewStatus    string `json:"accessViewStatus"`
			QuoteSharingAllowed bool   `json:"quoteSharingAllowed"`
		} `json:"accessInfo"`
		SearchInfo struct {
			TextSnippet string `json:"textSnippet"`
		} `json:"searchInfo"`
	} `json:"items"`
}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/health", Health)
	router.HandleFunc("/details/{id}", GetDetails)
	log.Fatal(http.ListenAndServe(":9080", router))
}

// Health returns health status of service
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{'status':'Details is healthy'}`)
}

// GetDetails returns book details
func GetDetails(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]

	if id == "" {
		http.Error(w, "please provide product id", 400)
		return
	}

	bookDetails := getBookDetails(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bookDetails)
}

func getBookDetails(id string) *BookDetails {
	if os.Getenv("ENABLE_EXTERNAL_BOOK_SERVICE") == "true" {
		return fetchDetailsFromExternalService("0486424618", id)
	} 

	bookDetails := BookDetails{ID: id, Author: "William Shakespeare", Year: "1595", Type: "paperback", Pages: 200, Publisher: "PublisherX", Language: "English", ISBN10: "1234567890", ISBN13: "123-1234567890"}
	return &bookDetails
}

func fetchDetailsFromExternalService(isbn string, id string) *BookDetails {
	response, err := http.Get("https://www.googleapis.com/books/v1/volumes?q=isbn:" + isbn)

	if err != nil {
		log.Printf("Fetching details from external service failed with error %s\n", err)
		return nil
	}

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		log.Printf("Can't read response from external service with error %s\n", err)
		return nil
	}

	var externalBookDetails ExternalBookDetails
	json.Unmarshal(body, &externalBookDetails)

	bookDetails := BookDetails{
		ID: id, 
		Author: externalBookDetails.Items[0].VolumeInfo.Authors[0], 
		Year: externalBookDetails.Items[0].VolumeInfo.PublishedDate, 
		Type: getPrintType(&externalBookDetails), 
		Pages: externalBookDetails.Items[0].VolumeInfo.PageCount, 
		Publisher: externalBookDetails.Items[0].VolumeInfo.Publisher, 
		Language: getLanguage(&externalBookDetails), 
		ISBN10: getISBN("ISBN_10", &externalBookDetails), 
		ISBN13: getISBN("ISBN_13", &externalBookDetails)}
	return &bookDetails
}

func getPrintType(externalBookDetails *ExternalBookDetails) string{
	if externalBookDetails.Items[0].VolumeInfo.PrintType == "BOOK" {
		return "paperback"
	} 
	return "unknown"
}

func getLanguage(externalBookDetails *ExternalBookDetails) string{
	if externalBookDetails.Items[0].VolumeInfo.Language == "en" {
		return "English"
	} 
	return "unknown"
}

func getISBN(kind string, externalBookDetails *ExternalBookDetails) string{
	industryIdentifiers := externalBookDetails.Items[0].VolumeInfo.IndustryIdentifiers

	for _,industryIdentifier := range industryIdentifiers {
		if industryIdentifier.Type == kind {
			return industryIdentifier.Identifier
		}
	}

	return "1234567890"
}
