package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"

	"net/http"

	_ "github.com/denisenkom/go-mssqldb"

	"strconv"

	_ "github.com/alexbrainman/odbc"
	"github.com/gorilla/mux"
)

const (
	// DBHost is the name of the server
	DBHost = "SAMHNB8CGZN22"
	// DBDbase is the name of the database
	DBDbase = "GoBlog"
	// PORT is the port used for the web service
	PORT = ":8080"
)

var database *sql.DB

// Page Defines data structure for the page.
type Page struct {
	Title      string
	RawContent string
	Content    template.HTML
	Date       string
	Comments   []Comment
	// Session    Session
	GUID string
}

// Comment struct
type Comment struct {
	ID          int
	Name        string
	Email       string
	CommentText string
}

// JSONResponse is a comment
type JSONResponse struct {
	Fields map[string]string
}

// ServePage comment here
func ServePage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageGUID := vars["guid"]
	thisPage := Page{}
	fmt.Println(pageGUID)
	err := database.QueryRow("SELECT page_title, page_content, page_date, page_guid FROM pages WHERE page_guid = ?", pageGUID).Scan(&thisPage.Title, &thisPage.RawContent, &thisPage.Date, &thisPage.GUID)
	thisPage.Content = template.HTML(thisPage.RawContent)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println("Couldn't get page!")
		return

	}
	t, _ := template.ParseFiles("templates/blog.html")
	t.Execute(w, thisPage)

}

// RedirIndex redirects to the home
func RedirIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/home", 301)
}

// ServeIndex This is it
func ServeIndex(w http.ResponseWriter, r *http.Request) {
	var Pages = []Page{}
	pages, err := database.Query("SELECT page_title, page_content, page_date, page_guid FROM pages ORDER BY page_date DESC")
	if err != nil {
		fmt.Fprintln(w, err.Error())
	}
	defer pages.Close()
	for pages.Next() {
		thisPage := Page{}
		pages.Scan(&thisPage.Title, &thisPage.RawContent, &thisPage.Date, &thisPage.GUID)
		thisPage.Content = template.HTML(thisPage.RawContent)
		Pages = append(Pages, thisPage)
	}
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(w, Pages)
}

// TruncatedText returns a smaller subset
func (p Page) TruncatedText() string {
	chars := 0
	for i := range p.RawContent {
		chars++
		if chars > 150 {
			return p.RawContent[:i] + ` ...`
		}
	}
	return p.RawContent
}

// APIPage has now been commented
func APIPage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pageGUID := vars["guid"]
	thisPage := Page{}
	fmt.Println(pageGUID)
	err := database.QueryRow("SELECT page_title, page_content, page_date FROM pages WHERE page_guid=?", pageGUID).Scan(&thisPage.Title, &thisPage.RawContent, &thisPage.Date)
	thisPage.Content = template.HTML(thisPage.RawContent)
	if err != nil {
		http.Error(w, http.StatusText(404), http.StatusNotFound)
		log.Println(err)
		return
	}

	APIOutput, err := json.Marshal(thisPage)
	fmt.Println(APIOutput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, thisPage)
}

//APICommentPost is a comment
func APICommentPost(w http.ResponseWriter, r *http.Request) {
	log.Println("APICommentPost")

	var commentAdded bool
	err := r.ParseForm()
	if err != nil {
		log.Println(err.Error())
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	comments := r.FormValue("comments")

	res, err := database.Exec("INSERT INTO comments (comment_name, comment_email, comment_text) VALUES (?, ?, ?)", name, email, comments)

	if err != nil {
		log.Println(err.Error())
	}

	id, err := res.LastInsertId()
	if err != nil {
		commentAdded = false
	} else {
		commentAdded = true
	}
	commentAddedBool := strconv.FormatBool(commentAdded)
	var resp JSONResponse
	resp.Fields["id"] = string(id)
	resp.Fields["added"] = commentAddedBool
	jsonResp, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, jsonResp)
}

func main() {
	// dbConn := fmt.Sprintf("server=%s;Database=%s;Integrated Security=true;", DBHost, DBDbase)

	dbConn := fmt.Sprintf("driver=sql server;server=%s;Database=%s;Integrated Security=true;", DBHost, DBDbase)

	fmt.Println(dbConn)
	db, err := sql.Open("odbc", dbConn)
	if err != nil {
		log.Println("Couldn't connect to" + DBDbase)
		log.Println(err.Error())
	}

	database = db

	routes := mux.NewRouter()
	routes.HandleFunc("/api/pages", APIPage).
		Methods("GET").
		Schemes("https")
	routes.HandleFunc("/api/pages/{guid:[0-9a-zA\\-]+}", APIPage).
		Methods("GET").
		Schemes("https")
	routes.HandleFunc("/api/comments", APICommentPost).
		Methods("POST")
	routes.HandleFunc("/page/{guid:[0-9a-zA\\-]+}", ServePage)
	routes.HandleFunc("/", RedirIndex)
	routes.HandleFunc("/home", ServeIndex)
	http.Handle("/", routes)

	http.ListenAndServe(PORT, nil)

	/*
		certificates, err := tls.LoadX509KeyPair("cert.pem", "key.pem")

		tlsConf := tls.Config{Certificates: []tls.Certificate{certificates}}
		tls.Listen("tcp", PORT, &tlsConf)
	*/

}
