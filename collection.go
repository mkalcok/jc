package jc

import (
	"reflect"
	"unicode"
	"fmt"
	"errors"
	"strings"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/mgo.v2"
	"github.com/kalcok/jc/tools"
)

type document interface {
	setCollection(string)
	CollectionName() string
	SetDatabase(string)
	Database() string
	Init(reflect.Value, reflect.Type)
	InitDB() error
	Info()
	Save(bool) (*mgo.ChangeInfo, error)
}

type Collection struct {
	_collectionName string                `bson:"-"json:"-"`
	_collectionDB   string                `bson:"-"json:"-"`
	_parent         reflect.Value         `bson:"-"json:"-"`
	_parentType     reflect.Type          `bson:"-"json:"-"`
	_explicitID     string                `bson:"-"json:"-"`
	_implicitID     bson.ObjectId         `bson:"-"json:"-"`
	_skeleton       []reflect.StructField `bson:"-"json:"-"`
}

func (c *Collection) setCollection(name string) {
	c._collectionName = name
}

func (c *Collection) CollectionName() string {
	return c._collectionName
}

func (c *Collection) SetDatabase(name string) {
	c._collectionDB = name

}

func (c *Collection) Database() string {
	return c._collectionDB
}

func (c *Collection) Info() {
	fmt.Printf("Database %s\n", c._collectionDB)
	fmt.Printf("Collection %s\n", c._collectionName)
	fmt.Printf("Parent__ %s\n", c._parent)
}

func (c *Collection) Save(reuseSocket bool) (info *mgo.ChangeInfo, err error) {
	var session *mgo.Session
	master_session, err := tools.GetSession()
	var documentID interface{}

	if err != nil {
		return info, err
	}

	if reuseSocket {
		session = master_session.Clone()
	} else {
		session = master_session.Copy()
	}

	if len(c._explicitID) > 0 {
		documentID = c._parent.Elem().FieldByName(c._explicitID).Interface()
	} else if len(c._implicitID) > 0 {
		documentID = c._implicitID
	} else {
		c._implicitID = bson.NewObjectId()
		documentID = c._implicitID
	}

	collection := session.DB(c._collectionDB).C(c._collectionName)
	info, err = collection.UpsertId(documentID, c._parent.Interface())

	return info, err
}

func (c *Collection) Init(parent reflect.Value, parentType reflect.Type) {

	c._parent = parent
	c._parentType = parentType
	fmt.Println(c._parent)
	fmt.Println(c._parentType)
	documentIdFound := false
	for i := 0; i < reflect.Indirect(c._parent).NumField(); i++ {
		field := c._parentType.Field(i)

		// Find explicit Collection name
		if field.Type == reflect.TypeOf(Collection{}) {
			explicitName := false
			odm_tag, tag_present := field.Tag.Lookup("odm")
			if tag_present {
				odm_fields := strings.Split(odm_tag, ",")
				if len(odm_fields) > 0 && odm_fields[0] != "" {
					c.setCollection(odm_fields[0])
					explicitName = true
				}
			}
			if !explicitName {
				c.setCollection(camelToSnake(parentType.Name()))
			}
		}

		// Find explicit index field
		bson_tag, tag_present := field.Tag.Lookup("bson")
		if tag_present {
			field_id := strings.Split(bson_tag, ",")
			switch field_id[0] {
			case "_id":
				c._explicitID = field.Name
				documentIdFound = true
				break
			case "-":
				continue
			default:
				break
			}
		}
		c._skeleton = append(c._skeleton, field)
	}
	if !documentIdFound {
		c._explicitID = "_id"
	}

}

func (c *Collection) InitDB() error {
	session, err := tools.GetSession()
	if err == nil {
		c.SetDatabase(session.DB("").Name)
	} else {
		err = errors.New("database not initialized")
	}
	return err
}

func NewDocument(c document) error {
	var err error
	objectType := reflect.TypeOf(c).Elem()
	objectValue := reflect.ValueOf(c)
	c.Init(objectValue, objectType)
	err = c.InitDB()

	return err
}

func camelToSnake(camel string) string {
	var (
		snake_name []rune
		next       rune
	)
	for i, c := range camel {
		if unicode.IsUpper(c) && i != 0 {
			snake_name = append(snake_name, '_')
		}
		next = unicode.ToLower(c)
		snake_name = append(snake_name, next)
	}
	return string(snake_name)
}

func initPrototype(prototype reflect.Value, internalType reflect.Type) reflect.Value {
	var inputs []reflect.Value
	inputs = append(inputs, reflect.ValueOf(prototype))
	inputs = append(inputs, reflect.ValueOf(internalType))
	prototype.MethodByName("Init").Call(inputs)
	prototype.MethodByName("InitDB").Call(nil)

	return reflect.Indirect(prototype)

}