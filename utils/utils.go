package utils

import (
	"github.com/dchest/uniuri"
	"reflect"
)

func RandomString(length int) string {
	return uniuri.NewLen(length)
}

//func SumMaps(maps ...map[interface{}]interface{}) map[interface{}]interface{} {
//	ret := map[interface{}]interface{}{} // create new map to return
//	for _, m := range maps {             // for each map
//		for k, v := range m { // add key-val pair to return value
//			ret[k] = v
//		}
//	}
//	return ret // return the result
//}

func SumMaps(maps ...map[string]interface{}) map[string]interface{} {
	ret := map[string]interface{}{} // create new map to return
	for _, m := range maps {        // for each map
		for k, v := range m { // add key-val pair to return value
			ret[k] = v
		}
	}
	return ret // return the result
}

// StructToGenericMap Takes a struct input and outputs a map[string]interface{}
func StructToGenericMap(in interface{}) map[string]interface{} {
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

// StructJsonKeys Takes a struct and returns its json tags as slice of strings
// Current implementation assumes that the tag entry is not comma-seperated string,
// we can cater for that when we need to use it
func StructJsonKeys(in interface{}) []string {
	var retVal []string
	tStruct := reflect.TypeOf(in)
	nKeys := tStruct.NumField()
	for i := 0; i < nKeys; i++ {
		tagVal := tStruct.Field(i).Tag.Get("json")
		retVal = append(retVal, tagVal)
	}
	return retVal
}
