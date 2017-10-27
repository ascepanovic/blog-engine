package comments

import (
	"fmt"
	"log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	//"strings"
	"time"
)

type Comment struct {
	ID          	bson.ObjectId `json:"id" bson:"_id,omitempty"`
	ParentId       	string        `json:"parentId" bson:"parentId"`
	ArticleId		string 		  `json:"articleId" bson:"articleId"`
	AuthorName		string 		  `json:"authorName" bson:"authorName"`
	AuthorEmail     string        `json:"authorEmail" bson:"authorEmail"`
	Body	        string        `json:"body" bson:"body"`
	DateAdded 		time.Time 	  `json:"dateAdded" bson:"dateAdded"`
	Status 			int64 		  `json:"status" bson:"status"`
}

type Comments []Comment

var (
	MgoSession      *mgo.Session
	MongoServer     = "localhost"
	MongoUser       = "" //os.Getenv("MGOUSER")
	MongoPass       = ""
	DB              = "blog"
	COMMENTSCollection = "Comments"
)

//init connection to mongo database
func init() {
	session, err := mgo.Dial(MongoServer)
	if err != nil {
		log.Fatal(err.Error())
	}
	MgoSession = session
}

//get comments for specific article
func GetCommentsForArticle(articleId string) (Comments, error) {
	sess := MgoSession.Clone()
	defer sess.Close()

	sess.SetMode(mgo.Monotonic, true)
	con := sess.DB(DB).C(COMMENTSCollection)

	fmt.Println("Getting comments for article with id: ", articleId)

	comments := Comments{}
	err := con.Find(bson.M{"articleId": articleId}).Sort("parentId").All(&comments)

	return comments, err
}

//Get all comments
func GetAllComments() (Comments, error) {
	sess := MgoSession.Clone()
	defer sess.Close()

	sess.SetMode(mgo.Monotonic, true)
	con := sess.DB(DB).C(COMMENTSCollection)

	query := bson.M{}

	fmt.Println("Query", query)

	allComments := Comments{}
	err := con.Find(query).Sort("dateAdded").All(&allComments)

	//fmt.Println(allComments)

	if err != nil {
		return allComments, err
	}

	return allComments, nil
}

func SaveComment(parentId string, articleId string, authorName string, authorEmail string, body string) (string, error) {
	sess := MgoSession.Clone()
	defer sess.Close()

	sess.SetMode(mgo.Monotonic, true)

	cmntTBL := sess.DB(DB).C(COMMENTSCollection)

	comment := Comment{}
	comment.ID = bson.NewObjectId()
	comment.ArticleId = articleId
	comment.ParentId = parentId
	comment.AuthorName = authorName
	comment.AuthorEmail = authorEmail
	comment.Body = body
	comment.DateAdded = bson.Now()
	comment.Status = 1

	//now insert
	err := cmntTBL.Insert(comment)

	if err != nil {
		return "Unable to save comment", err
	}

	return "Comment saved in database", nil
}

//get single comment
func GetComment(id string)(Comment, error){
	sess := MgoSession.Clone()
	defer sess.Close()

	sess.SetMode(mgo.Monotonic, true)
	con := sess.DB(DB).C(COMMENTSCollection)


	comment := Comment{}
	err := con.Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&comment)

	return comment, err
}

//delete comment
func DeleteComment(id string) (string, error){
	sess := MgoSession.Clone()
	defer sess.Clone()

	sess.SetMode(mgo.Monotonic, true)
	con := sess.DB(DB).C(COMMENTSCollection)

	err := con.Remove(bson.M{"_id": bson.ObjectIdHex(id)})

	if err != nil{
		return "Cant delete this comment from database", err
	}

	return "Comment deleted from database", nil
}


