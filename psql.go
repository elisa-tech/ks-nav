package main
 
import (
        "database/sql"
        "fmt"
        _ "github.com/lib/pq"
)
type Connect_token struct{
        Host    string
        Port    int
        User    string
        Pass    string
        Dbname  string
}

func Connect_db(t *Connect_token) (*sql.DB){
        fmt.Println("connect")
        psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", (*t).Host, (*t).Port, (*t).User, (*t).Pass, (*t).Dbname)
        db, err := sql.Open("postgres", psqlconn)
        if err!= nil {
                panic(err)
                }
        fmt.Println("connected")
        return db
}

func Insert_data(db *sql.DB, query string, test bool){

        fmt.Println(query)
        if !test {
                _, err := db.Exec(query)
                if err!= nil {
                        panic(err)
                        }
                } else {
                        fmt.Println(query)
                        }
}
