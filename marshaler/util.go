package marshaler

import (
	"encoding/json"
	"reflect"

	"github.com/golang/protobuf/ptypes/any"
	"github.com/mosuka/blast/protobuf"
	"github.com/mosuka/blast/registry"
)

func init() {
	registry.RegisterType("protobuf.LivenessCheckResponse", reflect.TypeOf(protobuf.LivenessCheckResponse{}))
	registry.RegisterType("protobuf.ReadinessCheckResponse", reflect.TypeOf(protobuf.ReadinessCheckResponse{}))
	registry.RegisterType("protobuf.Metadata", reflect.TypeOf(protobuf.Metadata{}))
	registry.RegisterType("protobuf.Node", reflect.TypeOf(protobuf.Node{}))
	registry.RegisterType("protobuf.Cluster", reflect.TypeOf(protobuf.Cluster{}))
	registry.RegisterType("protobuf.JoinRequest", reflect.TypeOf(protobuf.JoinRequest{}))
	registry.RegisterType("protobuf.LeaveRequest", reflect.TypeOf(protobuf.LeaveRequest{}))
	registry.RegisterType("protobuf.NodeResponse", reflect.TypeOf(protobuf.NodeResponse{}))
	registry.RegisterType("protobuf.ClusterResponse", reflect.TypeOf(protobuf.ClusterResponse{}))
	registry.RegisterType("protobuf.GetRequest", reflect.TypeOf(protobuf.GetRequest{}))
	registry.RegisterType("protobuf.GetResponse", reflect.TypeOf(protobuf.GetResponse{}))
	registry.RegisterType("protobuf.SetRequest", reflect.TypeOf(protobuf.SetRequest{}))
	registry.RegisterType("protobuf.DeleteRequest", reflect.TypeOf(protobuf.DeleteRequest{}))
	registry.RegisterType("protobuf.BulkIndexRequest", reflect.TypeOf(protobuf.BulkIndexRequest{}))
	registry.RegisterType("protobuf.BulkDeleteRequest", reflect.TypeOf(protobuf.BulkDeleteRequest{}))
	registry.RegisterType("protobuf.SetMetadataRequest", reflect.TypeOf(protobuf.SetMetadataRequest{}))
	registry.RegisterType("protobuf.DeleteMetadataRequest", reflect.TypeOf(protobuf.DeleteMetadataRequest{}))
	registry.RegisterType("protobuf.Event", reflect.TypeOf(protobuf.Event{}))
	registry.RegisterType("protobuf.WatchResponse", reflect.TypeOf(protobuf.WatchResponse{}))
	registry.RegisterType("protobuf.MetricsResponse", reflect.TypeOf(protobuf.MetricsResponse{}))
	registry.RegisterType("protobuf.Document", reflect.TypeOf(protobuf.Document{}))
	registry.RegisterType("map[string]interface {}", reflect.TypeOf((map[string]interface{})(nil)))
}

func MarshalAny(message *any.Any) (interface{}, error) {
	if message == nil {
		return nil, nil
	}

	typeUrl := message.TypeUrl
	value := message.Value

	instance := registry.TypeInstanceByName(typeUrl)

	if err := json.Unmarshal(value, instance); err != nil {
		return nil, err
	} else {
		return instance, nil
	}

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
