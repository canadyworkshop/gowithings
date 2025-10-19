package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/canadyworkshop/gowithings"
)

func main() {

	client := gowithings.NewClient(gowithings.Config{
		ClientID:     os.Getenv("GOWITHINGS_TEST_CLIENT_ID"),
		ClientSecret: os.Getenv("GOWITHINGS_TEST_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOWITHINGS_TEST_REDIRECT_URL"),
	})

	authURL, _, err := client.AuthCodeURL()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(authURL)
	fmt.Println("")
	fmt.Printf("Code: ")

	reader := bufio.NewReader(os.Stdin)
	code, _ := reader.ReadString('\n')
	code = strings.TrimSuffix(code, "\n")
	code = strings.TrimSuffix(code, "\r")

	token, err := client.RequestToken(context.Background(), code)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Refresh Token: %s\n, Expire: %d\n", token.Body.RefreshToken, token.Body.RefreshTokenCreationDate)
}
