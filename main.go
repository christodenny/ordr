package main

import (
  "fmt"
  "html/template"
  "log"
  "net/http"
  "net/url"
  "strconv"

  "github.com/gorilla/mux"
)

type Dish struct {
  Name string
  Price float64
  Count int
}

type Order struct {
  Name string
  Dishes []Dish
  Total float64
}

var (
  redirect = http.StatusFound

  totalOrder = Order{
    Name: "Total",
    Dishes: []Dish{
      Dish{
        Name: "Lamb Skewer",
        Price: 2.99,
      },
      Dish{
        Name: "Beef Skewer",
        Price: 3.99,
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
  c, err := r.Cookie("Name")
  _, ok := orders[c.Value]
  if err != nil || !ok {
	  t, _ := template.ParseFiles("tmpl/create.html")
	  t.Execute(w, nil)
  } else {
    t, _ := template.ParseFiles("tmpl/order.html")
    t.Execute(w, *orders[c.Value])
  }
}

func TotalHandler(w http.ResponseWriter, r *http.Request) {
  t, _ := template.ParseFiles("tmpl/orders.html")
  t.Execute(w, totalOrder)
}

func UpdateHandler(w http.ResponseWriter, r *http.Request) {
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
      if err == nil {
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
  vals := r.URL.Query()
  name := GetParam(vals, "name")
  c := http.Cookie{
    Name:    "Name",
    Value:   name,
    Path: "/",
  }
  http.SetCookie(w, &c)
  blankOrder := totalOrder
  blankOrder.Name = name
  blankOrder.Total = 0.0
  for i, _ := range blankOrder.Dishes {
    blankOrder.Dishes[i].Count = 0
  }
  orders[name] = &blankOrder
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

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/", HomeHandler)
  r.HandleFunc("/orders", TotalHandler)
  r.HandleFunc("/update", UpdateHandler)
  r.HandleFunc("/create_user", CreateUserHandler)
  r.HandleFunc("/logout", LogoutHandler)

  // Note that the path given to the http.Dir function is relative to the project
  // directory root.
  fileServer := http.FileServer(http.Dir("static/"))
  r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))

  log.Println("Starting server on :8080")
  log.Fatal(http.ListenAndServe(":8080", r))
}
