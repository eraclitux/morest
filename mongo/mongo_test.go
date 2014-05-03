package mongo

import (
	"testing"
	"fmt"
	"net/http"
)
type testCase struct {
	Req *http.Request
	expectedResult *mongoRequest
}
func buildCases() []*testCase {
	cases := []*testCase{}
	case1 := testCase{}
	case1.Req = &http.Request{
		Method : "GET",
		RequestURI:"/database.collection.find({name:pippo}).sort().limit(5)",
	}
	case1.expectedResult = &mongoRequest{"database", "collection", "find", "{name:pippo}", "sort", "", 5}
	cases = append(cases, &case1)
	case2 := testCase{}
	case2.Req = &http.Request{
		Method : "POST",
		RequestURI:"/database.collection.find({name:pippo}).sort().limit(5)",
	}
	case2.expectedResult = &mongoRequest{}
	cases = append(cases, &case2)
	return cases

}
func TestDecode(t *testing.T) {
	for _, singleCase := range buildCases() {
		testStruct := mongoRequest{}
		testStruct.Decode(singleCase.Req)
		if testStruct != *singleCase.expectedResult {
			fmt.Printf("[FAIL] %s\n", singleCase.Req.RequestURI)
			fmt.Printf("Expected: %+v\n", *singleCase.expectedResult)
			fmt.Printf("Got: %+v\n", testStruct)
			t.Fail()
		}
	}
}
