package main

import (
	"errors"
	"fmt"
)

/*
// operators
  $        -- root node (can be either object or array)
  @        -- the current node (in a filter)
  .node    -- dot-notated child
  [123]    -- array index
  [12:34]  -- array bound
  [?(<expression>)] -- filter expression. Applicable to arrays only.
// functions
  $.obj.length() -- array lengh or string length, depending on the obj type
  $.obj.size() -- object size in bytes (as is)
// definite
  $.obj
  $.obj.val
  // arrays: indexed
  $.obj[3]
  $.obj[3].val
  $.obj[-2]  -- second from the end
  // arrays: bounded
  $.obj[:]   -- == $.obj (all elements of the array)
  $.obj[2:]  -- items from index 2 (inclusive) till the end
  $.obj[:5]  -- items from the beginning to the index 5 (exclusive)
  $.obj[-2:] -- items from the second element from the end (inclusive) till the end
  $.obj[:-2] -- items from the beginning to the second element from the end (exclusive)
  $.obj[3:5] -- items from index 2 (inclusive) to the index 5 (exclusive)
// indefininte
  $.obj[any:any].something -- composite sub-query
  $.obj[3,5,7] -- multiple array indexes
  $.obj[?(@.price > 1000)] -- filter expression
// more examples
  $.obj[?(@.price > $.average)]
  $[0].compo[]
*/

const (
	// array node
	cArrayType = 1 << iota
	// array properties
	cArrBounded = 1 << iota // bounded [x:y] or indexed [x]
	// terminal node
	cIsTerminal = 1 << iota
)

type tToken struct {
	Key   string
	Type  int8 // properties
	Left  int  // >=0 index from the start, <0 backward index from the end
	Right int  // 0 till the end inclusive, >0 to index exclusive, <0 backward index from the end exclusive
	Next  *tToken
}

// Get the jsonpath subset of the input
func Get(input []byte, path string) ([]byte, error) {

	if path[0] != '$' {
		return nil, errors.New("path: $ expected")
	}

	tokens, err := parsePath([]byte(path))
	if err != nil {
		return nil, err
	}

	return getValue(input, tokens)
}

func parsePath(path []byte) (*tToken, error) {
	tok := &tToken{}
	i := 0
	ind := 0
	l := len(path)
	if l == 0 {
		return nil, errors.New("element expected")
	}
	// key
	for ; i < l && path[i] != '.' && path[i] != '['; i++ {
	}
	tok.Key = string(path[:i])
	// type
	if i == l {
		tok.Type |= cIsTerminal
		return tok, nil
	}
	if path[i] == '[' {
		tok.Type = cArrayType
		i++
		ind, i = readNumber(path, i)
		if i == l {
			return nil, errors.New("']' expected (1)")
		}
		if path[i] != ':' && path[i] != ']' {
			return nil, errors.New("invalid array bound (1)")
		}
		tok.Left = ind
		//
		if path[i] == ':' {
			tok.Type |= cArrBounded
			i++
			ind, i = readNumber(path, i)
			if i == l {
				return nil, errors.New("']' expected (2)")
			}
			if path[i] != ']' {
				return nil, errors.New("invalid array bound (2)")
			}
			tok.Right = ind
		}
		i++
		if i == l {
			tok.Type |= cIsTerminal
			return tok, nil
		}
	}
	if tok.Type&cArrBounded > 0 && tok.Type&cIsTerminal == 0 {
		return nil, errors.New("indefinite references are not yet supported")
	}
	if path[i] != '.' {
		return nil, errors.New("invalid element reference")
	}
	i++
	next, err := parsePath(path[i:])
	if err != nil {
		return nil, err
	}
	tok.Next = next

	return tok, nil
}

func getValue(input []byte, tok *tToken) (result []byte, err error) {
	if len(input) == 0 {
		return nil, errors.New("unexpected end of file")
	}
	if input[0] != '{' && input[0] != '[' {
		return nil, errors.New("object or array expected")
	}
	if tok.Key != "$" {
		// find the key and seek to the value
		input, err = getKeyValue(input, tok.Key)
		if err != nil {
			return nil, err
		}
	}
	// check value type
	if err = checkValueType(input, tok); err != nil {
		return nil, err
	}

	// here we are at the beginning of a value

	if tok.Type&cArrayType > 0 {
		input, err = sliceArray(input, tok)
	} else {
		input, err = sliceValue(input)
	}
	if tok.Type&cIsTerminal > 0 {
		return input, nil
	}
	if tok.Type&cArrayType == 0 {
		return getValue(input, tok.Next)
	}
	return getValues(input, tok)
}

// getKeyValue: seek to the key specified
func getKeyValue(input []byte, key string) ([]byte, error) {
	if input[0] != '{' {
		return nil, errors.New("object expected")
	}
	return nil, nil
}

// sliceArray: select node(s) by bound(s)
func sliceArray(input []byte, tok *tToken) ([]byte, error) {
	return nil, nil
}

// sliceValue: slice a single value
func sliceValue(input []byte) ([]byte, error) {
	return nil, nil
}

// getValues: get (sub-)values from array
func getValues(input []byte, tok *tToken) ([]byte, error) {
	return nil, nil
}

func checkValueType(input []byte, tok *tToken) error {
	if len(input) < 2 {
		return errors.New("unexpected end of input")
	}
	if input[0] == '[' && tok.Type&cArrayType == 0 {
		return errors.New("object expected at " + tok.Key)
	} else if input[0] == '{' && tok.Type&cArrayType > 0 {
		return errors.New("array expected at " + tok.Key)
	} else if tok.Type&cIsTerminal == 0 {
		return errors.New("complex type expected at " + tok.Key)
	}
	return nil
}

// get input and current value position
// return next key and new value position
const keySeek = 1
const keyOpen = 2
const keyClose = 4

func nextKey(input []byte, i int) ([]byte, int) {
	status := keySeek
	key := make([]byte, 0)
	for l := len(input); i < l; i++ {
		ch := input[i]
		switch {
		case status&keyOpen > 0:
			if ch == '"' {
				status -= keyOpen
				status |= keyClose
			} else {
				key = append(key, ch)
			}
		case status&keySeek > 0 && ch == '"':
			status -= keySeek
			status |= keyOpen
		case status&keyClose > 0 && ch == ':':
			return key, i + 1
		}
	}
	return nil, i
}

func readNumber(path []byte, i int) (int, int) {
	sign := 1
	l := len(path)
	ind := 0
	for ch := path[i]; i < l && (ch == '-' || (ch >= '0' && ch <= '9')); ch = path[i] {
		if ch == '-' {
			sign = -1
		} else {
			ind = ind*10 + int(ch-'0')
		}
		i++
	}
	return ind * sign, i
}

func main() {
	data := []byte(`
		{
			"store": {
				"book": [
					{
						"category": "reference",
						"author": "Nigel Rees",
						"title": "Sayings of the Century",
						"price": 8.95
					},
					{
						"category": "fiction",
						"author": "Evelyn Waugh",
						"title": "Sword of Honour",
						"price": 12.99
					},
					{
						"category": "fiction",
						"author": "Herman Melville",
						"title": "Moby Dick",
						"isbn": "0-553-21311-3",
						"price": 8.99
					},
					{
						"category": "fiction",
						"author": "J. R. R. Tolkien",
						"title": "The Lord of the Rings",
						"isbn": "0-395-19395-8",
						"price": 22.99
					}
				],
				"bicycle": {
					"color": "red",
					"price": 19.95
				}
			},
			"expensive": 10
		}
	`)

	s, err := Get(data, "$.book[0].title")

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(s))
}
