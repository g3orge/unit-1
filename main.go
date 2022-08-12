package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var templates *template.Template

var db *sql.DB

type Resp struct {
	Date  string `json:"date" db:"date"`
	Text  string `json:"explanation" db:"text"`
	HdUrl string `json:"hdurl" db:"hdurl"`
	Title string `json:"title" db:"title"`
	Img   []byte `db:"img"`
}

func NewDB(h, p, u, name, pass string) (*sql.DB, error) {
	connectionString := fmt.Sprintf("user=%s password=%s sslmode=disable", u, pass)
	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func httpClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second}
	return client
}

func init() {
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "unit2")
	os.Setenv("DB_USER", "postgres")
	os.Setenv("DB_PASS", "root")

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	ps := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")

	db, err := NewDB(host, port, user, dbname, ps)
	if err != nil {
		fmt.Print(err)
		return
	}

	dbName := "unit1"
	_, err = db.Exec("drop database if exists " + dbname)
	if err != nil {
		fmt.Print(err)
		return
	}

	_, err = db.Exec("create database " + dbName)
	if err != nil {
		fmt.Print("cannot create db")
		return
	}

	_, err = db.Exec("DROP TABLE images")
	if err != nil {
		fmt.Print(err)
		return
	}

	_, err = db.Exec("CREATE TABLE images (date text, txt text, hdurl text, title text, img bytea)")
	if err != nil {
		fmt.Print(err)
		return
	}

}

func apiRes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		client := httpClient()

		urlPath := "https://api.nasa.gov/planetary/apod?api_key=BJPj2Udy4BNShU1WgwypIBnVLG7yhGU9epgw4TDc"

		req, err := http.NewRequest("GET", urlPath, nil)
		if err != nil {
			fmt.Print(err)
			return
		}

		response, err := client.Do(req)
		if err != nil {
			fmt.Print(err)
			return
		}
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Print(err)
			return
		}

		var res Resp
		json.Unmarshal(body, &res)

		sqlState := `INSERT INTO images(date, txt, hdurl, title) VALUES ($1, $2, $3, $4)`
		db.QueryRow(sqlState, res.Date, res.Text, res.HdUrl, res.Title)

		templates = template.Must(template.ParseFiles("template/form.html"))
		err = templates.ExecuteTemplate(w, "form.html", res.HdUrl)
		if err != nil {
			fmt.Print(err)
			return
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "only get method"}`))
	}
}

func apiItem(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		date := r.URL.Query().Get("date")
		if date == "" {
			fmt.Fprint(w, "Please enter the date in yyyy-mm-dd format")
		}

		row := db.QueryRow("select * from images where date = $1", date)

		res := Resp{}
		err := row.Scan(&res.Date, &res.Text, &res.HdUrl, &res.Title, &res.Img)
		if err != nil {
			fmt.Print(err)
			fmt.Fprint(w, "no documents on this date")
			return
		}
		fmt.Fprint(w, res.Date, res.HdUrl, res.Title, res.Title)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "only get method"}`))
	}
}

func apiItems(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		var res = Resp{}

		rows, err := db.Query("select * from images")
		if err != nil {
			fmt.Print(err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			if err = rows.Scan(&res.Date, &res.Text, &res.HdUrl, &res.Title, &res.Img); err != nil {
				fmt.Print(err)
				return
			}
			fmt.Fprint(w, res.Date, res.HdUrl, res.Title, res.Text)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "only get method"}`))
	}
}

func apiSave(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		resp := Resp{}

		row := db.QueryRow("select hdurl from images")
		row.Scan(&resp.HdUrl)

		res1, err := http.Get(resp.HdUrl)
		if err != nil {
			fmt.Print(err)
			return
		}
		defer res1.Body.Close()

		buf := &bytes.Buffer{}
		buf.ReadFrom(res1.Body)
		data := buf.Bytes()

		db.QueryRow("INSERT INTO images(img) VALUES ($1)", data)
		fmt.Fprint(w, "file saved in db")
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message": "only get method"}`))
	}
	fmt.Fprint(w, "hello there")
}

func main() {
	http.HandleFunc("/", apiRes)
	http.HandleFunc("/item", apiItem)
	http.HandleFunc("/items", apiItems)
	http.HandleFunc("/saveimg", apiSave)

	log.Fatal(http.ListenAndServe(":8080", nil))
	defer db.Close()
}
