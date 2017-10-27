package main

//HOW TO BUILD FOR WINDOWS PLATFORM:
//GOOS=windows GOARCH=386 go build -o hello.exe hello.go

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"gopkg.in/mgo.v2/bson"
	"github.com/gin-contrib/cors"
	"github.com/pjebs/restgate"
	"strconv"
	"./translations"
	"./articles"
	"./comments"
	"./markdown"
	"fmt"
	"./helpers"
)

//TODO - this should be done differently
var adminUsername = "admin"
var adminPassword = "sifra2017"


func main() {
	router := gin.Default()

	//we are going to allow requests only from specific origin
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://127.0.0.1:8080", "http://localhost:3000", "http://blogfront.dev", "http://www.blogfront.dev", "*"}
	config.AllowMethods = []string{"GET","POST","PUT","DELETE","OPTIONS"}
	config.AllowHeaders = []string{"X-Auth-Key", "X-Auth-Secret", "X-Auth-Token", "Content-type"}

	//allow all origins to use our api :
	//router.Use(cors.Default())

	//limit it:
	router.Use(cors.New(config))

	// Initialize Restgate
	rg := restgate.New("X-Auth-Key", "X-Auth-Secret", restgate.Static,
		restgate.Config{
			Key: []string{adminUsername},
			Secret: []string{adminPassword},
			HTTPSProtectionOff: true,
		})

	// Create Gin middleware - integrate Restgate with Gin
	rgAdapter := func(c *gin.Context) {
		nextCalled := false
		nextAdapter := func(http.ResponseWriter, *http.Request) {
			nextCalled = true
			c.Next()
		}

		//we will allow only call to authorization method without headers
		if (c.Request.URL.Path=="/authorize"){
			//allow access without headers
			nextCalled = true
			c.Next()
			return
		}

		rg.ServeHTTP(c.Writer, c.Request, nextAdapter)
		if nextCalled == false {
			c.AbortWithStatus(403) //forbidden
		}
	}

	// Use Restgate with Gin
	router.Use(rgAdapter)


	//basic authorization endpoint
	router.POST("/authorize", Authorize)

	//get all articles
	router.GET("/articles", GetArticles)

	//get all articles on certain language
	router.GET("/articleson/:langCode", GetArticlesByLanguage)

	//get specific article
	router.GET("/articles/:id", GetArticle)

	//save article to database
	router.POST("/articles", SaveArticle)

	//update article
	router.PUT("/articles/:id", SaveArticle)

	//remove article from database
	router.DELETE("/articles/:id", DeleteArticle)

	//INSERT SOME TEST TRANSLATIONS INTO DATABASE 0 JUST TEST FUNCTION
	router.GET("/generatelabels", GenerateLabels)

	//get All translations
	router.GET("/translations", GetTranslations)

	//get SINGLE translation
	router.GET("/translations/:id", GetTranslation)

	//get all translations for certain langCode
	router.GET("/translationsfor/:langCode", GetTranslations)

	//Save Translation
	router.POST("/translations", SaveTranslations)

	//remove translation from database
	router.DELETE("/translations/:id", DeleteTranslation)

	//get all comments
	router.GET("/comments", GetAllComments)

	//get comments for specific article
	router.GET("/commentsfor/:articleId", GetCommentsForArticle)

	//get single comment
	router.GET("/comments/:id", GetComment)

	//save comment into the database
	router.POST("/comments", SaveComment)

	//delete comment from database
	router.DELETE("/comments/:id", DeleteComment)


	//ASSETS MANAGEMENT STARTS
	router.POST("/assets", upload)



	//TEST FOR MARKDOWN RENDERING
	router.GET("/markdown/:slug", Markdown)


	router.Run(":8000")
}

//TODO more sofisticated solution
func Authorize(c *gin.Context){

	username := c.PostForm("username")
	password := c.PostForm("password")

	fmt.Println("Podaci: ", username, password)

	if (username=="admin" && password=="sifra7182"){
		c.JSON(http.StatusOK, gin.H{"success": true, "msg": "now you are authorized", "username": adminUsername, "password": adminPassword})
		return
	}

	c.JSON(400, gin.H{"success": false, "msg": "Unable to authorize, please check given credentials", "errCode": 21})
	return
}

