package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Album struct {
	ID     int64
	Title  string
	Artist string
	Price  float32
}

var db_pool *pgxpool.Pool

func main() {
	// urlExample := "postgres://username:password@localhost:5432/database_name"
	db_pool = initDbPool()
	if ping_err := db_pool.Ping(context.Background()); ping_err != nil {
		log.Fatal(ping_err)
	}
	defer db_pool.Close()

	var albums []Album

	// var title string
	// var artist string
	rows, err := db_pool.Query(context.Background(), "select * from album")
	if err != nil {
		log.Fatalf("QueryRow failed: %v\n", err)
	}
	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		}
		albums = append(albums, alb)
	}
	fmt.Println(albums)
	albums, err = albumsByArtist("John Coltrane")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Albums found: %v\n", albums)

	// Hard-code ID 2 here to test the query.
	alb, err := albumByID(2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Album found: %v\n", alb)

	affected, err := addAlbum(Album{
		Title:  "The Modern Sound of Betty Carter",
		Artist: "Betty Carter",
		Price:  49.99,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Rows Affected: %v\n", affected)
}

// albumsByArtist queries for albums that have the specified artist name.
func albumsByArtist(name string) ([]Album, error) {
	var albums []Album

	rows, err := db_pool.Query(context.Background(), "select * from album where artist = $1", name)
	if err != nil {
		return nil, fmt.Errorf("Album by artist %v query failed: %v\n", name, err)
	}
	defer rows.Close()
	for rows.Next() {
		var alb Album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			return nil, fmt.Errorf("Album by artist %v query row failed: %v\n", name, err)

		}
		albums = append(albums, alb)
	}
	return albums, nil
}

// albumByID queries for the album with the specified ID.
func albumByID(id int64) (Album, error) {
	// An album to hold data from the returned row.
	var alb Album

	row := db_pool.QueryRow(context.Background(), "SELECT * FROM album WHERE id = $1", id)
	if err := row.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
		if err == pgx.ErrNoRows {
			return alb, fmt.Errorf("albumsById %d: no such album", id)
		}
		return alb, fmt.Errorf("albumsById %d: %v", id, err)
	}
	return alb, nil
}

// addAlbum adds the specified album to the database
func addAlbum(alb Album) (int64, error) {
	result, err := db_pool.Exec(context.Background(), "INSERT INTO album (title, artist, price) VALUES ($1, $2, $3)", alb.Title, alb.Artist, alb.Price)
	if err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}
	affected := result.RowsAffected()
	if affected == 0 {
		return 0, fmt.Errorf("addAlbum: failed")
	}
	return affected, nil
}

func initDbPool() *pgxpool.Pool {
	db_pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\nDB_URL: %v", err, os.Getenv("DATABASE_URL"))
	}
	return db_pool
}
