package main

import (
	"encoding/json"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"github.com/gorilla/websocket"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "9907"
	dbname   = "postgres"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Adjust this to a more secure check as needed
	},
}

func handleWebSocket(c *gin.Context) {
	w := c.Writer
	r := c.Request

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to set websocket upgrade:", err)
		return
	}
	defer func() {
		conn.Close()
		delete(clients, conn)
	}()
	
	clients[conn] = true
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}
		fmt.Printf("Received: %s\n", p)

		// Echo the message back or handle it as needed
		if err := conn.WriteMessage(messageType, p); err != nil {
			fmt.Println("Write error:", err)
			break
		}
		handleMessage(p, conn)
	}
	delete(clients, conn)
}

type WebSocketMessage struct {
    Action string          `json:"action"`
}

type EditMessage struct {
	Action string			`json:"action"`
	ID	   int				`json:"id"`
	Text   string			`json:"text"`
}

type AddMessage struct {
	Action string			`json:"action"`
	Text   string			`json:"text"`
}

type DeleteMessage struct {
	Action string			`json:"action"`
	ID     int				`json:"id"`
}

type CompleteMessage struct {
	Action string			`json:"action"`
	ID	   int				`json:"id"`
}

func handleMessage(message []byte, conn *websocket.Conn) {
    var msg WebSocketMessage
    err := json.Unmarshal(message, &msg)
    if err != nil {
        fmt.Println("Error unmarshalling message:", err)
        return
    }

    switch msg.Action {
    case "delete":
        var deleteMsg DeleteMessage
        err := json.Unmarshal(message, &deleteMsg)
        if err != nil {
            fmt.Println("Error unmarshalling delete message:", err)
            return
        }
        deleteTodo(deleteMsg.ID, conn)

    case "edit":
        var editMsg EditMessage
        err := json.Unmarshal(message, &editMsg)
        if err != nil {
            fmt.Println("Error unmarshalling edit message:", err)
            return
        }
        editTodo(editMsg.ID, editMsg.Text, conn)

    case "add":
        var addMsg AddMessage
        err := json.Unmarshal(message, &addMsg)
        if err != nil {
            fmt.Println("Error unmarshalling add message:", err)
            return
        }
        addTodo(addMsg.Text, conn)

	case "complete":
		var completeMsg CompleteMessage
		err := json.Unmarshal(message, &completeMsg)
        if err != nil {
            fmt.Println("Error unmarshalling add message:", err)
            return
        }
        complete(completeMsg.ID, conn)
	}
}

func complete(id int, conn *websocket.Conn) {

    // Check the current complete state of the to-do item
    var currentCompleteState bool
    err := db.QueryRow("SELECT complete FROM todos WHERE id = $1", id).Scan(&currentCompleteState)
    if err != nil {
        fmt.Println("Error querying todo:", err)
        return
    }

    // Toggle the complete state
    newCompleteState := !currentCompleteState

    // Update the complete state in the database
    _, err = db.Exec("UPDATE todos SET complete = $1 WHERE id = $2", newCompleteState, id)
    if err != nil {
        fmt.Println("Error updating todo:", err)
        return
    }

    fmt.Println("Todo complete state toggled successfully")

    // Broadcast the updated list of to-do items to all connected clients
    broadcastTodos()
}

func deleteTodo(id int, conn *websocket.Conn) {
	_, err := db.Exec("DELETE FROM todos WHERE id = $1", id)
    if err != nil {
        fmt.Println("Error deleting todo:", err)
        return
    }
    fmt.Println("Todo deleted successfully")
	broadcastTodos()
}

func editTodo(id int, text string, conn *websocket.Conn) {
    // Assuming `db` is your database connection
    _, err := db.Exec("UPDATE todos SET text = $1, complete = $2 WHERE id = $3", text, false, id)
    if err != nil {
        fmt.Println("Error updating todo:", err)
        return
    }
    fmt.Println("Todo updated successfully")
	broadcastTodos()
}

