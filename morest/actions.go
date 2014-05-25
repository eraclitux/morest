
package morest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/eraclitux/MoREST/external/mgo"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const DEBUG = false

//Mongodb supported actions.
//Declared global to save some memory
var supportedActions = []string{}
var supportedSubActions = []string{}

//Model the action requested from client to perform on mongodb
type mongoRequest struct {
	Database   string
	Collection string
	//mydb.mycoll.action(args1, args2, args3)
	Action string
	//REF convert into one slice
	Args1 map[string]interface{}
	Args2 map[string]interface{}
	Args3 map[string]interface{}
	//Unmarshaled Json data passed as request body
	JsonPayloadSlice []interface{}
	//REF convert into slices
	SubAction1 string
	SubArgs1   string
	SubAction2 string
	SubArgs2   string
}

//Check if decoded action is sopported and coherent with http method
func (s *mongoRequest) Check(r *http.Request) error {
	isSupported := false
	for _, v := range supportedActions {
		if s.Action == v {
			isSupported = true
		}
	}
	if !isSupported {
		return fmt.Errorf("%s action is invalid or not supported", s.Action)
	}
	isSupported = false
	for _, v := range supportedSubActions {
		if s.SubAction1 == v {
			isSupported = true
		}
	}
	if !isSupported {
		return fmt.Errorf("%s action is invalid or not supported", s.SubAction1)
	}
	isSupported = false
	for _, v := range supportedSubActions {
		if s.SubAction2 == v {
			isSupported = true
		}
	}
	if !isSupported {
		return fmt.Errorf("%s action is invalid or not supported", s.SubAction2)
	}
	switch r.Method {
	case "GET":
		if !(s.Action == "find" || s.Action == "count") {
			return fmt.Errorf("Action %s not coherent with http method", s.Action)
		}
	case "POST":
		if s.Action != "insert" {
			return fmt.Errorf("Action %s not coherent with http method", s.Action)
		}
	case "DELETE":
		if s.Action != "remove" {
			return fmt.Errorf("Action %s not coherent with http method", s.Action)
		}
	default:
		return fmt.Errorf("Action %s not coherent with http method", s.Action)
	}
	return nil
}

//Helps decode mongodb functions and arguments
//This is used for find, insert, update
func getActionArgs(s string) (action string, args1, args2, args3 map[string]interface{}, er error) {
	argsPointerSlice := []*map[string]interface{}{&args1, &args2, &args3}
	actiond := strings.Split(s, "(")
	action = actiond[0]
	args := strings.Trim(actiond[1], ")")
	if len(args) == 0 {
		return
	}
	argsSlice := strings.SplitAfter(args, "},")
	for i, v := range argsSlice {
		v = strings.TrimRight(v, ",")
		//Here we dont need to args1.assert(map[string]interface{})
		//becuse we dont pass interface{} to Unmarshal but the right type
		err := json.Unmarshal([]byte(v), argsPointerSlice[i])
		if err != nil {
			er = err
			return
		}
	}
	return
}

//Helps decode mongodb functions and its arguments
//This is used for sort, limit
func getSubActionArgs(s string) (action, args string) {
	actiond := strings.Split(s, "(")
	action = actiond[0]
	args = strings.Trim(actiond[1], ")")
	return
}

//Gets json data passed as body and unmarshal it
//Works on raw []byte to avoing string conversion overhead
func unmarshalPayload(r *http.Request) ([]interface{}, error) {
	if r.ContentLength > 0 {
		interfaceSlice := []interface{}{}
		data := make([]byte, r.ContentLength)
		r.Body.Read(data)
		//[][]byte
		splittedByteData := bytes.SplitAfter(data, []byte("},"))
		for _, single := range splittedByteData {
			var mData interface{}
			single = bytes.TrimRight(single, ",")
			err := json.Unmarshal(single, &mData)
			if err != nil {
				return nil, err
			}
			interfaceSlice = append(interfaceSlice, mData)
		}
		return interfaceSlice, nil
	}
	return nil, nil
}

func (s *mongoRequest) Decode(r *http.Request) error {
	mongoQuery := strings.Split(r.RequestURI, "/")[1]
	parameters := strings.Split(mongoQuery, ".")
	if len(parameters) < 3 {
		return fmt.Errorf("Too few arguments")
	} else if len(parameters) > 5 {
		return fmt.Errorf("Too much arguments")
	}
	for i, v := range parameters {
		switch i {
		case 0:
			s.Database = v
		case 1:
			s.Collection = v
		//mongodb main function (find, insert etc)
		case 2:
			var err error
			s.Action, s.Args1, s.Args2, s.Args3, err = getActionArgs(v)
			if err != nil {
				return err
			}
			s.JsonPayloadSlice, err = unmarshalPayload(r)
			if err != nil {
				return err
			}
		//sub action (es sort, limit)
		case 3:
			s.SubAction1, s.SubArgs1 = getSubActionArgs(v)
		//sub action (es sort, limit)
		case 4:
			s.SubAction2, s.SubArgs2 = getSubActionArgs(v)
		}
	}
	if DEBUG {
		log.Printf("[DEBUG] %+v\n", s)
	}
	return s.Check(r)
}

