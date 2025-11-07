//go:build !solution

package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type User struct {
	Name  string
	Email string
}

func ContextUser(ctx context.Context) (*User, bool) {
	us, ok := ctx.Value("user").(*User)
	return us, ok
}

var ErrInvalidToken = errors.New("invalid token")

type TokenChecker interface {
	CheckToken(ctx context.Context, token string) (*User, error)
}

// type LazyChecker map[string]struct {
// 	user *User
// 	err  error
// }

// func (lc LazyChecker) CheckToken(ctx context.Context, token string) (*User, error) {
// 	res := lc[token]
// 	return res.user, res.err
// }

func CheckAuth(checker TokenChecker) func(next http.Handler) http.Handler {

	getToken := func(r *http.Request) (string, error) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			return "", errors.New("authorization header is required")
		}
		if !strings.HasPrefix(auth, "Bearer") {
			return "", errors.New("bearer token is required")
		}

		parts := strings.Split(auth, "Bearer")
		if len(parts) < 2 {
			return "", errors.New("invalid token")
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			return "", errors.New("invalid token")
		}
		return token, nil
	}

	return func(next http.Handler) http.Handler {
		f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			t, err := getToken(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			u, err := checker.CheckToken(r.Context(), t)

			if err != nil {
				http.Error(w, "token check error", http.StatusInternalServerError)
				return
			}

			cr := r.WithContext(context.WithValue(r.Context(), "user", u))

			next.ServeHTTP(w, cr)
		})
		return f
	}

}

// func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
// 	u, ok := ContextUser(r.Context())
// 	fmt.Println(u)
// 	fmt.Println(ok)
// 	fmt.Fprintf(w, "Hello Worlde4kiy")
// }

// func main() {

// 	m := chi.NewRouter()

// 	c := LazyChecker{
// 		"token0": {
// 			user: &User{Name: "Fedor", Email: "dartslon@gmail.com"},
// 		},

// 		"token1": {
// 			err: fmt.Errorf("database offline"),
// 		},

// 		"token2": {
// 			err: fmt.Errorf("token expired: %w", ErrInvalidToken),
// 		},
// 	}

// 	middleware := CheckAuth(c)
// 	m.Use(middleware)

// 	m.Get("/", HelloWorldHandler)

// 	log.Fatal(http.ListenAndServe(":8889", m))
// }
