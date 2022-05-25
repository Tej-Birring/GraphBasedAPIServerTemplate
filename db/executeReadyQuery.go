package db

import "github.com/neo4j/neo4j-go-driver/v4/neo4j"

type QueryParameters = map[string]interface{}

func (c Controller) ExecuteReadQuery1(query string, queryParams QueryParameters) ([]*neo4j.Record, error) {
	// create session
	sess := c.driver.NewSession(neo4j.SessionConfig{})
	defer sess.Close()
	// execute query
	res, err := sess.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		res, err := tx.Run(query, queryParams)
		if err != nil {
			return nil, err
		}
		rec, err := res.Collect()
		if err != nil {
			return nil, err
		}
		return rec, err
	})
	if err != nil {
		return nil, err
	}
	// return the raw result
	return res.([]*neo4j.Record), err
}

func (c Controller) ExecuteReadQuery2(query string, queryParams QueryParameters) (neo4j.ResultSummary, error) {
	// create session
	sess := c.driver.NewSession(neo4j.SessionConfig{})
	defer sess.Close()
	// execute query
	res, err := sess.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		res, err := tx.Run(query, queryParams)
		if err != nil {
			return nil, err
		}
		summary, err := res.Consume()
		if err != nil {
			return nil, err
		}
		return summary, err
	})
	if err != nil {
		return nil, err
	}
	// return the raw result
	return res.(neo4j.ResultSummary), err
}
