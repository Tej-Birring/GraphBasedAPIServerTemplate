package db

import (
	"HayabusaBackend/utils"
	"errors"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"reflect"
	"strconv"
)

type Node struct {
	MatchLabels     Labels
	MatchProperties Properties
}

func (n Node) Create(dbc *Controller) error {
	matchLabelsStr := n.MatchLabels.ToString("n")
	matchPropsString := n.MatchProperties.GetMatchQueryString()
	matchPropsParams := n.MatchProperties.GetMatchQueryParameters()
	query := fmt.Sprintf("CREATE (%s {%s}) RETURN n", matchLabelsStr, matchPropsString)
	results, err := dbc.ExecuteWriteQuery2(query, matchPropsParams)
	if err != nil {
		return err
	}
	nodesCreated := results.Counters().NodesCreated()
	if nodesCreated < 1 {
		return errors.New("no node created")
	}
	//else if nodesCreated > 1 {
	//	return errors.New("more than one node created")
	//}
	return nil
}

//func (n Node) ReadOneNoError(dbc *Controller) (*neo4j.Node, error) {
//	matchLabelsStr := n.MatchLabels.ToString("n")
//	matchPropsString := n.MatchProperties.GetMatchQueryString()
//	matchPropsParams := n.MatchProperties.GetMatchQueryParameters()
//	query := fmt.Sprintf("MATCH (%s {%s}) RETURN n", matchLabelsStr, matchPropsString)
//	rec, err := dbc.ExecuteReadQuery1(query, matchPropsParams)
//	if err != nil {
//		return nil, err
//	}
//	if len(rec) < 1 {
//		return nil, nil
//	}
//	val, ok := rec[0].Get("n")
//	if ok != true {
//		return nil, nil
//	}
//	_val, ok := val.(neo4j.Node)
//	if ok != true {
//		return nil, errors.New("the value returned is not of type Node")
//	}
//	return &_val, nil
//}

func (n Node) GetOne(dbc *Controller) (*neo4j.Node, error) {
	matchLabelsStr := n.MatchLabels.ToString("n")
	matchPropsString := n.MatchProperties.GetMatchQueryString()
	matchPropsParams := n.MatchProperties.GetMatchQueryParameters()
	query := fmt.Sprintf("MATCH (%s {%s}) RETURN n", matchLabelsStr, matchPropsString)
	rec, err := dbc.ExecuteReadQuery1(query, matchPropsParams)
	if err != nil {
		return nil, err
	}
	if len(rec) < 1 {
		return nil, errors.New("no records returned from query")
	}
	val, ok := rec[0].Get("n")
	if ok != true {
		return nil, errors.New("queried node not found in records")
	}
	_val, ok := val.(neo4j.Node)
	if ok != true {
		return nil, errors.New("the value returned from record is not a node")
	}
	return &_val, nil
}

func GetById(controller *Controller, labels Labels, id string) (*neo4j.Node, error) {
	n := Node{MatchLabels: labels, MatchProperties: QueryParameters{"id": id}}
	return n.GetOne(controller)
}

func (n Node) Update(dbc *Controller, updateData map[string]interface{}) error {
	var props Properties = updateData
	//fmt.Println("props", props)
	// produce the query
	matchLabelsStr := n.MatchLabels.ToString("n")
	matchPropsString := n.MatchProperties.GetMatchQueryString()
	matchPropsParams := n.MatchProperties.GetMatchQueryParameters()
	assignPropsString := props.GetQueryAssignString("n")
	assignPropsParams := props.GetQueryAssignParameters()
	queryParams := utils.AddGenericMaps(matchPropsParams, assignPropsParams)
	query := fmt.Sprintf("MATCH (%s {%s}) SET %s RETURN n", matchLabelsStr, matchPropsString, assignPropsString)
	summary, err := dbc.ExecuteWriteQuery2(query, queryParams)
	if err != nil {
		return err
	}
	if summary.Counters().PropertiesSet() < 1 {
		return nil // don't treat this as an error
		//return errors.New("node(s) not updated")
	}
	return nil
}

