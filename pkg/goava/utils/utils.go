package utils

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/itskovichanton/goava/pkg/goava/errs"
	"github.com/lingdor/stackerror"
	"github.com/spf13/cast"
	"golang.org/x/text/encoding/charmap"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"text/template"
	"time"
)

func UpcastMapOfSlicesStr(m map[string][]string) map[string][]interface{} {
	p := map[string][]interface{}{}
	for k, q := range m {
		v := []interface{}{}
		for _, vl := range q {
			v = append(v, vl)
		}
		p[k] = v
	}
	return p
}

func ToSliceStringMapE(i interface{}) ([]map[string]interface{}, error) {
	var m = []map[string]interface{}{}

	switch v := i.(type) {
	case []interface{}:
		for k := range v {
			m = append(m, cast.ToStringMap(v[k]))
		}
		return m, nil
	default:
		return m, fmt.Errorf("unable to cast %#v of type %T to []map[string]interface{}", i, i)
	}
}

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func Ife(statement bool, a, b interface{}) interface{} {
	if statement {
		return a
	}
	return b
}

func ArrayToString(a []int, delim string) string {
	return strings.Trim(strings.Replace(cast.ToString(a), " ", delim, -1), "[]")
	//return strings.Trim(strings.Join(strings.Split(fmt.Sprint(a), " "), delim), "[]")
	//return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(a)), delim), "[]")
}

func GetByRoute(m map[string]interface{}, route ...string) interface{} {
	for _, p := range route {
		x := m[p]
		switch v := x.(type) {
		case map[string]interface{}:
			m = v
		default:
			return v
		}
	}
	return m
}

func ChopOffStringBounded(s string, rightOffset, lengthRemain int) string {
	firstCharIndex := len(s) - rightOffset
	if firstCharIndex < 0 {
		firstCharIndex = 0
	}
	return ChopOffString(s[firstCharIndex:], lengthRemain)
}

func ChopOffString(s string, lengthRemain int) string {
	if len(s) > lengthRemain {
		s = s[:lengthRemain] + "..."
	}
	return s
}

func CurrentTimeMillis() int64 {
	return time.Now().UnixNano() / 1e6
}

func EncodeWindows1251(ba []uint8) []uint8 {
	enc := charmap.Windows1251.NewEncoder()
	out, _ := enc.String(string(ba))
	return []uint8(out)
}

func GetErrorFullInfo(e error) string {
	if e == nil {
		return ""
	}
	var stackHolder stackerror.StackError
	be := errs.FindBaseError(e)
	if be != nil {
		stackHolder = be.Stack
	}
	if stackHolder == nil {
		stackHolder = stackerror.New(e.Error())
	}
	msg := e.Error() + ": " + ToJson(e)
	return fmt.Sprintf("%v\nStack:\n%v", msg, stackHolder.Error())
}

func GetType(x interface{}) string {
	if t := reflect.TypeOf(x); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

func CalcAndGetPtr(f func() interface{}, returnNilPredicate func(interface{}) bool) interface{} {
	res := f()
	if returnNilPredicate(res) {
		return nil
	}
	return &res
}

func ParseMemory(s string) (uint64, error) {
	var v datasize.ByteSize
	err := v.UnmarshalText([]byte(s))
	if err != nil {
		return 0, err
	}
	return v.Bytes(), nil
}

func MD5(arg string) string {
	h := md5.New()
	h.Write([]byte(arg))
	return hex.EncodeToString(h.Sum(nil))
}

func HasElem(s interface{}, elem interface{}) bool {
	if s == nil {
		return false
	}
	arrV := reflect.ValueOf(s)

	if arrV.Kind() == reflect.Slice {
		for i := 0; i < arrV.Len(); i++ {

			// XXX - panics if slice element points to an unexported struct field
			// see https://golang.org/pkg/reflect/#Value.Interface
			if arrV.Index(i).Interface() == elem {
				return true
			}
		}
	}

	return false
}

func EncodeDBBool(b bool) int {
	if b {
		return 1
	}
	return 0
}

func XmlToJson(xmlData []byte) ([]byte, error) {

	var data map[interface{}]interface{}

	err := xml.Unmarshal(xmlData, data)
	if nil != err {
		return nil, errs.NewBaseErrorFromCauseMsg(err, "Error unmarshalling from XML")
	}

	result, err := json.Marshal(data)
	if nil != err {
		return nil, errs.NewBaseErrorFromCauseMsg(err, "Error marshalling to JSON")
	}

	return result, nil
}

func GetFirstInStringMap(m []map[string]interface{}) interface{} {
	if len(m) == 0 {
		return nil
	}
	return m[0]
}

func ContainsStr(s []string, e string, caseInsensitive bool) bool {
	for _, a := range s {
		if caseInsensitive && strings.EqualFold(a, e) || !caseInsensitive && a == e {
			return true
		}
	}
	return false
}

func Contains(s []interface{}, e interface{}) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func CreateFileIfNotExists(filename string) (*os.File, error) {
	if !FileExists(filename) {
		return os.Create(filename)
	}
	return os.NewFile(3, filename), nil
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func ToJson(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func UnmarshalYaml(filename string, out interface{}) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bytes, out)
}

func GetFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}

