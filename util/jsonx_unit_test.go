// Copyright © 2019 Splunk Inc.
// SPLUNK CONFIDENTIAL – Use or disclosure of this material in whole or in part
// without a valid written license from Splunk Inc. is PROHIBITED.
//

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ManyTypes struct {
	// Various types and methods tags
	Bool                    bool        `json:"bool" methods:"POST,PUT,PATCH"`
	BoolEmpty               bool        `json:"boolEmpty,omitempty" methods:"PUT,PATCH"`
	BoolPointerEmpty        *bool       `json:"boolPointerEmpty,omitempty" methods:"PATCH"`
	Float                   float32     `json:"float" methods:"POST"`
	FloatEmpty              float64     `json:"floatEmpty,omitempty" methods:"POST"`
	FloatPointerEmpty       *float32    `json:"floatPointerEmpty,omitempty" methods:"POST"`
	Int                     int         `json:"int" methods:"POST"`
	IntEmpty                int         `json:"intEmpty,omitempty" methods:"POST"`
	IntPointerArray         [3]*int     `json:"intPointerArray" methods:"POST"`
	IntPointerEmpty         *int        `json:"intPointerEmpty,omitempty" methods:"POST"`
	IntPointerSlice         []*int      `json:"intPointerSlice" methods:"POST,PUT"`
	IntSliceEmpty           []int       `json:"intSliceEmpty,omitempty" methods:"POST,PUT,PATCH"`
	Interface               interface{} `json:"interface" methods:"POST"`
	InterfaceEmpty          interface{} `json:"interfaceEmpty,omitempty" methods:"POST"`
	String                  string      `json:"string" methods:"POST"`
	StringArray             [4]string   `json:"stringArray"` // no methods tag
	StringEmpty             string      `json:"stringEmpty,omitempty" methods:"POST"`
	StringPointerEmpty      *string     `json:"stringPointerEmpty,omitempty" methods:"POST"`
	StringPointerSliceEmpty []*string   `json:"stringPointerSliceEmpty,omitempty"` // no methods tag
	StringSlice             []string    `json:"stringSlice" methods:"POST"`
	// Special cases: tags
	EmptyJSONTag    string `json:"" methods:"POST"`
	EmptyMethodsTag string `json:"emptyMethodsTag" methods:""`
	HyphenJSONTag   string `json:"-,"`
	OmitJSONTag     string `json:"-"`
	MethodTagOnly   string `methods:"GET,POST"`
	OtherTags       string `foo:"foo" bar:"bar"`
	NoTags          string
}

type SimpleA struct {
	A string `json:"a" methods:"POST"`
}

type SimpleB struct {
	B string `json:"b,omitempty" methods:"POST"`
}

type SimpleCD struct {
	C string `json:"c"`
	D string `json:"d,omitempty" methods:"POST"`
}

type Embedded struct {
	SimpleA `json:"ignoredA" methods:"POST"`
}

type EmbeddedPointer struct {
	*SimpleB `json:"ignoredB" methods:"POST"`
}

type EmbeddedMulti struct {
	*Embedded
	EmbeddedPointer
	SimpleCD
}

type NoMethodsTags struct {
	A string `json:"a"`
}

type InvalidMethodsTag struct {
	A string `json:"a" methods:"GET,POTS"`
}

type StringTag struct {
	A string `json:"a,omitempty,string" methods:"POST"`
}

var a = "a"
var b = "b"
var c = "c"
var d = "d"
var one = 1
var two = 2
var three = 3
var yes = true
var no = false
var pi32 = float32(3.141)
var pi64 = float64(3.14159)
var map1 = map[string]interface{}{"a": a, "one": one, "yes": yes, "pi": pi64}
var map2 = map[string]interface{}{"b": b, "two": two, "no": no, "pi": pi32}

var manyTypesAll = ManyTypes{
	Bool:                    no,
	BoolEmpty:               yes,
	BoolPointerEmpty:        &no,
	Float:                   pi32,
	FloatEmpty:              pi64,
	FloatPointerEmpty:       &pi32,
	Int:                     one,
	IntEmpty:                two,
	IntPointerArray:         [3]*int{&one, &two, &three},
	IntPointerEmpty:         &three,
	IntPointerSlice:         []*int{&two, &three},
	IntSliceEmpty:           []int{one, three},
	Interface:               map1,
	InterfaceEmpty:          map2,
	String:                  a,
	StringArray:             [4]string{a, b, c, d},
	StringEmpty:             b,
	StringPointerEmpty:      &c,
	StringPointerSliceEmpty: []*string{&b, &c, &d},
	StringSlice:             []string{b, c},
	EmptyJSONTag:            "EmptyJSONTagValue",
	EmptyMethodsTag:         "EmptyMethodsTagValue",
	HyphenJSONTag:           "HyphenJSONTagValue",
	OmitJSONTag:             "OmitJSONTagValue",
	MethodTagOnly:           "MethodTagOnlyValue",
	OtherTags:               "OtherTagsValue",
	NoTags:                  "NoTagsValue",
}

