package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
	
	"context"
	"log"
)

var collection *mongo.Collection
var ctx = context.TODO()

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	collection = client.Database("curso_go").Collection("movies")
}


var movies = Movies{
	Movie{"Sin limites", 2013, "Desconocido"},
	Movie{"Batman", 1999, "Scorsese"},
	Movie{"Rapido y furioso", 2005, "Juan Antonio"},
}

func responseMovie(w http.ResponseWriter, status int, results Movie){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(results)
}

func responseMovies(w http.ResponseWriter, status int, results []Movie){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(results)
}

func Index(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w,"Hola mundo desde mi servidor web Go")
}

func MovieList(w http.ResponseWriter, r *http.Request){
	//fmt.Fprintf(w,"Listado de Peliculas")
	options := options.Find().SetSort(bson.D{{"_id", -1}})
	//cursor, err := collection.Find(context.TODO(), bson.D{}).SetSort(bson.D{{"_id", -1}})
	filter := bson.D{}
	cursor, err := collection.Find(ctx, filter, options)
	if err != nil {
        panic(err)
	}
	var results []Movie
	if err = cursor.All(context.TODO(), &results); err != nil {
        panic(err)
	}
	
	responseMovies(w, 200, results)
}

func MovieShow(w http.ResponseWriter, r *http.Request){
	params := mux.Vars(r)
	movie_id := params["id"]

	objectID, err := primitive.ObjectIDFromHex(movie_id)
	if err != nil {
		log.Fatal(err)
	}
	filter := bson.M{"_id": objectID}

	var results Movie
	err = collection.FindOne(ctx, filter).Decode(&results)

	if err != nil{
		w.WriteHeader(404)
		return
	}
	responseMovie(w, 200, results)
	//fmt.Fprintf(w,"Has cargado la pelicula numero %s", movie_id)
}


func MovieUpdate(w http.ResponseWriter, r *http.Request) {
    params := mux.Vars(r)
    movieID := params["id"]

    objectID, err := primitive.ObjectIDFromHex(movieID)
    if err != nil {
        http.Error(w, "Invalid movie ID", http.StatusBadRequest)
        return
    }

    decoder := json.NewDecoder(r.Body)
    var movieData Movie
    err = decoder.Decode(&movieData)
    if err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    filter := bson.M{"_id": objectID}
    update := bson.M{"$set": movieData}

    result, err := collection.UpdateOne(context.TODO(), filter, update)
    if err != nil {
        http.Error(w, "Failed to update movie", http.StatusInternalServerError)
        return
    }

    if result.ModifiedCount == 0 {
        http.Error(w, "Movie not found", http.StatusNotFound)
        return
    }

    responseMovie(w, 200, movieData)
}

func MovieAdd(w http.ResponseWriter, r *http.Request){
	decoder := json.NewDecoder(r.Body)
	
	var movie_data Movie
	err := decoder.Decode(&movie_data)

	if(err != nil){
		panic(err)
	}

	defer r.Body.Close()
	
	res, err := collection.InsertOne(context.Background(), movie_data)
	fmt.Println(res)
	if err != nil{
		w.WriteHeader(500)
		return
	} 
	json.NewEncoder(w).Encode(movie_data)

	responseMovie(w, 200, movie_data)
}

type Message struct{
	Status string `json:"status"`
	Message string `json:"message"`
}

// setStatus sets the status field of the Message struct
func (m *Message) setStatus(data string) {
	m.Status = data
}

// setMessage sets the message field of the Message struct
func (m *Message) setMessage(data string) {
	m.Message = data
}

func MovieRemove(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	movieID := params["id"]

	objectID, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		log.Fatal(err)
	}

	filter := bson.M{"_id": objectID}

	// Use DeleteOne instead of RemoveId
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	if result.DeletedCount == 0 {
		// No documents matched the filter, indicating that the movie with the given ID was not found
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Create a new Message instance
	message := &Message{}
	// Set the status and message using the methods
	message.setStatus("success")
	message.setMessage("La película con ID " + movieID + " ha sido borrada correctamente")

	// Return the message in the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(message)
}