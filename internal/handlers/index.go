package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Contact struct {
	Name  string
	Email string
}

type Contacts = []Contact

type Data struct {
	Contacts Contacts
	Count    int
}

func (d *Data) hasEmail(email string) bool {
	for _, contact := range d.Contacts {
		if contact.Email == email {
			return true
		}
	}
	return false
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

func newContact(name, email string) Contact {
	return Contact{
		Name:  name,
		Email: email,
	}
}

func newData() Data {
	return Data{
		Contacts: []Contact{
			newContact("Json", "json@gamil"),
			newContact("Maria", "maria@gamil"),
			newContact("Eitan", "destroier@gamil"),
		},
		Count: 0,
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

var page = newPage()

func Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index", page)
}

func IndexCount(c *gin.Context) {
	page.Data.Count++
	c.HTML(http.StatusOK, "count", page.Data)
}

func IndexCreateContact(c *gin.Context) {
	name := c.PostForm("name")
	email := c.PostForm("email")

	if page.Data.hasEmail(email) {
		form := newFormData()
		form.Values["name"] = name
		form.Values["email"] = email
		form.Errors["email"] = "Email already exists"

		c.HTML(http.StatusUnprocessableEntity, "form", form)
		return
	}

	page.Data.Contacts = append(page.Data.Contacts, newContact(name, email))

	c.HTML(http.StatusOK, "display", page.Data)
}
