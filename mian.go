package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"sort"
	"strconv"

	_ "github.com/lib/pq"
)

var Id, hours, minutes, persons int
var db *sql.DB

func main() {
	connStr := "user=postgres password=Alex12313 dbname=Aero sslmode=disable"
	DB, err := sql.Open("postgres", connStr)
	db = DB
	if err != nil {
		panic(err)
	}
	defer db.Close()
	fmt.Println("Database connected")
	handleFunc()
}

//Начальная страница
func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/index.html", "templates/header.html")
	if err != nil {
		fmt.Println(w, err.Error())
	}

	t.ExecuteTemplate(w, "index", nil)
}

//Проверка времени и мест
func choseRest(w http.ResponseWriter, r *http.Request) {
	personsS := r.FormValue("persons")
	hourS := r.FormValue("hour")
	minutesS := r.FormValue("minutes")
	correct := true
	personsL, err := strconv.Atoi(personsS)
	if err != nil || personsL <= 0 {
		correct = false
	}
	hourL, err := strconv.Atoi(hourS)
	if err != nil || hourL < 9 || hourL > 21 {
		correct = false
	}
	minutesL, err := strconv.Atoi(minutesS)
	if err != nil || minutesL < 0 || minutesL > 59 {
		correct = false
	}
	if correct {
		persons = personsL
		hours = hourL
		minutes = minutesL
		http.Redirect(w, r, "/confirm", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/error", http.StatusSeeOther)
	}
}

//Обработка выбора из доступных вариантов
func confirm(w http.ResponseWriter, r *http.Request) {
	restaurants := make(map[int]Restaurant)
	restaurants = getRestaurants(db, hours, minutes, persons)
	copyRests := make(map[int]Restaurant) //Отображение для показа отсортированного списка ресторанов по времени ожидания и чеку
	keys := make([]int, 0, len(restaurants))
	for key := range restaurants {
		keys = append(keys, key)
	}
	//Сортировка по требуемым параметрам
	sort.SliceStable(keys, func(i, j int) bool {
		if restaurants[keys[i]].WaitTime != restaurants[keys[j]].WaitTime {
			return restaurants[keys[i]].WaitTime < restaurants[keys[j]].WaitTime
		} else {
			return restaurants[keys[i]].Bill < restaurants[keys[j]].Bill
		}
	})
	for i, k := range keys {
		copyRests[i] = restaurants[k]
	}
	t, err := template.ParseFiles("templates/confirm.html", "templates/header.html")
	if err != nil {
		fmt.Println(w, err.Error())
	}

	t.ExecuteTemplate(w, "confirm", copyRests)
}

//Обработка страницы указания имени и номера
func result(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/result.html", "templates/header.html")
	if err != nil {
		fmt.Println(w, err.Error())
	}
	id, err := strconv.Atoi(r.FormValue("restaurant"))
	if err != nil {
		http.Redirect(w, r, "/error", http.StatusSeeOther)
	}
	Id = id
	t.ExecuteTemplate(w, "result", nil)
}

//Проверка номера клиента и бронирование мест
func final_check(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	phone := r.FormValue("phone")
	if name == "" {
		http.Redirect(w, r, "/error", http.StatusSeeOther)
	} else {
		matched, _ := regexp.MatchString(`\d+`, phone) //Проверка номера клиента
		if !matched || len(phone) != 10 {
			http.Redirect(w, r, "/error", http.StatusSeeOther)
		} else {
			// Выполняем запросы с бд
			rests := getResTables(buildMapFromID(Id, hours, minutes, db), persons)
			visitorId := addVisitor(db, name, phone)
			bookTables(db, hours, minutes, Id, visitorId, rests)
			http.Redirect(w, r, "/success", http.StatusSeeOther)
		}
	}
}

//Функция для отображения окна об успешном бронировании
func success(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/success.html", "templates/header.html")
	if err != nil {
		fmt.Println(w, err.Error())
	}

	t.ExecuteTemplate(w, "success", nil)
}

//Функция для отображения окна о некорректном вводе
func myError(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/error.html", "templates/header.html")
	if err != nil {
		fmt.Println(w, err.Error())
	}

	t.ExecuteTemplate(w, "error", nil)
}

//Роутинг
func handleFunc() {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.HandleFunc("/", index)
	http.HandleFunc("/chose_restaurant", choseRest)
	http.HandleFunc("/confirm", confirm)
	http.HandleFunc("/result", result)
	http.HandleFunc("/final_check", final_check)
	http.HandleFunc("/success", success)
	http.HandleFunc("/error", myError)
	http.ListenAndServe(":8080", nil)
}
