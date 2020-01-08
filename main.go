package main

import (
	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"log" // checkErr

	"html/template"
	"net/http"
)

var client *redis.Client
var redisServerReachable bool = false
var templates *template.Template



func main() {
	client = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	// check connection with redis server
	pong, err := client.Ping().Result()
	if err != nil {
		log.Println("Redis server unreachable. Proceeding without redis service.")
		log.Println(pong, err)
		redisServerReachable = false
	} else {
		redisServerReachable = true
		log.Println("Redis server online.")
	}

	// init fs to get static files from path
	fs := http.FileServer(http.Dir("./static/"))

	templates = template.Must(template.ParseGlob("templates/*.html"))
	r := mux.NewRouter()
	r.HandleFunc("/", indexHandler).Methods("GET")
	r.HandleFunc("/", indexPostHandler).Methods("POST")
	// tell router to handle all requests with static prefix 
	// fs expects files in the same dir so we strip the prefix
	r.PathPrefix("/static").Handler(http.StripPrefix("/static/", fs))
	// localhost:8080/static/some_file.txt => read file`s content in broswer

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if redisServerReachable {
		// get first 10 from redis.comments list
		comments, err := client.LRange("comments", 0, 10).Result()
		checkErr(err)
		templates.ExecuteTemplate(w, "index.html", comments)
	} else {
		templates.ExecuteTemplate(w, "index.html", nil)
	}
}

func indexPostHandler (w http.ResponseWriter, r *http.Request) {
	if redisServerReachable {
		r.ParseForm()
		comment := r.PostForm.Get("comment")
		client.LPush("comments", comment)
		http.Redirect(w, r, "/", 302)
	} else {
		log.Println("Redis server unreachable. Cannot post comments!")
		http.Redirect(w, r, "/", 302)	
	}
}

// Utils
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
