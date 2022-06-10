package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

func Perform(args Arguments, writer io.Writer) error {
	err := checkFlags(args)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(args["fileName"], os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return errors.New(err.Error())
	}
	defer f.Close()

	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return errors.New(err.Error())
	}

	var users []User
	if len(bytes) > 0 {
		err = json.Unmarshal(bytes, &users)
		if err != nil {
			return errors.New(err.Error())
		}
	}
	switch args["operation"] {
	case "list":
		err = listOperation(bytes, writer)
		if err != nil {
			return err
		}
	case "add":
		err = addOperation(args, users, bytes, f, writer)
		if err != nil {
			return err
		}
	case "remove":
		err = removeOperation(args, users, bytes, f, writer)
		if err != nil {
			return err
		}
	case "findById":
		err = findByIdOperation(args, users, bytes, writer)
		if err != nil {
			return err
		}
	}

	return nil
}

func findByIdOperation(args Arguments, users []User, bytes []byte, writer io.Writer) error {
	id := args["id"]
	var err error
	var exists bool
	for _, v := range users {
		if v.Id == id {
			exists = true
			bytes, err = json.Marshal(v)
			if err != nil {
				return errors.New(err.Error())
			}
			_, err = writer.Write(bytes)
			if err != nil {
				return errors.New(err.Error())
			}
		}
	}
	if !exists {
		_, err = writer.Write([]byte(""))
		if err != nil {
			return errors.New(err.Error())
		}
	}

	return nil
}

func removeOperation(args Arguments, users []User, bytes []byte, f *os.File, writer io.Writer) error {
	id := args["id"]
	var err error
	var exists bool
	for i, v := range users {
		if v.Id == id {
			exists = true
			copy(users[i:], users[i+1:])
			users[len(users)-1] = User{}
			users = users[:len(users)-1]
		}
	}
	if !exists {
		_, err = writer.Write([]byte(fmt.Sprintf("Item with id %s not found", id)))
		if err != nil {
			return errors.New(err.Error())
		}
	}
	bytes, err = json.Marshal(users)
	if err != nil {
		return errors.New(err.Error())
	}

	err = ioutil.WriteFile(f.Name(), bytes, 0755)
	if err != nil {
		return errors.New(err.Error())
	}

	return nil
}

func addOperation(args Arguments, users []User, bytes []byte, f *os.File, writer io.Writer) error {
	var user User
	err := json.Unmarshal([]byte(args["item"]), &user)
	if err != nil {
		return errors.New(err.Error())
	}

	for _, v := range users {
		if v.Id == user.Id {
			_, err = fmt.Fprintf(writer, "Item with id %s already exists", v.Id)
			if err != nil {
				return errors.New(err.Error())
			}
			return nil
		}
	}

	users = append(users, user)
	bytes, err = json.Marshal(users)
	if err != nil {
		return errors.New(err.Error())
	}

	err = ioutil.WriteFile(f.Name(), bytes, 0755)
	if err != nil {
		return errors.New(err.Error())
	}

	return nil
}

func listOperation(bytes []byte, writer io.Writer) error {
	_, err := writer.Write(bytes)
	if err != nil {
		return errors.New(err.Error())
	}
	return nil
}

func main() {

	err := Perform(parseArgs(), os.Stdout)
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}

func checkFlags(args Arguments) error {
	switch o := args["operation"]; {
	case args["fileName"] == "":
		return errors.New("-fileName flag has to be specified")
	case o == "":
		return errors.New("-operation flag has to be specified")
	case o != "add" && o != "list" && o != "findById" && o != "remove":
		return errors.New(fmt.Sprintf("Operation %s not allowed!", o))
	case o == "add" && args["item"] == "":
		return errors.New("-item flag has to be specified")
	case o == "remove" && args["id"] == "":
		return errors.New("-id flag has to be specified")
	case o == "findById" && args["id"] == "":
		return errors.New("-id flag has to be specified")
	}

	return nil
}

func parseArgs() Arguments {
	id := flag.String("id", "", "to find by id use -id flag following with id number "+
		"provided in quotes")
	item := flag.String("item", "", "for adding new item to the array inside .json file - "+
		"it should be valid json file with the id, email and age fields")
	operation := flag.String("operation", "", "to pass type of operation you want to perform")
	fileName := flag.String("fileName", "", "to pass the json file name where json objects "+
		"are stored")
	flag.Parse()

	arguments := Arguments{
		"id":        *id,
		"item":      *item,
		"operation": *operation,
		"fileName":  *fileName,
	}

	return arguments
}
