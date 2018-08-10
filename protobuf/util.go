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
	"encoding/json"
	"reflect"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/registry"
)

func init() {
	registry.RegisterType("interface {}", reflect.TypeOf((map[string]interface{})(nil)))
	registry.RegisterType("map[string]interface {}", reflect.TypeOf((map[string]interface{})(nil)))
	registry.RegisterType("protobuf.PutRequest", reflect.TypeOf(PutRequest{}))
	registry.RegisterType("protobuf.DeleteRequest", reflect.TypeOf(DeleteRequest{}))
	registry.RegisterType("protobuf.BulkRequest", reflect.TypeOf(BulkRequest{}))
	registry.RegisterType("protobuf.JoinRequest", reflect.TypeOf(JoinRequest{}))
	registry.RegisterType("protobuf.LeaveRequest", reflect.TypeOf(LeaveRequest{}))
}

func MarshalAny(message *any.Any) (interface{}, error) {
	if message == nil {
		return nil, nil
	}

	typeUrl := message.TypeUrl
	value := message.Value

	instance := registry.TypeInstanceByName(typeUrl)

	err := json.Unmarshal(value, instance)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func UnmarshalAny(instance interface{}, message *any.Any) error {
	if instance == nil {
		return nil
	}

	value, err := json.Marshal(instance)
	if err != nil {
		return err
	}

	message.TypeUrl = registry.TypeNameByInstance(instance)
	message.Value = value

	return nil
}

// ----------------------------------------

func (m *GetResponse) GetFieldsMap() (map[string]interface{}, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var fieldsInstance interface{}
	if fieldsInstance, err = MarshalAny(m.Fields); err != nil {
		return nil, err
	}

	if fieldsInstance == nil {
		return nil, nil
	}

	return *fieldsInstance.(*map[string]interface{}), err
}

func (m *GetResponse) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var fieldsMap map[string]interface{}
	if fieldsMap, err = m.GetFieldsMap(); err != nil {
		return nil, err
	}

	getResponse := struct {
		Id      string                 `json:"id,omitempty"`
		Fields  map[string]interface{} `json:"fields,omitempty"`
		Success bool                   `json:"success,omitempty"`
		Message string                 `json:"message,omitempty"`
	}{
		Id:      m.Id,
		Fields:  fieldsMap,
		Success: m.Success,
		Message: m.Message,
	}

	var getResponseBytes []byte
	if getResponseBytes, err = json.Marshal(&getResponse); err != nil {
		return nil, err
	}

	return getResponseBytes, nil
}

// ----------------------------------------

func (m *PutRequest) GetFieldsMap() (map[string]interface{}, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var fieldsInstance interface{}
	if fieldsInstance, err = MarshalAny(m.Fields); err != nil {
		return nil, err
	}

	if fieldsInstance == nil {
		return nil, nil
	}

	return *fieldsInstance.(*map[string]interface{}), err
}

func (m *PutRequest) GetFieldsBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var fieldsMap map[string]interface{}
	if fieldsMap, err = m.GetFieldsMap(); err != nil {
		return nil, err
	}

	var fieldsBytes []byte
	if fieldsBytes, err = json.Marshal(fieldsMap); err != nil {
		return nil, err
	}

	return fieldsBytes, nil
}

// ----------------------------------------

func (m *PutResponse) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	putResponse := struct {
		Success bool   `json:"success,omitempty"`
		Message string `json:"message,omitempty"`
	}{
		Success: m.Success,
		Message: m.Message,
	}

	var putResponseBytes []byte
	if putResponseBytes, err = json.Marshal(&putResponse); err != nil {
		return nil, err
	}

	return putResponseBytes, nil
}

// ----------------------------------------

func (m *DeleteResponse) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	deleteResponse := struct {
		Success bool   `json:"success,omitempty"`
		Message string `json:"message,omitempty"`
	}{
		Success: m.Success,
		Message: m.Message,
	}

	var deleteResponseBytes []byte
	if deleteResponseBytes, err = json.Marshal(&deleteResponse); err != nil {
		return nil, err
	}

	return deleteResponseBytes, nil
}

// ----------------------------------------

func (m *Document) SetBytes(documentBytes []byte) error {
	var err error

	if m == nil {
		return nil
	}

	var documentMap map[string]interface{}
	if err = json.Unmarshal(documentBytes, &documentMap); err != nil {
		return err
	}

	m.SetMap(documentMap)

	return nil
}

func (m *Document) SetMap(documentMap map[string]interface{}) error {
	var err error

	if m == nil {
		return nil
	}

	m.Id = documentMap["id"].(string)

	if m.Fields == nil {
		m.Fields = &any.Any{}
	}

	if err = UnmarshalAny(documentMap["fields"].(map[string]interface{}), m.Fields); err != nil {
		return err
	}

	return nil
}

