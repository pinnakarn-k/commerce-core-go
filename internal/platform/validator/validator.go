package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/pinnakarn-k/commerce-core-go/internal/platform/response"

	govalidator "github.com/go-playground/validator/v10"
)

func ParseValidationError(err error, dto any) []response.FieldError {
	var validationErrors govalidator.ValidationErrors

	if !errors.As(err, &validationErrors) {
		return []response.FieldError{
			{
				Field:   "",
				Code:    "INVALID_JSON",
				Message: "invalid request body",
			},
		}
	}

	fields := make([]response.FieldError, 0, len(validationErrors))

	for _, fe := range validationErrors {
		field := jsonFieldPath(fe, dto)

		fields = append(fields, response.FieldError{
			Field:   field,
			Code:    codeFromTag(fe.Tag()),
			Message: messageFromFieldError(field, fe),
		})
	}

	return fields
}

func jsonFieldPath(fe govalidator.FieldError, dto any) string {
	namespace := fe.StructNamespace()
	rootType := reflect.TypeOf(dto)

	if rootType.Kind() == reflect.Ptr {
		rootType = rootType.Elem()
	}

	rootName := rootType.Name()
	namespace = strings.TrimPrefix(namespace, rootName+".")

	parts := strings.Split(namespace, ".")

	currentType := rootType
	jsonParts := make([]string, 0, len(parts))

	for _, part := range parts {
		fieldName, indexPart := splitIndex(part)

		if currentType.Kind() == reflect.Ptr {
			currentType = currentType.Elem()
		}

		if currentType.Kind() == reflect.Struct {
			sf, ok := currentType.FieldByName(fieldName)
			if ok {
				jsonName := jsonNameFromTag(sf)
				jsonParts = append(jsonParts, jsonName+indexPart)

				currentType = sf.Type
				if currentType.Kind() == reflect.Slice || currentType.Kind() == reflect.Array {
					currentType = currentType.Elem()
				}

				continue
			}
		}

		jsonParts = append(jsonParts, toSnakeCase(fieldName)+indexPart)
	}

	return strings.Join(jsonParts, ".")
}

func splitIndex(s string) (string, string) {
	idx := strings.Index(s, "[")
	if idx == -1 {
		return s, ""
	}

	return s[:idx], s[idx:]
}

func jsonNameFromTag(sf reflect.StructField) string {
	tag := sf.Tag.Get("json")
	if tag == "" {
		return toSnakeCase(sf.Name)
	}

	name := strings.Split(tag, ",")[0]
	if name == "" || name == "-" {
		return toSnakeCase(sf.Name)
	}

	return name
}

func codeFromTag(tag string) string {
	switch tag {
	case "required":
		return "REQUIRED"
	case "email":
		return "INVALID_EMAIL"
	case "min":
		return "MIN"
	case "max":
		return "MAX"
	case "len":
		return "LENGTH"
	default:
		return strings.ToUpper(tag)
	}
}

func messageFromFieldError(field string, fe govalidator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func toSnakeCase(s string) string {
	var b strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteRune('_')
		}

		b.WriteRune(r)
	}

	return strings.ToLower(b.String())
}