func addTodo(message string, conn *websocket.Conn) {
	_, err := db.Exec("INSERT INTO todos (text, complete) VALUES ($1, $2)", message, false)
	if err != nil {
        fmt.Println("Error adding todo:", err)
        return
    }
    fmt.Println("Todo added successfully")
	broadcastTodos()
}

var clients = make(map[*websocket.Conn]bool) // Keep track of connected clients

func broadcastTodos() {
    todos, err := getTodosFromDatabase() // Implement this function to fetch todos
    if err != nil {
        fmt.Printf("Error fetching todos: %v", err)
        return
    }

    jsonTodos, err := json.Marshal(todos)
    if err != nil {
        fmt.Printf("Error marshalling todos: %v", err)
        return
    }

    for client := range clients {
		fmt.Println("Broadcasting todos")
        if err := client.WriteMessage(websocket.TextMessage, jsonTodos); err != nil {
            fmt.Printf("Error sending todos: %v", err)
            client.Close()
            delete(clients, client)
        }
    }
}

func getTodosFromDatabase() ([]Todo, error) {
    var todos []Todo
    rows, err := db.Query("SELECT id, text, complete FROM todos")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    for rows.Next() {
        var todo Todo
        if err := rows.Scan(&todo.ID, &todo.Text, &todo.Complete); err != nil {
            return nil, err
        }
        todos = append(todos, todo)
    }
	fmt.Println("Preparing to broadcast todos")
    return todos, nil
}


var db *sql.DB

type Todo struct {
	ID      int    `json:"id"`
	Text    string `json:"text"`
	Complete bool  `json:"complete"`
}	

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	  "password=%s dbname=%s sslmode=disable",
	  host, port, user, password, dbname)	
	var err error
	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
	  panic(err)
	}
	fmt.Println("Established a successful connection!")

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.Use(cors.Default())

	router.GET("/ws", handleWebSocket)

	router.GET("/todos", func(c *gin.Context) {
		var todos []Todo
	
		rows, err := db.Query("SELECT id, text, complete FROM todos")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()
	
		for rows.Next() {
			var todo Todo
			if err := rows.Scan(&todo.ID, &todo.Text, &todo.Complete); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			todos = append(todos, todo)
		}
	
		c.JSON(http.StatusOK, todos)
	})
	

	router.DELETE("/todos/:id", func(c *gin.Context) {
		id := c.Param("id")
	
		_, err := db.Exec("DELETE FROM todos WHERE id = $1", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	
		c.JSON(http.StatusOK, gin.H{"message": "Todo deleted successfully"})
	})
	
	router.POST("/todos", func(c *gin.Context) {
		var newTodo Todo
	
		// Bind the incoming JSON to newTodo
		if err := c.BindJSON(&newTodo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	
		// Insert newTodo into the database
		sqlStatement := `INSERT INTO todos (text, complete) VALUES ($1, $2) RETURNING id`
		id := 0
		err := db.QueryRow(sqlStatement, newTodo.Text, newTodo.Complete).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	
		newTodo.ID = id
		c.JSON(http.StatusCreated, newTodo)
	})

	router.PATCH("/todos/:id", func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr) // Convert string to int
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
			return
		}
	
		var updatedTodo Todo
		if err := c.BindJSON(&updatedTodo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	
		_, err = db.Exec("UPDATE todos SET text = $1, complete = $2 WHERE id = $3", updatedTodo.Text, updatedTodo.Complete, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	
		updatedTodo.ID = id
		c.JSON(http.StatusOK, updatedTodo)
	})
	
	// Paths to your SSL certificate and private key
	certPath := "/etc/letsencrypt/live/hamburgler.xyz/fullchain.pem" // e.g., /etc/letsencrypt/live/yourdomain.com/fullchain.pem
	keyPath := "/etc/letsencrypt/live/hamburgler.xyz/privkey.pem"   // e.g., /etc/letsencrypt/live/yourdomain.com/privkey.pem
    http.ListenAndServeTLS(":8081", certPath, keyPath, router) // Run the server on port 8081
}
