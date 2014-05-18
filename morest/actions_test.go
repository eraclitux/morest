package morest

import (
	"fmt"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/eraclitux/morest/external/mgo"
	"log"
	"strings"
)


type testCase struct {
	Req            *http.Request
	//To test Decode method
	expectedResult *mongoRequest
	Err            error
	//To test MakeMainHandler 
	ExpectedJson []map[string]interface{}
}

var testCases []testCase

func buildTestCases() []testCase {
	cases := []testCase{}
	singleCase := testCase{}
	caseArgs1 := make(map[string]interface{})
	//================================================
	singleCase.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection.find({name:pippo}).sort().limit(5)",
	}
	singleCase.Err = fmt.Errorf("Error for invalid json formatting")
	cases = append(cases, singleCase)
	//================================================
	singleCase = testCase{}
	singleCase.Req = &http.Request{
		Method:     "POST",
		RequestURI: "/testing-db.testing-collection.find({\"name\":\"pippo\"}).sort().limit(5)",
	}
	singleCase.Err = fmt.Errorf("We expect an error")
	cases = append(cases, singleCase)
	//================================================
	singleCase = testCase{}
	singleCase.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection.fund({\"name\":\"pippo\"}).sort().limit(5)",
	}
	singleCase.Err = fmt.Errorf("We expect an error")
	cases = append(cases, singleCase)
	//================================================
	singleCase = testCase{}
	singleCase.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection.find({\"num\":{\"$gt\":4}}).sort().limit(2)",
	}
	caseArgs1 = make(map[string]interface{})
	caseArgs1["num"] = map[string]interface{}{"$gt":float64(4)}
	singleCase.expectedResult = &mongoRequest{
		Database:"testing-db", Collection:"testing-collection", Action:"find", Args1:caseArgs1,
		SubAction1:"sort", SubArgs1:"", SubAction2:"limit", SubArgs2:"2",
	}
	singleCase.ExpectedJson = append(
		singleCase.ExpectedJson,
		//Convert to float64 because so does Unmarshal()
		map[string]interface{}{"name":"Pippo-5", "num":float64(5)},
	)
	singleCase.ExpectedJson = append(
		singleCase.ExpectedJson,
		map[string]interface{}{"name":"Pippo-6", "num":float64(6)},
	)
	cases = append(cases, singleCase)
	//================================================
	singleCase = testCase{}
	singleCase.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection",
	}
	singleCase.Err = fmt.Errorf("We expect an error")
	cases = append(cases, singleCase)
	//================================================
	singleCase = testCase{}
	singleCase.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection.find().sort().limit(2)",
	}
	caseArgs1 = make(map[string]interface{})
	singleCase.expectedResult = &mongoRequest{
		Database:"testing-db", Collection:"testing-collection", Action:"find", Args1:caseArgs1,
		SubAction1:"sort", SubArgs1:"", SubAction2:"limit", SubArgs2:"2",
	}
	singleCase.ExpectedJson = append(
		singleCase.ExpectedJson,
		map[string]interface{}{"name":"Pippo-0", "num":float64(0)},
	)
	singleCase.ExpectedJson = append(
		singleCase.ExpectedJson,
		map[string]interface{}{"name":"Pippo-1", "num":float64(1)},
	)
	cases = append(cases, singleCase)
	//================================================
	singleCase = testCase{}
	singleCase.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection.find().limit(2).sort({\"name\":-1})",
	}
	caseArgs1 = make(map[string]interface{})
	singleCase.expectedResult = &mongoRequest{
		Database:"testing-db", Collection:"testing-collection", Action:"find", Args1:caseArgs1,
		SubAction1:"limit", SubArgs1:"2", SubAction2:"sort", SubArgs2:"{\"name\":-1}",
	}
	singleCase.ExpectedJson = append(
		singleCase.ExpectedJson,
		map[string]interface{}{"name":"Pippo-99", "num":float64(99)},
	)
	singleCase.ExpectedJson = append(
		singleCase.ExpectedJson,
		map[string]interface{}{"name":"Pippo-98", "num":float64(98)},
	)
	cases = append(cases, singleCase)
	//================================================
	singleCase = testCase{}
	singleCase.Req = &http.Request{
		Method:     "GET",
		RequestURI: "/testing-db.testing-collection.insert({\"name\":\"Pippo-XX\",\"num\":42})",
	}
	caseArgs1 = make(map[string]interface{})
	caseArgs1["name"] = "Pippo-XX"
	caseArgs1["num"] = float64(42)
	singleCase.expectedResult = &mongoRequest{Database:"testing-db", Collection:"testing-collection", Action:"insert", Args1:caseArgs1}
	singleCase.ExpectedJson = append(
		singleCase.ExpectedJson,
		map[string]interface{}{"nInserted":float64(1)},
	)
	cases = append(cases, singleCase)
	//================================================
	singleCase = testCase{}
	singleCase.Req = &http.Request{
		Method:     "DELETE",
		RequestURI: "/testing-db.testing-collection.remove({\"name\":\"Pippo-42\"})",
	}
	caseArgs1 = make(map[string]interface{})
	caseArgs1["name"] = "Pippo-42"
	singleCase.expectedResult = &mongoRequest{
		Database:"testing-db", Collection:"testing-collection", Action:"remove", Args1:caseArgs1,
	}
	singleCase.ExpectedJson = append(
		singleCase.ExpectedJson,
		map[string]interface{}{"nRemoved":float64(1)},
	)
	cases = append(cases, singleCase)
	//================================================
	singleCase = testCase{}
	singleCase.Req = &http.Request{
		Method:     "DELETE",
		RequestURI: "/testing-db.testing-collection.update({\"name\":\"mario\",\"num\":42},{\"param\":1},{\"param\":3})",
	}
	caseArgs1 = make(map[string]interface{})
	caseArgs1["name"] = "mario"
	caseArgs1["num"] = float64(42)
	singleCase.expectedResult = &mongoRequest{
		Database:"testing-db", Collection:"testing-collection", Action:"update", Args1:caseArgs1,
	}
	//cases = append(cases, singleCase)
	return cases
}

