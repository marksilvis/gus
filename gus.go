// author: Mark Silvis
package main

import (
    "fmt"
    "log"
    "strings"
    "net/http"
    "html/template"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
)

const port = ":4000"
var base62 [62]byte
var base10 map[byte]int = make(map[byte]int)

func embiggen(short string, db *sql.DB) (url string, err error) {
    query := "select url from url_mappings where short=?;"
    err = db.QueryRow(query, short).Scan(&url)
    return
}

// func getId(path string) int {
//     var id float64
//     split := strings.Split(path, "")
//     for i, v := range split {
//         rune := []byte(v)
//         w := base10[rune[0]]
//         id += float64(w)*math.Pow(62, float64(len(split)-i))
//     }
//     return int(id)
// }

func shrinkUrl(id int, c chan int) {
    for id != 0 {
        rem := id % 62
        id = id / 62
        c <- rem
    }
    close(c)
}

func newUrlMapping(url string, db *sql.DB) string {
    var lastId int
    query := "select id from url_mappings order by id desc limit 1;"
    err := db.QueryRow(query).Scan(&lastId)
    if err != nil {
        lastId = 0
    }

    lastId++
    c := make(chan int, 6)
    var shrunk []byte
    shrinkUrl(lastId, c)
    for i := range c {
        shrunk = append(shrunk, base62[i])
    }

    short := string(shrunk[:])
    query = "INSERT INTO url_mappings (url, short) Values(?, ?);"
    _, err = db.Exec(query, url, short)
    if err != nil {
        log.Fatal(err)
    }

    return short
}

func handlePath(w http.ResponseWriter, r *http.Request, db *sql.DB, path string) {
    // this will send a redirect to the full website
    redirect, err := embiggen(path, db)
    if err != nil {
        fmt.Fprintln(w, "404: page not found")
    } else {
        http.Redirect(w, r, "https://"+redirect, http.StatusFound)
    }
}

func handleRoot(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    if r.Method == "GET" {
        http.ServeFile(w, r, "./resources/html/index.html")
    } else if r.Method == "POST" {
        url := r.FormValue("url")
        escaped := template.JSEscapeString(url)
        if escaped != url {
            fmt.Fprintln(w, "Sorry, no XSS attacks")
            return
        }
        escaped = strings.TrimPrefix(escaped, "http://")
        escaped = strings.TrimPrefix(escaped, "https://")
        escaped = strings.TrimPrefix(escaped, "www.")

        var short string
        query := "SELECT short FROM url_mappings WHERE url=?;"
        err := db.QueryRow(query, escaped).Scan(&short)
        if err != nil {
            short = newUrlMapping(escaped, db)
        }
        fmt.Fprintf(w, "<a href=\"localhost%s/%s\">localhost%s/%s</a>", port, short, port, short)
    } else {
        fmt.Fprintf(w, "405: method not allowed")
    }
}

func handler(db *sql.DB) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        path := r.URL.Path[1:]
        if len(path) == 0 {
           handleRoot(w, r, db)
        } else {
            handlePath(w, r, db, path)
        }
    })
}

func main() {
    db, err := sql.Open("sqlite3", "./db/gus.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // build base62 array for mapping urls
    for i := 0; i < 62; i++ {
        if i < 26 {
            base62[i] = byte(i+65)
        } else if i < 52 {
            base62[i] = byte(i+71)
        } else {
            base62[i] = byte(i-52)
        }
    }

    // build base10 array for recovering urls
    for i, v := range base62 {
        base10[v] = i
    }

    http.Handle("/", handler(db))
    // http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./css/"))))
    log.Fatal(http.ListenAndServe(port, nil))
}