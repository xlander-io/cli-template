## cli-template

#### description: 
> this is a template project that you can fork or copy as a project to start with.


#### designed with:
- easy configuration (toml)
- command line support
- rich plugins


#### main entry point file
- main.go is the entry file
- other test entries are under 'test' folders


#### steps to start:

> step1. config your project 
- 1. check folder 'root_conf' where you can put different {config_name}.toml files
- 2. the 'default.toml' config is used if not explicited configured 
- 3. you can run 'go run ./ config set --https.enable=false' to setup your own config file (generated inside 'user_conf/default.toml' )
- 3. you can run 'go run ./ config --conf=test set --https.enable=false' to setup your own config file (generated inside 'user_conf/test.toml' )
- 4. you can also edit user_conf/*.toml files directly without the help of command line

> step2. check database initialization 
- 1. config your database in your *.toml file
- 2. open your database and construct the tables using file 'assets/sql/table.sql'
- 3. run 'go run ./ db init' which will call the function 'Initialize()' inside the 'cmd_db/initialize.go' file which initialize the db data

>step2.5 optional ,geoip download
- 1. 'go run ./ geoip download'

> step3. write api
- 1.go to `cmd_default` folder where your main program locate
- 2.go to 'http/api' folder to add your own api file
- 3.run 'go run ./ gen_api' to auto generate your api files
- 4.run 'go run ./' to start your main program with http server 
- 5.you can view the api pages with 'localhost'


#### Command line hints
```
go run . gen_api                //generate api
go run . config set ...         //set configs
go run . log                    //show all logs
go run . log --only_err=true    //show all error logs [error,panic,fatal]

```