package utils

import (
	"encoding/json"
	"github.com/dchest/uniuri"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"time"
)

// GetRandomString
// used alphanumeric, mixed case, useful for ids and hashes
func GetRandomString(length int) string {
	return uniuri.NewLen(length)
}

// GetRandomString2
type StringType int

const (
	Alpha StringType = iota + 1
	AlphaNumeric
	Numeric
)

type StringCase int

const (
	UpperCase StringCase = iota + 1
	LowerCase
	MixedCase
)

var alpha = []string{
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
}

var alphaNumeric = []string{
	"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
}

var numeric = []string{
	"0", "1", "2", "3", "4", "5", "6", "7", "8", "9",
}

// GetRandomString2 TODO: Complete implementation
func GetRandomString2(stringType StringType, stringCase StringCase, resultLength int) string {
	rand.Seed(time.Now().UnixNano())
	var chars []string
	switch stringType {
	case Alpha:
		chars = alpha
		break
	case AlphaNumeric:
		chars = alphaNumeric
		break
	case Numeric:
		chars = numeric
		break
	default:
		panic("Invalid option!")
	}
	lengthSelection := len(chars)
	var result = []string{}
	for i := 0; i < resultLength; i++ {
		sel := rand.Intn(lengthSelection)
		result = append(result, chars[sel])
	}
	if stringType == Numeric {
		return strings.Join(result, "")
	}
	switch stringCase {
	case UpperCase:
		resultStr := strings.Join(result, "")
		return strings.ToUpper(resultStr)
	case LowerCase:
		resultStr := strings.Join(result, "")
		return strings.ToLower(resultStr)
	default:
		return strings.Join(result, "")
	}
}

func AddGenericMaps(maps ...map[string]interface{}) map[string]interface{} {
	ret := map[string]interface{}{} // create new map to return
	for _, m := range maps {        // for each map
		for k, v := range m { // add key-val pair to return value
			ret[k] = v
		}
	}
	return ret // return the result
}

// ConvertStructToGenericMap Takes a struct input and outputs a map[string]interface{}
func ConvertStructToGenericMap(in interface{}) map[string]interface{} {
	// TODO: Check that in is indeed a struct!

	retVal := map[string]interface{}{}

	v := reflect.ValueOf(in)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		key := (typeOfS.Field(i)).Name
		val := v.Field(i).Interface()

		// for debug:
		//fmt.Println(key, reflect.TypeOf(key).Kind())
		//fmt.Println(val, reflect.TypeOf(val).Kind())

		// uncomment if needed:
		// special case: always omit if empty (zero val)
		//if v.Field(i).IsZero() {
		//	continue
		//}

		retVal[key] = val
	}

	return retVal
}

// GetJsonKeysUsedByStruct Takes a struct and returns its json tags as slice of strings
// Current implementation assumes that the tag entry is not comma-seperated string,
// we can cater for that when we need to use it
func GetJsonKeysUsedByStruct(in interface{}) []string {
	var retVal []string
	tStruct := reflect.TypeOf(in)
	nKeys := tStruct.NumField()
	for i := 0; i < nKeys; i++ {
		tagVal := tStruct.Field(i).Tag.Get("json")
		retVal = append(retVal, tagVal)
	}
	return retVal
}

func ReadJSONFile(path string) (*map[string]interface{}, error) {
	// Open JSON file
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	// Read the JSON data
	raw, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	result := &map[string]interface{}{}
	err = json.Unmarshal(raw, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func ReadJSON(reader io.Reader) (*map[string]interface{}, error) {
	raw, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	result := &map[string]interface{}{}
	err = json.Unmarshal(raw, result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