func (n Node) UpdateAllowedPropsOnly(dbc *Controller, allowProperties []string, updateData map[string]interface{}) error {
	// generate props for setting from allow list (for safety)
	props := Properties{}
	for _, key := range allowProperties {
		val, exists := updateData[key]
		if exists == true {
			props[key] = val
		}
	}
	//fmt.Println("props", props)
	//produce the query
	matchLabelsStr := n.MatchLabels.ToString("n")
	matchPropsString := n.MatchProperties.GetMatchQueryString()
	matchPropsParams := n.MatchProperties.GetMatchQueryParameters()
	assignPropsString := props.GetQueryAssignString("n")
	assignPropsParams := props.GetQueryAssignParameters()
	queryParams := utils.AddGenericMaps(matchPropsParams, assignPropsParams)
	query := fmt.Sprintf("MATCH (%s {%s}) SET %s RETURN n", matchLabelsStr, matchPropsString, assignPropsString)
	summary, err := dbc.ExecuteWriteQuery2(query, queryParams)
	if err != nil {
		return err
	}
	if summary.Counters().PropertiesSet() < 1 {
		return nil // don't treat this as an error
		//return errors.New("node(s) not updated")
	}
	return nil
}

type UpdatableProperty struct {
	Key  string
	Kind reflect.Kind // see: https://neo4j.com/docs/go-manual/current/cypher-workflow/
}

// UpdateFromString Is mainly for use with input retrieved from urlencoded forms (i.e. POST data)
func (n Node) UpdateFromString(dbc *Controller, allowProperties []UpdatableProperty, updateData map[string]string) error {
	if len(updateData) < 1 {
		return errors.New("updateData param is blank, nothing to update")
	}
	var err error
	var val interface{}
	// generate props for setting from allow list (for safety)
	props := Properties{}
	for _, prop := range allowProperties {
		strVal, exists := updateData[prop.Key]
		if len(strVal) < 1 {
			continue
		}
		if exists == true {
			switch prop.Kind {
			case reflect.String:
				val = strVal //fmt.Sprintf("'%s'", strVal)
			case reflect.Bool:
				val, err = strconv.ParseBool(strVal)
				if err != nil {
					return err
				}
			case reflect.Int64:
				val, err = strconv.ParseInt(strVal, 10, 64)
				if err != nil {
					return err
				}
			case reflect.Float64:
				val, err = strconv.ParseFloat(strVal, 64)
				if err != nil {
					return err
				}
			}
			props[prop.Key] = val
		}
	}
	// produce the query
	matchLabelsStr := n.MatchLabels.ToString("n")
	matchPropsString := n.MatchProperties.GetMatchQueryString()
	matchPropsParams := n.MatchProperties.GetMatchQueryParameters()
	assignPropsString := props.GetQueryAssignString("n")
	assignPropsParams := props.GetQueryAssignParameters()
	queryParams := utils.AddGenericMaps(matchPropsParams, assignPropsParams)
	query := fmt.Sprintf("MATCH (%s {%s}) SET %s RETURN n", matchLabelsStr, matchPropsString, assignPropsString)
	summary, err := dbc.ExecuteWriteQuery2(query, queryParams)
	if err != nil {
		return err
	}
	if summary.Counters().PropertiesSet() < 1 {
		return nil // don't treat this as an error
		//return errors.New("node(s) not updated")
	}
	return nil
}

func (n Node) Delete(dbc *Controller) error {
	matchLabelsStr := n.MatchLabels.ToString("n")
	matchPropsString := n.MatchProperties.GetMatchQueryString()
	matchPropsParams := n.MatchProperties.GetMatchQueryParameters()
	query := fmt.Sprintf("MATCH (%s {%s}) DELETE n", matchLabelsStr, matchPropsString)
	summary, err := dbc.ExecuteReadQuery2(query, matchPropsParams)
	if err != nil {
		return err
	}
	if summary.Counters().NodesDeleted() < 1 {
		return errors.New("node(s) not deleted")
	}
	return nil
}
