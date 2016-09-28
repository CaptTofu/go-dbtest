package main

import "database/sql"
import _ "github.com/go-sql-driver/mysql"
import (
    "encoding/json"
    "flag"
    "fmt"
    "math/rand"
    "time"
    "log"
    "net/http"
    "runtime"
)

type Message struct {
    Message string `json:"message"`
}
type rres struct {
    Id int
    Msg string
    Time string
}

var (
    debug   = flag.Bool("debug", false, "debug logging")
    port    = flag.Int("port", 9080, "port to serve on")
    mysql_port = flag.Int("mysql-port", 3306, "mysql db port")
    mysql_user = flag.String("mysql-user", "test", "mysql user")
    mysql_password = flag.String("mysql-password", "test", "mysql password")
    mysql_host = flag.String("mysql-host", "localhost", "mysql host")
    mysql_db = flag.String("mysql-db", "test", "mysql schema")
)

func dbsetup() {
    dsn := fmt.Sprintf("%s:%s@/%s", *mysql_user, *mysql_password, *mysql_db)
    //db, err := sql.Open("mysql", "user:password@/dbname")
    db, err := sql.Open("mysql", dsn)

    stmt, err := db.Prepare("drop table if exists t1")
    res, err := stmt.Exec()
    checkErr(err)

    stmt, err = db.Prepare("create table t1 (id int auto_increment, message varchar(256) not null default '', created datetime, primary key (id)) engine=innodb")
    res, err = stmt.Exec()
    checkErr(err)

    id, err := res.LastInsertId()
    checkErr(err)

    fmt.Println(id)

}

func random(min, max int) int {
    rand.Seed(time.Now().Unix())
    return rand.Intn(max - min) + min
}

func dbprocess(msg string) rres {
    dsn := fmt.Sprintf("%s:%s@/%s", *mysql_user, *mysql_password, *mysql_db)
    //db, err := sql.Open("mysql", "user:password@/dbname")
    db, err := sql.Open("mysql", dsn)

    // insert
    stmt, err := db.Prepare("INSERT into t1 SET message=?,created=now()")
    checkErr(err)

    _, err = stmt.Exec(msg)
    checkErr(err)

    // query
    var min int
    var max int
    err = db.QueryRow("SELECT min(id), max(id) FROM t1").Scan(&min, &max)
    checkErr(err)

    rnd := random(min, max)
    fmt.Println("random num: %d", rnd)

    var id int
    var created string
    var message string
    err = db.QueryRow("SELECT * FROM t1 where id = ?", rnd).Scan(&id, &message, &created)
    checkErr(err)

    resstruct := rres{id, msg, created}
    //result_string = fmt.Sprintf("|%d|%s|%s|", id, msg, created)

    //fmt.Println(resstruct)
    //fmt.Println(recs)

    db.Close()
    return resstruct
}

func main() {
    flag.Parse()

    runtime.GOMAXPROCS(runtime.NumCPU())

    dbsetup()
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
    msg := r.FormValue("msg")
    resstruct := dbprocess(msg)
    //json.NewEncoder(w).Encode(&Message{resstruct})
    json.NewEncoder(w).Encode(resstruct)
}

func checkErr(err error) {
    if err != nil {
        panic(err)
    }
}
