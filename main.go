package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bluesky-social/indigo/api/atproto"
	"github.com/bluesky-social/indigo/api/bsky"
	util "github.com/bluesky-social/indigo/lex/util"
	xrpc "github.com/bluesky-social/indigo/xrpc"
	"gopkg.in/yaml.v3"
)

func must(err error) {
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
}

// readCredentials reads the bot's username and password in from a file and returns them. The file
// should be in the structure of:
// ```yaml
// username: <username>
// password: <password>
// ```
// Or any equivalent yaml file.
func readCredentials(filename string) (username string, password string, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", "", err
	}

	var credentials struct {
		Username string
		Password string
	}

	err = yaml.Unmarshal([]byte(data), &credentials)
	must(err)

	return credentials.Username, credentials.Password, nil
}

func main() {
	// Read in username and password from file
	username, password, err := readCredentials("credentials.yaml")
	must(err)

	// Get the trainbot message data
	data, err := readTrainbotData("trainbot.yaml")
	must(err)

	// Create a new message and check if we want to post it
	// This just posts one time when you run the program. I'll turn this into an automated cron job thing later
	post := data.newMessage()

	fmt.Printf("bot wants to post: \"%s\"; are u sure u wanna post this (y/N)?", post)

	var response string
	fmt.Scan(&response)

	if !(response == "y" || response == "Y") {
		fmt.Println("fair call chief")
		return
	}

	// the client will be used for authentication
	client := xrpc.Client{
		Client: new(http.Client),
		Host:   "https://bsky.social",
	}

	ctx := context.Background()

	// Attempt to authenticate with the server
	createSessionInput := atproto.ServerCreateSession_Input{
		Identifier: username,
		Password:   password,
	}

	sesh, err := atproto.ServerCreateSession(ctx, &client, &createSessionInput)

	if err != nil {
		fmt.Printf("Couldn't create session: %v\n", err)
		return
	}

	fmt.Printf("Connected as %s\n", sesh.Handle)
	client.Auth = &xrpc.AuthInfo{
		AccessJwt:  sesh.AccessJwt,
		RefreshJwt: sesh.RefreshJwt,
		Handle:     sesh.Handle,
		Did:        sesh.Did,
	}

	// Make our post
	nowStr := time.Now().In(time.UTC).Format(time.RFC3339)

	postInput := atproto.RepoCreateRecord_Input{
		Collection: "app.bsky.feed.post",
		Repo:       username,
		Record: &util.LexiconTypeDecoder{Val: &bsky.FeedPost{
			CreatedAt: nowStr,
			Text:      post,
			Langs:     []string{"en"},
		}},
	}
	postOutput, err := atproto.RepoCreateRecord(ctx, &client, &postInput)

	if err != nil {
		fmt.Printf("Error making post: %v\n", err)
		return
	}

	fmt.Printf("Successfully posted at %s\n", postOutput.Uri)

	fmt.Println("Goodbye, bsky!")
}
