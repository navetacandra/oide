package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	_ "github.com/lib/pq"
	"github.com/navetacandra/oide/lib/cloudshell"
	"github.com/navetacandra/oide/lib/server"
	"github.com/navetacandra/oide/lib/server/git" // credits: https://github.com/asim
	"golang.org/x/crypto/bcrypt"
)

type FileCache struct {
	path    string
	content []byte
}

const domain = "navetacandraa.my.id"

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
var dockerProxyPattern = regexp.MustCompile(`^([\d\w]+)\-(\d+)$`)
var caches = []FileCache{}

func GetConnection(host string, port int, user string, password string, dbname string) *sql.DB {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname),
	)
	if err != nil {
		panic(err)
	}
	return db
}

func GetFile(path string) ([]byte, error) {
	// for _, file := range caches {
	//   if file.path == path {
	//     return file.content, nil
	//   }
	// }
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	caches = append(caches, FileCache{path, content})
	return content, nil
}

func PreventDirTree(fs http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subdomain := server.ParseSubdomain(domain, r)
		if subdomain == "" {
			if r.URL.Path[len(r.URL.Path)-1] == '/' {
				http.NotFound(w, r)
				return
			}
			fs.ServeHTTP(w, r)
			return
		}
		http.NotFound(w, r)
	}
}

func RenderHtml(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		content, err := GetFile(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, string(content))
	}
}

func WriteHtml(path string, w http.ResponseWriter, r *http.Request) {
	content, err := GetFile(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, string(content))
}

func GenerateToken(userId string, time int64) string {
	json := fmt.Sprintf("{\"user_id\": \"%s\", \"time\": %d}", userId, time)
	return base64.StdEncoding.EncodeToString([]byte(json))
}

func ErrorLoginForm(username string, usernameError string, passwordError string) string {
	return fmt.Sprintf("<form id=\"loginForm\" hx-post=\"/login\" hx-target=\"#loginForm\" hx-trigger=\"submit\">\n<div class=\"flex flex-col justify-center align-start px-4 mb-1\"><label class=\"my-1\" for=\"username\">Username:</label><input class=\"underline w-100 py-1 px-1\" type=\"text\" id=\"username\" name=\"username\" placeholder=\"Username\" value=\"%s\" />%s</div><div class=\"flex flex-col justify-center align-start px-4 mb-2\"> <label class=\"my-1\" for=\"password\">Password:</label><input class=\"underline w-100 py-1 px-1\" type=\"password\" id=\"password\" name=\"password\" placeholder=\"Password\" />%s</div><div class=\"flex justify-end\"> <button type=\"submit\" class=\"btn bg-blue dblock px-5 py-1 mx-3\">Login</button></div></form>", username, usernameError, passwordError)
}

func InvalidateToken(db *sql.DB, w http.ResponseWriter, r *http.Request) bool {
  cookie, err := r.Cookie("accesToken")
  if err != nil {
    return true
  }
  accesToken := cookie.Value
  if accesToken == "" {
    return true
  }
  cookie.Expires = time.Now().Add(time.Hour * -1)
  http.SetCookie(w, cookie)
  _, err = db.Exec("DELETE FROM tokens WHERE token = $1", accesToken)
  return true
}

func ValidateToken(db *sql.DB, w http.ResponseWriter, r *http.Request) bool {
  cookie, err := r.Cookie("accesToken")
  if err != nil {
    return false
  }
  accesToken := cookie.Value
  if accesToken == "" {
    return false
  }
  var id int
  var expire int64
  err = db.QueryRow("SELECT id, expire FROM tokens WHERE token = $1", accesToken).Scan(&id, &expire)
  if err != nil || expire < time.Now().Unix() {
    InvalidateToken(db, w, r)
    return false
  }
  return true
}

