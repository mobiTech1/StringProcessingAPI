package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	pusher "github.com/pusher/pusher-http-go/v5"
)

type requestBody struct {
	Lang string `json:"lang"`
	Data string `json:"data"`
}
type Response struct {
	Message string
}

func main() {
	app := fiber.New()
	// Default config
	app.Use(cors.New())
	// Or extend your config for customization
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Post("/data", func(c *fiber.Ctx) error {
		var reqBody requestBody
		err := json.Unmarshal(c.Body(), &reqBody)
		if err != nil {
			// handle error
			fmt.Println(err)
		}
		var codeFile string = reqBody.Data
		var langCode string = reqBody.Lang
		response := Response{
			Message: "data Recieved Successfully:" + codeFile + langCode,
		}
		return c.JSON(response)
	})
	app.Post("/", func(c *fiber.Ctx) error {
		Client := pusher.Client{
			AppID:   "1568045",
			Key:     "51f659ce3f43900892ff",
			Secret:  "2693c09337092248c022",
			Cluster: "eu",
			Secure:  true,
		}
		var reqBody requestBody
		err := json.Unmarshal(c.Body(), &reqBody)
		if err != nil {
			// handle error
			fmt.Println(err)
		}
		var codeFile string = reqBody.Data
		var langCode string = reqBody.Lang
		resultChan := make(chan struct{})
		for line_num, line := range strings.Split(codeFile, "\n") {
			// if this line has fewer characters then skip it
			if len(line) < 3 {
				sendEventData(Client, line, line_num)
				continue
			}
			// using concurrency to process each line for better performance
			go func(line string, num int) {
				line, err := postRequest(line, langCode)
				if err != nil {
					fmt.Println(err)
				}
				//fmt.Println(num+1, line)
				sendEventData(Client, line, num)
				resultChan <- struct{}{}
			}(line, line_num)
		}
		for i := 0; i < len(strings.Split(codeFile, "\n"))-1; i++ {
			<-resultChan
		}
		response := Response{
			Message: "data Recieved Successfully",
		}
		return c.JSON(response)
	})

	var port string = os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	//app.Listen(":3000")
	log.Fatal(app.Listen("0.0.0.0:" + port))
}

func sendEventData(Client pusher.Client, line string, line_number int) {
	data := map[string]string{
		"number": strconv.Itoa(line_number),
		"line":   line,
	}
	err := Client.Trigger("my-channel", "my-event", data)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func postRequest(text string, LangCode string) (string, error) {
	authKey := "e5ab02b3-3e3d-bfaa-acc8-bc4f34c70884:fx"
	link := "https://api-free.deepl.com/v2/translate"
	targetLang := LangCode

	data := url.Values{}
	data.Set("text", text)
	data.Set("target_lang", targetLang)

	req, err := http.NewRequest("POST", link, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", authKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", err
	}
	if result["translations"] == nil {
		return "", fmt.Errorf("no translations found")
	}
	translations := result["translations"].([]interface{})
	translation := translations[0].(map[string]interface{})
	return translation["text"].(string), nil
}

// Client := pusher.Client{
// 	AppID:   os.Getenv("APP_ID"),
// 	Key:     os.Getenv("APP_KEY"),
// 	Secret:  os.Getenv("APP_SECRET"),
// 	Cluster: os.Getenv("APP_CLUSTER"),
// }
