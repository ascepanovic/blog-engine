package pagination


import (
	"encoding/json"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	db = "blog"
	itemsPerPage = 25;
)

type PageMetadata struct {
	Page        int `json:"page"`
	Per_page    int `json:"per_page"`
	Page_count  int `json:"page_count"`
	Total_count int `json:"total_count"`
}

type QueryParam struct {
	Page      int    `json:"page"`
	Per_page  int    `json:"per_page"`
	Value     string `json:"value"`
	Sort      string `json:"sort"`
	Direction string `json:"direction"`
	Filter    bool   `json:"filter"`
}

type PaginatedResults struct {
	Pagination PageMetadata
	Results    json.RawMessage
}

func GetPaginatedResults(session *mgo.Session, collection string, q bson.M, sortField string, qp *QueryParam) (p *PaginatedResults, err error) {
	sess := session.Clone()
	defer sess.Close()

	pageReturn := &PaginatedResults{}

	itemsPerPage := itemsPerPage

	if qp.Page == 0 {
		pageReturn.Pagination.Page = 1
	} else {
		pageReturn.Pagination.Page = qp.Page
	}
	if qp.Per_page == 0 {
		pageReturn.Pagination.Per_page = itemsPerPage
	} else {
		pageReturn.Pagination.Per_page = qp.Per_page
	}

	c := sess.DB(db).C(collection)
	count, err := c.Find(q).Count()
	if err != nil {
		return pageReturn, err
	}

	var results []interface{}
	midQuery := c.Find(q).Skip((pageReturn.Pagination.Page - 1) * pageReturn.Pagination.Per_page).Limit(pageReturn.Pagination.Per_page)
	if sortField != "" {
		midQuery = midQuery.Sort(sortField)
	}

	err = midQuery.All(&results)

	pageReturn.Results, err = json.Marshal(results)

	if err != nil {
		return pageReturn, err
	}

	pageReturn.Pagination.Total_count = count
	pageReturn.Pagination.Page_count = count / pageReturn.Pagination.Per_page

	if pageReturn.Pagination.Page_count*pageReturn.Pagination.Per_page < pageReturn.Pagination.Total_count {
		pageReturn.Pagination.Page_count++
	}

	return pageReturn, err

}

