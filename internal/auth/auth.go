package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"spotify-tui/internal/config"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const (
	redirectURI = "http://127.0.0.1:8080/callback"
	state       = "spotify-tui-state"
)

var (
	auth   *spotifyauth.Authenticator
	ch     = make(chan *oauth2.Token)
	scopes = []string{
		spotifyauth.ScopeUserReadPrivate,
		spotifyauth.ScopeUserReadPlaybackState,
		spotifyauth.ScopeUserModifyPlaybackState,
		spotifyauth.ScopeUserReadCurrentlyPlaying,
		spotifyauth.ScopePlaylistReadPrivate,
		spotifyauth.ScopeUserLibraryRead,
	}
)

func Authenticate(cfg *config.Config) (*spotify.Client, error) {
	auth = spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(scopes...),
		spotifyauth.WithClientID(cfg.ClientID),
		spotifyauth.WithClientSecret(cfg.ClientSecret),
	)

	// 既存のトークンがあれば使用
	if cfg.AccessToken != "" {
		token := &oauth2.Token{
			AccessToken:  cfg.AccessToken,
			RefreshToken: cfg.RefreshToken,
			Expiry:       time.Unix(cfg.TokenExpiry, 0),
		}

		client := spotify.New(auth.Client(context.Background(), token))

		// トークンが有効かチェック
		_, err := client.CurrentUser(context.Background())
		if err == nil {
			return client, nil
		}
	}

	// 新規認証
	http.HandleFunc("/callback", completeAuth)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Got request for:", r.URL.String())
	})

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	url := auth.AuthURL(state)
	fmt.Println("Please log in to Spotify by visiting the following page in your browser:")
	fmt.Println(url)

	// Wait for auth to complete
	token := <-ch

	// Save token to config
	cfg.AccessToken = token.AccessToken
	cfg.RefreshToken = token.RefreshToken
	cfg.TokenExpiry = token.Expiry.Unix()
	if err := cfg.Save(); err != nil {
		return nil, err
	}

	client := spotify.New(auth.Client(context.Background(), token))
	return client, nil
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	token, err := auth.Token(r.Context(), state, r)
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		log.Fatal(err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		log.Fatalf("State mismatch: %s != %s\n", st, state)
	}

	fmt.Fprintf(w, "Login Completed! You can close this window and return to the terminal.")
	ch <- token
}
