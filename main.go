package main

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "strconv"

    _ "github.com/mattn/go-sqlite3"
)

// Task структура для хранения данных задачи
type Task struct {
    ID          int    `json:"id"`
    Description string `json:"description"`
    Completed   bool   `json:"completed"`
}

var db *sql.DB

func main() {
    var err error
    db, err = sql.Open("sqlite3", "./tasks.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            description TEXT,
            completed BOOLEAN
        )
    `)
    if err != nil {
        log.Fatal(err)
    }

    http.HandleFunc("/tasks", handleTasks)
    http.HandleFunc("/tasks/", handleTask)

    log.Println("Server is running on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleTasks(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case "GET":
        getTasks(w, r)
    case "POST":
        createTask(w, r)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func handleTask(w http.ResponseWriter, r *http.Request) {
    id, err := strconv.Atoi(r.URL.Path[len("/tasks/"):])
    if err != nil {
        http.Error(w, "Invalid task ID", http.StatusBadRequest)
        return
    }

    switch r.Method {
    case "GET":
        getTask(w, r, id)
    case "PUT":
        updateTask(w, r, id)
    case "DELETE":
        deleteTask(w, r, id)
    default:
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func getTasks(w http.ResponseWriter, r *http.Request) {
    rows, err := db.Query("SELECT id, description, completed FROM tasks")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var tasks []Task
    for rows.Next() {
        var task Task
        if err := rows.Scan(&task.ID, &task.Description, &task.Completed); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        tasks = append(tasks, task)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(tasks)
}

func createTask(w http.ResponseWriter, r *http.Request) {
    var task Task
    if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    result, err := db.Exec("INSERT INTO tasks (description, completed) VALUES (?, ?)", task.Description, task.Completed)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    id, _ := result.LastInsertId()
    task.ID = int(id)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(task)
}
func getTask(w http.ResponseWriter, r *http.Request, id int) {
    var task Task
    err := db.QueryRow("SELECT id, description, completed FROM tasks WHERE id = ?", id).Scan(&task.ID, &task.Description, &task.Completed)
    if err != nil {
        if err == sql.ErrNoRows {
            http.Error(w, "Task not found", http.StatusNotFound)
        } else {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        }
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(task)
}

func updateTask(w http.ResponseWriter, r *http.Request, id int) {
    var task Task
    if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    _, err := db.Exec("UPDATE tasks SET description = ?, completed = ? WHERE id = ?", task.Description, task.Completed, id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    task.ID = id
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(task)
}

func deleteTask(w http.ResponseWriter, r *http.Request, id int) {
    _, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}
