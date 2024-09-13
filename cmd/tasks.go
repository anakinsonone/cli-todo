package cmd

import (
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	file                    string = "tasks.db"
	INSERT_QUERY            string = "INSERT INTO tasks (task) VALUE (?)"
	SELECT_INCOMPLETE_QUERY string = "SELECT id, task, created FROM tasks WHERE done = false"
	SELECT_ALL_QUERY        string = "SELECT * FROM tasks"
	MARK_COMLETE_QUERY      string = "UPDATE tasks SET done = true WHERE id = ?"
)

func getWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.TabIndent)
}

func printTasks(w *tabwriter.Writer, showCompletion bool) error {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return fmt.Errorf("Error opening database: %w\n", err)
	}
	defer db.Close()

	query := SELECT_INCOMPLETE_QUERY
	if showCompletion {
		query = SELECT_ALL_QUERY
	}

	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("Error querying database: %w\n", err)
	}
	defer rows.Close()

	if showCompletion {
		fmt.Fprintf(w, "ID\tTask\tCreated\tDone\n")
	} else {
		fmt.Fprintf(w, "ID\tTask\tCreated\n")
	}

	for rows.Next() {
		var (
			id      int
			task    string
			created time.Time
			done    sql.NullBool
		)

		if showCompletion {
			err = rows.Scan(&id, &task, &created, &done)
		} else {
			err = rows.Scan(&id, &task, &created)
		}
		if err != nil {
			return fmt.Errorf("Error scanning row: %w\n", err)
		}

		age := time.Since(created).Round(time.Minute)

		if showCompletion {
			completed := "No"
			if done.Valid && done.Bool {
				completed = "Yes"
			}
			fmt.Fprintf(w, "%d\t%s\t%v\t%s\n", id, task, age, completed)
		} else {
			fmt.Fprintf(w, "%d\t%s\t%v\n", id, task, age)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("Error iterating rows: %w\n", err)
	}

	return nil
}

func List(showCompletion bool) error {
	w := getWriter()
	defer w.Flush()

	return printTasks(w, showCompletion)
}

func Add(task string) error {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return fmt.Errorf("Error opening database: %w\n", err)
	}
	defer db.Close()

	_, err = db.Exec(INSERT_QUERY, task)
	if err != nil {
		return fmt.Errorf("Error inserting todo: %w\n", err)
	}

	w := getWriter()
	defer w.Flush()

	return printTasks(w, false)
}

func Complete(id int) error {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return fmt.Errorf("Error opening db: %w\n", err)
	}
	defer db.Close()

	result, err := db.Exec(MARK_COMLETE_QUERY, id)
	if err != nil {
		return fmt.Errorf("Error updating row: %w\n", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Error getting affected rows: %w\n", err)
	}

	if affectedRows == 0 {
		return fmt.Errorf("No task found with id %d\n", id)
	}

	fmt.Printf("Marked task %d as completed.", id)
	return nil
}

func Delete(id int) error {
	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return fmt.Errorf("Error opening db: %w\n", err)
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("Error deleting from db: %w\n", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Error getting affected rows: %w\n", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("No task found with id %d\n", id)
	}

	fmt.Printf("Deleted task %d\n", id)
	return nil
}
