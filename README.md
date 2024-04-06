# qp

`qp` is a standalone CLI utility to analyze a SQL query (Postgres or Mysql) with the context of the database schema, amount of data, and more to help identify ways to imrpove the query, or understand the performance of the query and the effects on the database in general.

This utility can(optionally) use OpenAI's gpt-4 model to construct specific remediation recommandations. 

## Installing / Getting Started

`qp` is a statically compiled single binary that you can download and execute on most systems. 
Head over to the [releases](https://github.com/queryplan-ai/qp/releases) page to get the latest.

### Environment Variables

No env vars are required to use `qp`, but setting some will make runs easier in the future.

| Optional Environment Variable Name | Description |
|------------------------------------|-------------|
| `QP_DB_URI` | The connection string (URI) to automatically connect to |
| `OPENAI_KEY` | Your OpenAI API Key to automatically use |


## Connecting

Use a connection uri:

Mysql:
mysql://username:password@host:port/database

Postgres: 
postgres://username@host:port/database

For a local containerized version of postgres, it's common to use `?sslmode=disable` at the end of the connection string.

## FAQ

What about transactions?
Unfortunately, transactions aren't supported today. Track this issue and if you have suggestions or ideas on how to plan a transaction using QueryPlan, comment.
