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
    // poglej če se splača dodati naslednjo linijo
    // kot argument temu: log.Fatal(http.ListenAndServe....)
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

        ///// TODO: EMAIL VERIFICATION!!! /////
        //update: i wont be doing this

        /*  1. Generate verification token for user
            2. Send email with verification token as part of the url and
               store the time the token was generated. Decide how long
               the token should be active.
            3. When the user clicks the url, parse the url and verify that
               the token is valid
            4. Update user status to valid
            5. Send confirmation email
        */

        sqlStatement := "INSERT INTO subscribers (email) VALUES ($1)"
        _, query_err := db.Query(sqlStatement, email.Email)

        if query_err != nil {
            log.Println("could not insert into db", query_err)
            w.WriteHeader(http.StatusInternalServerError)
            return
        }

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

///// CREATE GET REQUEST TO GET EMAILS FOR MAILING LIST /////

//func handleGetQuery() *string {
//
//    rows, err := db.Query("SELECT * FROM subscribers")
//
//    tableString := "\nid  Entry_time  email\n"
//
//    for rows.Next() {
//        var id int
//        var entry_time time.Time
//        var email string
//
//
//        err := rows.Scan(&id, &entry_time, &email)
//        if err != nil {
//            log.Println(err)
//        }
//
//        tableString += fmt.Sprintf("id: %d\t Entry_time: %v\t email: %s\n", id, entry_time.Format("01-02-2006 03:04"), email)
//    }
//
//
//    log.Print(tableString)
//
//    return &tableString
//}

func validateEmailFormat(email string) bool {
    regex := regexp.MustCompile(`^([a-zA-Z0-9_\.-]+)@([a-zA-Z0-9_\.-]+)(\.[a-zA-Z0-9_\.-]+)*$`)
    return regex.MatchString(email)
}