//Decode json sort argumets to be passed to mgo Sort()
func decodeSortArgs(s string) []string {
	//FIXME parse $natural key
	returnValue := []string{}
	glob := new(map[string]interface{})
	if len(s) != 0 {
		err := json.Unmarshal([]byte(s), glob)
		if err != nil {
			return []string{""}
		}
		for k, v := range *glob {
			i := v.(float64)
			if int(i) < 0 {
				k := "-" + k
				returnValue = append(returnValue, k)
			} else {
				returnValue = append(returnValue, k)
			}
		}
	}
	return returnValue
}

//Setup the query to exucute primary action
func bakeAction(queryP **mgo.Query, s *mongoRequest, coll *mgo.Collection) error {
	switch s.Action {
	case "find":
		*queryP = coll.Find(s.Args1)
		return nil
	case "count":
		return nil
	case "insert":
		return nil
	default:
		return fmt.Errorf("Unable to execute %s", s.Action)
	}
}

//Setup the query to exucute secondary actions
func bakeSubActions(queryP **mgo.Query, s *mongoRequest, coll *mgo.Collection) error {
	//No subactions on these
	if s.Action == "count" ||
		s.Action == "insert" ||
		s.Action == "remove" {
		return nil
	}
	if s.SubAction1 != "" {
		switch s.SubAction1 {
		case "sort":
			*queryP = queryP.Sort(decodeSortArgs(s.SubArgs1)...)
		case "limit":
			num, err := strconv.Atoi(s.SubArgs1)
			if err != nil {
				return fmt.Errorf("Unable to convert limit argument")
			} else {
				*queryP = queryP.Limit(num)
			}
		default:
			return fmt.Errorf("Unable to execute %s", s.SubAction1)
		}
	}
	if s.SubAction2 != "" {
		switch s.SubAction2 {
		case "sort":
			*queryP = queryP.Sort(decodeSortArgs(s.SubArgs2)...)
			return nil
		case "limit":
			num, err := strconv.Atoi(s.SubArgs2)
			if err != nil {
				return fmt.Errorf("Unable to convert limit argument")
			} else {
				*queryP = queryP.Limit(num)
				return nil
			}
		default:
			return fmt.Errorf("Unable to execute %s", s.SubAction1)
		}
	}
	return nil
}

//Exectute query on mongodb
func executeQuery(query *mgo.Query, s *mongoRequest, coll *mgo.Collection) (interface{}, error) {
	gdata := new([]interface{})
	switch s.Action {
	case "find":
		err := query.All(gdata)
		if err != nil {
			return []byte{}, err
		}
		return json.Marshal(gdata)
	case "insert":
		payloadLen := len(s.JsonPayloadSlice)
		if payloadLen > 0 {
			err := coll.Insert(s.JsonPayloadSlice...)
			if err != nil {
				return []byte{}, err
			}
			res := fmt.Sprintf("{\"nInserted\":%d}", payloadLen)
			return []byte(res), nil
		} else {
			err := coll.Insert(s.Args1)
			if err != nil {
				return []byte{}, err
			}
			return []byte(`{"nInserted":1}`), nil
		}
	case "remove":
		//TODO add tests
		if v, ok := s.Args2["justOne"]; ok && v.(float64) == 1 {
			err := coll.Remove(s.Args1)
			if err != nil {
				return []byte{}, err
			}
			return []byte(`{"nRemoved":1}`), nil
		}
		info, err := coll.RemoveAll(s.Args1)
		if err != nil {
			return []byte{}, err
		}
		returnString := fmt.Sprintf("{\"nRemoved\":%d}", info.Removed)
		return []byte(returnString), nil
	case "count":
		n, err := coll.Count()
		if err != nil {
			return []byte{}, err
		}
		number := strconv.Itoa(n)
		return number, nil
	default:
		return []byte{}, fmt.Errorf("Unable to execute %s", s.Action)
	}
}

//Performs decoded action on mongodb.
func (s *mongoRequest) Execute(msession *mgo.Session, r *http.Request) (interface{}, error) {
	//FIXME add session to mongoRequest struct?
	//TODO test copy/clone/new against consistency modes
	err := s.Decode(r)
	if err != nil {
		return nil, err
	}
	session := msession.Copy()
	defer session.Close()
	coll := session.DB(s.Database).C(s.Collection)
	query := new(mgo.Query)
	bakeAction(&query, s, coll)
	bakeSubActions(&query, s, coll)
	jdata, err := executeQuery(query, s, coll)
	if err != nil {
		return nil, err
	}
	return jdata, nil
}

func MakeMainHandler(msession *mgo.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if DEBUG {
			log.Printf("[DEBUG] Request struct: %+v\n", r)
		}
		mReq := mongoRequest{}
		iData, err := mReq.Execute(msession, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		switch aData := iData.(type) {
		case string:
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "%s\n", aData)
		case []byte:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "%s\n", string(aData))
		}
	}
}

func init() {
	//To check against user requests
	supportedActions = []string{"find", "insert", "remove", "count"}
	supportedSubActions = []string{"sort", "limit", ""}
}
