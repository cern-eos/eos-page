package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/apeters/eospage/internal/api"
	"github.com/apeters/eospage/internal/store"
	"github.com/apeters/eospage/web"
)

func main() {
	dataDir := env("DATA_DIR", "data")
	secret := os.Getenv("CONTROLLER_SECRET")
	if secret == "" {
		log.Println("warning: CONTROLLER_SECRET is empty; controller login will be unavailable")
	}
	if err := os.MkdirAll(filepath.Join(dataDir, "media"), 0o755); err != nil {
		log.Fatal(err)
	}

	st, err := store.Open(filepath.Join(dataDir, "eos.db"), filepath.Join(dataDir, "media"))
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	publicFS, err := fs.Sub(web.Public, "public")
	if err != nil {
		log.Fatal(err)
	}
	controllerFS, err := fs.Sub(web.Controller, "controller")
	if err != nil {
		log.Fatal(err)
	}

	cert := strings.TrimSpace(os.Getenv("TLS_CERT"))
	key := strings.TrimSpace(os.Getenv("TLS_KEY"))
	tlsOn := cert != ""
	if tlsOn && key == "" {
		key = inferKey(cert)
		if key == "" {
			log.Fatal("TLS key is required: set TLS_KEY or place a sibling .key next to the certificate")
		}
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		if tlsOn {
			addr = ":443"
		} else {
			addr = ":8080"
		}
	}

	srv := api.New(st, api.Options{
		ControllerSecret: secret,
		PublicBaseURL:    strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")), "/"),
	}, publicFS, controllerFS)

	scheme := "http"
	if tlsOn {
		scheme = "https"
	}
	log.Printf("EOS page listening on %s", addr)
	log.Printf("public site:  %s://localhost%s/", scheme, addr)
	log.Printf("controller:   %s://localhost%s/controller", scheme, addr)

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	if tlsOn {
		err = httpSrv.ListenAndServeTLS(cert, key)
	} else {
		err = httpSrv.ListenAndServe()
	}
	if err != nil {
		log.Fatal(err)
	}
}

func inferKey(cert string) string {
	ext := filepath.Ext(cert)
	base := strings.TrimSuffix(cert, ext)
	dir := filepath.Dir(cert)
	for _, p := range []string{base + ".key", base + "-key" + ext, filepath.Join(dir, "key.pem"), filepath.Join(dir, "privkey.pem")} {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
