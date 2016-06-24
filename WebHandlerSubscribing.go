package main

import (
	"fmt"
	"github.com/boltdb/bolt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/smtp"
)

func WebSubscribe(w http.ResponseWriter, r *http.Request, db *bolt.DB, FQDN string) {

	if r.Method == "POST" {
		subscribeErr := CreateSubscription(r, db)

		if subscribeErr == nil {
			tmpl, err := template.New("subscribed").ParseFiles("assets/templates/subscribe.html")
			err = tmpl.Execute(w, nil)

			if err != nil {
				log.Fatal(err)
			}
		} else {
			w.Header().Set("ow-success", "false")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "There was an error processing your subscription: %s", subscribeErr)
		}

	} else if r.Method == "GET" {
		url := strings.Split(r.URL.String(), "/")
		log.Println(url[2])
		log.Println(url[3])

		//Check whether we are monitoring a hidden service or a relay
		if strings.ToLower(url[2]) == "relay" {
			relayDetails, relayErr := FetchRelay(url[3])

			//We don't really care if fetching the relay info worked or not
			if relayErr != nil {
				log.Println(relayErr)
			} else {
				log.Println("Relay: " + relayDetails.NickName + " / " + relayDetails.Fingerprint)
			}

			tmpl, err := template.New("subscribe-relay").ParseFiles("assets/templates/subscribe.html")
			err = tmpl.Execute(w, relayDetails)

			if err != nil {
				log.Fatal(err)
			}
		} else if strings.ToLower(url[2]) == "hidden-service" {
			hs := HiddenService{HSAddr: url[3]}

			tmpl, err := template.New("subscribe-hs").ParseFiles("assets/templates/subscribe.html")
			err = tmpl.Execute(w, hs)

			if err != nil {
				log.Fatal(err)
			}

		} else {
			tmpl, err := template.New("error").ParseFiles("assets/templates/subscribe.html")
			err = tmpl.Execute(w, nil)

			if err != nil {
				log.Fatal(err)
			}

		}
	} else {
		//We don't accept PUT / DELETE etc on the /create/ URL
		w.Header().Set("ow-success", "false")
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Method %s is not allowed!", r.Method)
		fmt.Fprintf(w, "Method %s is not allowed!", r.Method)
	}
}

func WebSubscribeVerify(w http.ResponseWriter, r *http.Request, db *bolt.DB, FQDN string) {

	url := strings.Split(r.URL.String(), "/")
	log.Println(url[3])

	dbErr := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("verificationlookup"))
		md5Hash := b.Get([]byte(url[3]))

		if md5Hash != nil {
			b := tx.Bucket([]byte("watchedrelays"))
			wrJSON := b.Get([]byte(md5Hash))

			if wrJSON != nil {
				var relayWatch RelayWatch
				jsonUnMarshalErr := json.Unmarshal(wrJSON, &relayWatch)

				if jsonUnMarshalErr == nil {
					relayWatch.EmailVerified = true

					buf, jsonMarshalErr := json.Marshal(relayWatch)

					if jsonMarshalErr != nil {
						return jsonMarshalErr
					}

					return b.Put([]byte(md5Hash), buf)
				} else {
					return jsonUnMarshalErr
				}
			} else {
				//Possible this is a hidden service key...
				log.Println("Couldn't find a key in the watched relays DB, checking watched hiddenservices")

				b := tx.Bucket([]byte("watchedhiddenservices"))
				wrJSON := b.Get([]byte(md5Hash))

				if wrJSON == nil {
					log.Println("Couldn't find the MD5 hash in hiddenservices either :/")
					return nil
				}

				var hsWatch HSWatch
				jsonUnMarshalErr := json.Unmarshal(wrJSON, &hsWatch)

				if jsonUnMarshalErr == nil {
					hsWatch.EmailVerified = true

					buf, jsonMarshalErr := json.Marshal(hsWatch)

					if jsonMarshalErr != nil {
						return jsonMarshalErr
					}

					return b.Put([]byte(md5Hash), buf)
				} else {
					return jsonUnMarshalErr
				}

			}
		} else {
			return errors.New("No such verification key")
		}
	})

	if dbErr == nil {
		tmpl, err := template.New("verified").ParseFiles("assets/templates/subscribe.html")
		err = tmpl.Execute(w, nil)

		if err != nil {
			log.Fatal(err)
		}
	} else {
		w.Header().Set("ow-success", "false")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "There was an error (%s) processing your verification request", dbErr)

	}

}

func WebUnsubscribe(w http.ResponseWriter, r *http.Request, db *bolt.DB, FQDN string) {

	if r.Method == "POST" {
		tmpl, err := template.New("create").ParseFiles("assets/templates/create.html")
		err = tmpl.Execute(w, nil)

		if err != nil {
			log.Fatal(err)
		}

	} else if r.Method == "GET" {

		tmpl, err := template.New("create").ParseFiles("assets/templates/create.html")
		err = tmpl.Execute(w, nil)

		if err != nil {
			log.Fatal(err)
		}

	} else {
		//We don't accept PUT / DELETE etc on the /create/ URL
		w.Header().Set("ow-success", "false")
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Method %s is not allowed!", r.Method)
		fmt.Fprintf(w, "Method %s is not allowed!", r.Method)
	}
}

