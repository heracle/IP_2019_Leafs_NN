package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strings"

    "github.com/dgrijalva/jwt-go"
    "github.com/gorilla/context"
    "github.com/mitchellh/mapstructure"
)

type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type JwtToken struct {
    Token string `json:"token"`
}

type Exception struct {
    Message string `json:"message"`
}

// 	mux.HandleFunc("/authenticate", createTokenEndpoint)
func createTokenEndpoint(w http.ResponseWriter, req *http.Request) {
    fmt.Printf("here in auth\n")

    if req.Method != "POST" {
		fmt.Println("\t Error: request is not of type POST")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

    var user User
    if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
        fmt.Printf("error while decoding user: %v", err)
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": user.Username,
        "password": user.Password,
    })
    tokenString, err := token.SignedString([]byte("secret"))
    if err != nil {
        fmt.Printf("error while signing token string: %v", err)
    }

    if err := userServerInteraction(
        user.Username,
        "Login",
    ); err != nil {
        fmt.Printf("error while writing in user log file: %v", err)
    }

    json.NewEncoder(w).Encode(JwtToken{Token: tokenString})
}

// mux.HandleFunc("/protected", protectedEndpoint)
func protectedEndpoint(w http.ResponseWriter, req *http.Request) {
    params := req.URL.Query()
    token, _ := jwt.Parse(params["token"][0], func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("There was an error")
        }
        return []byte("secret"), nil
    })
    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        var user User
        mapstructure.Decode(claims, &user)
        json.NewEncoder(w).Encode(user)
    } else {
        json.NewEncoder(w).Encode(Exception{Message: "Invalid authorization token"})
    }
}

// mux.HandleFunc("/test", validateMiddleware(testEndpoint))
func validateMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        authorizationHeader := req.Header.Get("authorization")
        if authorizationHeader != "" {
            bearerToken := strings.Split(authorizationHeader, " ")
            if len(bearerToken) == 2 {
                token, error := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
                    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                        return nil, fmt.Errorf("There was an error")
                    }
                    return []byte("secret"), nil
                })
                if error != nil {
                    json.NewEncoder(w).Encode(Exception{Message: error.Error()})
                    return
                }
                if token.Valid {
                    context.Set(req, "decoded", token.Claims)
                    next(w, req)
                } else {
                    json.NewEncoder(w).Encode(Exception{Message: "Invalid authorization token"})
                }
            }
        } else {
            json.NewEncoder(w).Encode(Exception{Message: "An authorization header is required"})
        }
    })
}

func testEndpoint(w http.ResponseWriter, req *http.Request) {
    decoded := context.Get(req, "decoded")
    var user User
    mapstructure.Decode(decoded.(jwt.MapClaims), &user)
    json.NewEncoder(w).Encode(user)
}
