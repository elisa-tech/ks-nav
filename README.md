# NAV

## Sample configuration:
```
{
"DBURL":"dbs.hqhome163.com",
"DBPort":5432,
"DBUser":"alessandro",
"DBPassword":"<password>",
"DBTargetDB":"kernel_bin",
"Symbol":"__arm64_sys_getppid",
"Instance":1,
"Mode":1,
"Excluded": ["rcu_.*", "kmalloc", "kfree"],
"MaxDepth":0,
"Jout": "JsonOutputPlain"
}
```
Configuration is a file containing a JSON serialized conf object

|Field     |description                                                               |type    |sample value       |
|----------|--------------------------------------------------------------------------|--------|-------------------|
|DBURL     |Host name ot ip address of the psql instance                              |string  |dbs.hqhome163.com  |
|DBPort    |tcp port where psql instance is listening                                 |integer |5432               |
|DBUser    |Valid username on the psql instance                                       |string  |alessandro         |
|DBPassword|Valid password on the psql instance                                       |string  |<password>         |
|DBTargetDB|The identifier for the DB containing symbols                              |string  |kernel_bin         |
|Symbol    |The symbol where start the navigation                                     |string  |__arm64_sys_getppid|
|Instance  |The interesting symbols instance identifier                               |integer |1                  |
|Mode      |Mode of plotting: 1 symbols, 2 subsystems                                 |integer |1                  |
|Excluded  |List of symbols/subsystem not to be expanded                              |string[]|["rcu_.*", "kfree"]|
|MaxDepth  |Max number of levels to explore 0 no limit                                |integer |0                  |
|Jout      |Type of output: GraphOnly, JsonOutputPlain, JsonOutputB64, JsonOutputGZB64|enum    | GraphOnly         |
