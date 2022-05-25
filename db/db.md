This is a facade for the database system used by this backend system: neo4j.

The facade does NOT attempt to wrap the query capability of neo4j (cypher syntax) in any way by using a 
query builder for the reason that, unless a lot of time can be spent developing the perfectly intuitive and
full-featured query-builder implementation, it will severely restrict the flexibility that cypher is
designed to provide. 

If anything, the facade should provide capability to elimate or *reduces* the need to write 'model' code. 
A thin layer that reduces the need for repetitive and boring boilerplate code. Also to maintain uniformity.