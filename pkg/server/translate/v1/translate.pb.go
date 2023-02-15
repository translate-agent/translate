// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: translate/v1/translate.proto

package translatev1

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Schema int32

const (
	Schema_UNSPECIFIED   Schema = 0
	Schema_NG_LOCALISE   Schema = 1
	Schema_NGX_TRANSLATE Schema = 2
	Schema_GO            Schema = 3
	Schema_ARB           Schema = 4
)

// Enum value maps for Schema.
var (
	Schema_name = map[int32]string{
		0: "UNSPECIFIED",
		1: "NG_LOCALISE",
		2: "NGX_TRANSLATE",
		3: "GO",
		4: "ARB",
	}
	Schema_value = map[string]int32{
		"UNSPECIFIED":   0,
		"NG_LOCALISE":   1,
		"NGX_TRANSLATE": 2,
		"GO":            3,
		"ARB":           4,
	}
)

func (x Schema) Enum() *Schema {
	p := new(Schema)
	*p = x
	return p
}

func (x Schema) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Schema) Descriptor() protoreflect.EnumDescriptor {
	return file_translate_v1_translate_proto_enumTypes[0].Descriptor()
}

func (Schema) Type() protoreflect.EnumType {
	return &file_translate_v1_translate_proto_enumTypes[0]
}

func (x Schema) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Schema.Descriptor instead.
func (Schema) EnumDescriptor() ([]byte, []int) {
	return file_translate_v1_translate_proto_rawDescGZIP(), []int{0}
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Message     string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
	Description string `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	Fuzzy       bool   `protobuf:"varint,4,opt,name=fuzzy,proto3" json:"fuzzy,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_translate_v1_translate_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_translate_v1_translate_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_translate_v1_translate_proto_rawDescGZIP(), []int{0}
}

func (x *Message) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Message) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *Message) GetDescription() string {
	if x != nil {
		return x.Description
	}
	return ""
}

func (x *Message) GetFuzzy() bool {
	if x != nil {
		return x.Fuzzy
	}
	return false
}

type Messages struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Language string     `protobuf:"bytes,1,opt,name=language,proto3" json:"language,omitempty"`
	Messages []*Message `protobuf:"bytes,2,rep,name=messages,proto3" json:"messages,omitempty"`
}

func (x *Messages) Reset() {
	*x = Messages{}
	if protoimpl.UnsafeEnabled {
		mi := &file_translate_v1_translate_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Messages) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Messages) ProtoMessage() {}

func (x *Messages) ProtoReflect() protoreflect.Message {
	mi := &file_translate_v1_translate_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Messages.ProtoReflect.Descriptor instead.
func (*Messages) Descriptor() ([]byte, []int) {
	return file_translate_v1_translate_proto_rawDescGZIP(), []int{1}
}

func (x *Messages) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

func (x *Messages) GetMessages() []*Message {
	if x != nil {
		return x.Messages
	}
	return nil
}

type UploadTranslationFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Language string `protobuf:"bytes,1,opt,name=language,proto3" json:"language,omitempty"`
	Data     []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Schema   Schema `protobuf:"varint,3,opt,name=schema,proto3,enum=translate.v1.Schema" json:"schema,omitempty"`
}

func (x *UploadTranslationFileRequest) Reset() {
	*x = UploadTranslationFileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_translate_v1_translate_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadTranslationFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadTranslationFileRequest) ProtoMessage() {}

func (x *UploadTranslationFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_translate_v1_translate_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadTranslationFileRequest.ProtoReflect.Descriptor instead.
func (*UploadTranslationFileRequest) Descriptor() ([]byte, []int) {
	return file_translate_v1_translate_proto_rawDescGZIP(), []int{2}
}

func (x *UploadTranslationFileRequest) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

func (x *UploadTranslationFileRequest) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

func (x *UploadTranslationFileRequest) GetSchema() Schema {
	if x != nil {
		return x.Schema
	}
	return Schema_UNSPECIFIED
}

type DownloadTranslationFileRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Language string `protobuf:"bytes,1,opt,name=language,proto3" json:"language,omitempty"`
	Schema   Schema `protobuf:"varint,2,opt,name=schema,proto3,enum=translate.v1.Schema" json:"schema,omitempty"`
}

func (x *DownloadTranslationFileRequest) Reset() {
	*x = DownloadTranslationFileRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_translate_v1_translate_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloadTranslationFileRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadTranslationFileRequest) ProtoMessage() {}

