Helios - Service Oriented Authentication
==========

## Dependencies ##
* code.google.com/p/go-uuid        - UUID generator for Go
* github.com/RangelReale/osin      - OAuth2 server library for Go
* github.com/garyburd/redigo/redis - Redis client for Go 
* github.com/go-sql-driver/mysql   - MySQL Driver for Go
* github.com/coopernurse/gorp      - ORM-ish library for Go


## Running ##
Pass mysql config as param:
```bash
go run main.go "user:pass@tcp(host:port)/dbname"
```
