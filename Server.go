package main

import (
	//"html/template"
	//"encoding/json"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	//"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

const (
	HSTSEXPIRY  = 94670856
	APIVERSION  = 1
	HSTSENABLED = false
	HASHSEED    = "3qy5r0wierfhjwiejisdh0whrt8wh08rh0swdhf0ss0f0sdfhsd0fhsdf8h0d"
	FROM_EMAIL  = ""
	SMTP_HOST   = "localhost:25"
)

func main() {
	log.Println("---------------------------------------------")

	var conf CoreConf

	//Grab all our command line config
	flag.StringVar(&conf.DbPath, "dbpath", "", "path to the bolt database")
	flag.StringVar(&conf.ListenIP, "listenip", "127.0.0.1", "IPv4 address to listen on")
	flag.StringVar(&conf.ListenIPv6, "listenipv6", "::1", "IPv4 address to listen on")
	flag.Int64Var(&conf.ListenPort, "port", 8080, "IP port to bind on")
	flag.StringVar(&conf.SOCKSConfig, "socks", "127.0.0.1:9150", "SOCKS IP:Port to use (optional)")
	flag.StringVar(&conf.FQDN, "fqdn", "onionwatch.email", "FQDN for absolute URLs etc")
	flag.Parse()

	if conf.DbPath == "" {
		log.Fatal("At a minimum please specify the database path with -dbpath. See -h for full command line argument list.")
	}

	//We need a DB for holding Relay / email details
	db, err := bolt.Open(conf.DbPath, 0600, nil)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("DB opened, all is OK")
	}
	defer db.Close()

	//Make sure our buckets exist
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("watchedrelays"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("verificationlookup"))

		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		_, err = tx.CreateBucketIfNotExists([]byte("watchedhiddenservices"))

		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}

		return nil
	})

	//Start our watcher process
	go RelayWatcher(db)
	go HSWatcher(db)

	//Handle static web pages
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { WebIndex(w, r) })
	http.HandleFunc("/about/", func(w http.ResponseWriter, r *http.Request) { WebAbout(w, r) })

	//Handle functionality
	http.HandleFunc("/manage/", func(w http.ResponseWriter, r *http.Request) { WebManage(w, r, db, conf.FQDN) })
	http.HandleFunc("/subscribe/", func(w http.ResponseWriter, r *http.Request) { WebSubscribe(w, r, db, conf.FQDN) })
	http.HandleFunc("/subscribe/verify/", func(w http.ResponseWriter, r *http.Request) { WebSubscribeVerify(w, r, db, conf.FQDN) })
	http.HandleFunc("/unsubscribe/", func(w http.ResponseWriter, r *http.Request) { WebUnsubscribe(w, r, db, conf.FQDN) })

	//Serve up static resources
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("assets/css/"))))
	http.Handle("/font/", http.StripPrefix("/font/", http.FileServer(http.Dir("assets/font/"))))

	ListenPort := strconv.FormatInt(conf.ListenPort, 10)
	if conf.TLS {
		http.ListenAndServeTLS(":"+ListenPort, conf.TLSCert, conf.TLSKey, nil)
	} else {
		http.ListenAndServe(":"+ListenPort, nil)
	}
}