func TestGetFieldsNil(t *testing.T) {
	fields := getFieldsByTag(nil, "json")
	assert.Len(t, fields, 0)
}

func TestGetFieldsNonStruct(t *testing.T) {
	fields := getFieldsByTag(one, "json")
	assert.Len(t, fields, 0)
}

func TestGetFieldsPointerToNonStruct(t *testing.T) {
	fields := getFieldsByTag(&map1, "json")
	assert.Len(t, fields, 0)
}

func TestGetFieldsByTagManyTypes(t *testing.T) {
	fields := getFieldsByTag(manyTypesAll, "json")
	// flatten to field names
	names := make([]string, len(fields))
	for i, v := range fields {
		names[i] = v.Name
	}
	// Assert that all "json" tagged fields are here
	assert.Contains(t, names, "StringArray")
	assert.Contains(t, names, "StringArray")
	assert.Contains(t, names, "IntPointerArray")
	assert.Contains(t, names, "StringSlice")
	assert.Contains(t, names, "IntPointerSlice")
	assert.Contains(t, names, "IntSliceEmpty")
	assert.Contains(t, names, "StringPointerSliceEmpty")
	assert.Contains(t, names, "Bool")
	assert.Contains(t, names, "BoolEmpty")
	assert.Contains(t, names, "BoolPointerEmpty")
	assert.Contains(t, names, "Float")
	assert.Contains(t, names, "FloatEmpty")
	assert.Contains(t, names, "FloatPointerEmpty")
	assert.Contains(t, names, "Int")
	assert.Contains(t, names, "IntEmpty")
	assert.Contains(t, names, "IntPointerEmpty")
	assert.Contains(t, names, "String")
	assert.Contains(t, names, "StringEmpty")
	assert.Contains(t, names, "StringPointerEmpty")
	assert.Contains(t, names, "Interface")
	assert.Contains(t, names, "InterfaceEmpty")
	assert.Contains(t, names, "EmptyJSONTag")
	assert.Contains(t, names, "EmptyMethodsTag")
	assert.Contains(t, names, "HyphenJSONTag")
	assert.Contains(t, names, "OmitJSONTag")
	// fields without "json" tag should not be here
	assert.NotContains(t, names, "MethodTagOnly")
	assert.NotContains(t, names, "NoTags")
	assert.NotContains(t, names, "OtherTags")
}

func TestMarshalByMethodNil(t *testing.T) {
	// This should marshal only the fields without any methods tags
	_, err := MarshalByMethod(nil, "POST")
	assert.NotNil(t, err)
}

func TestMarshalByMethodNonExistentMethod(t *testing.T) {
	// This should marshal only the fields without any methods tags
	bytes, err := MarshalByMethod(manyTypesAll, "not_a_real_method")
	require.Nil(t, err)
	require.NotEmpty(t, bytes)
	assert.Equal(t, `{"-":"HyphenJSONTagValue","stringArray":["a","b","c","d"],"stringPointerSliceEmpty":["b","c","d"]}`, string(bytes))
}

func TestMarshalByMethodPATCH(t *testing.T) {
	// This should marshal only the fields without any methods tags or with a PATCH method tag
	bytes, err := MarshalByMethod(manyTypesAll, "PATCH")
	require.Nil(t, err)
	require.NotEmpty(t, bytes)
	assert.Equal(t, `{"-":"HyphenJSONTagValue","bool":false,"boolEmpty":true,"boolPointerEmpty":false,"intSliceEmpty":[1,3],"stringArray":["a","b","c","d"],"stringPointerSliceEmpty":["b","c","d"]}`, string(bytes))
}

func TestMarshalByMethodPOST(t *testing.T) {
	// This should marshal only the fields without any methods tags or with a POST method tag
	bytes, err := MarshalByMethod(manyTypesAll, "POST")
	require.Nil(t, err)
	require.NotEmpty(t, bytes)
	assert.Equal(t, `{"-":"HyphenJSONTagValue","EmptyJSONTag":"EmptyJSONTagValue","bool":false,"float":3.141,"floatEmpty":3.14159,"floatPointerEmpty":3.141,"int":1,"intEmpty":2,"intPointerArray":[1,2,3],"intPointerEmpty":3,"intPointerSlice":[2,3],"intSliceEmpty":[1,3],"interface":{"a":"a","one":1,"pi":3.14159,"yes":true},"interfaceEmpty":{"b":"b","no":false,"pi":3.141,"two":2},"string":"a","stringArray":["a","b","c","d"],"stringEmpty":"b","stringPointerEmpty":"c","stringPointerSliceEmpty":["b","c","d"],"stringSlice":["b","c"]}`, string(bytes))
}

