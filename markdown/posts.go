package markdown

import (
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"fmt"
	"strings"
)


type Post struct {
	Title string `json:"title"`
	Date string  `json:"date"`
	Summary string `json:"summary"`
	Body string `json:"body"`
	File string `json:"file"`
}

func GetPost(slug string) (Post, error){
	f := "markdown/"+slug+".md"

	fileread, err := ioutil.ReadFile(f)

	if err != nil {
		return Post{}, err
	}

	fmt.Println(fileread)
	lines := strings.Split(string(fileread), "\n")
	title := string(lines[0])
	date := string(lines[1])
	summary := string(lines[2])
	body := strings.Join(lines[3:len(lines)], "\n")

	fmt.Println(body)

	unsafe := string(blackfriday.MarkdownCommon([]byte(body)))
	safe := bluemonday.UGCPolicy().Sanitize(unsafe)

	post := Post{title, date, summary, safe, ""}

	return post, nil
}