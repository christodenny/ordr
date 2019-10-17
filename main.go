package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Dish struct {
	Name        string  `json:"name"`
	ChineseName string  `json:"chinese_name"`
	Price       float64 `json:"price"`
	ImageURL    string  `json:"image_url"`
	Count       int
}

type Order struct {
	Name   string
	Dishes []Dish
	Total  float64
}

var (
	redirect = http.StatusFound

	lock = sync.RWMutex{}

	totalOrder = Order{
		Name:   "Total",
		Dishes: nil,
		Total:  0.0,
	}

	orders = map[string]*Order{}
)

func InitMenu() {
	// Open our jsonFile
	menuJSONFile, err := os.Open("menu.json")

	// If we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}

	// Read in the menu
	var dishes []Dish
	menuJSONByteValue, _ := ioutil.ReadAll(menuJSONFile)
	json.Unmarshal(menuJSONByteValue, &dishes)
	totalOrder.Dishes = dishes

	// Defer the closing of our jsonFile so that we can parse it later on
	defer menuJSONFile.Close()
}

func GetParam(vals url.Values, s string) string {
	fields, ok := vals[s]
	var field string = ""
	if ok && len(fields) > 0 {
		field = fields[0]
	}
	return field
}

func printDish(d *Dish) {
	fmt.Printf("%s: %d\n", d.Name, d.Count)
}

func printOrder(o *Order) {
	fmt.Printf("Name: %s\nTotal: %.2f\n", o.Name, o.Total)
	for _, d := range o.Dishes {
		if d.Count > 0 {
			printDish(&d)
		}
	}
}

func printState() {
	fmt.Println("Total:")
	printOrder(&totalOrder)
	fmt.Println("Individual orders:")
	for orderer, order := range orders {
		fmt.Printf("User %s\n", orderer)
		printOrder(order)
	}
	fmt.Println()
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	lock.RLock()
	defer lock.RUnlock()
	c, err := r.Cookie("Name")
	if err != nil {
		t, _ := template.ParseFiles("tmpl/create.html")
		t.Execute(w, nil)
	} else {
		_, ok := orders[c.Value]
		if !ok {
			t, _ := template.ParseFiles("tmpl/create.html")
			t.Execute(w, nil)
		} else {
			t, _ := template.ParseFiles("tmpl/order.html")
			t.Execute(w, *orders[c.Value])
		}
	}
}

func TotalHandler(w http.ResponseWriter, r *http.Request) {
	lock.RLock()
	defer lock.RUnlock()
	printState()
	copyOfTotalOrder := totalOrder
	filtered := []Dish{}
	for _, dish := range totalOrder.Dishes {
		if dish.Count > 0 {
			filtered = append(filtered, dish)
		}
	}
	copyOfTotalOrder.Dishes = filtered
	t, _ := template.ParseFiles("tmpl/orders.html")
	t.Execute(w, copyOfTotalOrder)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	c, err := r.Cookie("Name")
	vals := r.URL.Query()
	if err != nil {
		t, _ := template.ParseFiles("tmpl/create.html")
		t.Execute(w, nil)
	} else {
		name := c.Value
		total := 0.0
		order := orders[name]
		for i, dish := range order.Dishes {
			count, err := strconv.Atoi(GetParam(vals, dish.Name))
			if err == nil && count >= 0 {
				totalOrder.Dishes[i].Count += count - order.Dishes[i].Count
				order.Dishes[i].Count = count
			}
			total += float64(order.Dishes[i].Count) * order.Dishes[i].Price
		}
		total, err := strconv.ParseFloat(fmt.Sprintf("%.2f", total), 64)
		if err == nil {
			totalOrder.Total += total - order.Total
			order.Total = total
		}
		printState()
		http.Redirect(w, r, "/", redirect)
	}
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	vals := r.URL.Query()
	name := GetParam(vals, "name")
	c := http.Cookie{
		Name:  "Name",
		Value: name,
		Path:  "/",
	}
	http.SetCookie(w, &c)
	_, ok := orders[name]
	if !ok {
		blankOrder := totalOrder
		blankOrder.Name = name
		blankOrder.Total = 0.0
		blankOrder.Dishes = append([]Dish(nil), blankOrder.Dishes...)
		for i, _ := range blankOrder.Dishes {
			blankOrder.Dishes[i].Count = 0
		}
		orders[name] = &blankOrder
	}
	printState()
	http.Redirect(w, r, "/", redirect)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	c := http.Cookie{
		Name:   "Name",
		MaxAge: -1,
	}
	printState()
	http.SetCookie(w, &c)
	http.Redirect(w, r, "/", redirect)
}

func ResetHandler(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()
	orders = map[string]*Order{}
	for i, _ := range totalOrder.Dishes {
		totalOrder.Dishes[i].Count = 0
	}
	totalOrder.Total = 0.0
	printState()
	http.Redirect(w, r, "/", redirect)
}

func main() {
	// Read the menu from JSON file
	InitMenu()

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/orders", TotalHandler)
	r.HandleFunc("/update", UpdateHandler)
	r.HandleFunc("/create_user", CreateUserHandler)
	r.HandleFunc("/logout", LogoutHandler)
	r.HandleFunc("/reset", ResetHandler)

	// Note that the path given to the http.Dir function is relative to the project
	// directory root.
	fileServer := http.FileServer(http.Dir("static/"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
