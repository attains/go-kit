package structs

import (
	"fmt"
	"github.com/attains/go-kit/helper/intersect"
	"reflect"
	"strings"
)

type GormNamer interface {
	ColumnName(table, column string) string
}

func Assign(dst interface{}, src interface{}) error {
	dstType, dstValue := reflect.TypeOf(dst), reflect.ValueOf(dst)
	srcType, srcValue := reflect.TypeOf(src), reflect.ValueOf(src)
	if dstType.Kind() != reflect.Ptr {
		return fmt.Errorf("dst must be type %q", reflect.Ptr.String())
	}
	dstType, dstValue = dstType.Elem(), dstValue.Elem()
	if dstType.Kind() != reflect.Struct {
		return fmt.Errorf("dst elem must be type %q", reflect.Struct.String())
	}

	if srcType.Kind() == reflect.Ptr {
		srcType, srcValue = srcType.Elem(), srcValue.Elem()
	}
	if srcType.Kind() != reflect.Struct {
		return fmt.Errorf("src elem must be type %q", reflect.Struct.String())
	}

	for i := 0; i < dstType.NumField(); i++ {
		dstFieldType, dstFieldValue := dstType.Field(i), dstValue.Field(i)
		if !dstFieldValue.CanSet() || !dstFieldValue.CanAddr() {
			continue
		}
		srcFieldType, ok := srcType.FieldByName(dstFieldType.Name)
		if !ok || srcFieldType.Type != dstFieldType.Type {
			continue
		}
		srcFieldValue := srcValue.FieldByName(dstFieldType.Name)
		if srcFieldValue.Type() != dstFieldValue.Type() {
			continue
		}
		dstFieldValue.Set(srcFieldValue)
	}
	return nil
}

func Diff(dst interface{}, src interface{}) (map[string]interface{}, []string, error) {
	dstType, dstValue := reflect.TypeOf(dst), reflect.ValueOf(dst)
	srcType, srcValue := reflect.TypeOf(src), reflect.ValueOf(src)
	if dstType.Kind() == reflect.Ptr {
		dstType, dstValue = dstType.Elem(), dstValue.Elem()
	}
	if srcType.Kind() == reflect.Ptr {
		srcType, srcValue = srcType.Elem(), srcValue.Elem()
	}
	if dstType.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("dst elem must be type %q", reflect.Struct.String())
	}
	if srcType.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("src elem must be type %q", reflect.Struct.String())
	}
	if dstType != srcType {
		return nil, nil, fmt.Errorf("dst and src must be same type")
	}
	result, fields := make(map[string]interface{}), make([]string, 0)
	for i := 0; i < dstType.NumField(); i++ {
		dstFieldType, dstFieldValue := dstType.Field(i), dstValue.Field(i)
		srcFieldType, srcFieldValue := srcType.Field(i), srcValue.Field(i)
		if dstFieldType.Type != srcFieldType.Type || dstFieldType.Name != srcFieldType.Name {
			continue
		}
		if dstFieldValue.Kind() == reflect.Ptr {
			dstFieldValue = dstFieldValue.Elem()
		}
		if srcFieldValue.Kind() == reflect.Ptr {
			srcFieldValue = srcFieldValue.Elem()
		}
		if dstFieldValue.Kind() != srcFieldValue.Kind() ||
			!reflect.DeepEqual(dstFieldValue.Interface(), srcFieldValue.Interface()) {
			result[dstFieldType.Name] = dstFieldValue.Interface()
			fields = append(fields, dstFieldType.Name)
			continue
		}
	}
	return result, fields, nil
}

func Fields(dst interface{}) ([]string, error) {
	dstType := reflect.TypeOf(dst)
	if dstType.Kind() == reflect.Ptr {
		dstType = dstType.Elem()
	}
	if dstType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("dst elem must be type %q", reflect.Struct.String())
	}
	result := make([]string, 0)
	for i := 0; i < dstType.NumField(); i++ {
		dstFieldType := dstType.Field(i)
		result = append(result, dstFieldType.Name)
	}
	return result, nil
}

func FieldsDiff(dst, src, reference interface{}) ([]string, error) {
	_, differenceFields, err := Diff(dst, src)
	if err != nil {
		return nil, err
	}
	if len(differenceFields) != 0 {
		canUpdateFields, err := Fields(reference)
		if err != nil {
			return nil, err
		}
		differenceFields = intersect.String(differenceFields, canUpdateFields)
	}
	return differenceFields, nil
}

func GormColumnName(st reflect.StructField, namer GormNamer) string {
	columnName := namer.ColumnName("", st.Name)
	if gormTag, ok := st.Tag.Lookup("gorm"); ok {
		gormTagSlice := strings.Split(gormTag, ",")
		for _, s := range gormTagSlice {
			ss := strings.Split(s, ":")
			if ss[0] == "column" && ss[1] != "" {
				columnName = ss[1]
				break
			}
		}
	}
	return columnName
}

func GormMapToStruct(data map[string]interface{}, dst interface{}, namer GormNamer) (interface{}, map[string]interface{}, error) {
	dstValue := reflect.ValueOf(dst)
	dstType := reflect.TypeOf(dst)
	if dstValue.Type().Kind() != reflect.Ptr {
		return nil, nil, fmt.Errorf("dst must be %s type", reflect.Ptr)
	}
	dstValue = dstValue.Elem()
	dstType = dstType.Elem()
	if dstValue.Type().Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("dst must be %s type", reflect.Struct)
	}
	for i := 0; i < dstValue.NumField(); i++ {
		fieldValue := dstValue.Field(i)
		if !fieldValue.CanSet() {
			continue
		}
		fieldType := dstType.Field(i)
		columnName := GormColumnName(fieldType, namer)
		if value, ok := data[columnName]; ok {
			fieldValue.Set(reflect.ValueOf(value))
			delete(data, columnName)
		}
	}
	return dstValue.Interface(), data, nil
}

func MapToStruct(namingStrategy GormNamer, data map[string]interface{}, dst interface{}) (interface{}, error) {
	dstValue := reflect.ValueOf(dst)
	dstType := reflect.TypeOf(dst)
	if dstValue.Type().Kind() != reflect.Ptr {
		return nil, fmt.Errorf("dst must be %s type", reflect.Ptr)
	}
	dstValue = dstValue.Elem()
	dstType = dstType.Elem()
	if dstValue.Type().Kind() != reflect.Struct {
		return nil, fmt.Errorf("dst must be %s type", reflect.Struct)
	}
	for i := 0; i < dstValue.NumField(); i++ {
		fieldValue := dstValue.Field(i)
		if !fieldValue.CanSet() {
			continue
		}
		fieldType := dstType.Field(i)
		columnName := namingStrategy.ColumnName("", fieldType.Name)
		if gormTag, ok := fieldType.Tag.Lookup("gorm"); ok {
			gormTagSlice := strings.Split(gormTag, ",")
			for _, s := range gormTagSlice {
				ss := strings.Split(s, ":")
				if ss[0] == "column" && ss[1] != "" {
					columnName = ss[1]
					break
				}
			}
		}
		if value, ok := data[columnName]; ok {
			fieldValue.Set(reflect.ValueOf(value))
		}
	}
	return dstValue.Interface(), nil
}

func GetName(src interface{}) string {
	t := reflect.TypeOf(src)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return ""
	}
	return t.Name()
}