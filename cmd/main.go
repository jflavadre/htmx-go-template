package main

import (
	"fmt"
	"html/template"
	"io"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Templates struct {
	templates *template.Template
}

// Funcion para renderizar el HTML
func (t *Templates) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func NewTemplates() *Templates {
	return &Templates{
		templates: template.Must(template.ParseGlob("views/*.html")),
	}
}

type Count struct {
	Count  int
	Saludo string
}

// Creamos un nuevo contacto (usuario)
var id = 0

type Contact struct {
	Username string
	Email    string
	Id       int
}

// Creamos una funcion para guardar un nuevo contacto
func newContact(username, email string) Contact {
	id++ //Incrementamos el ID mas 1 para diferenciar del resto
	return Contact{
		Username: username, Email: email, Id: id,
	}
}

// Creamos un tipo Contacts con un array de Contact
type Contacts = []Contact // Cada valor es un estructura de Contact [[33]]

// Creamos un tipo de Data para devolver el contenido al frontend
type Data struct {
	Contacts Contacts // Contiente un array de Contact -> []Contact [[46]]
}

func (d *Data) indexOf(id int) int {
	for i, contact := range d.Contacts {
		if contact.Id == id {
			return i
		}
	}
	return -1
}

func (d *Data) verifyEmail(email string) bool {
	for _, contact := range d.Contacts {
		if contact.Email == email {
			return true
		}
	}
	return false
}

func newData() Data {
	return Data{
		Contacts: []Contact{
			newContact("John", "john@gmail.com"),
			newContact("Claudia", "claudia@gmail.com"),
		},
	}
}

type FormData struct {
	Values map[string]string
	Errors map[string]string
}

func newFormData() FormData {
	return FormData{
		Values: make(map[string]string),
		Errors: make(map[string]string),
	}
}

type Page struct {
	Data Data
	Form FormData
}

func newPage() Page {
	return Page{
		Data: newData(),
		Form: newFormData(),
	}
}

func main() {

	e := echo.New()
	e.Use(middleware.Logger())

	toTheHtml := Count{Count: 0, Saludo: "Probando Templates en GO con HTMX!"}

	page := newPage()
	e.Renderer = NewTemplates()

	e.Static("/images", "images")
	e.Static("/css", "css")

	e.GET("/", func(c echo.Context) error {
		toTheHtml.Count = 0
		// toTheHtml.Saludo += ", " + strconv.Itoa(toTheHtml.Count)
		return c.Render(200, "index", toTheHtml)
	})

	e.POST("/count", func(c echo.Context) error {
		toTheHtml.Count++
		return c.Render(200, "count", toTheHtml)
	})

	e.GET("/contacts", func(c echo.Context) error {
		return c.Render(200, "contact-page", page)
	})

	e.POST("/contacts", func(c echo.Context) error {
		username := c.FormValue("username")
		email := c.FormValue("email")

		if page.Data.verifyEmail(email) {
			formData := newFormData()
			formData.Values["username"] = username
			formData.Values["email"] = email
			formData.Errors["email"] = "Email already exists!"

			// Se tiene que devolver con un 422 para ver el resultao con HTMX
			return c.Render(422, "form-contact", formData)
		}
		contact := newContact(username, email)

		page.Data.Contacts = append(page.Data.Contacts, contact)

		fmt.Println("-------------------------")
		fmt.Println(contact)
		fmt.Println("-------------------------")

		c.Render(200, "form-contact", newFormData())
		return c.Render(200, "oob-contact", contact)
	})

	e.DELETE("/contacts/:id", func(c echo.Context) error {
		time.Sleep(3 * time.Second)

		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.String(400, "Invalid id")
		}

		index := page.Data.indexOf(id)
		if index == -1 {
			return c.String(404, "Contact not found")
		}

		page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)

		return c.NoContent(200)
	})

	e.Logger.Fatal(e.Start(":4000"))
}