func StructToMap(i interface{}) (values url.Values) {
	values = url.Values{}
	iVal := reflect.ValueOf(i).Elem()
	typ := iVal.Type()
	for i := 0; i < iVal.NumField(); i++ {
		values.Set(typ.Field(i).Name, fmt.Sprint(iVal.Field(i)))
	}
	return
}

func CalcSha256(arg string) string {
	h := sha256.New()
	h.Write([]byte(arg))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func TryCalcByTryings(maxTryings int, duration time.Duration, f func() (interface{}, error)) interface{} {
	if maxTryings == -1 {
		maxTryings = 2147483647 - 10
	}
	for i := 0; i < maxTryings; i++ {
		r, err := f()
		if err == nil {
			return r
		}
		time.Sleep(duration)
	}

	return nil
}

func ToStringMapE(i interface{}) (map[string]interface{}, error) {
	var m = map[string]interface{}{}

	switch v := i.(type) {
	case map[string]string:
		for k, val := range v {
			m[cast.ToString(k)] = val
		}
		return m, nil
	default:
		return cast.ToStringMapE(i)
	}
}

func ToStringMap(i interface{}) map[string]interface{} {
	v, _ := ToStringMapE(i)
	return v
}

func ByteCountIEC(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func RetrieveIPs(s string) []string {
	ipRegexp := regexp.MustCompile("(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)")
	return ipRegexp.FindAllString(s, -1)
}

func ToStringSliceInts(a []int) []string {
	var r []string
	if a == nil {
		return r
	}
	for _, v := range a {
		r = append(r, cast.ToString(v))
	}
	return r
}

func ToStringSlice(a []interface{}) []string {
	var r []string
	if a == nil {
		return r
	}
	for _, v := range a {
		r = append(r, cast.ToString(v))
	}
	return r
}

func GetFileSize(fileName string) uint64 {
	fi, err := os.Stat(fileName)
	if err != nil {
		return 0
	}
	return uint64(fi.Size())
}

func WriteToFile(filename string, writer func(w *bufio.Writer) error) (*os.File, error) {

	f, err := os.Create(filename)

	if err != nil {
		return nil, err
	}

	w := bufio.NewWriter(f)
	err = writer(w)

	if err != nil {
		return nil, err
	}

	err = w.Flush()
	if err != nil {
		return nil, err
	}

	return f, f.Close()

}

func CloseAndRemove(f *os.File) error {
	err := f.Close()
	if err != nil {
		println(err.Error())
	}
	err = os.Remove(f.Name())
	if err != nil {
		println(err.Error())
	}
	return err
}

func GetFirstElement(a []interface{}) interface{} {
	if a == nil || len(a) == 0 {
		return nil
	}
	return a[0]
}

func GetFirstElementStr(a []string) string {
	if a == nil || len(a) == 0 {
		return ""
	}
	return a[0]
}

func IterateFields(v interface{}, action func(fieldName string, value interface{}, f reflect.Value) error) error {
	return iterateFields(nil, v, "", action)
}

func iterateFields(root, v interface{}, fieldName string, action func(fieldName string, value interface{}, f reflect.Value) error) error {
	f := reflect.Indirect(reflect.ValueOf(v))
	if f.Kind() == reflect.Ptr {
		f = f.Elem()
	}
	if !f.IsValid() {
		return nil
	}
	if f.Kind() == reflect.Slice {
		for k := 0; k < f.Len(); k++ {
			iterateFields(v, f.Index(k).Interface(), fieldName, action)
		}
	} else if f.Kind() == reflect.Struct {
		for k := 0; k < f.NumField(); k++ {
			iterateFields(v, f.Field(k).Interface(), f.Type().Field(k).Name, action)
		}
	} else {
		err := action(fieldName, f.Interface(), reflect.ValueOf(root).Elem().FieldByName(fieldName))
		if err != nil {
			return err
		}
	}
	return nil
}

func FirstEntryStr(a map[string]interface{}) (string, interface{}) {
	for k, v := range a {
		return k, v
	}
	return "", nil
}

func FirstEntry(a map[interface{}]interface{}) (interface{}, interface{}) {
	for k, v := range a {
		return k, v
	}
	return nil, nil
}

type FormatTp struct {
	tp *template.Template
}

func (f FormatTp) Template() *template.Template {
	return f.tp
}

// Exec  Pass in map Fill the predetermined template
func (f FormatTp) Exec(args interface{}) string {
	s := new(strings.Builder)
	f.tp.Execute(s, args)
	return s.String()
}

/*
	Format  Custom naming format, Strictly in accordance with  {{.CUSTOMNAME}}  As a predetermined parameter , Don't write anything else template grammar

usage:

	s = Format("{{.name}} hello.").Exec(map[string]interface{}{
	    "name": "superpig",
	}) // s: superpig hello.
*/
func Format(fmt string) FormatTp {
	temp, _ := template.New("").Parse(fmt)
	return FormatTp{tp: temp}
}

func Concat[T any](first []T, second []T) []T {
	n := len(first)
	return append(first[:n:n], second...)
}

func MapSlice[A, B any](a []A, mapper func(x A) B) []B {
	var r []B
	for _, x := range a {
		r = append(r, mapper(x))
	}
	return r
}

func IsWindows() bool {
	return strings.EqualFold(runtime.GOOS, "windows")
}
