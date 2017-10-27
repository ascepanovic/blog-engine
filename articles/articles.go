package articles

import (
	"time"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"fmt"
	"../helpers"
	"encoding/json"
)

type Article struct {
	ID bson.ObjectId `json:"_id" bson:"_id,omitempty"`
	LangCode string `json:"langCode" bson:"langCode"`
	AuthorId string `json:"authorId" bson:"authorId"`
	AuthorName string `json:"authorName" bson:"authorName"`
	ArticleType int64 `json:"type" bson:"type"`
	Title string `json:"title" bson:"title"`
	Intro string `json:"intro" bson:"intro"`
	Body string `json:"body" bson:"body"`
	MainPic string `json:"mainPic" bson:"mainPic"`
	Tags string `json:"tags" bson:"tags"`
	Slug string `json:"slug" bson:"slug"`
	DateAdded time.Time `json:"dateAdded" bson:"dateAdded"`
	Status int64 `json:"status" bson:"status"`
}

type Articles []Article

type ArticlesPagination struct {
	Pagination pagination.PageMetadata `json:"pagination"`
	Articles    []Article           `json:"articles"`
}


var allPostsData = Articles{

	//Article{1, 1, "sunt aut facere repellat provident occaecati excepturi optio reprehenderit", "quia et suscipit\nsuscipit recusandae consequuntur expedita et cum\nreprehenderit molestiae ut ut quas totam\nnostrum rerum est autem sunt rem eveniet architecto"},
	//Article{2, 2, "Second title", "And long body of second"},
	//Article{3, 4, "Nas treci post", "Nas treci post text se nalazi u ovom djelu"},
}

var (
	MgoSession      *mgo.Session
	MongoServer     = "localhost"
	MongoUser       = "" //os.Getenv("MGOUSER")
	MongoPass       = ""
	DB              = "blog"
	ARTICLESCollection = "Articles"
)

//init connection to mongo database
func init() {
	session, err := mgo.Dial(MongoServer)
	if err != nil {
		log.Fatal(err.Error())
	}
	MgoSession = session
}

//Will update or create new record
func SaveArticle(id bson.ObjectId, langCode string, authorId string, authorName string, articleType int64, title string, intro string, body string, mainPic string, tags string, slug string, status int64) (Article, error) {
	sess := MgoSession.Clone()
	defer sess.Close()

	sess.SetMode(mgo.Monotonic, true)

	articleTBL := sess.DB(DB).C(ARTICLESCollection)
	article :=Article{}

	//check does article exist in database
	err := articleTBL.Find(bson.M{"_id": id}).One(&article)

	article.LangCode = langCode
	article.AuthorId = authorId
	article.AuthorName = authorName
	article.ArticleType = articleType
	article.Title = title
	article.Intro = intro
	article.Body = body
	article.MainPic = mainPic
	article.Tags = tags
	article.Slug = slug

	if err == nil {
		// Update article
		err = articleTBL.Update(bson.M{"_id": id}, article)
		fmt.Println("Update odradjen", article.Title)
	} else {
		// Insert new article
		article.ID = bson.NewObjectId()
		article.DateAdded = bson.Now()
		fmt.Println("ID je: ", article.ID)

		err = articleTBL.Insert(article)
	}

	if err != nil {
		return article, err
	}

	return article, nil
}

func GetAllArticles(langCode string, qp *pagination.QueryParam)(p *ArticlesPagination, err error){
	query := bson.M{}

	//TODO add fetching based on language code
	if langCode != "" {
		query = bson.M{"langCode": langCode}
	}

	fmt.Println("Query", query)

	pageReturn := &ArticlesPagination{}

	resultsList, err := pagination.GetPaginatedResults(MgoSession, ARTICLESCollection, query, "_id", qp)
	if err != nil {
		return pageReturn, err
	}

	pageReturn.Pagination = resultsList.Pagination

	err = json.Unmarshal(resultsList.Results, &pageReturn.Articles)

	return pageReturn, err
}

func GetArticle(id string)(Article, error){
	sess := MgoSession.Clone()
	defer sess.Close()

	sess.SetMode(mgo.Monotonic, true)
	con := sess.DB(DB).C(ARTICLESCollection)


	article := Article{}
	err := con.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&article)

	return article, err
}


//delete article from database
func DeleteArticle(id string) (string, error){
	sess := MgoSession.Clone()
	defer sess.Clone()

	sess.SetMode(mgo.Monotonic, true)
	con := sess.DB(DB).C(ARTICLESCollection)

	err := con.Remove(bson.M{"_id": bson.ObjectIdHex(id)})

	if err != nil{
		return "Cant delete this article from database", err
	}

	return "Article deleted from database", nil
}