func (x *DownloadTranslationFileRequest) ProtoReflect() protoreflect.Message {
	mi := &file_translate_v1_translate_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadTranslationFileRequest.ProtoReflect.Descriptor instead.
func (*DownloadTranslationFileRequest) Descriptor() ([]byte, []int) {
	return file_translate_v1_translate_proto_rawDescGZIP(), []int{3}
}

func (x *DownloadTranslationFileRequest) GetLanguage() string {
	if x != nil {
		return x.Language
	}
	return ""
}

func (x *DownloadTranslationFileRequest) GetSchema() Schema {
	if x != nil {
		return x.Schema
	}
	return Schema_UNSPECIFIED
}

type DownloadTranslationFileResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *DownloadTranslationFileResponse) Reset() {
	*x = DownloadTranslationFileResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_translate_v1_translate_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloadTranslationFileResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadTranslationFileResponse) ProtoMessage() {}

func (x *DownloadTranslationFileResponse) ProtoReflect() protoreflect.Message {
	mi := &file_translate_v1_translate_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadTranslationFileResponse.ProtoReflect.Descriptor instead.
func (*DownloadTranslationFileResponse) Descriptor() ([]byte, []int) {
	return file_translate_v1_translate_proto_rawDescGZIP(), []int{4}
}

func (x *DownloadTranslationFileResponse) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

var File_translate_v1_translate_proto protoreflect.FileDescriptor

var file_translate_v1_translate_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x31, 0x2f, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x1a, 0x1c, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67,
	0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74,
	0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x6b, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02,
	0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x20, 0x0a, 0x0b,
	0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x14,
	0x0a, 0x05, 0x66, 0x75, 0x7a, 0x7a, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x05, 0x66,
	0x75, 0x7a, 0x7a, 0x79, 0x22, 0x59, 0x0a, 0x08, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73,
	0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x12, 0x31, 0x0a, 0x08,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15,
	0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x08, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x73, 0x22,
	0x7c, 0x0a, 0x1c, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12,
	0x2c, 0x0a, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0e, 0x32,
	0x14, 0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x52, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x22, 0x6a, 0x0a,
	0x1e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12,
	0x1a, 0x0a, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x12, 0x2c, 0x0a, 0x06, 0x73,
	0x63, 0x68, 0x65, 0x6d, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x14, 0x2e, 0x74, 0x72,
	0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x53, 0x63, 0x68, 0x65, 0x6d,
	0x61, 0x52, 0x06, 0x73, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x22, 0x35, 0x0a, 0x1f, 0x44, 0x6f, 0x77,
	0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x46, 0x69, 0x6c, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x2a, 0x4e, 0x0a, 0x06, 0x53, 0x63, 0x68, 0x65, 0x6d, 0x61, 0x12, 0x0f, 0x0a, 0x0b, 0x55, 0x4e,
	0x53, 0x50, 0x45, 0x43, 0x49, 0x46, 0x49, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0f, 0x0a, 0x0b, 0x4e,
	0x47, 0x5f, 0x4c, 0x4f, 0x43, 0x41, 0x4c, 0x49, 0x53, 0x45, 0x10, 0x01, 0x12, 0x11, 0x0a, 0x0d,
	0x4e, 0x47, 0x58, 0x5f, 0x54, 0x52, 0x41, 0x4e, 0x53, 0x4c, 0x41, 0x54, 0x45, 0x10, 0x02, 0x12,
	0x06, 0x0a, 0x02, 0x47, 0x4f, 0x10, 0x03, 0x12, 0x07, 0x0a, 0x03, 0x41, 0x52, 0x42, 0x10, 0x04,
	0x32, 0xa4, 0x02, 0x0a, 0x10, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x79, 0x0a, 0x15, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x54,
	0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x2a,
	0x2e, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x70,
	0x6c, 0x6f, 0x61, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46,
	0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x22, 0x1c, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x16, 0x1a, 0x14, 0x2f, 0x76, 0x31, 0x2f,
	0x66, 0x69, 0x6c, 0x65, 0x73, 0x2f, 0x7b, 0x6c, 0x61, 0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x7d,
	0x12, 0x94, 0x01, 0x0a, 0x17, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x54, 0x72, 0x61,
	0x6e, 0x73, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x6c, 0x65, 0x12, 0x2c, 0x2e, 0x74,
	0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x6f, 0x77, 0x6e,
	0x6c, 0x6f, 0x61, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46,
	0x69, 0x6c, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2d, 0x2e, 0x74, 0x72, 0x61,
	0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x2e, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f,
	0x61, 0x64, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x46, 0x69, 0x6c,
	0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x1c, 0x82, 0xd3, 0xe4, 0x93, 0x02,
	0x16, 0x12, 0x14, 0x2f, 0x76, 0x31, 0x2f, 0x66, 0x69, 0x6c, 0x65, 0x73, 0x2f, 0x7b, 0x6c, 0x61,
	0x6e, 0x67, 0x75, 0x61, 0x67, 0x65, 0x7d, 0x42, 0xc1, 0x01, 0x0a, 0x10, 0x63, 0x6f, 0x6d, 0x2e,
	0x74, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x76, 0x31, 0x42, 0x0e, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x4c,
	0x67, 0x6f, 0x2e, 0x65, 0x78, 0x70, 0x65, 0x63, 0x74, 0x2e, 0x64, 0x69, 0x67, 0x69, 0x74, 0x61,
	0x6c, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x70, 0x6b, 0x67, 0x2f,
	0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65,
	0x2f, 0x76, 0x31, 0x2f, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x31,
	0x3b, 0x74, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x54,
	0x58, 0x58, 0xaa, 0x02, 0x0c, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x56,
	0x31, 0xca, 0x02, 0x0c, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x5c, 0x56, 0x31,
	0xe2, 0x02, 0x18, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x5c, 0x56, 0x31, 0x5c,
	0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x0d, 0x54, 0x72,
	0x61, 0x6e, 0x73, 0x6c, 0x61, 0x74, 0x65, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_translate_v1_translate_proto_rawDescOnce sync.Once
	file_translate_v1_translate_proto_rawDescData = file_translate_v1_translate_proto_rawDesc
)

