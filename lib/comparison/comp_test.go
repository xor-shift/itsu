package comparison

import (
	"log"
	"testing"
)

func TestTokenize(t *testing.T) {
	strs := []string{
		"test 123 \"asdasdsad asdasd\" asd\"asd asd",
	}

	tokens := [][]string{
		[]string{"test", "123", "\"asdasdsad asdasd\"", "asd\"asd", "asd"},
	}

	for k, v := range strs {
		r := tokenizeForProgram(v)

		if len(tokens[k]) == len(r) {
			for k2, v2 := range r {
				if v2 != tokens[k][k2] {
					t.Error("test", k, "failed at index", k2, ":", r)
				}
			}
		} else {
			t.Error("test", k, "failed due to slice size (", len(r), "!=", len(tokens[k]), "): ", r)
		}
	}
}

func TestStringToProgram(t *testing.T) {
	strs := []string{
		"0 FIELD DUP 1 >= 3 <= AND 1 FIELD \"asdasd asd\" == OR",
	}

	for _, str := range strs {
		program, err := StringToProgram(str)

		if err != nil {
			t.Error(err)
			continue
		}

		log.Println(program)

		if str, err = StringFromProgram(program); err != nil {
			t.Error(err)
		} else {
			log.Println(str)
		}
	}
}

func TestComparer_Run(t *testing.T) {
	prog, err := StringToProgram("0u FIELD DUP 1 >= ROL 3 <= AND 1u FIELD \"asdasd asd\" == OR")

	if err != nil {
		t.Error(err)
		return
	}

	log.Println(prog)

	c := NewComparer(prog)

	c.Fields[0] = 2
	c.Fields[1] = "asdasd asd"

	res, err := c.Run()

	log.Println(res)
	log.Println(err)

	log.Println("instrPtr: ", c.instrPtr)
	log.Println("stack: ", c.stack)
}
