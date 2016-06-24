package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"log"
	//  "math"
	"strconv"
	"time"
)

// RelayWatcher is the go routine that iterates over all the RelayWatcher objects in the DB and checks their status
// If the state has changed from the last run the user is emailed a notification (optionally GPG encrypted).
func RelayWatcher(db *bolt.DB) {
	log.Println("Hello")
	for {
		checkTime := time.Now().Unix()
		log.Println("Check Time " + strconv.FormatInt(checkTime, 10))
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("watchedrelays"))

			var relayWatch RelayWatch
			var whatsChanged RelayStateTemplate

			b.ForEach(func(watchedID, watchJSON []byte) error {
				jsonErr := json.Unmarshal(watchJSON, &relayWatch)
				whatsChanged = RelayStateTemplate{}

				if jsonErr == nil {
					if relayWatch.EmailVerified == true {
						log.Printf("Relay: %s | Email: %s", relayWatch.Fingerprint, relayWatch.ContactEmail)

						//Check their relay
						relay, relayFetchErr := FetchRelay(relayWatch.Fingerprint)

						log.Println("Fetched relay! Processing...")

						//Check if we've ever seen the relay before
						if relayWatch.RelayDetails.Fingerprint == "" {
							log.Println("This relay has never been seen before, storing state")
							relayWatch.RelayDetails = relay
						}

						if relayFetchErr == nil {
							if relay.Running != relayWatch.RelayDetails.Running {
								//State has changed
								whatsChanged.anythingChanged = true
								whatsChanged.HasRunningChanged = true
							}

							if relay.ASNumber != relayWatch.RelayDetails.ASNumber {
								//State has changed
								whatsChanged.anythingChanged = true
								whatsChanged.HasASNChanged = true
							}

							//etc etc
							//etc

							//Email the user if neccessary
							if whatsChanged.anythingChanged == true {
								log.Println("Change detected!")
								log.Println(whatsChanged)

								notifyErr := NotifyUser(whatsChanged, relay)

								if notifyErr != nil {
									log.Println(notifyErr)
								}
							} else {
								log.Println("Nothing has changed!")
								log.Println(relay)
							}

							//Update the DB
							relayWatch.RelayDetails = relay
							buf, jsonMarshalErr := json.Marshal(relayWatch)

							if jsonMarshalErr != nil {
								log.Println(jsonMarshalErr)
								return jsonMarshalErr
							}

							return b.Put([]byte(watchedID), buf)

						} else {
							log.Println("Error:")
							log.Println(relayFetchErr)
						}
					} else {
						log.Printf("Email %s not yet verified, skipping check of %s", relayWatch.ContactEmail, relayWatch.Fingerprint)
					}
					return nil
				} else {
					log.Printf("There was an issue unmarshaling %s", watchedID)
					return jsonErr
				}
			})

			return nil
		})
		time.Sleep(300000 * time.Millisecond)
	}
}

func NotifyUser(whatsChanged RelayStateTemplate, relay Relay) error {
	return nil
}
