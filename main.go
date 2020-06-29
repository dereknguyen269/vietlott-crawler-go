package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gocolly/colly"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type Reward struct {
	LotteryType string
	DateOpen    string
	Code        string
	Results     []string
}

type ResponseResult struct {
	URL    string   `json:"URL"`
	Status string   `json:"Status"`
	Data   []Reward `json:"Reward"`
}

type Mega645 struct {
	LotteryType string
	DateOpen string
	Code string
	Results []string
}

func getResultsMega645(url string, lotteryType string) []Reward {
	c := colly.NewCollector()
	allRewards := []Reward{}
	c.OnHTML("#divResultContent table tbody tr", func(e *colly.HTMLElement) {
		reward := Reward{}
		reward.LotteryType = lotteryType
		e.ForEach("td", func(index int, el *colly.HTMLElement) {
			switch index {
			case 0:
				reward.DateOpen = el.Text
			case 1:
				reward.Code = el.Text
			case 2:
				var pairNumbers = []string{}
				el.ForEach("span", func(index2 int, elm *colly.HTMLElement) {
					if elm.Text != "|" {
					  pairNumbers = append(pairNumbers, elm.Text)
		      }
				})
				reward.Results = append(reward.Results, strings.Join(pairNumbers, ","))
			}
		})
		allRewards = append(allRewards, reward)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit(url)
	c.Wait()
	// fmt.Print("%v", allRewards)
	return allRewards
}

func getResultsMax3D(url string, lotteryType string) []Reward {
	c := colly.NewCollector()
	allRewards := []Reward{}
	c.OnHTML(".doso_output_nd table tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(index int, el *colly.HTMLElement) {
			reward := Reward{}
			reward.LotteryType = lotteryType
			el.ForEach("a", func(_ int, el1 *colly.HTMLElement) {
				reward.Code = el1.Text
			})
			el.ForEach("div", func(index1 int, el2 *colly.HTMLElement) {
				if index1 == 0 {
					var t1 = strings.Split(el2.Text, "Ngày: ")
					reward.DateOpen = t1[1]
				}
			})
			el.ForEach(".day_so_ket_qua_v2", func(index2 int, el3 *colly.HTMLElement) {
				reward.Results = append(reward.Results, el3.Text)
			})
			allRewards = append(allRewards, reward)
		})
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit(url)
	c.Wait()
	// fmt.Print("%v", allRewards)
	return allRewards
}

func getResultsKeno(url string, lotteryType string) []Reward {
	c := colly.NewCollector()
	allRewards := []Reward{}
	c.OnHTML(".doso_output_nd table tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(index int, el *colly.HTMLElement) {
			if index > 0 {
				reward := Reward{}
				reward.LotteryType = lotteryType
				el.ForEach("td", func(index1 int, el2 *colly.HTMLElement) {
					if index1 == 0 {
						var t1 = strings.Split(el2.Text, "#")
						reward.DateOpen = t1[0]
						reward.Code = t1[1]
					} else {
						el2.ForEach(".day_so_ket_qua_v2", func(index2 int, el3 *colly.HTMLElement) {
							reward.Results = append(reward.Results, el3.Text)
						})
					}
				})
				allRewards = append(allRewards, reward)
			}
		})
	})
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.Visit(url)
	c.Wait()
	// fmt.Print("%v", allRewards)
	return allRewards
}

func getResults(lotteryType string, url string) []Reward {
	res := []Reward{}
	switch lotteryType {
	case "MEGA645", "POWER655":
		res = getResultsMega645(url, lotteryType)
	case "MAX3D", "MAX4D":
		res = getResultsMax3D(url, lotteryType)
	case "KENO":
		res = getResultsKeno(url, lotteryType)
	}
	return res
}

func getResultsRoute(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	lotteryType := query["type"][0]
	var url = os.Getenv(strings.ToUpper(lotteryType))
	result := ResponseResult{}
	result.URL = url
	result.Status = "OK"
	result.Data = getResults(strings.ToUpper(lotteryType), url)
	json.NewEncoder(w).Encode(result)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/getResults", getResultsRoute).Methods("GET")
	fmt.Println("Server running on port " + port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
