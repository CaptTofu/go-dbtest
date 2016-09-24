package main

import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import (
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "runtime"
)

type Message struct {
    Message string `json:"message"`
}

var (
    debug   = flag.Bool("debug", false, "debug logging")
    message = flag.String("message", "Hello, World!", "message to return")
    port    = flag.Int("port", 9080, "port to serve on")
    mysql_port = flag.Int("mysql-port", 3306, "mysql db port")
    mysql_user = flag.String("mysql-user", "test", "mysql user")
    mysql_password = flag.String("mysql-password", "test", "mysql password")
    mysql_host = flag.String("mysql-host", "localhost", "mysql host")
    mysql_db = flag.String("mysql-db", "test", "mysql schema")
)

func dbprocess() {
    dsn := fmt.Sprintf("%s:%s@/%s", *mysql_user, *mysql_password, *mysql_db)
    //db, err := sql.Open("mysql", "user:password@/dbname")
    db, err := sql.Open("mysql", dsn)

    // insert
    stmt, err := db.Prepare("INSERT users SET first=?,last=?,age=?,created=now()")
    checkErr(err)

    res, err := stmt.Exec("first", "last", 48)
    checkErr(err)

    id, err := res.LastInsertId()
    checkErr(err)

    fmt.Println(id)

    // query
    rows, err := db.Query("SELECT * FROM users")
    checkErr(err)

    for rows.Next() {
        var id int
        var age int
        var first string
        var last string
        var created string
        err = rows.Scan(&id, &first, &last, &age, &created)
        checkErr(err)
        fmt.Println(id)
        fmt.Println(first)
        fmt.Println(last)
        fmt.Println(age)
        fmt.Println(created)
    }

    db.Close()
}

func main() {
    flag.Parse()
    dbprocess()

    runtime.GOMAXPROCS(runtime.NumCPU())

    http.HandleFunc("/json", jsonHandler)
    http.ListenAndServe(fmt.Sprintf(":%d", *port), Log(http.DefaultServeMux))
}

func Log(handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if *debug == true {
            log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
        }
        handler.ServeHTTP(w, r)
    })
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    msg := r.FormValue("test")
    json.NewEncoder(w).Encode(&Message{msg})
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
