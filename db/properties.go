package db

import "reflect"

type Properties map[string]interface{}

/*
	GetMatchQueryParameters
	Converts map[<key>]<val> to map[MATCH<key>]<val>
*/
func (props Properties) GetMatchQueryParameters() map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range props {
		// special case to omit empty string values
		if reflect.TypeOf(v).Kind() == reflect.String && len(v.(string)) < 1 {
			m["MATCH"+k] = nil
			continue
		}
		m["MATCH"+k] = v
	}
	return m
}

/*
	GetMatchQueryString
	Converts map[<key>]<val> to "<key>:$MATCH<key>, ..."
*/
func (props Properties) GetMatchQueryString() string {
	if len(props) < 1 {
		return ""
	}
	// build
	count := 0
	retStr := "" //"{"
	for k, _ := range props {
		if count > 0 {
			retStr += ", "
		}
		retStr += k + ":" + "$MATCH" + k
		count++
	}
	//retStr += "}"
	// return
	return retStr
}

/*
	GetMatchQueryParameters
	Converts map[<key>]<val> to map[MATCH<key>]<val>
*/
func (props Properties) GetQueryAssignParameters() map[string]interface{} {
	m := map[string]interface{}{}
	for k, v := range props {
		// special case to omit empty string values
		//if reflect.TypeOf(v).Kind() == reflect.String && len(v.(string)) < 1 {
		//	m["MATCH"+k] = nil
		//	continue
		//}
		m["SET"+k] = v
	}
	return m
}

/*
	GetMatchQueryString
	Converts map[<key>]<val> to "<key>:$MATCH<key>, ..."
*/
func (props Properties) GetQueryAssignString(varPrefix string) string {
	if len(props) < 1 {
		return ""
	}
	// build
	count := 0
	retStr := "" //"{"
	for k, _ := range props {
		if count > 0 {
			retStr += ", "
		}
		retStr += varPrefix + "." + k + "=" + "$SET" + k
		count++
	}
	//retStr += "}"
	// return
	return retStr
}