func (m *Document) GetFieldsMap() (map[string]interface{}, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var fieldsInstance interface{}
	if fieldsInstance, err = MarshalAny(m.Fields); err != nil {
		return nil, err
	}

	if fieldsInstance == nil {
		return nil, nil
	}

	return *fieldsInstance.(*map[string]interface{}), err
}

func (m *Document) GetFieldsBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var fieldsMap map[string]interface{}
	if fieldsMap, err = m.GetFieldsMap(); err != nil {
		return nil, err
	}

	var fieldsBytes []byte
	if fieldsBytes, err = json.Marshal(fieldsMap); err != nil {
		return nil, err
	}

	return fieldsBytes, nil
}

func (m *Document) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var fieldsMap map[string]interface{}
	if fieldsMap, err = m.GetFieldsMap(); err != nil {
		return nil, err
	}

	document := struct {
		Id     string                 `json:"id,omitempty"`
		Fields map[string]interface{} `json:"fields,omitempty"`
	}{
		Id:     m.Id,
		Fields: fieldsMap,
	}

	var documentBytes []byte
	if documentBytes, err = json.Marshal(&document); err != nil {
		return nil, err
	}

	return documentBytes, nil
}

func (m *Document) GetMap() (map[string]interface{}, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var documentBytes []byte
	if documentBytes, err = m.GetBytes(); err != nil {
		return nil, err
	}

	var documentMap map[string]interface{}
	if err = json.Unmarshal(documentBytes, &documentMap); err != nil {
		return nil, err
	}

	return documentMap, nil
}

// ----------------------------------------

func (m *UpdateRequest) SetBytes(updateRequestBytes []byte) error {
	var err error

	if m == nil {
		return nil
	}

	var updateRequestMap map[string]interface{}
	if err = json.Unmarshal(updateRequestBytes, &updateRequestMap); err != nil {
		return err
	}

	m.SetMap(updateRequestMap)

	return nil
}

func (m *UpdateRequest) SetMap(updateRequestMap map[string]interface{}) error {
	var err error

	if m == nil {
		return nil
	}

	if updateRequestMap == nil {
		return nil
	}

	if updateRequestType, ok := updateRequestMap["type"]; ok {
		m.Type = UpdateRequest_Type(UpdateRequest_Type_value[updateRequestType.(string)])
	}

	if updateRequestDocument, ok := updateRequestMap["document"]; ok {
		m.Document = &Document{}
		if err = m.Document.SetMap(updateRequestDocument.(map[string]interface{})); err != nil {
			return err
		}
	}

	return nil
}

func (m *UpdateRequest) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var documentMap map[string]interface{}
	if documentMap, err = m.Document.GetMap(); err != nil {
		return nil, err
	}

	updateRequest := struct {
		Type     string                 `json:"type,omitempty"`
		Document map[string]interface{} `json:"document,omitempty"`
	}{
		Type:     m.Type.String(),
		Document: documentMap,
	}

	var updateRequestBytes []byte
	if updateRequestBytes, err = json.Marshal(&updateRequest); err != nil {
		return nil, err
	}

	return updateRequestBytes, nil
}

func (m *UpdateRequest) GetMap() (map[string]interface{}, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var updateRequestBytes []byte
	if updateRequestBytes, err = m.GetBytes(); err != nil {
		return nil, err
	}

	var updateRequestMap map[string]interface{}
	if err = json.Unmarshal(updateRequestBytes, &updateRequestMap); err != nil {
		return nil, err
	}

	return updateRequestMap, nil
}

// ----------------------------------------

func (m *BulkResponse) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	bulkResponse := struct {
		PutCount    int    `json:"put_count,omitempty"`
		DeleteCount int    `json:"delete_count,omitempty"`
		Success     bool   `json:"success,omitempty"`
		Message     string `json:"message,omitempty"`
	}{
		PutCount:    int(m.PutCount),
		DeleteCount: int(m.DeleteCount),
		Success:     m.Success,
		Message:     m.Message,
	}

	var bulkResponseBytes []byte
	if bulkResponseBytes, err = json.Marshal(&bulkResponse); err != nil {
		return nil, err
	}

	return bulkResponseBytes, nil
}

// ----------------------------------------

func (m *SearchRequest) SetBytes(searchRequestBytes []byte) error {
	var err error

	if m == nil {
		return nil
	}

	var searchRequestMap map[string]interface{}
	if err = json.Unmarshal(searchRequestBytes, &searchRequestMap); err != nil {
		return err
	}

	if err = m.SetMap(searchRequestMap); err != nil {
		return err
	}

	return nil
}

func (m *SearchRequest) SetMap(searchRequestMap map[string]interface{}) error {
	var err error

	if m == nil {
		return nil
	}

	if m.SearchRequest == nil {
		m.SearchRequest = &any.Any{}
		if err = UnmarshalAny(searchRequestMap["search_request"], m.SearchRequest); err != nil {
			return err
		}
	}

	return nil
}

