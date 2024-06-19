package main

import(
  "os"
  "log"
  "regexp"
  "html/template"
  "net/http"
  "path/filepath"
  //"database/sql"
  "golang.org/x/crypto/bcrypt"
  //"io/ioutil"
  //"strings"
  )
// DATA STRUCTURE
type Page struct {
  Title string
  Body []byte
}

type User struct {
  UserName string
  Password []byte
}

// INITIALIZING

// Compile known expressions and point to templates
var validPath = regexp.MustCompile("^/(edit|save|view|new|user)/([a-zA-Z0-9]+)$")
// Initialize and parse multiple templates
var templates *template.Template
func init() {
  templateDir := "../templ"

  // Debugging: check if template directory and files exist
  if _, err := os.Stat(templateDir); os.IsNotExist(err) {
    log.Fatalf("Template directory does not exist: %s", templateDir)
  }
  // Load all templates
  templatesToLoad := []string{
    filepath.Join(templateDir, "edit.html"),
    filepath.Join(templateDir, "view.html"),
    filepath.Join(templateDir, "new.html"),
    filepath.Join(templateDir, "user.html"),
  }
  // Handle errors for each template
  for _, tmpl := range templatesToLoad {
    if _, err := os.Stat(tmpl); os.IsNotExist(err) {
      log.Fatalf("Template file does not exist: %s", tmpl)
    }
  }

  var err error
  templates, err = template.ParseFiles(templatesToLoad...)
  if err != nil {
    log.Fatalf("Error parsing templates: %v", err)
  }
}
// BASE FUNCTIONALITY

// Save page data to database
func (p *Page) pSave() error {
  db, err := connectDB()
  if err != nil{
    return err
  }
  defer db.Close()

  _, err = db.Exec("INSERT INTO lime (Title, Body) VALUES ($1, $2) ON CONFLICT (Title) DO UPDATE SET Body = EXCLUDED.Body", p.Title, p.Body)
  return err
}

// Save user data to database
func saveUser(w http.ResponseWriter, r *http.Request) {
  db, err := connectDB()
  if err != nil {
    log.Println("no connection to database")
    return
  }

  r.ParseForm() 
  name := r.FormValue("UserName")
  pass := r.FormValue("Password")
  log.Println(name, pass)
  /*selection := "SELECT UserName FROM resident WHERE UserName = ?"
  row := db.QueryRow(selection, name)
  log.Println(row)
  var uID string
  scanErr := row.Scan(&uID)
  if scanErr != sql.ErrNoRows{
    log.Println("username taken", scanErr)
    return
  }*/
  
  var hash []byte
  hash, err = bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
  if err != nil{
    log.Println("could not register account")
  }

  query := "INSERT INTO resident (UserName, Password) VALUES ($1, $2)"
  _, err = db.Exec(query, name, hash)
  if err != nil {
    log.Println(err)
  }
  defer db.Close()
  
  http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

// Load page from the database
func loadPage(title string) (*Page, error) {
  db, err := connectDB()
  if err != nil {
    return nil, err
  }
  defer db.Close()

  var page Page
  err = db.QueryRow("SELECT Title, Body FROM lime WHERE Title = $1", title).Scan(&page.Title, &page.Body)
  if err != nil {
    return nil, err
  }
  return &page, nil
}

// Render the template that init returns
func renderTemplate(w http.ResponseWriter, tmpl string, p *Page){
  err := templates.ExecuteTemplate(w, tmpl + ".html", p)
  if err != nil{
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
}


// TEMPLATE LOGIC

// Fishes correct handler based on title string
func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc{
  return func(w http.ResponseWriter, r*http.Request){
    // Extract page title
    m := validPath.FindStringSubmatch(r.URL.Path)
    if m == nil{
      http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
      return
    }
    // Call correct handler
    fn(w, r, m[2])
  }
}
// NEW PAGE
func newHandler(w http.ResponseWriter, r *http.Request, title string){
  page, err := loadPage(title)
  if err != nil{
    page = &Page{Title: title}
  }
  renderTemplate(w, "new", page)
}
// VIEW PAGE
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
  p, err := loadPage(title)
  if err != nil{
    http.Redirect(w, r, "/new/"+title, http.StatusFound)
    return
  }
  renderTemplate(w, "view", p)
}
// EDIT PAGE
func editHandler(w http.ResponseWriter, r *http.Request, title string){
  p, err := loadPage(title)
  if err != nil{
    p = &Page{Title: title}
  }
  renderTemplate(w, "edit", p)
}
// Is not actually a page but saves to file and renders to /view/
func pSaveHandler(w http.ResponseWriter, r *http.Request, title string){
  body := r.FormValue("body")
  p := &Page{Title: title, Body: []byte(body)}
  err := p.pSave()
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w, r, "/view/"+title, http.StatusFound)
}
// USER PAGE
func userHandler(w http.ResponseWriter, r *http.Request, title string){
  /*if r.Method != http.MethodPost {
    http.Error(w, "Username and password cannot be blank", http.StatusBadRequest)
    return
  }*/
  p, _ := loadPage(title)
  
  renderTemplate(w, "user", p)
}

// ROOT REDIRECT
func rootHandler(w http.ResponseWriter, r *http.Request){
  http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func match(){
  
}


func main(){
// Check if directory exists
  /*cwd, err := os.Getwd()
  if err != nil {
    log.Fatal(err)
  }*/
//  log.Printf("Current working directory: %s", cwd)
  connectDB()

  // Serve static files from the styles directory
  http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("../styles"))))

  http.HandleFunc("/", rootHandler)
  http.HandleFunc("/view/", makeHandler(viewHandler))
  http.HandleFunc("/edit/", makeHandler(editHandler))
  http.HandleFunc("/save/", makeHandler(pSaveHandler))
  http.HandleFunc("/user/register", saveUser)
  //http.HandleFunc("/user/login", match)
  http.HandleFunc("/user/sign", makeHandler(userHandler))
  http.HandleFunc("/new/", makeHandler(newHandler))
  
  table()
  versionDB()

  
  log.Fatal(http.ListenAndServe(":8080", nil))
}
