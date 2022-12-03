/*
.
.
SERVER CODE
.
.
*/
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var err error

type Album struct {
	ID    int64
	Title string
}

var conn_backup net.Conn

func main() {

	dp_connection()
	conn_backup, _ = net.Dial("tcp", "192.168.43.11:8000")
	_, _ = conn_backup.Write([]byte("hello backup"))

	fmt.Println("server listening on 3000")
	listener, err := net.Listen("tcp", "0.0.0.0:3000")
	if err != nil {
		log.Fatalln(err)
	}
	defer listener.Close()
	// listening for incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		buffer := make([]byte, 1400)
		dataSize, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("connection closed" + err.Error())
			return
		}
		// the actual message
		data := buffer[:dataSize]
		fmt.Println("received message: ", string(data))
		// listen to connections in another gorutine
		go listenConnection(conn, db)
	}
}

func dp_connection() {
	dpDriver := "mysql"
	dpUser := "root"
	dpPass := ""
	dpName := "serverdb"

	db, err = sql.Open(dpDriver, dpUser+":"+dpPass+"@tcp(127.0.0.1:3306)/"+dpName)
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()
	fmt.Println("succussfuly connect")
}

// listening for messages from connection
func listenConnection(conn net.Conn, db *sql.DB) {
	for {
		choose, err := read_fromClient(conn)
		write_to_backup(conn_backup, choose)
		if err != nil {
			fmt.Println("Connection closed..!" + err.Error())
			return
		} else if choose == "1" {
			_Title, _ := read_fromClient(conn)
			albID, err := addAlbum(db, Album{
				Title: _Title,
			})
			fmt.Printf("ID of added album: %v \n", albID)
			if err != nil {
				panic(err.Error())
			}
			fmt.Println("successfully insert  sir ")
			write_to_backup(conn_backup, _Title)

		} else if choose == "2" {
			_id, _ := read_fromClient(conn)
			id, _ := strconv.Atoi(_id)
			Delete_Album(id)
			write_to_backup(conn_backup, _id)

		} else if choose == "3" {
			_id, _ := read_fromClient(conn)
			_Title, _ := read_fromClient(conn)
			id, _ := strconv.Atoi(_id)
			Update_Album(id, _Title)
			write_to_backup(conn_backup, _id)
			write_to_backup(conn_backup, _Title)

		}
	}
}

func read_fromClient(conn net.Conn) (string, error) {
	buffer_command := make([]byte, 1400)
	comm_Size, err := conn.Read(buffer_command)

	if err != nil {
		// fmt.Println("Connection closed")
		return "", err
	}
	command := string(buffer_command[:comm_Size])
	fmt.Println("read " + command)
	return command, nil
}
func write_to_backup(conn net.Conn, str string) error {
	_, err := conn.Write([]byte(str))

	if err != nil {
		// fmt.Println("Connection closed")
		return err
	}

	return nil
}

func addAlbum(db *sql.DB, alb Album) (int64, error) {
	result, err := db.Exec("INSERT INTO album (title) VALUES (?)", alb.Title)
	if err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("addAlbum: %v", err)
	}
	return id, nil
}

func GetAllAlbums() []Album {
	row, err := db.Query("select * from album")
	if err != nil {
		log.Fatal(err)
	}
	alb := Album{}      //Create instanc from Album
	albums := []Album{} //Create Array of Album
	for row.Next() {
		err := row.Scan(&alb.ID, &alb.Title)
		if err != nil {
			log.Fatal(err)
		}
		albums = append(albums, alb)
	}
	return albums
}

func Update_Album(id int, title string) {
	statment, err := db.Prepare("update serverdb.album set Title=? where ID=?")
	if err != nil {
		log.Fatal(err)
	}
	r, err := statment.Exec(title, id)
	if err != nil {
		log.Fatal(err)
	}
	affectedRow, err := r.RowsAffected()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("Query OK, %d rows affected.\n", affectedRow)
		fmt.Println("Rows matched: 1  Changed: 1  Warnings: 0")
	}
}

func Delete_Album(id int) {
	statment, err := db.Prepare("delete from serverdb.album where ID=?")
	if err != nil {
		log.Fatal(err)
	}
	r, err := statment.Exec(id)
	if err != nil {
		log.Fatal(err)
	}
	affectedRow, err := r.RowsAffected()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("Query OK, %d rows affected.\n", affectedRow)
		fmt.Println("Rows matched: 1  Changed: 1  Warnings: 0")

	}
}
