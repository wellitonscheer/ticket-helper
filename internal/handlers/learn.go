package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Learn struct {
}

func NewLearnHandlers() *Learn {
	return &Learn{}
}

var id = 0

type Contact struct {
	Name  string
	Email string
	Id    int
}

type Contacts = []Contact

type Data struct {
	Contacts Contacts
	Count    int
}

func (d *Data) indexOf(id int) int {
	for i, contact := range d.Contacts {
		if contact.Id == id {
			return i
		}
	}

	return -1
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
	id++
	return Contact{
		Name:  name,
		Email: email,
		Id:    id,
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

func (l *Learn) Learn(c *gin.Context) {
	c.HTML(http.StatusOK, "learn", page)
}

func (l *Learn) Count(c *gin.Context) {
	page.Data.Count++
	c.HTML(http.StatusOK, "count", page.Data)
}

func (l *Learn) CreateContact(c *gin.Context) {
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

	contact := newContact(name, email)
	page.Data.Contacts = append(page.Data.Contacts, contact)

	c.HTML(http.StatusOK, "form", newFormData())
	c.HTML(http.StatusOK, "oob-contact", contact)
}

func (l *Learn) DeleteContact(c *gin.Context) {
	time.Sleep(2 * time.Second)

	idStr := c.Params.ByName("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid id")
		return
	}

	index := page.Data.indexOf(id)
	if index == -1 {
		c.String(http.StatusBadRequest, "Contact not found")
	}

	page.Data.Contacts = append(page.Data.Contacts[:index], page.Data.Contacts[index+1:]...)

	c.Status(http.StatusOK)
}
