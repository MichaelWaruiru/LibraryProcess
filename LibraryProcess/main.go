package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var Pgs string

type Student struct {
	StudentID  int
	Name       string
	Department string
}

type Book struct {
	BookID   int
	Name     string
	Author   string
	Genre    string
	Category string
}

type BorrowedBook struct {
	BorrowingID   int
	BookID        int
	StudentID     int
	ReturnDate    NullTime
	ExceedingDate NullTime
	Penalties     float64
}

type Catalog struct {
	CatalogID   int
	CatalogName string
	ShelfNo     string
	NoBooks     int
	BookID      int
}

type Department struct {
	DepartmentID   int
	DepartmentName string
}

// Struct that convert dates
type NullTime struct {
	Time  time.Time
	Valid bool
}

// Converting date from strings to time.Time
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}
	if v, ok := value.([]uint8); ok {
		t, err := time.Parse("2006-01-02", string(v))
		if err != nil {
			return err
		}
		nt.Time, nt.Valid = t, true
		return nil
	}
	return fmt.Errorf("unsupported Scan: %T", value)
}
func (nt NullTime) String() string {
	if nt.Valid {
		return nt.Time.Format("2006-01-02")
	}
	return ""
}

func menu(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.URL.Path != "/menu" {
		http.NotFound(w, r)
		return
	}

	files := []string{
		"templates/menu.gohtml",
	}

	pgsMap := map[string]string{
		"3":   "3",
		"3-1": "3-1",
		"3-2": "3-2",
		"3-3": "3-3",
		"3-4": "3-4",
		"3-5": "3-5",
	}

	Pgs = "3"

	queryParamValue := r.FormValue("pgs_")
	if pgsValue, found := pgsMap[queryParamValue]; found {
		Pgs = pgsValue
	}

	switch Pgs {
	case "3":
		files = append(files, "templates/menu.gohtml")
	case "3-1":
		files = append(files, "templates/menu_catalog.gohtml")
	case "3-2":
		files = append(files, "templates/menu_books.gohtml")
	case "3-3":
		files = append(files, "templates/menu_borrowed_books.gohtml")
	case "3-4":
		files = append(files, "templates/menu_students.gohtml")
	case "3-5":
		files = append(files, "templates/menu_departments.gohtml")
	}

	// Query students data
	studentRows, err := db.Query("SELECT student_id, student_name, department FROM students")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer studentRows.Close()

	var students []Student

	for studentRows.Next() {
		var student Student
		err := studentRows.Scan(&student.StudentID, &student.Name, &student.Department)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		students = append(students, student)
	}

	// Query borrowed books data
	borrowedRows, err := db.Query("SELECT borrowing_id, student_id, book_id, return_date, exceeding_date, penalties FROM borrowed_books")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer borrowedRows.Close()

	var borrowedBooks []BorrowedBook

	for borrowedRows.Next() {
		var borrowedBook BorrowedBook

		err := borrowedRows.Scan(
			&borrowedBook.BorrowingID,
			&borrowedBook.StudentID,
			&borrowedBook.BookID,
			&borrowedBook.ReturnDate,
			&borrowedBook.ExceedingDate,
			&borrowedBook.Penalties,
		)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		borrowedBooks = append(borrowedBooks, borrowedBook)
	}

	// Query books data
	bookRows, err := db.Query("SELECT book_id, name, author, genre, category FROM books")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer bookRows.Close()

	var books []Book

	for bookRows.Next() {
		var book Book
		err := bookRows.Scan(&book.BookID, &book.Name, &book.Author, &book.Genre, &book.Category)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		books = append(books, book)
	}

	// Query catalog data
	catalogRows, err := db.Query("SELECT catalog_id, catalog_name, shelf_no, no_of_books FROM catalog")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer catalogRows.Close()

	var catalogs []Catalog

	for catalogRows.Next() {
		var catalog Catalog
		err := catalogRows.Scan(&catalog.CatalogID, &catalog.CatalogName, &catalog.ShelfNo, &catalog.NoBooks)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		catalogs = append(catalogs, catalog)
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Query departments data
	departmentRows, err := db.Query("SELECT department_id, department_name FROM department")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer departmentRows.Close()

	var departments []Department

	for departmentRows.Next() {
		var department Department
		err := departmentRows.Scan(&department.DepartmentID, &department.DepartmentName)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		departments = append(departments, department)
	}

	m := map[string]interface{}{
		"curMenu":       "3",
		"pgs_":          Pgs,
		"students":      students,
		"books":         books,
		"borrowedBooks": borrowedBooks,
		"catalog":       catalogs,
		"departments":   departments,
	}

	err = ts.ExecuteTemplate(w, "menu", m)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func addCatalog(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.URL.Path != "/addCatalog" {
		http.NotFound(w, r)
		return
	}
	// Handle form submission for adding a catalog
	if r.Method == http.MethodPost {
		catalogName := r.FormValue("catalogName")
		shelfNo := r.FormValue("shelfNo")
		noBooks := r.FormValue("noBooks")

		// Validate form data
		if catalogName == "" || shelfNo == "" || noBooks == "" {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Insert the data into the database
		_, err := db.Exec("INSERT INTO catalog (catalog_name, shelf_no, no_of_books) VALUES (?, ?, ?)", catalogName, shelfNo, noBooks)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect user back to the page
		http.Redirect(w, r, "/menu?pgs_=3-1", http.StatusSeeOther)
	}
}

func editCatalog(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		catalogID := r.URL.Path[len("/editCatalog/"):]
		catalogName := r.FormValue("catalogName")
		shelfNo := r.FormValue("shelfNo")
		noBooks := r.FormValue("noBooks")

		// Validate form data
		if catalogID == "" || catalogName == "" || shelfNo == "" || noBooks == "" {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Update the catalog data in the database
		_, err := db.Exec("UPDATE catalog SET catalog_name = ?, shelf_no = ?, no_of_books = ? WHERE catalog_id = ?", catalogName, shelfNo, noBooks, catalogID)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Return a success response
		w.Write([]byte("Catalog updated successfully"))
		return
	}
}

func deleteCatalog(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.URL.Path != "/deleteCatalog/" {
		http.NotFound(w, r)
		return
	}

	// Handle deleting a catalog
	if r.Method == http.MethodDelete {
		catalogID := r.FormValue("catalogID")
		if catalogID != "" {
			// Delete the catalog from the database using catalogID
			_, err := db.Exec("DELETE FROM catalog WHERE catalog_id = ?", catalogID)
			if err != nil {
				log.Print(err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// Return a success message or response
			w.Write([]byte("Catalog deleted successfully"))
			return
		}
	}
}

func addBook(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.URL.Path != "/addBook" {
		http.NotFound(w, r)
		return
	}
	// Handle adding book form submission
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		author := r.FormValue("author")
		genre := r.FormValue("genre")
		category := r.FormValue("category")

		// Validate form data
		if name == "" || author == "" || genre == "" || category == "" {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Insert the data into the database
		_, err := db.Exec("INSERT INTO books (name, author, genre, category) VALUES (?, ?, ?, ?)", name, author, genre, category)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect user back to the page
		http.Redirect(w, r, "/menu?pgs_=3-2", http.StatusSeeOther)
	}
}

func editBook(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		bookID := r.URL.Path[len("/editBook/"):]
		name := r.FormValue("name")
		author := r.FormValue("author")
		genre := r.FormValue("genre")
		category := r.FormValue("category")

		// Validate form data
		if bookID == "" || name == "" || author == "" || genre == "" || category == "" {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Update the book data in the database
		_, err := db.Exec("UPDATE books SET name = ?, author = ?, genre = ?, category = ? WHERE book_id = ?", name, author, genre, category, bookID)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Return a success response
		w.Write([]byte("Book updated successfully"))
		return
	}
}

func deleteBook(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.URL.Path != "/deleteBook/" {
		http.NotFound(w, r)
		return
	}

	// Handle deleting a book
	if r.Method == http.MethodDelete {
		bookID := r.FormValue("bookID")
		if bookID != "" {
			// Delete the book from the database using bookID
			_, err := db.Exec("DELETE FROM books WHERE book_id = ?", bookID)
			if err != nil {
				log.Print(err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// Return a success message or response
			w.Write([]byte("Book deleted successfully"))
			return
		}
	}
}

func addBorrowedBook(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.URL.Path != "/addBorrowedBook" {
		http.NotFound(w, r)
		return
	}
	// Handle form submission for adding a student
	if r.Method == http.MethodPost {
		studentID := r.FormValue("studentID")
		bookID := r.FormValue("bookID")
		returnDate := r.FormValue("returnDate")
		exceedingDate := r.FormValue("exceedingDate")
		penalties := r.FormValue("penalties")

		// Validate form data
		if studentID == "" || bookID == "" || returnDate == "" || exceedingDate == "" || penalties == "" {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Convert date strings to time.Time format
		parsedReturnDate, _ := time.Parse("2006-01-02", returnDate)
		parsedExceedingDate, _ := time.Parse("2006-01-02", exceedingDate)

		// Insert the data into the database
		_, err := db.Exec("INSERT INTO borrowed_books (student_id, book_id, return_date, exceeding_date, penalties) VALUES (?, ?, ?, ?, ?)", studentID, bookID, parsedReturnDate, parsedExceedingDate, penalties)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect user back to the page
		http.Redirect(w, r, "/menu?pgs_=3-3", http.StatusSeeOther)
	}
}

func editBorrowedBook(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		borrowingID := r.URL.Path[len("/editBorrowedBook/"):]
		studentID := r.FormValue("studentID")
		bookID := r.FormValue("bookID")
		returnDate := r.FormValue("returnDate")
		exceedingDate := r.FormValue("exceedingDate")
		penalties := r.FormValue("penalties")

		// Validate form data
		if borrowingID == "" || studentID == "" || bookID == "" || returnDate == "" || exceedingDate == "" || penalties == "" {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Convert date strings to time.Time format
		parsedReturnDate, _ := time.Parse("2006-01-02", returnDate)
		parsedExceedingDate, _ := time.Parse("2006-01-02", exceedingDate)

		// Update the borrowed book data in the database
		_, err := db.Exec("UPDATE borrowed_books SET student_id = ?, book_id = ?, return_date = ?, exceeding_date = ?, penalties = ? WHERE borrowing_id = ?", studentID, bookID, parsedReturnDate, parsedExceedingDate, penalties, borrowingID)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Return a success response
		w.Write([]byte("Borrowed book updated successfully"))
		return
	}
}

func deleteBorrowedBook(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get the borrowing ID from the query parameter
	borrowingID := r.URL.Query().Get("borrowingID")

	// Validate the borrowing ID to ensure it's not empty
	if borrowingID == "" {
		http.Error(w, "Invalid borrowing ID", http.StatusBadRequest)
		return
	}

	// Handle deleting a borrowed book by borrowing ID
	if r.Method == http.MethodDelete {
		if borrowingID != "" {
			// Delete the borrowed book from the database using borrowingID
			_, err := db.Exec("DELETE FROM borrowed_books WHERE borrowing_id = ?", borrowingID)
			if err != nil {
				log.Print(err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Return a success message or response
			w.Write([]byte("Borrowed book deleted successfully"))
			return
		}
	}
}

func addStudent(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.URL.Path != "/addStudent" {
		http.NotFound(w, r)
		return
	}

	// Handle adding student form submission
	if r.Method == http.MethodPost {
		studentName := r.FormValue("name")
		department := r.FormValue("department")

		// Validate form data
		if studentName == "" || department == "" {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Insert the data into the database
		_, err := db.Exec("INSERT INTO students (student_name, department) VALUES (?, ?)", studentName, department)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect user back to the page
		http.Redirect(w, r, "/menu?pgs_=3-4", http.StatusSeeOther)
		return
	}
}

func editStudent(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		studentID := r.URL.Path[len("/editStudent/"):]
		name := r.FormValue("name")
		department := r.FormValue("department")

		// Validate form data
		if studentID == "" || name == "" || department == "" {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Update the student data in the database
		_, err := db.Exec("UPDATE students SET student_name = ?, department = ? WHERE student_id = ?", name, department, studentID)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Return a success response
		w.Write([]byte("Student updated successfully"))
		return
	}
}

func deleteStudent(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Get the student ID from the query parameter
	studentID := r.URL.Query().Get("studentID")

	// Validate the student ID to ensure it's not empty
	if studentID == "" {
		http.Error(w, "Invalid student ID", http.StatusBadRequest)
		return
	}

	// Handle deleting a student by student ID
	if r.Method == http.MethodDelete {
		studentID := r.FormValue("studentID")
		if studentID != "" {
			// Check for and delete related borrowed book records
			_, err := db.Exec("DELETE FROM borrowed_books WHERE student_id = ?", studentID)
			if err != nil {
				log.Print(err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Delete the student from the database using studentID
			_, err = db.Exec("DELETE FROM students WHERE student_id = ?", studentID)
			if err != nil {
				log.Print(err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Return a success message or response
			w.Write([]byte("Student and related borrowed books deleted successfully"))
			return
		}
	}
}

func addDepartment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.URL.Path != "/addDepartment" {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		departmentName := r.FormValue("departmentName")

		// Validate form data
		if departmentName == "" {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Insert the data into the database
		_, err := db.Exec("INSERT INTO department (department_name) VALUES (?)", departmentName)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Redirect user back to the page
		http.Redirect(w, r, "/menu?pgs_=3-5", http.StatusSeeOther)
		return
	}
}

func editDepartment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == http.MethodPost {
		departmentID := r.URL.Path[len("/editDepartment/"):]
		departmentName := r.FormValue("departmentName")

		// Validate form data
		if departmentID == "" || departmentName == "" {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Update the department data in the database
		_, err := db.Exec("UPDATE department SET department_name = ? WHERE department_id = ?", departmentName, departmentID)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Return a success response
		w.Write([]byte("Department updated successfully"))
		return
	}
}

func deleteDepartment(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.URL.Path != "/deleteDepartment/" {
		http.NotFound(w, r)
		return
	}

	// Handle deleting a department
	if r.Method == http.MethodDelete {
		departmentID := r.FormValue("departmentID")
		if departmentID != "" {
			// Delete the department from the database using departmentID
			_, err := db.Exec("DELETE FROM department WHERE department_id = ?", departmentID)
			if err != nil {
				log.Print(err.Error())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			// Return a success message or response
			w.Write([]byte("Department deleted successfully"))
			return
		}
	}
}

func main() {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MySQL")

	// Pass the db connection to the functions
	http.HandleFunc("/menu", func(w http.ResponseWriter, r *http.Request) {
		menu(w, r, db)
	})

	http.HandleFunc("/addBook", func(w http.ResponseWriter, r *http.Request) {
		addBook(w, r, db)
	})

	http.HandleFunc("/editBook/", func(w http.ResponseWriter, r *http.Request) {
		editBook(w, r, db)
	})

	http.HandleFunc("/deleteBook/", func(w http.ResponseWriter, r *http.Request) {
		deleteBook(w, r, db)
	})

	http.HandleFunc("/addBorrowedBook", func(w http.ResponseWriter, r *http.Request) {
		addBorrowedBook(w, r, db)
	})

	http.HandleFunc("/editBorrowedBook/", func(w http.ResponseWriter, r *http.Request) {
		editBorrowedBook(w, r, db)
	})

	http.HandleFunc("/deleteBorrowedBook/", func(w http.ResponseWriter, r *http.Request) {
		deleteBorrowedBook(w, r, db)
	})

	http.HandleFunc("/addCatalog", func(w http.ResponseWriter, r *http.Request) {
		addCatalog(w, r, db)
	})

	http.HandleFunc("/editCatalog/", func(w http.ResponseWriter, r *http.Request) {
		editCatalog(w, r, db)
	})

	http.HandleFunc("/deleteCatalog/", func(w http.ResponseWriter, r *http.Request) {
		deleteCatalog(w, r, db)
	})

	http.HandleFunc("/addStudent", func(w http.ResponseWriter, r *http.Request) {
		addStudent(w, r, db)
	})

	http.HandleFunc("/editStudent/", func(w http.ResponseWriter, r *http.Request) {
		editStudent(w, r, db)
	})

	http.HandleFunc("/deleteStudent/", func(w http.ResponseWriter, r *http.Request) {
		deleteStudent(w, r, db)
	})

	http.HandleFunc("/addDepartment", func(w http.ResponseWriter, r *http.Request) {
		addDepartment(w, r, db)
	})

	http.HandleFunc("/editDepartment/", func(w http.ResponseWriter, r *http.Request) {
		editDepartment(w, r, db)
	})

	http.HandleFunc("/deleteDepartment/", func(w http.ResponseWriter, r *http.Request) {
		deleteDepartment(w, r, db)
	})

	fs := http.FileServer(http.Dir("/static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", nil)
}
