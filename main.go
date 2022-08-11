package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/line/line-bot-sdk-go/linebot"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
)

func ConnectDB() {
	databaseURL := "mongodb://localhost:27017"
	clientOptions := options.Client().ApplyURI(databaseURL)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	database := client.Database("mymongodb")
	collection = database.Collection("mycollection")
	fmt.Println("Connected to MongoDB!")
}

type User struct {
	Name    string `bson:"name"`
	Message string `bson:"message"`
}

func insertOne(u User) {
	insertResult, err := collection.InsertOne(context.TODO(), u)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
}

var filter = bson.D{{"name", "kai-@"}}

func FindAll() ([]User, error) {
	var result User
	var results []User
	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		defer cursor.Close(context.TODO())
		log.Fatal(err)
	}

	for cursor.Next(context.TODO()) {
		err := cursor.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, result)
	}
	fmt.Printf("Found document: %+v\n", results)
	return results, nil
}

var user = User{
	Name:    "John",
	Message: "563",
}

func main() {
	ConnectDB()

	// FindAll()
	// 建立客戶端
	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_ACCESS_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}
	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		defer req.Body.Close()
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Print(err)
		}
		decoded, err := base64.StdEncoding.DecodeString(req.Header.Get("x-line-signature"))
		fmt.Println(decoded)
		if err != nil {
			log.Print(err)
		}
		hash := hmac.New(sha256.New, []byte(os.Getenv("CHANNEL_SECRET")))
		hash.Write(body)
		// Compare decoded signature and `hash.Sum(nil)` by using `hmac.Equal`
		hmac.Equal(decoded, hash.Sum(nil))
		bot, err := linebot.New(
			os.Getenv("CHANNEL_SECRET"),
			os.Getenv("CHANNEL_ACCESS_TOKEN"),
		)
		if err != nil {
			log.Print(err)
		}
		res, err := bot.GetProfile("Udbb56cfb66612054e7a42f715d939e78").Do()
		if err != nil {
			log.Print(err)
		}
		println(res.DisplayName)
		println(res.PictureURL)
		println(res.StatusMessage)
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					Userinfo := User{res.DisplayName, message.Text}
					insertOne(Userinfo)
					fmt.Println(event.ReplyToken)
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				case *linebot.StickerMessage:
					replyMessage := fmt.Sprintf(
						"sticker id is %s, stickerResourceType is %s", message.StickerID, message.StickerResourceType)
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	})

	http.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		b := r.Form["id"]
		filter := bson.D{{"name", b[0]}}
		fmt.Println(filter)
		var result User
		var results []User
		cursor, err := collection.Find(context.TODO(), filter)
		if err != nil {
			defer cursor.Close(context.TODO())
			log.Fatal(err)
		}
		for cursor.Next(context.TODO()) {
			err := cursor.Decode(&result)
			if err != nil {
				log.Fatal(err)
			}
			results = append(results, result)
		}
		fmt.Printf("Found document: %+v\n", results)
		fmt.Fprintln(w, results)
		// return results, nil
	})

	// This is just sample code.
	// For actual use, you must support HTTPS by using `ListenAndServeTLS`, a reverse proxy or something else.
	if err := http.ListenAndServe(":"+os.Getenv("PORT"), nil); err != nil {
		log.Fatal(err)
	}
	// router := routes.NewRouter()
	// http.ListenAndServe(":3000", router)
}
