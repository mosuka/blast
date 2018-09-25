//  Copyright (c) 2018 Minoru Osuka
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 		http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package protobuf

import (
	"testing"

	"github.com/golang/protobuf/ptypes/any"
)

func TestMarshalAny(t *testing.T) {
	a := &any.Any{
		TypeUrl: "map[string]interface {}",
		Value:   []byte(`{"name":"foo"}`),
	}

	i, err := MarshalAny(a)
	if err != nil {
		t.Errorf("unexpected error. %v", a)
	}

	d := *i.(*map[string]interface{})

	expectedName := "foo"
	actualName := d["name"]
	if expectedName != actualName {
		t.Errorf("unexpected error.  expected %s, actual %s", expectedName, actualName)
	}
}

func TestUnmarshalAny(t *testing.T) {
	m := map[string]interface{}{
		"name": "foo",
	}

	a := &any.Any{}
	err := UnmarshalAny(&m, a)
	if err != nil {
		t.Errorf("unexpected error. %v", m)
	}

	expectedTypeUrl := "map[string]interface {}"
	actualTypeUrl := a.TypeUrl
	if expectedTypeUrl != actualTypeUrl {
		t.Errorf("unexpected error.  expected %s, actual %s", expectedTypeUrl, actualTypeUrl)
	}

	expectedValue := `{"name":"foo"}`
	actualValue := a.Value
	if string(expectedValue) != string(actualValue) {
		t.Errorf("unexpected error.  expected %s, actual %s", expectedValue, actualValue)
	}
}

