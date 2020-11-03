package main

import (
    "context"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"

    "github.com/go-chi/chi"

    "golang.org/x/oauth2"
)

const apiBase = "https://sso.compsoc.ie/auth/realms/base";

var oauth = oauth2.Config{
	RedirectURL:  "http://localhost:8080/callback",
	ClientID:     "codel",
	ClientSecret: "aa70e5d8-fb0d-4135-b2a9-d0c7751f97c2",
	Endpoint: oauth2.Endpoint{
		AuthURL: apiBase + "/protocol/openid-connect/auth",
		TokenURL: apiBase + "/protocol/openid-connect/token",
	},
}

var httpClient = http.Client{}

func main() {
    r := chi.NewRouter()
    r.Get("/", handleRedirect)
    r.Get("/callback", handleCallback)
    log.Fatal(http.ListenAndServe(":3200", r))
}

func handleRedirect(w http.ResponseWriter, r *http.Request) {
    uri := oauth.AuthCodeURL("")
    http.Redirect(w, r, uri, http.StatusTemporaryRedirect)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
    code := r.URL.Query().Get("code")
    if code == "" {
        w.Write([]byte("No code"))
        w.WriteHeader(400)
        return
    }

    token, err := oauth.Exchange(context.Background(), code)
    if err != nil {
        w.Write([]byte("Failed to exchange token"))
        w.WriteHeader(400)
        return
    }

    channel, err := getChannel(token.AccessToken)
    if err != nil {
        w.Write([]byte("Failed to get channel"))
        w.WriteHeader(400)
        return
    }
    bs, err := json.Marshal(channel)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    w.Write(bs)
}

func getChannel(authToken string) (*Channel, error) {
    url := fmt.Sprintf("%s/kappa/v2/channels/me", apiBase)
    req, err := http.NewRequest(http.MethodGet, url, nil)
    if err != nil {
        return nil, err
    }
    req.Header.Add("Authorization", fmt.Sprintf("OAuth %s", authToken))
    res, err := httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    bs, err := ioutil.ReadAll(res.Body)
    if err != nil {
        return nil, err
    }
    channel := &Channel{}
    err = json.Unmarshal(bs, channel)
    if err != nil {
        return nil, err
    }
    return channel, nil
}