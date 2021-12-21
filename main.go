package main

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"time"
)

func main() {
	// It is unknown whether storing cookies is necessary, but this is just a precaution
	j, err := cookiejar.New(nil)

	if err != nil {
		panic(err)
	}

	c := &http.Client{
		Timeout: 1000 * time.Millisecond,
		Jar: j,
	}

	sid, err := GetSessionID(c)

	if err != nil {
		panic(err)
	}

	fmt.Println("sid", sid)

	rid, err := CreateCrosstabCSVRequest(c, sid, DocIDDailyAndTotalCases)

	if err != nil {
		panic(err)
	}

	fmt.Println("rid", rid)

	dlu, err := GetCrosstabCSVRequestURL(sid, rid)

	if err != nil {
		panic(err)
	}

	fmt.Println("dlu", dlu)
}