func file_translate_v1_translate_proto_rawDescGZIP() []byte {
	file_translate_v1_translate_proto_rawDescOnce.Do(func() {
		file_translate_v1_translate_proto_rawDescData = protoimpl.X.CompressGZIP(file_translate_v1_translate_proto_rawDescData)
	})
	return file_translate_v1_translate_proto_rawDescData
}

var file_translate_v1_translate_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_translate_v1_translate_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_translate_v1_translate_proto_goTypes = []interface{}{
	(Schema)(0),                             // 0: translate.v1.Schema
	(*Message)(nil),                         // 1: translate.v1.Message
	(*Messages)(nil),                        // 2: translate.v1.Messages
	(*UploadTranslationFileRequest)(nil),    // 3: translate.v1.UploadTranslationFileRequest
	(*DownloadTranslationFileRequest)(nil),  // 4: translate.v1.DownloadTranslationFileRequest
	(*DownloadTranslationFileResponse)(nil), // 5: translate.v1.DownloadTranslationFileResponse
	(*emptypb.Empty)(nil),                   // 6: google.protobuf.Empty
}
var file_translate_v1_translate_proto_depIdxs = []int32{
	1, // 0: translate.v1.Messages.messages:type_name -> translate.v1.Message
	0, // 1: translate.v1.UploadTranslationFileRequest.schema:type_name -> translate.v1.Schema
	0, // 2: translate.v1.DownloadTranslationFileRequest.schema:type_name -> translate.v1.Schema
	3, // 3: translate.v1.TranslateService.UploadTranslationFile:input_type -> translate.v1.UploadTranslationFileRequest
	4, // 4: translate.v1.TranslateService.DownloadTranslationFile:input_type -> translate.v1.DownloadTranslationFileRequest
	6, // 5: translate.v1.TranslateService.UploadTranslationFile:output_type -> google.protobuf.Empty
	5, // 6: translate.v1.TranslateService.DownloadTranslationFile:output_type -> translate.v1.DownloadTranslationFileResponse
	5, // [5:7] is the sub-list for method output_type
	3, // [3:5] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_translate_v1_translate_proto_init() }
func file_translate_v1_translate_proto_init() {
	if File_translate_v1_translate_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_translate_v1_translate_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_translate_v1_translate_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Messages); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_translate_v1_translate_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadTranslationFileRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_translate_v1_translate_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloadTranslationFileRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_translate_v1_translate_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloadTranslationFileResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_translate_v1_translate_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_translate_v1_translate_proto_goTypes,
		DependencyIndexes: file_translate_v1_translate_proto_depIdxs,
		EnumInfos:         file_translate_v1_translate_proto_enumTypes,
		MessageInfos:      file_translate_v1_translate_proto_msgTypes,
	}.Build()
	File_translate_v1_translate_proto = out.File
	file_translate_v1_translate_proto_rawDesc = nil
	file_translate_v1_translate_proto_goTypes = nil
	file_translate_v1_translate_proto_depIdxs = nil
}
