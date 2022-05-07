package healthcheck

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Healthcheck is a helper function for dockers healthcheck.
func Healthcheck(port int) {
	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)
	if _, err := http.Get(url); err != nil {
		log.Fatalf("we have an error %v", err)
	}
	os.Exit(0)
}
