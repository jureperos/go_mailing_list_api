package main

import (
    "os"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	_ "github.com/lib/pq"
    "regexp"
)

    ///// GENERAL TODOS /////
/*
    - Make the frontend work
    - Make tests
    - Deploy to fly.io
*/

type Email struct {
    Email string `json:"Email"`
}

func main() {

    dbConnStr := os.Getenv("LOCAL_DB_CONNECTION_STR")
    localDevPort:= os.Getenv("LOCAL_DEV_PORT")

    var db *sql.DB
    db, err := sql.Open("postgres", dbConnStr)

    if err != nil {
        log.Fatal(err)
    }

    defer db.Close()

    //for a potential dashboard if needed
    //fs := http.FileServer(http.Dir("./public/"))
    //http.Handle("/static/", http.StripPrefix("/static/", fs))


    http.HandleFunc("/email", dbWrap(db))
    // TODO: Make new endpoint to handle unsubscribe

    // Listen from all adresses(0.0.0.0) when deploying to fly.io
    http.ListenAndServe("0.0.0.0:" + localDevPort, nil)

}

func dbWrap(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {

        // TODO: Enforce request limits
        log.Println("Post function called")

        if r.Method != "POST" {
            http.Error(w, "invalid request method", http.StatusMethodNotAllowed)
            return
        }

        contentType := r.Header.Get("Content-Type")
        if contentType != "application/json" {
            http.Error(w, "Content must be application/json", http.StatusUnsupportedMediaType)
            return
        }

        body, read_err := io.ReadAll(r.Body)
        if read_err != nil {
            http.Error(w, "Error reading request body", http.StatusInternalServerError)
            return
        }

        var email Email

        unmarshal_err := json.Unmarshal(body, &email)
        if unmarshal_err != nil {
            log.Println("Error unmarshalling request body")
            w.WriteHeader(http.StatusInternalServerError)
        }

        if validateEmailFormat(email.Email) == false {
            log.Printf("Invalid email format for email: %s", email.Email)
            w.WriteHeader(http.StatusBadRequest)
        }

        sqlStatement := "INSERT INTO subscribers (email) VALUES ($1)"
        _, query_err := db.Query(sqlStatement, email.Email)

        if query_err != nil {
            log.Println("could not insert into db", query_err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

        //template for confirmation email
        emailConfirmErr := confirmationEmail(email.Email)
        if emailConfirmErr != nil {
            log.Println("Error sending confirmation email", emailConfirmErr)
            // How do i actualy handle this because the person was arleady
            // added in the database???
        }

        w.WriteHeader(http.StatusOK)
    }
}

func confirmationEmail(email string) error {
    log.Println("confirmation email")

    var sendEmailErr error

    // send email that returns email_err if unsuccesful
    var emailErr error // replace this with actual email unsuccesfull error
    if emailErr != nil {
        sendEmailErr = emailErr
    }

    return sendEmailErr
}

func validateEmailFormat(email string) bool {
    regex := regexp.MustCompile(`^([a-zA-Z0-9_\.-]+)@([a-zA-Z0-9_\.-]+)(\.[a-zA-Z0-9_\.-]+)*$`)
    return regex.MatchString(email)
}

