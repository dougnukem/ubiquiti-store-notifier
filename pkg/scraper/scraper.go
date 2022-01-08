package scraper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

type UbiquitiCredentials struct {
	Username string
	Password string
}

type TokenPayload struct {
	Token string `json:"token"`
}

type Product struct {
	Name      string  `db:"name"`
	Price     float64 `db:"price"`
	Available bool    `db:"available"`
	Link      string  `db:"link"`
}

type Products []Product

func GetProducts(config UbiquitiCredentials) (Products, error) {
	c := colly.NewCollector()
	err := login(c, config)
	if err != nil {
		return nil, err
	}

	products := getProducts(c)
	return products, nil
}

func login(c *colly.Collector, config UbiquitiCredentials) error {
	token := TokenPayload{}

	c.OnResponse(func(r *colly.Response) {
		if r.Request.URL.String() == "https://sso.ui.com/api/sso/v1/jwt/token" {
			err := json.Unmarshal(r.Body, &token)
			if err != nil {
				log.Fatal(fmt.Errorf("Error decoding token: %v", err))
			}
		}
	})

	credentials := map[string]string{"user": config.Username, "password": config.Password}
	loginPayload, err := json.Marshal(credentials)
	if err != nil {
		return fmt.Errorf("Error creating payload for login: %v", err)
	}
	reader := bytes.NewReader(loginPayload)

	err = c.Request("POST", "httddps://sso.ui.com/api/sso/v1/login", reader, nil, http.Header{"Content-Type": []string{"application/json"}})
	if err != nil {
		return fmt.Errorf("Error logging in: %v", err)
	}

	err = c.Request("GET", "https://sso.ui.com/api/sso/v1/jwt/token", nil, nil, nil)
	if err != nil {
		return fmt.Errorf("Error getting JWT token: %v", err)
	}

	err = c.Request("GET", "https://sso.ubnt.com/api/sso/v1/jwt/token/login/"+token.Token, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("Error logging in with token: %v", err)
	}

	err = c.Request("GET", "https://sso.ubnt.com/api/sso/v1/shopify_login?region=eu", nil, nil, nil)
	if err != nil {
		return fmt.Errorf("Error logging in to shopify: %v", err)
	}

	return nil
}

func getProducts(c *colly.Collector) Products {
	products := Products{}

	c.OnHTML("body", func(body *colly.HTMLElement) {
		matches := regexp.MustCompile(`<a(.|\n)*?<\/a>`).FindAllString(body.Text, -1)
		for i := 0; i < len(matches); i++ {
			dom, err := goquery.NewDocumentFromReader(strings.NewReader(matches[i]))
			if err != nil {
				log.Println(err)
				continue
			}

			product := Product{}
			title := dom.Find(".comProductTile__title > .smaller > .link")
			price := dom.Find(".price")
			link := dom.Find(".comProductTile")

			if link.Length() < 1 && title.Length() < 1 || price.Length() < 1 {
				continue
			}

			title.Each(func(i int, s *goquery.Selection) {
				product.Name = s.Text()
			})

			if price, err := strconv.ParseFloat(price.AttrOr("data-price-cents", "0"), 64); err == nil {
				product.Price = price / 100
			}

			soldOut := dom.Find(".comProductTile__soldOut")
			product.Available = soldOut.Length() < 1
			product.Link = link.AttrOr("href", "No link found")

			products = append(products, product)
		}
	})

	c.Visit("https://eu.store.ui.com/collections/early-access")

	return products
}
