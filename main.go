package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
)

type Dish struct {
	Name  string
	Price float64
	Count int
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
		Name: "Total",
		Dishes: []Dish{
			Dish{
				Name:  "Lamb",
				Price: 1.75,
			},
			Dish{
				Name:  "Beef",
				Price: 1.50,
			},
			Dish{
				Name:  "Chicken Wing",
				Price: 2.00,
			},
			Dish{
				Name:  "Chicken",
				Price: 1.50,
			},
			Dish{
				Name:  "Chicken Gizzard",
				Price: 1.00,
			},
			Dish{
				Name:  "Chicken Gristle",
				Price: 1.99,
			},
			Dish{
				Name:  "Chicken Heart",
				Price: 1.00,
			},
			Dish{
				Name:  "Shrimp",
				Price: 1.00,
			},
			Dish{
				Name:  "Squid",
				Price: 1.00,
			},
			Dish{
				Name:  "Yellow Croaker",
				Price: 1.00,
			},
			Dish{
				Name:  "Pollack",
				Price: 5.99,
			},
			Dish{
				Name:  "Pig Feet",
				Price: 4.99,
			},
			Dish{
				Name:  "Sausage",
				Price: 1.00,
			},
			Dish{
				Name:  "Lamb Kidney",
				Price: 2.99,
			},
			Dish{
				Name:  "Quail",
				Price: 4.99,
			},
			Dish{
				Name:  "Mushroom",
				Price: 1.00,
			},
			Dish{
				Name:  "Taiwan Sausage",
				Price: 2.75,
			},
			Dish{
				Name:  "Eggplant",
				Price: 1.49,
			},
			Dish{
				Name:  "A Vegetable",
				Price: 1.49,
			},
			Dish{
				Name:  "Chinese Chives",
				Price: 1.49,
			},
			Dish{
				Name:  "String Beans",
				Price: 1.49,
			},
			Dish{
				Name:  "Beef Tendon",
				Price: 1.75,
			},
			Dish{
				Name:  "Steam Bun",
				Price: 1.00,
			},
		},
		Total: 0.0,
	}

	orders = map[string]*Order{}
)

func GetParam(vals url.Values, s string) string {
	fields, ok := vals[s]
	var field string = ""
	if ok && len(fields) > 0 {
		field = fields[0]
	}
	return field
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
		blankOrder := Order{
			Name:   name,
			Total:  0.0,
			Dishes: make([]Dish, len(totalOrder.Dishes)),
		}
		for i, _ := range blankOrder.Dishes {
			blankOrder.Dishes[i] = Dish{
				Name:  totalOrder.Dishes[i].Name,
				Price: totalOrder.Dishes[i].Price,
			}
		}
		orders[name] = &blankOrder
	}
	http.Redirect(w, r, "/", redirect)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	c := http.Cookie{
		Name:   "Name",
		MaxAge: -1,
	}
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
	http.Redirect(w, r, "/", redirect)
}

func main() {
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
