package main

import (
	//"html/template"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"io/ioutil"
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

	//Grab all our command line config
	configuration := flag.String("conf", "", "path to configuration file")
	flag.Parse()
	conf := readConfig(*configuration)

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

// Reads our JSON formatted config file
// and returns a struct
func readConfig(filename string) CoreConf {
	var conf CoreConf

	if filename == "" {
		conf.DbPath = "./nionyn.bolt"
		conf.ListenIP = "127.0.0.1"
		conf.ListenIPv6 = "::1"
		conf.ListenPort = 8080
		conf.SOCKSConfig = "127.0.0.1:9150"
		conf.FQDN = "onionwatch.email"
	} else {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatal("Cannot read configuration file ", filename)
		}
		err = json.Unmarshal(b, &conf)
		if err != nil {
			log.Fatal("Cannot parse configuration file ", filename)
		}

		if conf.DbPath == "" {
			conf.DbPath = "./nionyn.bolt"
		}

		if conf.ListenIP == "" {
			conf.ListenIP = "127.0.0.1"
		}

		if conf.ListenIPv6 == "" {
			conf.ListenIPv6 = "::1"
		}

		if conf.ListenPort == 0 {
			conf.ListenPort = 8080
		}

		if conf.SOCKSConfig == "" {
			conf.SOCKSConfig = "127.0.0.1:9150"
		}

		if conf.FQDN == "" {
			conf.FQDN = "onionwatch.email"
		}
	}
	return conf
}
