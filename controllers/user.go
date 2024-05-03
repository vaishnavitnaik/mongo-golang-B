package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/vaishnavitnaik/mongo-golang/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	client *mongo.Client
}

func NewUserController(client *mongo.Client) *UserController {
	return &UserController{client}
}

func (uc UserController) GetUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	// Convert the ID to an ObjectID
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Define a filter to find the user by ID
	filter := bson.M{"_id": oid}

	// Find the user by ID
	var user models.User
	err = uc.client.Database("mongo-golang").Collection("users").FindOne(context.Background(), filter).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Marshal the user object to JSON
	userJSON, err := json.Marshal(user)
	if err != nil {
		log.Println("Error marshaling user to JSON:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set response headers and write the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(userJSON)
}

func (uc UserController) CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user models.User

	// Decode JSON request body into User struct
	// user.Id = bson.ObjectIDFromHex()
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Println("Error decoding JSON request:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Insert the user into the database
	insertResult, err := uc.client.Database("mongo-golang").Collection("users").InsertOne(context.Background(), user)
	if err != nil {
		log.Println("Error inserting user into database:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get the inserted ID from the result
	insertedID := insertResult.InsertedID.(primitive.ObjectID)

	// Set response headers and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, `{"_id": "%s"}`, insertedID.Hex())
}

func (uc UserController) DeleteUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	id := p.ByName("id")

	// Convert the ID to an ObjectID
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Define a filter to find the user by ID
	filter := bson.M{"_id": oid}

	// Delete the user from the database
	_, err = uc.client.Database("mongo-golang").Collection("users").DeleteOne(context.Background(), filter)
	if err != nil {
		log.Println("Error deleting user from database:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Write the success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Deleted user with ID %s", id)
}
