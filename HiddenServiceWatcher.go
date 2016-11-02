package main

import (
	"encoding/json"
	"github.com/boltdb/bolt"
	"github.com/btcsuite/go-socks/socks"
	//"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"strconv"
	"time"
)

func HSWatcher(db *bolt.DB) {
	for {
		checkTime := time.Now().Unix()
		log.Println("HS | Hidden Service Check Time " + strconv.FormatInt(checkTime, 10))
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte("watchedhiddenservices"))

			var hsWatch HSWatch

			b.ForEach(func(watchedID, watchJSON []byte) error {
				jsonErr := json.Unmarshal(watchJSON, &hsWatch)

				if jsonErr == nil {
					if hsWatch.EmailVerified == true {
						log.Printf("HS | Verified | HS: %s | Email: %s", hsWatch.HSDetails.HSAddr, hsWatch.ContactEmail)

						isReachable, hsFetchErr := FetchHS(hsWatch.HSDetails.HSAddr)

						if isReachable == false || hsFetchErr != nil {

							if hsFetchErr != nil {
								log.Println(hsFetchErr)
							}

							notifyErr := NotifyHSWatcher(hsWatch.ContactEmail, hsWatch.GPGKey, hsWatch.HSDetails.HSAddr, false)

							if notifyErr == nil {
								log.Println("HS | Notificiation sent successfully")
							} else {
								log.Println("HS | Error sending notification")
							}
						} else {
							log.Printf("HS | Hidden Service %s is up and running", hsWatch.HSDetails.HSAddr)
						}

					} else {
						log.Printf("HS | Unverified | HS: %s | Email: %s", hsWatch.HSDetails.HSAddr, hsWatch.ContactEmail)
					}
				} else {
					log.Println(jsonErr)
				}
				return nil
			})
			return nil
		})

		time.Sleep(300000 * time.Millisecond)
	}
}

func NotifyHSWatcher(email, gpgkey, hs string, isUp bool) error {

	to := []string{email}
	msg := []byte("To: " + email + "\r\n" +
		"Subject: RelayWatch Subscription for New Hidden Service\r\n" +
		"Hello,\r\n\r\n" +
		"Hidden service " + hs + " has gone down" +
		"\r\n")

	smtpErr := smtp.SendMail(SMTP_HOST, nil, "OnionWatch <notification@onionwatch.email>", to, msg)

	return nil
}

func FetchHS(url string) (bool, error) {
	var client http.Client

	proxy := &socks.Proxy{"127.0.0.1:9150", "", "", true}
	tr := &http.Transport{
		Dial: proxy.Dial,
	}
	client = http.Client{Transport: tr}

	req, httpReqErr := http.NewRequest("GET", "http://"+url, nil)
	if httpReqErr != nil {
		return false, httpReqErr
	}

	req.Header.Set("User-Agent", "OnionWatch.email/1.0 (+https://onionwatch.email/bot/)")

	resp, httpErr := client.Do(req)
	if httpErr != nil {
		return false, httpErr
	}

	defer resp.Body.Close()

	//body, _ := ioutil.ReadAll(resp.Body)
	//log.Println(string(body))

	if resp.StatusCode == 200 {
		return true, nil
	} else {
		return false, nil
	}
}
