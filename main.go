package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	firebase "firebase.google.com/go"
	"github.com/rs/cors"
	"google.golang.org/api/option"
	db "upper.io/db.v3"
	"upper.io/db.v3/mysql"
)

var settings = mysql.ConnectionURL{
	Database: `cuco_api`,
	Host:     `dev.db.culturacolectiva.com`,
	User:     `root`,
	Password: `root`,
}

type InvitedUser struct {
	ID       uint   `db:"id"`
	Email    string `db:"email"`
	Realm    string `db:"realm"`
	Status   string `db:"status"`
	Username string `db:"username"`
}

func showDate(w http.ResponseWriter, r *http.Request) {
	sess, err := mysql.Open(settings)
	if err != nil {
		log.Fatal(err)
	}

	opt := option.WithCredentialsFile("./key.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	idToken := r.Header.Get("Authorization")

	token, err := client.VerifyIDToken(idToken)
	if err != nil {
		log.Fatalf("error verifying ID token: %v\n", err)
	}

	email := token.Claims["email"]

	defer sess.Close()

	sess.SetLogging(true)

	invitedUsers := sess.Collection("InvitedUsers")

	var invitedUser InvitedUser

	err = invitedUsers.Find(db.Cond{"email": email}).One(&invitedUser)

	message := "{\"realm\": " + invitedUser.Realm + "}"

	if err != nil {
		log.Printf("%v\n", err)
		message = "{\"error\": \"User not found\"}"
	}

	fmt.Printf("%v\n", invitedUser.Realm)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(message))
}

func main() {
	c := cors.New(cors.Options{
		AllowedHeaders: []string{"Authorization"},
		AllowedOrigins: []string{
			"https://cms.culturacolectiva.com",
			"http://dev.cms.culturacolectiva.com",
			"https://staging.cms.culturacolectiva.com",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"OPTIONS",
		},
		Debug: true,
	})

	handler := http.HandlerFunc(showDate)

	err := http.ListenAndServe(":80", c.Handler(handler))
	if err != nil {
		log.Fatal(err)
	}
}