// ----------------------------------------

func (m *SearchResponse) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var searchResultInstance interface{}
	if m.SearchResult != nil {
		if searchResultInstance, err = MarshalAny(m.SearchResult); err != nil {
			return nil, err
		}
	}

	searchResponse := struct {
		SearchResult map[string]interface{} `json:"search_result,omitempty"`
		Success      bool                   `json:"success,omitempty"`
		Message      string                 `json:"message,omitempty"`
	}{
		SearchResult: *searchResultInstance.(*map[string]interface{}),
		Success:      m.Success,
		Message:      m.Message,
	}

	var searchResponseBytes []byte
	if searchResponseBytes, err = json.Marshal(searchResponse); err != nil {
		return nil, err
	}

	return searchResponseBytes, nil
}

// ----------------------------------------

func (m *Metadata) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	metadata := struct {
		GrpcAddress string `json:"grpc_address,omitempty"`
		HttpAddress string `json:"http_address,omitempty"`
	}{
		GrpcAddress: m.GrpcAddress,
		HttpAddress: m.HttpAddress,
	}

	var metadataBytes []byte
	if metadataBytes, err = json.Marshal(metadata); err != nil {
		return nil, err
	}

	return metadataBytes, nil
}

func (m *Metadata) GetMap() (map[string]interface{}, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var metadataBytes []byte
	if metadataBytes, err = m.GetBytes(); err != nil {
		return nil, err
	}
	var metadataMap map[string]interface{}
	if err = json.Unmarshal(metadataBytes, &metadataMap); err != nil {
		return nil, err
	}

	return metadataMap, nil
}

// ----------------------------------------

func (m *JoinResponse) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	joinResponse := struct {
		Success bool   `json:"success,omitempty"`
		Message string `json:"message,omitempty"`
	}{
		Success: m.Success,
		Message: m.Message,
	}

	var joinResponseBytes []byte
	if joinResponseBytes, err = json.Marshal(joinResponse); err != nil {
		return nil, err
	}

	return joinResponseBytes, nil
}

// ----------------------------------------

func (m *LeaveResponse) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	leaveResponse := struct {
		Success bool   `json:"success,omitempty"`
		Message string `json:"message,omitempty"`
	}{
		Success: m.Success,
		Message: m.Message,
	}

	var leaveResponseBytes []byte
	if leaveResponseBytes, err = json.Marshal(leaveResponse); err != nil {
		return nil, err
	}

	return leaveResponseBytes, nil
}

// ----------------------------------------

func (m *Peer) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var metadataMap map[string]interface{}
	if metadataMap, err = m.Metadata.GetMap(); err != nil {
		return nil, err
	}

	peer := struct {
		NodeId   string                 `json:"node_id,omitempty"`
		Address  string                 `json:"address,omitempty"`
		Leader   bool                   `json:"leader,omitempty"`
		Metadata map[string]interface{} `json:"metadata,omitempty"`
	}{
		NodeId:   m.NodeId,
		Address:  m.Address,
		Leader:   m.Leader,
		Metadata: metadataMap,
	}

	var peerBytes []byte
	if peerBytes, err = json.Marshal(peer); err != nil {
		return nil, err
	}

	return peerBytes, nil
}

func (m *Peer) GetMap() (map[string]interface{}, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	var peerBytes []byte
	if peerBytes, err = m.GetBytes(); err != nil {
		return nil, err
	}
	var peerMap map[string]interface{}
	if err = json.Unmarshal(peerBytes, &peerMap); err != nil {
		return nil, err
	}

	return peerMap, nil
}

// ----------------------------------------

func (m *PeersResponse) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	peers := make([]map[string]interface{}, 0)
	for _, peer := range m.Peers {
		var peerMap map[string]interface{}
		if peerMap, err = peer.GetMap(); err != nil {
			return nil, err
		}
		peers = append(peers, peerMap)
	}

	peersResponse := struct {
		Peers   []map[string]interface{} `json:"peers,omitempty"`
		Success bool                     `json:"success,omitempty"`
		Message string                   `json:"message,omitempty"`
	}{
		Peers:   peers,
		Success: m.Success,
		Message: m.Message,
	}

	var peersResponseBytes []byte
	if peersResponseBytes, err = json.Marshal(peersResponse); err != nil {
		return nil, err
	}

	return peersResponseBytes, nil
}

// ----------------------------------------

func (m *SnapshotResponse) GetBytes() ([]byte, error) {
	var err error

	if m == nil {
		return nil, nil
	}

	snapshotResponse := struct {
		Success bool   `json:"success,omitempty"`
		Message string `json:"message,omitempty"`
	}{
		Success: m.Success,
		Message: m.Message,
	}

	var snapshotResponseBytes []byte
	if snapshotResponseBytes, err = json.Marshal(snapshotResponse); err != nil {
		return nil, err
	}

	return snapshotResponseBytes, nil
}