type dummyMongoData struct {
	Name string
	Num int
}
//Fill test db with dummy data
func arrangeDB(session *mgo.Session) {
	session.DB("testing-db").DropDatabase()
	for i := 0; i < 100; i++ {
		dummy := dummyMongoData{}
		dummy.Name = fmt.Sprintf("Pippo-%d", i)
		dummy.Num = i
		session.DB("testing-db").C("testing-collection").Insert(dummy)
	}
}

func compareJsonResponses(r string, expectSlice []map[string]interface{}) bool {
	defer func() bool {
		if r := recover(); r != nil {
			fmt.Println("Panic in compareJsonResponses", r)
		}
		return false
	}()
	resp := map[string]interface{}{}
	r = strings.TrimPrefix(r, "[")
	r = strings.TrimSuffix(r, "]\n")
	resultSlice := strings.SplitAfter(r, "},")
	if len(resultSlice) != len(expectSlice) {
		return false
	}
	for i, v := range resultSlice {
		v = strings.TrimSuffix(v, ",")
		err := json.Unmarshal([]byte(v), &resp)
		if err != nil {
			fmt.Println("[ERROR] Unmarshalling", err, v)
			return false
		}
		//We cannot predict _id value so drop it
		//Sadly mgo Find() doesnt support projection so we need 
		//to unmarshal responses for a comparison
		delete(resp, "_id")
		if !compareMapInterfaces(resp, expectSlice[i]) {
			return false
		}
	}
	return true
}

//This test needs mongodb running @ localhost
func TestMakeMainHandler (t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fail()
			log.Printf("Panic in TestMakeMainHandler: %s. Maybe mongodb unreacheble?", r)
		}
	}()
	msession, err := mgo.Dial("localhost")
	if err != nil {
		log.Print("Error connecting to Mongodb ", err)
	}
	defer msession.Close()
	arrangeDB(msession)

	handler := MakeMainHandler(msession)
	for i, singleCase := range testCases {
		if singleCase.Err != nil {
			continue
		}
		recorder := httptest.NewRecorder()
		handler(recorder, singleCase.Req)
		if !compareJsonResponses(recorder.Body.String(), singleCase.ExpectedJson) {
			fmt.Println("In case", i + 1, singleCase.Req.RequestURI)
			fmt.Printf("Got: %+v\nExpect: %+v\n", recorder.Body.String(), singleCase.ExpectedJson)
			t.Fail()
		}
	}
}
func compareMapInterfaces(o, p map[string]interface{}) bool {
	//TODO add more types
	if len(o) != len(p) {
		return false
	}
	for k, v := range o {
		switch vv := v.(type) {
		case string:
			//fmt.Println(k, "is string", vv)
			if vv != p[k].(string) {
				return false
			}
		case float64:
			//fmt.Println(k, "is float64", vv)
			if vv != p[k].(float64) {
				return false
			}
		case map[string]interface{}:
			//Recursion rocks
			return compareMapInterfaces(vv, p[k].(map[string]interface{}))
		default:
			fmt.Printf("%v is of a type I don't know how to handle %T(%v)\n", k, v, v)
			return false
		}
	}
	return true
}
func TestDecode(t *testing.T) {
	for i, singleCase := range testCases {
		testStruct := mongoRequest{}
		err := testStruct.Decode(singleCase.Req)
		if singleCase.Err != nil && err != nil {
			continue
		} else if err != nil {
			fmt.Println("[FAIL] unexpected error")
			fmt.Println(singleCase.Req.RequestURI)
			fmt.Println(err)
			t.Fail()
		} else if !(testStruct.Database == singleCase.expectedResult.Database &&
			testStruct.Collection == singleCase.expectedResult.Collection &&
			testStruct.Action == singleCase.expectedResult.Action &&
			testStruct.SubAction1 == singleCase.expectedResult.SubAction1 &&
			testStruct.SubArgs1 == singleCase.expectedResult.SubArgs1 &&
			testStruct.SubAction2 == singleCase.expectedResult.SubAction2 &&
			testStruct.SubArgs2 == singleCase.expectedResult.SubArgs2 &&
			compareMapInterfaces(testStruct.Args1, singleCase.expectedResult.Args1)) {
			fmt.Println("===============")
			fmt.Println("Checking case", i+1, singleCase.Req.RequestURI)
			fmt.Printf("Expected: %+v\n", *singleCase.expectedResult)
			fmt.Printf("Got: %+v\n", testStruct)
			t.Fail()
		}
	}
}

type decodeSortCase struct {
	Case     string
	Expected []string
}

func compareSortSlices(o, p []string) bool {
	if len(o) != len(p) {
		return false
	}
	for i, v := range o {
		if v != p[i] {
			return false
		}
	}
	return true
}
func TestDecodeSortArgs(t *testing.T) {
	//TODO add $natural case
	cases := []decodeSortCase{}
	oneCase := decodeSortCase{"{\"name\":-1,\"age\":1}", []string{"-name", "age"}}
	cases = append(cases, oneCase)
	for _, singleCase := range cases {
		got := decodeSortArgs(singleCase.Case)
		if !compareSortSlices(got, singleCase.Expected) {
			fmt.Printf("expected: %+v\n", singleCase)
			fmt.Printf("got: %+v\n", got)
			t.Fail()
		}
	}
}
func init() {
	testCases = buildTestCases()
}
