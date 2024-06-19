package main

import(
  "log"
  "golang.org/x/crypto/bcrypt"
  )

func hashPass(password []byte){
  

  hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
  if err != nil {
    panic(err)
  }

  log.Println(string(hash))

  err = bcrypt.CompareHashAndPassword(hash, password)

  if err != nil{
    panic(err)
  } // nil is a match

  return hash
}