func main() {
	db := GetConnection("localhost", 5432, "postgres", "postgres", "oide")
	defer db.Close()

	assetsFile := http.FileServer(http.Dir("./web/assets/"))
	http.Handle("/assets/*", PreventDirTree(http.StripPrefix("/assets/", assetsFile)))

	http.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
		handled :=
			server.HandlerChain(
				server.HandleSubdomain(domain, "", func(w http.ResponseWriter, r *http.Request) bool {
					if !server.HandlerChain(
						server.HandlePath("/", RenderHtml("./web/home.html")),
						server.HandlePath("/login", func(w http.ResponseWriter, r *http.Request) {
              if ValidateToken(db, w, r) {
                w.Header().Set("HX-Redirect", "/dashboard")
                http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
                return
							}
							if r.Method == http.MethodGet {
								WriteHtml("./web/login.html", w, r)
							}

							if r.Method == "POST" {
								username := r.FormValue("username")
								password := r.FormValue("password")
								usernameError := ""
								passwordError := ""

								if username == "" {
									usernameError = "<small class=\"text-red py-1\">Username is required!</small>"
								} else if !usernameRegex.MatchString(username) {
									usernameError = "<small class=\"text-red py-1\">Username is invalid!</small>"
								}
								if password == "" {
									passwordError = "<small class=\"text-red py-1\">Password is required!</small>"
								}

								if usernameError == "" && passwordError == "" {
									var uid string
									var dbuname string
									var dbpassword string
									err := db.QueryRow("SELECT id, username, password FROM users WHERE username = $1", username).Scan(&uid, &dbuname, &dbpassword)
									if err != nil || bcrypt.CompareHashAndPassword([]byte(dbpassword), []byte(password)) != nil {
										usernameError = "<small class=\"text-red py-1\">Username or password is incorrect!</small>"
										w.Write([]byte(ErrorLoginForm(username, usernameError, passwordError)))
									} else {
                    token := GenerateToken(uid, time.Now().Unix())
                    expire := time.Now().Add(365 * 24 * time.Hour)
										cookie := http.Cookie{}
										cookie.Name = "accesToken"
										cookie.Value = token
										cookie.Expires = expire
										http.SetCookie(w, &cookie)

                    _, _ = db.Exec("INSERT INTO tokens (user_id, token, expire) VALUES ($1, $2, $3)", uid, token, expire.Unix())
                    w.Header().Set("HX-Redirect", "/dashboard")
										return
									}
								} else {
									w.Write([]byte(ErrorLoginForm(username, usernameError, passwordError)))
								}
							}
						}),
						server.HandlePath("/logout", func(w http.ResponseWriter, r *http.Request) {
							currentCookie, _ := r.Cookie("accesToken")
              currentCookie.Expires = time.Now().Add(time.Hour * -1)

              db.Exec("DELETE FROM tokens WHERE token = $1", currentCookie.Value)
              http.SetCookie(w, currentCookie)
              w.Header().Set("HX-Redirect", "/login")
              http.Redirect(w, r, "/login", http.StatusSeeOther)
						}),
						server.HandlePath("/dashboard", func(w http.ResponseWriter, r *http.Request) {
							if !ValidateToken(db, w, r) {
						    w.Header().Set("HX-Redirect", "/login")
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
							}
							fmt.Fprintf(w, "Hello, World!") 
						}),
					)(w, r) {
						http.NotFound(w, r)
					}

					return true
				}),

				server.HandleSubdomain(domain, "git", func(w http.ResponseWriter, r *http.Request) bool {
					git.Handler(w, r, func(dir string, repo string, branch string) {
						fmt.Printf("Pushed to %s:%s to %s", repo, branch, dir)
					})
					return true
				}),

				server.HandleSubdomain(domain, "ssh", func(w http.ResponseWriter, r *http.Request) bool {
					cloudshell.Handle(w, r, 10*time.Second, []string{"bash", "-l"})
					return true
				}),

				server.HandleSubdomainRegexp(domain, dockerProxyPattern, func(w http.ResponseWriter, r *http.Request, match []string) {
					var ip string
					row := db.QueryRow("SELECT container_ip FROM containers WHERE user_id=(SELECT id FROM users WHERE username=$1)", match[1])
					err := row.Scan(&ip)

					if err != nil {
						http.Error(w, fmt.Sprintf("Failed get container: %s", err.Error()), http.StatusInternalServerError)
						return
					}
					server.HandleProxy(w, r, fmt.Sprintf("http://%s:%s", ip, match[2]))
				}),
			)(w, r)

		if !handled {
			http.NotFound(w, r)
		}
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
