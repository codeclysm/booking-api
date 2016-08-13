package calendar

import (
	"time"

	"gopkg.in/dancannon/gorethink.v1"

	"github.com/codeclysm/rdbutils"
	"github.com/juju/errors"
)

// Client is the way you use this module. Just instantiate it with a db instance
// and call its methods
type Client struct {
	DB *rdbutils.Database
}

// Between returns all the appointments in a certain location, after the time start and before the time end, included.
func (c *Client) Between(location string, start, end time.Time) ([]Appointment, error) {
	// Build the query
	query := c.DB.Query().Filter(map[string]string{"where": location})
	query = query.Filter(func(row gorethink.Term) gorethink.Term {
		return row.Field("when").During(start, end)
	})

	// Execute the query
	cursor, err := c.DB.Run(query)
	if err != nil {
		return nil, errors.Annotatef(err, "while executing the query %v", query)
	}

	// Build the array
	appointments := []Appointment{}
	err = cursor.All(&appointments)
	if err != nil {
		return nil, errors.Annotatef(err, "while executing the query %v", query)
	}
	return appointments, nil
}

// Get an appointment with a specific id
func (c *Client) Get(id string) (*Appointment, error) {
	query := c.DB.Query().Get(id)
	cursor, err := c.DB.Run(query)
	if err != nil {
		return nil, errors.Annotatef(err, "while executing the query %v", query)
	}

	if cursor.IsNil() {
		return nil, errors.NotFoundf("with id %s", id)
	}

	var app Appointment
	err = cursor.One(&app)
	if err != nil {
		return nil, errors.Annotatef(err, "while unmarshaling cursor %v", cursor)
	}

	return &app, nil
}

// Save persists the appointment in database
func (c *Client) Save(app *Appointment) error {
	options := gorethink.InsertOpts{Conflict: "replace"}
	query := c.DB.Query().Insert(app, options)
	_, err := c.DB.RunWrite(query)
	if err != nil {
		return errors.Annotatef(err, "while executing the query %v", query)
	}
	return nil
}

// Delete removes an appointment from the database
func (c *Client) Delete(app *Appointment) error {
	return c.DB.Exec(c.DB.Query().Get(app.ID).Delete())
}