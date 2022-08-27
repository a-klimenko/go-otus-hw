package hw09structvalidator

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int      `validate:"min:18|max:50"`
		Email  string   `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole `validate:"in:admin,stuff"`
		Phones []string `validate:"len:11"`
		meta   json.RawMessage
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}

	Person struct {
		Name string `validate:""`
	}

	Product struct {
		Title string
		Price float32 `validate:"min:1"`
	}
)

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{
				ID:     "hyPjYyO2bhMyhHSU07ezeD5XjpStfIJL34n9",
				Name:   "name",
				Age:    19,
				Email:  "mymail@mail.ru",
				Role:   "stuff",
				Phones: []string{"89123456789", "89123456788"},
				meta:   nil,
			},
			expectedErr: nil,
		},
		{
			in: User{
				ID:     "hyPjYyO2bhMyhHSU07ezeD5XjpStfIJL34n",
				Name:   "name",
				Age:    17,
				Email:  "mymail",
				Role:   "role",
				Phones: []string{"89123456789", "8912345678"},
				meta:   nil,
			},
			expectedErr: ValidationErrors{
				{Field: "ID", Err: fmt.Errorf("%w 36 symbols", ErrStringLength)},
				{Field: "Age", Err: fmt.Errorf("%w 18", ErrNumberLessThanMin)},
				{Field: "Email", Err: fmt.Errorf("%w ^\\w+@\\w+\\.\\w+$", ErrStringNotMatchRegexp)},
				{Field: "Role", Err: fmt.Errorf("%w [admin stuff]", ErrStringNotInSet)},
				{Field: "Phones", Err: fmt.Errorf("%w 11 symbols", ErrStringLength)},
			},
		},
		{
			in: User{
				ID:     "hyPjYyO2bhMyhHSU07ezeD5XjpStfIJL34n9",
				Name:   "name",
				Age:    51,
				Email:  "mymail@mail.ru",
				Role:   "stuff",
				Phones: []string{"89123456789", "89123456788"},
				meta:   nil,
			},
			expectedErr: ValidationErrors{
				{Field: "Age", Err: fmt.Errorf("%w 50", ErrNumberMoreThanMax)},
			},
		},
		{
			in:          App{Version: "12345"},
			expectedErr: nil,
		},
		{
			in: App{Version: "1234"},
			expectedErr: ValidationErrors{
				{Field: "Version", Err: fmt.Errorf("%w 5 symbols", ErrStringLength)},
			},
		},
		{
			in:          Response{Code: 200, Body: "body"},
			expectedErr: nil,
		},
		{
			in:          Response{Code: 404, Body: "body"},
			expectedErr: nil,
		},
		{
			in:          Response{Code: 500, Body: "body"},
			expectedErr: nil,
		},
		{
			in: Response{Code: 300, Body: "body"},
			expectedErr: ValidationErrors{
				{Field: "Code", Err: fmt.Errorf("%w [200 404 500]", ErrNumberNotInSet)},
			},
		},
		{
			in:          Token{},
			expectedErr: nil,
		},
		{
			in:          Token{[]byte{1}, []byte{1}, []byte{1}},
			expectedErr: nil,
		},
		{
			in:          []int{1},
			expectedErr: ErrInterfaceIsNotStruct,
		},
		{
			in:          Person{Name: "Ivan"},
			expectedErr: ErrValidateTagIsEmpty,
		},
		{
			in:          Product{Title: "title", Price: 32.4},
			expectedErr: ErrUnsupportedType,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()

			result := Validate(tt.in)
			require.Equal(t, tt.expectedErr, result)
		})
	}
}
