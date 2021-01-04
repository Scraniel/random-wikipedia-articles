package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

const wikipediaEndpoint = "https://en.wikipedia.org/w/api.php"
const queryPrefix = "action=query&format=json"

type PageData struct {
	PageID      int    `json:"pageid,omitempty"`
	ID          int    `json:"id,omitempty"`
	NamespaceID int    `json:"ns"`
	Title       string `json:"title"`
	Extract     string `json:"extract,omitempty"`
}

type ArticleExtract struct {
	Batchcomplete string `json:"batchcomplete"`
	Query         struct {
		Pages map[string]PageData `json:"pages"`
	} `json:"query"`
}

// RandomResponse is a DTO to hold info about random pages
type RandomResponse struct {
	Batchcomplete string `json:"batchcomplete"`
	Continue      struct {
		Rncontinue string `json:"rncontinue"`
		Continue   string `json:"continue"`
	} `json:"continue"`
	Query struct {
		Random []PageData `json:"random"`
	} `json:"query"`
}

func getArticleDescriptions(ids []int) (ArticleExtract, error) {
	var builder strings.Builder
	_, err := fmt.Fprintf(&builder, "%v?%v&prop=extracts&exintro&explaintext&redirects=1&pageids=", wikipediaEndpoint, queryPrefix)

	if err != nil {
		return ArticleExtract{}, err
	}

	for index, id := range ids {
		_, err = fmt.Fprintf(&builder, "%d", id)

		if err != nil {
			return ArticleExtract{}, err
		}

		if index < len(ids)-1 {
			_, err = builder.WriteString("%7C")

			if err != nil {
				return ArticleExtract{}, err
			}
		}
	}

	url := builder.String()
	resp, err := http.Get(url)
	if err != nil {
		return ArticleExtract{}, err
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data ArticleExtract
	err = decoder.Decode(&data)

	return data, err
}

func getRandomArticles(count int) (RandomResponse, error) {
	url := fmt.Sprintf("%v?%v&list=random&rnnamespace=0&rnlimit=%v", wikipediaEndpoint, queryPrefix, count)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return RandomResponse{}, err
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var data RandomResponse
	err = decoder.Decode(&data)

	if err != nil {
		return RandomResponse{}, err
	}

	if len(data.Query.Random) == 0 {
		return RandomResponse{}, errors.New("random pages not found")
	}

	return data, err
}

func main() {
	articles, err := getRandomArticles(3)

	if err != nil {
		fmt.Println("error getting random articles:", err)
		return
	}

	articleIds := make([]int, len(articles.Query.Random))
	for index, article := range articles.Query.Random {
		articleIds[index] = article.ID
	}
	summaries, err := getArticleDescriptions(articleIds)

	if err != nil {
		fmt.Println("error getting article summaries:", err)
		return
	}

	for _, article := range articles.Query.Random {
		fmt.Println("Id:", article.ID)
		fmt.Println("Title:", article.Title)
		fmt.Println("Summary:", summaries.Query.Pages[strconv.Itoa(article.ID)].Extract)
	}

}
