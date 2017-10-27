package translations

import (
	"fmt"
	"log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	//"strings"
)

type Translation struct {
	ID          	bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Label       	string        `json:"label" bson:"label"`
	LanguageShort	string 		  `json:"languageShort" bson:"languageShort"`
	Value        	string        `json:"value" bson:"value"`
}

type TranslationValues []Translation

var (
	MgoSession      *mgo.Session
	MongoServer     = "localhost"
	MongoUser       = "" //os.Getenv("MGOUSER")
	MongoPass       = ""
	DB              = "blog"
	TRANSLATIONSCollection = "Translations"
)

//init connection to mongo database
func init() {
	session, err := mgo.Dial(MongoServer)
	if err != nil {
		log.Fatal(err.Error())
	}
	MgoSession = session
}

//should be called only once - here for test purposes only
//NOTE this is just test function to insert some dummy data into db
func CreateTranslationLabels() (string, error) {
	sess := MgoSession.Clone()
	defer sess.Close()

	sess.SetMode(mgo.Monotonic, true)

	transTBL := sess.DB(DB).C(TRANSLATIONSCollection)

	trans := Translation{}
	trans.ID = bson.NewObjectId()
	trans.Label = "HOME"
	trans.LanguageShort = "US"
	trans.Value = "Home"

	//now insert
	err := transTBL.Insert(trans)

	if err != nil {
		return "Unable to save translations into database", err
	}

	return "Translations saved into database", nil
}

func GetTranslation(label string, id bool) (Translation, error) {
	sess := MgoSession.Clone()
	defer sess.Close()


	sess.SetMode(mgo.Monotonic, true)
	con := sess.DB(DB).C(TRANSLATIONSCollection)
	trans := Translation{}

	//fetch it by id
	if (id){
		err := con.Find(bson.M{"_id": bson.ObjectIdHex(label)}).One(&trans)

		return trans, err
	}

	//otherwise fetch by label string

	fmt.Println("Translation String: ", label)
	err := con.Find(bson.M{"label": label}).One(&trans)

	return trans, err
}

func GetAllTranslation(langCode string) (TranslationValues, error) {
	sess := MgoSession.Clone()
	defer sess.Close()

	sess.SetMode(mgo.Monotonic, true)
	con := sess.DB(DB).C(TRANSLATIONSCollection)

	query := bson.M{}

	if langCode != "" {
		query = bson.M{"languageShort": langCode}
	}

	fmt.Println("Query", query)

	allTranslations := TranslationValues{}
	err := con.Find(query).Sort("languageShort").All(&allTranslations)

	//fmt.Println(allScores)

	if err != nil {
		return allTranslations, err
	}

	return allTranslations, nil
}

func SaveTranslation(label string, languageShort string, value string) (string, error) {
	sess := MgoSession.Clone()
	defer sess.Close()

	sess.SetMode(mgo.Monotonic, true)

	transTBL := sess.DB(DB).C(TRANSLATIONSCollection)

	trans := Translation{}
	trans.ID = bson.NewObjectId()
	trans.Label = label
	trans.LanguageShort = languageShort
	trans.Value = value
	//extLB.AddedTime = bson.Now()

	//fmt.Println("ALL: userId: " + userId.String() + " username: " + username + " labelName: " + labelName)

	//now insert
	err := transTBL.Insert(trans)

	if err != nil {
		return "Unable to save translation string", err
	}

	return "Translation string saved in database", nil
}


//delete translation from database
func DeleteTranslation(id string) (string, error){

	sess := MgoSession.Clone()
	defer sess.Clone()

	sess.SetMode(mgo.Monotonic, true)
	con := sess.DB(DB).C(TRANSLATIONSCollection)

	err := con.Remove(bson.M{"_id": bson.ObjectIdHex(id)})

	if err != nil {
		return "Unable to delete translation ", err
	}

	return "Translation removed from database", nil
}