/*** -- ARTICLES HANDLERS START -- ***/

func GetArticles(c *gin.Context){

	query_data := c.Request.URL.Query()
	qp := query_param(query_data)

	content, err := articles.GetAllArticles("", qp)

	if err != nil {
		c.JSON(400, gin.H{"success": false, "msg": "Unable to retrieve articles", "errCode": 38})
		return
	}

	total :=len(content.Articles)
	c.Header("X-Total-Count", strconv.Itoa(total))
	c.Header("Access-Control-Expose-Headers","X-Total-Count")
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "ok", "data": content})
}

func GetArticlesByLanguage(c *gin.Context){

	query_data := c.Request.URL.Query()
	qp := query_param(query_data)

	content, err := articles.GetAllArticles(c.Param("langCode"), qp)

	if err != nil {
		c.JSON(400, gin.H{"success": false, "msg": "Unable to retrieve articles", "errCode": 38})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "ok", "data": content})
}

func GetArticle(c *gin.Context){
	id := c.Param("id")

	article, err := articles.GetArticle(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error(), "errCode": 66})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "serverTime": bson.Now(), "data": article})
}

func SaveArticle(c *gin.Context){
	id := c.PostForm("id")
	langCode := c.PostForm("langCode")
	authorId := c.PostForm("authorId")
	authorName := c.PostForm("authorName")
	articleType := "1"//c.PostForm("articleType")
	title := c.PostForm("title")
	intro := c.PostForm("intro")
	body := c.PostForm("body")
	mainPic := c.PostForm("mainPic")
	tags := c.PostForm("tags")
	slug := c.PostForm("slug")
	status := c.PostForm("status")

	articleNewId := bson.NewObjectId()

	if id != "" {
		articleNewId = bson.ObjectIdHex(id)
	}

	//convert to integer string value coming from form
	articleTypeInt, err := strconv.ParseInt(articleType, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error(), "errCode": 70})
		return
	}

	statusBool, err := strconv.ParseBool(status)

	var statusInt int64 = 0

	if statusBool {
		statusInt = 1
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error(), "errCode": 71})
		return
	}

	article, err := articles.SaveArticle(articleNewId, langCode, authorId, authorName, articleTypeInt, title, intro, body, mainPic, tags, slug, statusInt)

	if err != nil {
		c.JSON(400, gin.H{"success": false, "msg": "problem with saving article to database", "errCode": 40})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "ok", "data": article})
}

func DeleteArticle(c *gin.Context){
	id := c.Param("id")

	msg, err := articles.DeleteArticle(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": msg, "errCode": 67})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success":true,"msg":msg, "data":"ok"})
}

/*** -- ARTICLES HANDLERS START -- ***/


/*** -- TRANSLATION HANDLERS START -- ***/

// GenerateLabels should be used only on dev - creates translation lables in mongodb
func GenerateLabels(c *gin.Context) {
	content, err := translations.CreateTranslationLabels()

	if err != nil {
		c.JSON(400, gin.H{"success": false, "msg": "Unable to create labels", "errCode": 38})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "ok", "data": content})
}

func GetTranslations(c *gin.Context) {
	content, err := translations.GetAllTranslation(c.Param("langCode"))

	fmt.Println("CODE", c.Param("langCode"))

	if err != nil {
		c.JSON(400, gin.H{"success": false, "msg": "Unable to fetch translations", "errCode": 38})
		return
	}

	total :=len(content)
	c.Header("X-Total-Count", strconv.Itoa(total))
	c.Header("Access-Control-Expose-Headers","X-Total-Count")
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "ok", "data": content})
}

func GetTranslation(c *gin.Context){
	id := c.Param("id")

	translation, err := translations.GetTranslation(id, true)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error(), "errCode": 66})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "serverTime": bson.Now(), "data": translation})
}