func CreateSubscription(r *http.Request, db *bolt.DB) error {

	r.ParseForm()

	if r.FormValue("type") == "hs" {
		return createHSSubscription(r, db)
	} else {
		return createRelaySubscription(r, db)
	}

}

func createHSSubscription(r *http.Request, db *bolt.DB) error {
	r.ParseForm()
	hs := HiddenService{HSAddr: r.FormValue("hiddenservice")}

	hsWatch := HSWatch{ContactEmail: r.FormValue("email"),
		GPGKey:        r.FormValue("gpg"),
		EmailVerified: false,
		HSDetails:     hs}

	//Dedupe key
	md5Hash := GetMD5Hash(hsWatch.HSDetails.HSAddr + hsWatch.ContactEmail)
	sha256Hash := GetSHA256Hash(hsWatch.ContactEmail + HASHSEED)

	//Ignore dedupe for the moment

	buf, jsonErr := json.Marshal(hsWatch)

	if jsonErr == nil {
		dbErr := db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("watchedhiddenservices"))
			return b.Put([]byte(md5Hash), buf)
		})

		if dbErr == nil {
			dbErr = db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("verificationlookup"))
				return b.Put([]byte(sha256Hash), []byte(md5Hash))
			})
		}

		if dbErr == nil {
			to := []string{hsWatch.ContactEmail}
			msg := []byte("To: " + hsWatch.ContactEmail + "\r\n" +
				"Subject: RelayWatch Subscription for New Hidden Service\r\n" +
				"\r\n" +
				"You (or someone else) has requested that your email address receive status notifications regarding the hidden service " + hsWatch.HSDetails.HSAddr +
				"\r\n" +
				"If you do want to receive these email notificaitions please visit the following URL:\r\n" +
				"https://onionwatch.email/subscribe/verify/" + sha256Hash + "/" +
				"\r\n")

			smtpErr := smtp.SendMail("localhost:25", nil, "OnionWatch Verification <verify@onionwatch.email>", to, msg)
			if smtpErr != nil {
				log.Println(smtpErr)
				return smtpErr
			} else {
				return nil
			}
		} else {
			return dbErr
		}
	} else {
		return jsonErr
	}

	return nil
}

func createRelaySubscription(r *http.Request, db *bolt.DB) error {

	r.ParseForm()
	relayWatch := RelayWatch{Fingerprint: r.FormValue("fingerprint"),
		ContactEmail:  r.FormValue("email"),
		GPGKey:        r.FormValue("gpg"),
		EmailVerified: false}

	//Dedupe key
	md5Hash := GetMD5Hash(relayWatch.Fingerprint + relayWatch.ContactEmail)
	sha256Hash := GetSHA256Hash(r.FormValue("email") + HASHSEED)

	//However we shouldn't expose that a given email address is watching a given relay
	duplicateCheckErr := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("watchedrelays"))
		wrJSON := b.Get([]byte(md5Hash))

		if wrJSON == nil {
			//This is empty which is good
			return nil
		} else {
			//Something is there
			log.Println("Duplicate attempt for " + relayWatch.Fingerprint + relayWatch.ContactEmail)
			//We should use a custom error
			return errors.New("Duplicate")
		}
	})

	if duplicateCheckErr != nil {
		//Don't store to the DB and don't send a new email
		return nil
	}

	buf, jsonErr := json.Marshal(relayWatch)

	if jsonErr == nil {
		dbErr := db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("watchedrelays"))
			return b.Put([]byte(md5Hash), buf)
		})

		if dbErr == nil {
			dbErr = db.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("verificationlookup"))
				return b.Put([]byte(sha256Hash), []byte(md5Hash))
			})
		}

		if dbErr == nil {
			to := []string{relayWatch.ContactEmail}
			msg := []byte("To: " + relayWatch.ContactEmail + "\r\n" +
				"Subject: RelayWatch Subscription for Tor Relay " + relayWatch.Fingerprint + "\r\n" +
				"\r\n" +
				"You (or someone else) has requested that your email address receive status notifications regarding the Tor Relay " + relayWatch.Fingerprint +
				"\r\n" +
				"If you do want to receive these email notificaitions please visit the following URL:\r\n" +
				"https://onionwatch.email/subscribe/verify/" + sha256Hash + "/" +
				"\r\n")

			smtpErr := smtp.SendMail("localhost:25", nil, "OnionWatch Verification <verify@onionwatch.email>", to, msg)
			if smtpErr != nil {
				log.Println(smtpErr)
				return smtpErr
			} else {
				return nil
			}
		} else {
			return dbErr
		}
	} else {
		return jsonErr
	}

}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func GetSHA256Hash(text string) string {
	hash := sha256.New()
	hash.Write([]byte(text))
	md := hash.Sum(nil)
	return hex.EncodeToString(md)
}
