package flagconfig

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	timeDurationType            = reflect.TypeOf(time.Duration(0))
	flagsValueVar               flag.Value
	flagsValueInterfaceType     = reflect.ValueOf(&flagsValueVar).Elem().Type()
	encodingBinaryVar           encodingBinary
	encodingBinaryInterfaceType = reflect.ValueOf(&encodingBinaryVar).Elem().Type()
	encodingTextVar             encodingText
	encodingTextInterfaceType   = reflect.ValueOf(&encodingTextVar).Elem().Type()
)

var (
	DefaultTagsNaming = TagsSettings{
		Default:      "default",
		Description:  "desc",
		NameOverride: "name",
		Ignored:      "ignored",
	}
	EnvconfigTagsNaming = TagsSettings{
		Default:      "default",
		Description:  "desc",
		NameOverride: "envconfig",
		Ignored:      "ignored",
	}
)

// MakeFlags - make flag for each field in structure
// s - pointer to structure
func MakeFlags(s interface{}, name string, handling flag.ErrorHandling) (*flag.FlagSet, error) {
	return MakeFlagsWithCustomTags(s, name, handling, DefaultTagsNaming)
}

func MakeFlagsEnvconfig(s interface{}, name string, handling flag.ErrorHandling) (*flag.FlagSet, error) {
	return MakeFlagsWithCustomTags(s, name, handling, EnvconfigTagsNaming)
}

func MakeFlagsWithCustomTags(s interface{}, name string, handling flag.ErrorHandling, settings TagsSettings) (*flag.FlagSet, error) {
	flagSet := flag.NewFlagSet(name, handling)
	return flagSet, fieldsAsFlags(s, settings, "", flagSet)
}

// TagsSettings contains names of tags
type TagsSettings struct {
	Default, Description, NameOverride, Ignored string
}

func (s TagsSettings) getName(field reflect.StructField) string {
	if override, ok := field.Tag.Lookup(s.NameOverride); ok {
		return override
	}
	return strings.ToLower(field.Name)
}

type IncorrectDefaultValue struct {
	Field  string
	Type   reflect.Type
	Value  string
	Reason error
}

func (err IncorrectDefaultValue) Error() string {
	return fmt.Sprintf(
		"can't set default value %q for field %s type %s: %s",
		err.Value, err.Field, err.Type, err.Reason)
}

func (err IncorrectDefaultValue) Unwrap() error {
	return err.Reason
}

func asValue(ptr interface{}, t reflect.Type) flag.Value {
	switch {
	case t.Implements(flagsValueInterfaceType):
		return ptr.(flag.Value)
	case t.Implements(encodingTextInterfaceType):
		return textWrapper{ptr.(encodingText)}
	case t.Implements(encodingBinaryInterfaceType):
		return binaryWrapper{ptr.(encodingBinary)}
	}
	return nil
}

func fieldsAsFlags(s interface{}, settings TagsSettings, prefix string, flags *flag.FlagSet) error {
	v := reflect.ValueOf(s).Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		pointerToFieldValue := getPointerToField(v, i)
		if !pointerToFieldValue.Elem().CanSet() || f.Tag.Get(settings.Ignored) == "true" {
			continue
		}
		valueType := pointerToFieldValue.Elem().Type()
		pointerToField := pointerToFieldValue.Interface()
		flagName := settings.getName(f)
		if prefix != "" {
			flagName = prefix + "." + flagName
		}
		desc := f.Tag.Get(settings.Description)
		defaultString := defaultStringFlag(f, settings)
		errorSetDefault := func(err error) IncorrectDefaultValue {
			return IncorrectDefaultValue{f.Name, valueType, defaultString, err}
		}
		if value := asValue(pointerToField, pointerToFieldValue.Type()); value != nil {
			if defaultString != "" {
				if err := value.Set(defaultString); err != nil {
					return errorSetDefault(err)
				}
			}
			// if usage contains `name`, flags package will use it as a type name
			if !strings.ContainsRune(desc, '`') {
				desc = fmt.Sprintf("`%s` %s", valueType, desc)
			}
			flags.Var(value, flagName, desc)
			continue
		}
		if pointerToFieldValue.Elem().Kind() == reflect.Struct {
			if f.Anonymous {
				flagName = prefix
			}
			err := fieldsAsFlags(pointerToField, settings, flagName, flags)
			if err != nil {
				return fmt.Errorf("in field %s with type %s: %w", f.Name, valueType, err)
			}
			continue
		}
		if valueType == timeDurationType {
			d, err := defaultDuration(defaultString)
			if err != nil {
				return errorSetDefault(err)
			}
			flags.DurationVar(pointerToField.(*time.Duration), flagName, d, desc)
			continue
		}
		err := byKind(valueType, flags, pointerToField, flagName, defaultString, desc)
		if err != nil {
			if errors.Is(err, unsupportedTypeErr) {
				return fmt.Errorf("field %s has unsupported type %s, type should implements flag.Value or be bool, (u)int(64), string, float64",
					f.Name, valueType)
			}
			if badDefault := new(IncorrectDefaultValue); errors.As(err, badDefault) {
				return errorSetDefault(badDefault.Reason)
			}
			return err
		}
	}
	return nil
}

func defaultStringFlag(f reflect.StructField, settings TagsSettings) string {
	if fromTag := f.Tag.Get(settings.Default); fromTag != "" {
		return fromTag
	}
	return ""
}

func byKind(ft reflect.Type, flags *flag.FlagSet, pointerToField interface{}, flagName string, defaultString string, desc string) error {
	switch ft.Kind() {
	case reflect.String:
		flags.StringVar(pointerToField.(*string), flagName, defaultString, desc)
	case reflect.Bool:
		d, err := defaultBool(defaultString)
		if err != nil {
			return IncorrectDefaultValue{Reason: err}
		}
		flags.BoolVar(pointerToField.(*bool), flagName, d, desc)
	case reflect.Int:
		d, err := defaultInt(defaultString)
		if err != nil {
			return IncorrectDefaultValue{Reason: err}
		}
		flags.IntVar(pointerToField.(*int), flagName, d, desc)
	case reflect.Int64:
		d, err := defaultInt64(defaultString)
		if err != nil {
			return IncorrectDefaultValue{Reason: err}
		}
		flags.Int64Var(pointerToField.(*int64), flagName, d, desc)
	case reflect.Uint:
		d, err := defaultUint(defaultString)
		if err != nil {
			return IncorrectDefaultValue{Reason: err}
		}
		flags.UintVar(pointerToField.(*uint), flagName, d, desc)
	case reflect.Uint64:
		d, err := defaultUint64(defaultString)
		if err != nil {
			return IncorrectDefaultValue{Reason: err}
		}
		flags.Uint64Var(pointerToField.(*uint64), flagName, d, desc)
	case reflect.Float64:
		d, err := defaultFloat64(defaultString)
		if err != nil {
			return IncorrectDefaultValue{Reason: err}
		}
		flags.Float64Var(pointerToField.(*float64), flagName, d, desc)
	default:
		return unsupportedTypeErr
	}
	return nil
}

var (
	unsupportedTypeErr = errors.New("unsupported type: type should implements flag.Value or be bool, (u)int(64), string, float64")
)

func getPointerToField(v reflect.Value, i int) reflect.Value {
	field := v.Field(i)
	if field.Type().Kind() == reflect.Interface && !field.IsNil() {
		if elem := field.Elem(); elem.Kind() == reflect.Ptr {
			return elem
		}
		return field
	}
	if field.Type().Kind() != reflect.Ptr {
		return field.Addr()
	}
	if field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
	}
	return field
}
