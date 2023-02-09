package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	Image []byte `json:"image"`
	Text  string `json:"text"`
}

func storeDataHandler(c *fiber.Ctx) error {
	var store Storage
	if err := c.BodyParser(&store); err != nil {
		return c.Status(500).SendString(err.Error())
	}

	client, err := mongo.Connect(c.Context(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	defer client.Disconnect(c.Context())

	collection := client.Database("WebStorageApp").Collection("images")
	_, err = collection.InsertOne(c.Context(), bson.M{
		"image": store.Image,
		"text":  store.Text,
	})

	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	return c.Status(200).SendString("Data stored successfully")
}

func retrieveDataHandler(c *fiber.Ctx) error {
	client, err := mongo.Connect(c.Context(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return c.Status(500).SendString("Failed to connect to MongoDB: " + err.Error())
	}

	defer client.Disconnect(c.Context())

	collection := client.Database("WebStorageApp").Collection("images")
	cur, err := collection.Find(c.Context(), bson.M{})
	if err != nil {
		return c.Status(500).SendString("Failed to retrieve data: " + err.Error())
	}

	defer cur.Close(c.Context())

	var images []bson.M
	if err := cur.All(c.Context(), &images); err != nil {
		return c.Status(500).SendString("Failed to retrieve data: " + err.Error())
	}

	if len(images) == 0 {
		return c.Status(404).SendString("No data found in the database")
	}

	message := "Data from MongoDB\n"
	for i, image := range images {
		message += fmt.Sprintf("Image %d: %s\nText: %s\n", i+1, string(image["image"].([]byte)), image["text"].(string))
	}

	return c.SendString(message)
}

func main() {
	app := fiber.New()

	//serve frontend files
	app.Static("/", "./views")

	//Set up Routing group api
	api := app.Group("/api")
	{
		api.Get("/", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"message": "pong",
			})
		})
	}

	//this api will consists of two routes:
	// /retrieveData - which will retrieve the data stored in mongoDB database
	// /storeData - which will store the data in mongoDB storage
	api.Get("/print", retrieveDataHandler)
	api.Post("/store", storeDataHandler)

	log.Fatal(app.Listen(":5060"))

}
