package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

var (
	ErrInterfaceIsNotStruct = errors.New("can't validate value because it's not struct")
	ErrValidateTagIsEmpty   = errors.New("validate tag is empty")
	ErrUnsupportedType      = errors.New("unsupported type of field")

	ErrStringLength         = errors.New("string length must be exactly")
	ErrStringNotMatchRegexp = errors.New("string does not match regexp")
	ErrStringNotInSet       = errors.New("string must be in the set of strings")
	ErrNumberLessThanMin    = errors.New("number must be equal or greater than")
	ErrNumberMoreThanMax    = errors.New("number must be equal or less than")
	ErrNumberNotInSet       = errors.New("number must be in the set of numbers")
)

const (
	ValidateTagName      = "validate"
	ValidationRuleLen    = "len"
	ValidationRuleRegexp = "regexp"
	ValidationRuleIn     = "in"
	ValidationRuleMin    = "min"
	ValidationRuleMax    = "max"
)

func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "no errors"
	}

	sb := strings.Builder{}
	sb.WriteString("Validation errors: ")
	for _, validationErr := range v {
		sb.WriteString(fmt.Sprintf("Field %v is invalid %s \n", validationErr.Field, validationErr.Err))
	}

	return sb.String()
}

func Validate(v interface{}) error { //nolint:gocognit
	data := reflect.ValueOf(v)
	typeOfData := data.Type()

	if typeOfData.Kind() != reflect.Struct {
		return ErrInterfaceIsNotStruct
	}

	fields := reflect.VisibleFields(typeOfData)
	res := make(ValidationErrors, 0, len(fields))

	for _, field := range fields {
		if validateTag, ok := field.Tag.Lookup(ValidateTagName); ok {
			if validateTag == "" {
				return ErrValidateTagIsEmpty
			}

			var fieldErrs ValidationErrors
			validationRules := strings.Split(validateTag, "|")
			switch field.Type.Kind() { //nolint:exhaustive
			case reflect.Int:
				fieldValue := int(data.FieldByIndex(field.Index).Int())
				fieldErrs = validateInt(field.Name, fieldValue, validationRules)
			case reflect.String:
				fieldValue := data.FieldByIndex(field.Index).String()
				fieldErrs = validateString(field.Name, fieldValue, validationRules)
			case reflect.Slice:
				switch field.Type.Elem().Kind() { //nolint:exhaustive
				case reflect.Int:
					s := data.FieldByIndex(field.Index).Interface().([]int)

					for _, item := range s {
						itemErrs := validateInt(field.Name, item, validationRules)
						if itemErrs != nil {
							fieldErrs = append(fieldErrs, itemErrs...)
						}
					}
				case reflect.String:
					s := data.FieldByIndex(field.Index).Interface().([]string)

					for _, item := range s {
						itemErrs := validateString(field.Name, item, validationRules)
						if itemErrs != nil {
							fieldErrs = append(fieldErrs, itemErrs...)
						}
					}
				default:
					return ErrUnsupportedType
				}
			default:
				return ErrUnsupportedType
			}
			if fieldErrs != nil {
				res = append(res, fieldErrs...)
			}
		}
	}

	if len(res) > 0 {
		return res
	}

	return nil
}

func validateInt(fieldName string, v int, validationRules []string) ValidationErrors {
	var fieldErrs ValidationErrors

	for _, validationRuleString := range validationRules {
		validationData := strings.Split(validationRuleString, ":")
		validationRule := validationData[0]
		validationValue := validationData[1]
		fieldErr := ValidationError{Field: fieldName, Err: nil}

		switch validationRule {
		case ValidationRuleMin:
			t, _ := strconv.Atoi(validationValue)
			fieldErr.Err = validateMin(v, t)
		case ValidationRuleMax:
			t, _ := strconv.Atoi(validationValue)
			fieldErr.Err = validateMax(v, t)
		case ValidationRuleIn:
			stringSet := strings.Split(validationValue, ",")
			intSet := make([]int, len(stringSet))
			for i, s := range stringSet {
				intSet[i], _ = strconv.Atoi(s)
			}
			fieldErr.Err = validateIntSet(v, intSet)
		}
		if fieldErr.Err != nil {
			fieldErrs = append(fieldErrs, fieldErr)
		}
	}

	if len(fieldErrs) > 0 {
		return fieldErrs
	}

	return nil
}

func validateString(fieldName string, v string, validationRules []string) ValidationErrors {
	var fieldErrs ValidationErrors

	for _, validationString := range validationRules {
		validationData := strings.Split(validationString, ":")
		validationRule := validationData[0]
		validationValue := validationData[1]
		fieldErr := ValidationError{Field: fieldName, Err: nil}

		switch validationRule {
		case ValidationRuleLen:
			t, _ := strconv.Atoi(validationValue)
			fieldErr.Err = validateLen(v, t)
		case ValidationRuleRegexp:
			fieldErr.Err = validateRegexp(v, validationValue)
		case ValidationRuleIn:
			stringSet := strings.Split(validationValue, ",")
			fieldErr.Err = validateStringSet(v, stringSet)
		}
		if fieldErr.Err != nil {
			fieldErrs = append(fieldErrs, fieldErr)
		}
	}

	if len(fieldErrs) > 0 {
		return fieldErrs
	}

	return nil
}

func validateMin(v int, t int) error {
	if v < t {
		return fmt.Errorf("%w %v", ErrNumberLessThanMin, t)
	}

	return nil
}

func validateMax(v int, t int) error {
	if v > t {
		return fmt.Errorf("%w %v", ErrNumberMoreThanMax, t)
	}

	return nil
}

func validateIntSet(v int, t []int) error {
	for _, ti := range t {
		if v == ti {
			return nil
		}
	}

	return fmt.Errorf("%w %v", ErrNumberNotInSet, t)
}

func validateLen(v string, t int) error {
	if len(v) != t {
		return fmt.Errorf("%w %v symbols", ErrStringLength, t)
	}

	return nil
}

func validateRegexp(v string, t string) error {
	validString, _ := regexp.Compile(t)

	if !validString.MatchString(v) {
		return fmt.Errorf("%w %v", ErrStringNotMatchRegexp, t)
	}

	return nil
}

func validateStringSet(v string, t []string) error {
	for _, ti := range t {
		if v == ti {
			return nil
		}
	}

	return fmt.Errorf("%w %v", ErrStringNotInSet, t)
}