//func TestGetResponse_GetFieldsAsMap(t *testing.T) {
//	getResponse := &GetResponse{
//		Fields: &any.Any{
//			TypeUrl: "map[string]interface {}",
//			Value:   []byte(`{"name":"foo"}`),
//		},
//	}
//
//	m, err := getResponse.GetFieldsMap()
//	if err != nil {
//		t.Errorf("unexpected error. %v", getResponse)
//	}
//
//	expectedName := "foo"
//	actualName := m["name"]
//	if expectedName != actualName {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedName, actualName)
//	}
//}
//
//func TestGetResponse_Bytes(t *testing.T) {
//	getResponse := &GetResponse{
//		Success: true,
//		Message: "Test",
//		Fields: &any.Any{
//			TypeUrl: "map[string]interface {}",
//			Value:   []byte(`{"name":"foo"}`),
//		},
//	}
//
//	b, err := getResponse.GetBytes()
//	if err != nil {
//		t.Errorf("unexpected error. %v", getResponse)
//	}
//
//	expected := []byte(`{"fields":{"name":"foo"},"success":true,"message":"Test"}`)
//	actual := b
//	if string(expected) != string(actual) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expected, actual)
//	}
//}
//
//func TestPutRequest_GetFieldsAsMap(t *testing.T) {
//	putRequest := &PutRequest{
//		Fields: &any.Any{
//			TypeUrl: "map[string]interface {}",
//			Value:   []byte(`{"name":"foo"}`),
//		},
//	}
//
//	m, err := putRequest.GetFieldsMap()
//	if err != nil {
//		t.Errorf("unexpected error. %v", putRequest)
//	}
//
//	expectedName := "foo"
//	actualName := m["name"]
//	if expectedName != actualName {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedName, actualName)
//	}
//}
//
//func TestPutRequest_GetFieldsAsBytes(t *testing.T) {
//	putRequest := &PutRequest{
//		Fields: &any.Any{
//			TypeUrl: "map[string]interface {}",
//			Value:   []byte(`{"name":"foo"}`),
//		},
//	}
//
//	b, err := putRequest.GetFieldsBytes()
//	if err != nil {
//		t.Errorf("unexpected error. %v", putRequest)
//	}
//
//	expected := []byte(`{"name":"foo"}`)
//	actual := b
//	if string(expected) != string(actual) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expected, actual)
//	}
//}
//
//func TestPutResponse_Bytes(t *testing.T) {
//	putResponse := &PutResponse{
//		Message: "test",
//		Success: true,
//	}
//
//	b, err := putResponse.GetBytes()
//	if err != nil {
//		t.Errorf("unexpected error. %v", putResponse)
//	}
//
//	expected := []byte(`{"success":true,"message":"test"}`)
//	actual := b
//	if string(expected) != string(actual) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expected, actual)
//	}
//}
//
//func TestDeleteResponse_Bytes(t *testing.T) {
//	deleteResponse := &DeleteResponse{
//		Message: "test",
//		Success: true,
//	}
//
//	b, err := deleteResponse.GetBytes()
//	if err != nil {
//		t.Errorf("unexpected error. %v", deleteResponse)
//	}
//
//	expected := []byte(`{"success":true,"message":"test"}`)
//	actual := b
//	if string(expected) != string(actual) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expected, actual)
//	}
//}
//
//func TestDocument_FromBytes(t *testing.T) {
//	b := []byte(`{"id":"1","fields":{"name":"foo"}}`)
//
//	doc := &Document{}
//	err := doc.SetBytes(b)
//	if err != nil {
//		t.Errorf("unexpected error. %v", b)
//	}
//
//	expectedId := "1"
//	actualId := doc.Id
//	if expectedId != actualId {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedId, actualId)
//	}
//
//	f := doc.Fields
//
//	expectedType := "interface {}"
//	actualType := f.TypeUrl
//	if expectedType != actualType {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedType, actualType)
//	}
//
//	expectedValue := `{"name":"foo"}`
//	actualValue := f.Value
//	if string(expectedValue) != string(actualValue) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedValue, actualValue)
//	}
//}
//
//func TestDocument_FromMap(t *testing.T) {
//	m := map[string]interface{}{
//		"id": "1",
//		"fields": map[string]interface{}{
//			"name": "foo",
//		},
//	}
//
//	doc := &Document{}
//	err := doc.SetMap(m)
//	if err != nil {
//		t.Errorf("unexpected error. %v", m)
//	}
//
//	expectedId := "1"
//	actualId := doc.Id
//	if expectedId != actualId {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedId, actualId)
//	}
//
//	f := doc.Fields
//
//	expectedType := "interface {}"
//	actualType := f.TypeUrl
//	if expectedType != actualType {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedType, actualType)
//	}
//
//	expectedValue := `{"name":"foo"}`
//	actualValue := f.Value
//	if string(expectedValue) != string(actualValue) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedValue, actualValue)
//	}
//}
//
//func TestDocument_GetFieldsAsMap(t *testing.T) {
//	doc := &Document{
//		Id: "1",
//		Fields: &any.Any{
//			TypeUrl: "map[string]interface {}",
//			Value:   []byte(`{"1": "a"}`),
//		},
//	}
//
//	m, err := doc.GetFieldsMap()
//	if err != nil {
//		t.Errorf("unexpected error. %v", doc)
//	}
//
//	expectedValue := "a"
//	actualValue := m["1"]
//	if expectedValue != actualValue {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedValue, actualValue)
//	}
//
//}
//
//func TestDocument_Bytes(t *testing.T) {
//	doc := &Document{
//		Id: "1",
//		Fields: &any.Any{
//			TypeUrl: "map[string]interface {}",
//			Value:   []byte(`{"1": "a"}`),
//		},
//	}
//
//	actual, err := doc.GetBytes()
//	if err != nil {
//		t.Errorf("unexpected error. %v", doc)
//	}
//
//	expected := []byte(`{"id":"1","fields":{"1":"a"}}`)
//	if string(expected) != string(actual) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expected, actual)
//	}
//}
//
//func TestDocument_Map(t *testing.T) {
//	doc := &Document{
//		Id: "1",
//		Fields: &any.Any{
//			TypeUrl: "map[string]interface {}",
//			Value:   []byte(`{"1": "a"}`),
//		},
//	}
//
//	m, err := doc.GetMap()
//	if err != nil {
//		t.Errorf("unexpected error. %v", doc)
//	}
//
//	expectedId := "1"
//	actualId := m["id"]
//	if expectedId != actualId {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedId, actualId)
//	}
//
//	fields := m["fields"].(map[string]interface{})
//
//	expectedValue := "a"
//	actualValue := fields["1"]
//	if expectedValue != actualValue {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedValue, actualValue)
//	}
//}
//
//func TestUpdateRequest_FromBytes(t *testing.T) {
//	b := []byte(`{"type":"PUT","document":{"id":"1","fields":{"name":"foo"}}}`)
//
//	updateRequest := &UpdateRequest{}
//	err := updateRequest.SetBytes(b)
//	if err != nil {
//		t.Errorf("unexpected error. %v", b)
//	}
//
//	expectedType := "PUT"
//	actualType := updateRequest.Type.String()
//	if expectedType != actualType {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedType, actualType)
//	}
//
//	doc := updateRequest.Document
//
//	expectedId := "1"
//	actualId := doc.Id
//	if expectedId != actualId {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedId, actualId)
//	}
//
//	fields := doc.Fields
//
//	expectedTypeUrl := "interface {}"
//	actualTypeUrl := fields.TypeUrl
//	if expectedTypeUrl != actualTypeUrl {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedTypeUrl, actualTypeUrl)
//	}
//
//	expectedValue := `{"name":"foo"}`
//	actualValue := fields.Value
//	if string(expectedValue) != string(actualValue) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedValue, actualValue)
//	}
//}
//
//func TestUpdateRequest_FromMap(t *testing.T) {
//	m := map[string]interface{}{
//		"type": "PUT",
//		"document": map[string]interface{}{
//			"id": "1",
//			"fields": map[string]interface{}{
//				"name": "foo",
//			},
//		},
//	}
//
//	updateRequest := &UpdateRequest{}
//	err := updateRequest.SetMap(m)
//	if err != nil {
//		t.Errorf("unexpected error. %v", m)
//	}
//
//	expectedType := "PUT"
//	actualType := updateRequest.Type.String()
//	if expectedType != actualType {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedType, actualType)
//	}
//
//	doc := updateRequest.Document
//
//	expectedId := "1"
//	actualId := doc.Id
//	if expectedId != actualId {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedId, actualId)
//	}
//
//	fields := doc.Fields
//
//	expectedTypeUrl := "interface {}"
//	actualTypeUrl := fields.TypeUrl
//	if expectedTypeUrl != actualTypeUrl {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedTypeUrl, actualTypeUrl)
//	}
//
//	expectedValue := `{"name":"foo"}`
//	actualValue := fields.Value
//	if string(expectedValue) != string(actualValue) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedValue, actualValue)
//	}
//}
//
//func TestUpdateRequest_Bytes(t *testing.T) {
//	b := []byte(`{"type":"PUT","document":{"fields":{"name":"foo"},"id":"1"}}`)
//
//	updateRequest := &UpdateRequest{}
//	err := updateRequest.SetBytes(b)
//	if err != nil {
//		t.Errorf("unexpected error. %v", b)
//	}
//
//	u, err := updateRequest.GetBytes()
//	if err != nil {
//		t.Errorf("unexpected error. %v", b)
//	}
//
//	expected := b
//	actual := u
//	if string(expected) != string(actual) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expected, actual)
//	}
//}
//
//func TestUpdateRequest_Map(t *testing.T) {
//	m := map[string]interface{}{
//		"type": "PUT",
//		"document": map[string]interface{}{
//			"id": "1",
//			"fields": map[string]interface{}{
//				"name": "foo",
//			},
//		},
//	}
//
//	updateRequest := &UpdateRequest{}
//	err := updateRequest.SetMap(m)
//	if err != nil {
//		t.Errorf("unexpected error. %v", m)
//	}
//
//	u, err := updateRequest.GetMap()
//	if err != nil {
//		t.Errorf("unexpected error. %v", m)
//	}
//
//	expectedType := "PUT"
//	actualType := u["type"].(string)
//	if expectedType != actualType {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedType, actualType)
//	}
//
//	expectedId := "1"
//	actualId := updateRequest.Document.Id
//	if expectedId != actualId {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedId, actualId)
//	}
//
//	expectedFieldsTypeUrl := `interface {}`
//	actualFieldsTypeUrl := updateRequest.Document.Fields.TypeUrl
//	if expectedFieldsTypeUrl != actualFieldsTypeUrl {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedFieldsTypeUrl, actualFieldsTypeUrl)
//	}
//
//	expectedFieldsValue := []byte(`{"name":"foo"}`)
//	actualFieldsValue := updateRequest.Document.Fields.Value
//	if string(expectedFieldsValue) != string(actualFieldsValue) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedFieldsValue, actualFieldsValue)
//	}
//}
//
//func TestBulkResponse_Bytes(t *testing.T) {
//	bulkResponse := &BulkResponse{
//		Success:     true,
//		Message:     "test",
//		PutCount:    1,
//		DeleteCount: 2,
//	}
//
//	b, err := bulkResponse.GetBytes()
//	if err != nil {
//		t.Errorf("unexpected error. %v", bulkResponse)
//	}
//
//	expected := []byte(`{"put_count":1,"delete_count":2,"success":true,"message":"test"}`)
//	actual := b
//	if string(expected) != string(actual) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expected, actual)
//	}
//}
//
//func TestSearchRequest_FromBytes(t *testing.T) {
//	b := []byte(`{"search_request":{"facets":{"Category count":{"field":"category","size":10},"Popularity range":{"field":"popularity","numeric_ranges":[{"max":1,"name":"less than 1"},{"max":2,"min":1,"name":"more than or equal to 1 and less than 2"},{"max":3,"min":2,"name":"more than or equal to 2 and less than 3"},{"max":4,"min":3,"name":"more than or equal to 3 and less than 4"},{"max":5,"min":4,"name":"more than or equal to 4 and less than 5"},{"min":5,"name":"more than or equal to 5"}],"size":10},"Release date range":{"date_ranges":[{"end":"2010-12-31T23:59:59Z","name":"2001 - 2010","start":"2001-01-01T00:00:00Z"},{"end":"2020-12-31T23:59:59Z","name":"2011 - 2020","start":"2011-01-01T00:00:00Z"}],"field":"release","size":10}},"fields":["*"],"from":0,"highlight":{"fields":["name","description"],"style":"html"},"query":{"query":"name:*"},"size":10,"sort":["-_score"]}}`)
//
//	searchRequest := SearchRequest{}
//	searchRequest.SetBytes(b)
//
//	expectedTypeUrl := "interface {}"
//	actualTypeUrl := searchRequest.SearchRequest.TypeUrl
//	if expectedTypeUrl != actualTypeUrl {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedTypeUrl, actualTypeUrl)
//	}
//
//	expectedValue := []byte(`{"facets":{"Category count":{"field":"category","size":10},"Popularity range":{"field":"popularity","numeric_ranges":[{"max":1,"name":"less than 1"},{"max":2,"min":1,"name":"more than or equal to 1 and less than 2"},{"max":3,"min":2,"name":"more than or equal to 2 and less than 3"},{"max":4,"min":3,"name":"more than or equal to 3 and less than 4"},{"max":5,"min":4,"name":"more than or equal to 4 and less than 5"},{"min":5,"name":"more than or equal to 5"}],"size":10},"Release date range":{"date_ranges":[{"end":"2010-12-31T23:59:59Z","name":"2001 - 2010","start":"2001-01-01T00:00:00Z"},{"end":"2020-12-31T23:59:59Z","name":"2011 - 2020","start":"2011-01-01T00:00:00Z"}],"field":"release","size":10}},"fields":["*"],"from":0,"highlight":{"fields":["name","description"],"style":"html"},"query":{"query":"name:*"},"size":10,"sort":["-_score"]}`)
//	actualValue := searchRequest.SearchRequest.Value
//	if string(expectedValue) != string(actualValue) {
//		t.Errorf("unexpected error.  expected %s, actual %s", expectedValue, actualValue)
//	}
//}