func TestMarshalByMethodPUT(t *testing.T) {
	// This should marshal only the fields without any methods tags or with a PUT method tag
	bytes, err := MarshalByMethod(manyTypesAll, "PUT")
	require.Nil(t, err)
	require.NotEmpty(t, bytes)
	assert.Equal(t, `{"-":"HyphenJSONTagValue","bool":false,"boolEmpty":true,"intPointerSlice":[2,3],"intSliceEmpty":[1,3],"stringArray":["a","b","c","d"],"stringPointerSliceEmpty":["b","c","d"]}`, string(bytes))
}

func TestMarshalByMethodOmitEmpty(t *testing.T) {
	// This should marshal only the fields without `omitempty` and without any methods tags or with a POST method tag
	manyTypesEmpty := ManyTypes{}
	bytes, err := MarshalByMethod(manyTypesEmpty, "POST")
	require.Nil(t, err)
	require.NotEmpty(t, bytes)
	assert.Equal(t, `{"-":"","EmptyJSONTag":"","bool":false,"float":0,"int":0,"intPointerArray":[null,null,null],"intPointerSlice":null,"interface":null,"string":"","stringArray":["","","",""],"stringSlice":null}`, string(bytes))
}

func TestMarshalByMethodCaseInsensitive(t *testing.T) {
	// Verify that the case of the method does not matter w.r.t. marshaling behavior
	bytesUpper, err := MarshalByMethod(manyTypesAll, "POST")
	require.Nil(t, err)
	require.NotEmpty(t, bytesUpper)
	bytesLower, err := MarshalByMethod(manyTypesAll, "post")
	require.Nil(t, err)
	require.NotEmpty(t, bytesLower)
	assert.Equal(t, bytesLower, bytesUpper)
}

func TestMarshalByMethodEmbedded(t *testing.T) {
	a := SimpleA{A: "valueA"}
	bytes, err := MarshalByMethod(a, "POST")
	require.Nil(t, err)
	require.NotEmpty(t, bytes)
	assert.Equal(t, `{"a":"valueA"}`, string(bytes))

	b := SimpleB{B: "valueB"}
	bytes, err = MarshalByMethod(b, "POST")
	require.Nil(t, err)
	require.NotEmpty(t, bytes)
	assert.Equal(t, `{"b":"valueB"}`, string(bytes))

	cd := SimpleCD{C: "valueC", D: "valueD"}
	bytes, err = MarshalByMethod(cd, "POST")
	require.Nil(t, err)
	require.NotEmpty(t, bytes)
	assert.Equal(t, `{"c":"valueC","d":"valueD"}`, string(bytes))

	embedded := Embedded{SimpleA: a}
	bytes, err = MarshalByMethod(embedded, "POST")
	require.Nil(t, err)
	require.NotEmpty(t, bytes)
	assert.Equal(t, `{"a":"valueA"}`, string(bytes))

	embeddedPtr := EmbeddedPointer{SimpleB: &b}
	bytes, err = MarshalByMethod(embeddedPtr, "POST")
	require.Nil(t, err)
	require.NotEmpty(t, bytes)
	assert.Equal(t, `{"b":"valueB"}`, string(bytes))

	embeddedMulti := EmbeddedMulti{Embedded: &embedded, EmbeddedPointer: embeddedPtr, SimpleCD: cd}
	bytes, err = MarshalByMethod(embeddedMulti, "POST")
	require.Nil(t, err)
	require.NotEmpty(t, bytes)
	assert.Equal(t, `{"a":"valueA","b":"valueB","c":"valueC","d":"valueD"}`, string(bytes))
}

func TestMarshalByMethodNoMethodsTags(t *testing.T) {
	a := NoMethodsTags{A: "valueA"}
	_, err := MarshalByMethod(a, "POST")
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "should only be used on structs with fields containing at least one `methods:` tag")
}

func TestMarshalByMethodInvalidMethodsTag(t *testing.T) {
	a := InvalidMethodsTag{A: "valueA"}
	_, err := MarshalByMethod(a, "GET")
	// there is no err because GET is found before POTS is encountered
	require.Nil(t, err)
	_, err = MarshalByMethod(a, "POST")
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "`methods:` field tag found with invalid method: \"POTS\"")
}

func TestMarshalByMethodUnsupportedStringTag(t *testing.T) {
	a := StringTag{A: "valueA"}
	_, err := MarshalByMethod(a, "PATCH")
	// there is no err because json field is not read since no PATCH field to marshal
	require.Nil(t, err)
	_, err = MarshalByMethod(a, "POST")
	require.NotNil(t, err)
	assert.Contains(t, err.Error(), "the \"string\" tag (`json:\",string\"`) is not supported")
}
