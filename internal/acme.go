package internal

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"log"
	"os"
	"sync"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
)

type AccountStore struct {
	mu    sync.Mutex
	Path  string
	Users map[string]*MyUser
}

// Save writes all users back to disk
func (s *AccountStore) Save() error {

	// refresh PrivateKeyPEM before save
	for _, u := range s.Users {
		if u.key != nil {
			u.KeyPEM = string(certcrypto.PEMEncode(u.key))
		}
	}

	data, err := json.MarshalIndent(s.Users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0600)
}

type MyUser struct {
	Email        string
	Registration *registration.Resource
	KeyPEM       string `json:"key"`
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

const userFilePath = "./data/accounts.json"

var (
	globalStore *AccountStore
	storeMu     sync.Mutex
)

func init() {
	store, err := LoadAccountStore(userFilePath)
	if err != nil {
		log.Fatalf("Failed to load account store: %v", err)
	}
	globalStore = store
}

// LoadAccountStore loads all users from file (if exists)
func LoadAccountStore(path string) (*AccountStore, error) {
	store := &AccountStore{Path: path, Users: map[string]*MyUser{}}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return store, nil // empty store
		}
		return nil, err
	}

	if err := json.Unmarshal(data, &store.Users); err != nil {
		return nil, err
	}

	// decode PEM keys
	for email, u := range store.Users {
		if u.KeyPEM != "" {
			block, _ := pem.Decode([]byte(u.KeyPEM))
			if block == nil {
				continue
			}
			key, err := x509.ParseECPrivateKey(block.Bytes)
			if err == nil {
				u.key = key
				u.Email = email
			}
		}
	}

	return store, nil
}

func GetLegoClient(user *MyUser) (*lego.Client, error) {
	config := lego.NewConfig(user)

	// config.CADirURL = "https://acme-v02.api.letsencrypt.org/directory"
	config.CADirURL = "https://acme-staging-v02.api.letsencrypt.org/directory"
	config.Certificate.KeyType = certcrypto.RSA2048

	return lego.NewClient(config)
}

func RegisterUser(email string) (*MyUser, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	myUser := MyUser{
		Email: email,
		key:   privateKey,
	}

	client, err := GetLegoClient(&myUser)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	myUser.Registration = reg

	return &myUser, nil
}

func GetUser(email string) (*MyUser, error) {
	storeMu.Lock()
	defer storeMu.Unlock()

	if user, ok := globalStore.Users[email]; ok && user.key != nil {
		return user, nil
	}

	user, err := RegisterUser(email)

	if err != nil {
		return nil, err
	}
	globalStore.Users[email] = user
	if err := globalStore.Save(); err != nil {
		return nil, err
	}

	return user, nil
}

func GenerateCertificate(domains []string) (*certificate.Resource, error) {
	acmeEmail := os.Getenv("ACME_EMAIL")

	user, err := GetUser(acmeEmail)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	client, err := GetLegoClient(user)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	provider, err := cloudflare.NewDNSProvider()
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	err = client.Challenge.SetDNS01Provider(provider)

	// New users will need to register

	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return certificates, nil
}