func SaveTranslations(c *gin.Context) {
	label := c.PostForm("label")
	langShort := c.PostForm("languageShort")
	value := c.PostForm("value")

	//check user
	/*hash := c.PostForm("authToken")

	user, err := tmusers.GetUserByHashId(hash)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "msg": "unable to find user", "authToken": hash, "errCode": 39})
		return
	}*/

	trans, err := translations.SaveTranslation(label, langShort, value)
	if err != nil {
		c.JSON(400, gin.H{"success": false, "msg": "problem with saving translation item in database", "errCode": 40})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "ok", "data": trans})
}

//TODO delete logic = AND CORS for DELETE method
func DeleteTranslation(c *gin.Context){
	id := c.Param("id")

	msg, err := translations.DeleteTranslation(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error(), "errCode": 69})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success":true,"msg":msg, "data":"ok"})
}

/*** -- TRANSLATION HANDLERS END -- ***/


/*** -- COMMENTS HANDLERS START -- ***/
func GetAllComments(c *gin.Context) {
	content, err := comments.GetAllComments()

	if err != nil {
		c.JSON(400, gin.H{"success": false, "msg": "Unable to fetch translations", "errCode": 38})
		return
	}

	total :=len(content)
	c.Header("X-Total-Count", strconv.Itoa(total))
	c.Header("Access-Control-Expose-Headers","X-Total-Count")
	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "ok", "data": content})
}

//TODO check is articleId valid
func GetCommentsForArticle(c *gin.Context){
	content, err := comments.GetCommentsForArticle(c.Param("articleId"))

	if err != nil {
		c.JSON(400, gin.H{"success": false, "msg": "Unable to retrieve articles", "errCode": 38})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "ok", "data": content})
}

func GetComment(c *gin.Context){
	id := c.Param("id")

	comment, err := comments.GetComment(id)

	if err !=nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error(), "errCode": 71})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success":true, "data":comment})
}

func SaveComment(c *gin.Context){
	parentId := c.PostForm("parentId")
	articleId := c.PostForm("articleId")
	authorName := c.PostForm("authorName")
	authorEmail := c.PostForm("authorEmail")
	body := c.PostForm("body")

	if len(body) <10 {
		c.JSON(400, gin.H{"success":false, "msg": "Comment body must be at least 10 characters long"})
		return
	}

	comment, err := comments.SaveComment(parentId, articleId, authorName, authorEmail, body)

	if err != nil {
		c.JSON(400, gin.H{"success": false, "msg": "problem with saving comment to database", "errCode": 40})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": comment})
}

func DeleteComment(c *gin.Context){
	id := c.Param("id")

	msg, err := comments.DeleteComment(id)

	if err !=nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "msg": err.Error(), "errCode": 81})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success":true,"msg":msg, "data":"ok"})
}

/*** -- COMMENT HANDLERS END -- ***/



/*** -- MARKDOWN HANDLERS START -- ***/

func Markdown(c *gin.Context){
	slug := c.Param("slug")


	data, err := markdown.GetPost(slug)

	if err != nil {
		c.JSON(400, gin.H{"success": false, "msg": "We are not able to locate item with given parameter", "errCode": 40})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "msg": "ok", "data": data})
}

/*** -- MARKDOWN HANDLERS END -- ***/



//hendle assets upload
func upload(c *gin.Context) {

	// single file
	file, _ := c.FormFile("file")
	fmt.Println(file.Filename)
}

//NOTE put it to helper
func query_param(query_data map[string][]string) *pagination.QueryParam {
	qp := new(pagination.QueryParam)
	if len(query_data["page"]) > 0 {
		page, err := strconv.Atoi(query_data["page"][0])
		if err == nil {
			qp.Page = page
		}
	}

	if len(query_data["per_page"]) > 0 {
		page, err := strconv.Atoi(query_data["per_page"][0])
		if err == nil {
			qp.Per_page = page
		}
	}

	if len(query_data["value"]) > 0 {
		qp.Value = query_data["value"][0]
	}

	if len(query_data["filter"]) > 0 {
		qp.Filter, _ = strconv.ParseBool(query_data["filter"][0])
	}

	return qp
}


