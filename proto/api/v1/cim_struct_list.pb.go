package apiv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

type CIMElementList struct {
}

func (x *CIMElementList) Reset() {}

func (x *CIMElementList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CIMElementList) ProtoMessage() {}

var file_CIMElementList_proto_msgTypes = make([]protoimpl.MessageInfo, 1)

func (x *CIMElementList) ProtoReflect() protoreflect.Message {
	mi := &file_CIMElementList_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}
