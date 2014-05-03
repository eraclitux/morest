package mongo

import (
	"encoding/json"
	"fmt"
	"github.com/eraclitux/morest/external/mgo"
	"github.com/eraclitux/morest/external/mgo/bson"
	"log"
	"net/http"
	"strings"
	"strconv"
)

const DEBUG = true

//model the action requested from client to perform on mongodb
type mongoRequest struct {
	Database   string
	Collection string
	Action     string
	Args       string
	SubAction  string
	SubArgs	   string
	Limit	   int
}

func (s *mongoRequest) Decode(r *http.Request) error {
	r.ParseForm() //debug
	mongoQuery := strings.Split(r.RequestURI, "/")[1]
	for i, v := range strings.Split(mongoQuery, ".") {
		if DEBUG {
			fmt.Println(i, v)
		}
		switch i {
		case 0:
			s.Database = v
		case 1:
			s.Collection = v
		case 2:
			//mongodb function to permform on data (find, insert etc)
			//TODO a function!
			action := strings.Split(v, "(")
			s.Action = action[0]
			s.Args = strings.Trim(action[1], ")")
		case 3:
			//sub action (es sort, limit)
			action := strings.Split(v, "(")
			s.SubAction = action[0]
			s.SubArgs = strings.Trim(action[1], ")")
		case 4:
			action := strings.Split(v, "(")
			s.Limit, _ = strconv.Atoi(strings.Trim(action[1], ")"))
		}
	}
	return nil
}

func (s *mongoRequest) Execute(msession *mgo.Session) ([]byte, error) {
	session := msession.Copy()
	defer session.Close()
	coll := session.DB(s.Database).C(s.Collection)
	gdata := new([]interface{})
	coll.Find(bson.M{"my-key": "my_value"}).Limit(6).All(gdata)
	jdata, err := json.Marshal(gdata)
	if err != nil {
		return nil, err
	}
	return jdata, nil
}

func MainHandler(msession *mgo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if DEBUG {
			log.Printf("[DEBUG] Request.URL struct: %+v\n", r.URL)
			log.Printf("[DEBUG] Request struct: %+v\n", r)
		}
		mReq := mongoRequest{}
		mReq.Decode(r)
		jdata, _ := mReq.Execute(msession)
		fmt.Fprintf(w, "Data: %v\n", string(jdata))
	}
}
