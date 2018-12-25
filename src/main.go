package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/go-redis/redis"
	"github.com/satori/go.uuid"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	templates = template.Must(template.ParseFiles("views/index.html"))
	validPath = regexp.MustCompile("^/([a-zA-Z0-9-_]{36})/([a-zA-Z0-9-_]+)$")

	ErrEncryptionFailed = errors.New("encryption failed")
	ErrWriteFailed      = errors.New("writing	failed")
	ErrUnpad            = errors.New("unpad error")
	ErrHealthFail       = errors.New("health: could not reach db")

	client        *redis.Client
	REDIS_ADDRESS string
	HOST_URL      string
)

func set(token string, payload []byte, ttl time.Duration) error {
	if err := client.Set(token, payload, ttl).Err(); err != nil {
		return ErrWriteFailed
	}
	return nil
}

func get(token string) (string, error) {
	v, err := client.Get(token).Result()
	if err != nil {
		return "", ErrWriteFailed
	}
	client.Del(token)
	return v, nil
}

func generateKey() ([]byte, error) {
	key := make([]byte, aes.BlockSize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func encrypt(plaintext, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decrypt(cipherText, key []byte) ([]byte, error) {
	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherText) < nonceSize {
		return nil, errors.New("cipherText too short")
	}

	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]
	return gcm.Open(nil, nonce, cipherText, nil)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	if err := templates.ExecuteTemplate(w, tmpl+".html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		secretHandler(w, r)
		return
	}
	d := validPath.FindStringSubmatch(r.URL.Path)
	if d != nil {
		revealHandler(w, r, d[1], d[2])
		return
	}
	renderTemplate(w, "index", nil)
}

func secretHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Not Allowed", http.StatusBadRequest)
	}
	r.ParseForm()
	pw := r.FormValue("secret")
	ttl, err := time.ParseDuration(r.FormValue("ttl"))
	if err != nil {
		ttl = 30 * time.Minute
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}
	storageKey := uuid.String()
	hashKey, err := generateKey()
	if err != nil {
		http.Error(w, "Unexpected error", http.StatusInternalServerError)
		return
	}
	c, err := encrypt([]byte(pw), hashKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := set(storageKey, c, ttl); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token := strings.Join([]string{HOST_URL, storageKey, hex.EncodeToString(hashKey)}, "/")
	renderTemplate(w, "index", token)
}

func revealHandler(w http.ResponseWriter, r *http.Request, token, key string) {
	k, _ := hex.DecodeString(key)
	c, err := get(token)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
		return
	}
	pw, _ := decrypt([]byte(c), k)

	renderTemplate(w, "index", string(pw))
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := client.Ping().Result(); err != nil {
		http.Error(w, ErrHealthFail.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	client = redis.NewClient(&redis.Options{
		Addr:     REDIS_ADDRESS,
		Password: "",
		DB:       0})
	_, err := client.Ping().Result()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/health", healthCheckHandler)
	go http.ListenAndServeTLS(":443", "/run/secrets/server.cert", "/run/secrets/server.key", nil)
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func init() {
	REDIS_ADDRESS = os.Getenv("REDIS_ADDRESS")
	HOST_URL = os.Getenv("HOST_URL")
}
