package main

import(
  "database/sql"
  "fmt"
  "net/url"
  "log"
  _ "github.com/lib/pq"
)
// DATABASE URI
var serviceURI = "postgres://avnadmin:AVNS_U5YeK0f2K_xawVcnA2i@gowiki-111-gowiki-111.g.aivencloud.com:14428/defaultdb?sslmode=require"

// DATABASE SERVICE

// make a connection to the database
func connectDB() (*sql.DB, error) {
  conn, err := url.Parse(serviceURI)
  if err != nil {
    return nil, err
  }
  //conn.RawQuery = "sslmode=verify-full"

  db, err := sql.Open("postgres", conn.String())
  if err != nil {
    return nil, err
  }
  return db, nil
}

// Creates table for storing web data

func table(){
  //conn.RawQuery = "sslmode=verify-full"

  db, err := connectDB()
  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()
  log.Println("connected")

  _, err = db.Exec("CREATE TABLE IF NOT EXISTS lime(Title varchar(30) NOT NULL UNIQUE, Body BYTEA);")
  if err != nil{
    log.Fatal(err)
  }
    _, err = db.Exec("CREATE TABLE IF NOT EXISTS resident(UserName varchar(50) PRIMARY KEY NOT NULL UNIQUE, Password BYTEA NOT NULL);")
  if err != nil{
    log.Fatal(err)
  }
  log.Println("table created")

  rows, err := db.Query("SELECT Title FROM lime WHERE Title = 'FrontPage'")
  if err != nil{
    log.Fatal(err)
  }
  defer rows.Close()
  log.Println(rows)

  for rows.Next(){
    var title string
    if err := rows.Scan(&title); err != nil{
      log.Fatal("Failed to retrieve title", err)
    }
    fmt.Println(title, "lime")
  }
  if err := rows.Err(); err != nil {
    log.Fatal("Could not iterate", err)
  }
  defer db.Close()
}

func databaseUpdate(updateFunc func(db *sql.DB) error){
  conn, err := url.Parse(serviceURI)
  if err != nil {
    log.Fatalf("Failed to parse URL: %v", err)
  }
  conn.RawQuery = "sslmode=verify-full&sslrootcert=ca.pem"

  db, err := sql.Open("postgres", conn.String())
  if err != nil {
    log.Fatalf("Failed to open database connection: %v", err)
  }
  defer db.Close()

// Call the update function
  err = updateFunc(db)
  if err != nil {
    log.Fatalf("Failed to perform database update: %v", err)
  }
  log.Println("Database update performed successfully")
}


func changedBody(db *sql.DB, title string, uBody []byte) error {
	query := "UPDATE lime SET Body::BYTEA = $1 WHERE Title::text = $2"
  _, err := db.Exec(query, uBody, title)
  if err != nil {
    return fmt.Errorf("failed to update body content: %w", err)
  }
  log.Println("Body content updated successfully")
  return nil
}


func versionDB() {
  
  conn, _ := url.Parse(serviceURI)
  conn.RawQuery = "sslmode=verify-ca;sslrootcert=ca.pem"

  db, err := sql.Open("postgres", conn.String())

  if err != nil {
    log.Fatal(err)
  }
  defer db.Close()

  rows, err := db.Query("SELECT version()")
  if err != nil {
    panic(err)
  }

  for rows.Next() {
    var result string
    err = rows.Scan(&result)
    if err != nil {
      panic(err)
    }
  fmt.Printf("Version: %s\n", result)
  }
}
