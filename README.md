# qp

`qp` is a standalone CLI utility to analyze a SQL query (Postgres or Mysql) with the context of the database schema, amount of data, and more to help identify ways to imrpove the query, or understand the performance of the query and the effects on the database in general.

This utility can use OpenAI's gpt-4 model to construct specific remediation recommandations. 

## Connecting

Use a connection uri:

Mysql:
mysql://username:password@host:port/database

Postgres: 

## FAQ

What about transactions?
Yeah, these aren't supported today. Track this issue and if you have suggestions or ideas on how to plan a transaction using QueryPlan, comment.
