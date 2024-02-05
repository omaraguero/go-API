package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const uri = "mongodb+srv://Omar:dbOmar1@cluster0.xhwlox6.mongodb.net/?retryWrites=true&w=majority"

var mongoClient *mongo.Client

func init() {
	if err := connect_to_mongodb(); err != nil {
		log.Fatal("No se pudo conectar a MongoDB")
	}
	println("Conexion Establecida")
}

func main() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Pedro es un buen chico!",
		})
	})

	r.GET("/books", getBooks)
	r.GET("/books/:id", getBookByID)
	r.POST("/books/aggregate", aggregateBooks)
	r.DELETE("/books/:id", deleteBookByID)
	r.POST("/books", addBook)

	r.Run()
}

func connect_to_mongodb() error {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	err = client.Ping(context.TODO(), nil)
	mongoClient = client
	return err
}

func getBooks(c *gin.Context) {
	cursor, err := mongoClient.Database("bookshop").Collection("books").Find(context.TODO(), bson.D{{}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var books []bson.M
	if err = cursor.All(context.TODO(), &books); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error})
		return
	}

	c.JSON(http.StatusOK, books)
}

func getBookByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := primitive.ObjectIDFromHex(idStr)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var book bson.M
	err = mongoClient.Database("bookshop").Collection("books").FindOne(context.TODO(), bson.D{{"_id", id}}).Decode(&book)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, book)

}

func aggregateBooks(c *gin.Context) {
	var pipeline interface{}
	if err := c.ShouldBindJSON(&pipeline); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cursor, err := mongoClient.Database("bookshop").Collection("books").Aggregate(context.TODO(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var result []bson.M
	if err = cursor.All(context.TODO(), &result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)

}

func deleteBookByID(c *gin.Context) {
	idStr := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := mongoClient.Database("bookshop").Collection("books").DeleteOne(context.TODO(), bson.D{{"_id", id}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "Libro No Encontrado"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Libro eliminado exitosamente"})
}

func addBook(c *gin.Context) {
	var pipeline interface{}
	if err := c.ShouldBindJSON(&pipeline); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := mongoClient.Database("bookshop").Collection("books").InsertOne(context.TODO(), pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	insertedID := result.InsertedID

	c.JSON(http.StatusOK, gin.H{"message": "Libro agregado exitosamente", "insertedID": insertedID})
}
