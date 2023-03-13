// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             (unknown)
// source: translate/v1/translate.proto

package translatev1

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	TranslateService_LoadService_FullMethodName             = "/translate.v1.TranslateService/LoadService"
	TranslateService_LoadServices_FullMethodName            = "/translate.v1.TranslateService/LoadServices"
	TranslateService_SaveService_FullMethodName             = "/translate.v1.TranslateService/SaveService"
	TranslateService_DeleteService_FullMethodName           = "/translate.v1.TranslateService/DeleteService"
	TranslateService_UploadTranslationFile_FullMethodName   = "/translate.v1.TranslateService/UploadTranslationFile"
	TranslateService_DownloadTranslationFile_FullMethodName = "/translate.v1.TranslateService/DownloadTranslationFile"
)

// TranslateServiceClient is the client API for TranslateService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TranslateServiceClient interface {
	LoadService(ctx context.Context, in *LoadServiceRequest, opts ...grpc.CallOption) (*Service, error)
	LoadServices(ctx context.Context, in *LoadServicesRequest, opts ...grpc.CallOption) (*LoadServicesResponse, error)
	SaveService(ctx context.Context, in *SaveServiceRequest, opts ...grpc.CallOption) (*Service, error)
	DeleteService(ctx context.Context, in *DeleteServiceRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	UploadTranslationFile(ctx context.Context, in *UploadTranslationFileRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	DownloadTranslationFile(ctx context.Context, in *DownloadTranslationFileRequest, opts ...grpc.CallOption) (*DownloadTranslationFileResponse, error)
}

type translateServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewTranslateServiceClient(cc grpc.ClientConnInterface) TranslateServiceClient {
	return &translateServiceClient{cc}
}

func (c *translateServiceClient) LoadService(ctx context.Context, in *LoadServiceRequest, opts ...grpc.CallOption) (*Service, error) {
	out := new(Service)
	err := c.cc.Invoke(ctx, TranslateService_LoadService_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *translateServiceClient) LoadServices(ctx context.Context, in *LoadServicesRequest, opts ...grpc.CallOption) (*LoadServicesResponse, error) {
	out := new(LoadServicesResponse)
	err := c.cc.Invoke(ctx, TranslateService_LoadServices_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *translateServiceClient) SaveService(ctx context.Context, in *SaveServiceRequest, opts ...grpc.CallOption) (*Service, error) {
	out := new(Service)
	err := c.cc.Invoke(ctx, TranslateService_SaveService_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *translateServiceClient) DeleteService(ctx context.Context, in *DeleteServiceRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, TranslateService_DeleteService_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *translateServiceClient) UploadTranslationFile(ctx context.Context, in *UploadTranslationFileRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, TranslateService_UploadTranslationFile_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *translateServiceClient) DownloadTranslationFile(ctx context.Context, in *DownloadTranslationFileRequest, opts ...grpc.CallOption) (*DownloadTranslationFileResponse, error) {
	out := new(DownloadTranslationFileResponse)
	err := c.cc.Invoke(ctx, TranslateService_DownloadTranslationFile_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TranslateServiceServer is the server API for TranslateService service.
// All implementations must embed UnimplementedTranslateServiceServer
// for forward compatibility
type TranslateServiceServer interface {
	LoadService(context.Context, *LoadServiceRequest) (*Service, error)
	LoadServices(context.Context, *LoadServicesRequest) (*LoadServicesResponse, error)
	SaveService(context.Context, *SaveServiceRequest) (*Service, error)
	DeleteService(context.Context, *DeleteServiceRequest) (*emptypb.Empty, error)
	UploadTranslationFile(context.Context, *UploadTranslationFileRequest) (*emptypb.Empty, error)
	DownloadTranslationFile(context.Context, *DownloadTranslationFileRequest) (*DownloadTranslationFileResponse, error)
	mustEmbedUnimplementedTranslateServiceServer()
}

// UnimplementedTranslateServiceServer must be embedded to have forward compatible implementations.
type UnimplementedTranslateServiceServer struct {
}

func (UnimplementedTranslateServiceServer) LoadService(context.Context, *LoadServiceRequest) (*Service, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LoadService not implemented")
}
func (UnimplementedTranslateServiceServer) LoadServices(context.Context, *LoadServicesRequest) (*LoadServicesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method LoadServices not implemented")
}
func (UnimplementedTranslateServiceServer) SaveService(context.Context, *SaveServiceRequest) (*Service, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SaveService not implemented")
}
func (UnimplementedTranslateServiceServer) DeleteService(context.Context, *DeleteServiceRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteService not implemented")
}
func (UnimplementedTranslateServiceServer) UploadTranslationFile(context.Context, *UploadTranslationFileRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UploadTranslationFile not implemented")
}
func (UnimplementedTranslateServiceServer) DownloadTranslationFile(context.Context, *DownloadTranslationFileRequest) (*DownloadTranslationFileResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DownloadTranslationFile not implemented")
}
func (UnimplementedTranslateServiceServer) mustEmbedUnimplementedTranslateServiceServer() {}

// UnsafeTranslateServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TranslateServiceServer will
// result in compilation errors.
type UnsafeTranslateServiceServer interface {
	mustEmbedUnimplementedTranslateServiceServer()
}

func RegisterTranslateServiceServer(s grpc.ServiceRegistrar, srv TranslateServiceServer) {
	s.RegisterService(&TranslateService_ServiceDesc, srv)
}

func _TranslateService_LoadService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoadServiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TranslateServiceServer).LoadService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TranslateService_LoadService_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TranslateServiceServer).LoadService(ctx, req.(*LoadServiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TranslateService_LoadServices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoadServicesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TranslateServiceServer).LoadServices(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TranslateService_LoadServices_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TranslateServiceServer).LoadServices(ctx, req.(*LoadServicesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TranslateService_SaveService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SaveServiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TranslateServiceServer).SaveService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TranslateService_SaveService_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TranslateServiceServer).SaveService(ctx, req.(*SaveServiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TranslateService_DeleteService_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DeleteServiceRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TranslateServiceServer).DeleteService(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TranslateService_DeleteService_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TranslateServiceServer).DeleteService(ctx, req.(*DeleteServiceRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TranslateService_UploadTranslationFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadTranslationFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TranslateServiceServer).UploadTranslationFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TranslateService_UploadTranslationFile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TranslateServiceServer).UploadTranslationFile(ctx, req.(*UploadTranslationFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _TranslateService_DownloadTranslationFile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DownloadTranslationFileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TranslateServiceServer).DownloadTranslationFile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TranslateService_DownloadTranslationFile_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TranslateServiceServer).DownloadTranslationFile(ctx, req.(*DownloadTranslationFileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TranslateService_ServiceDesc is the grpc.ServiceDesc for TranslateService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TranslateService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "translate.v1.TranslateService",
	HandlerType: (*TranslateServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "LoadService",
			Handler:    _TranslateService_LoadService_Handler,
		},
		{
			MethodName: "LoadServices",
			Handler:    _TranslateService_LoadServices_Handler,
		},
		{
			MethodName: "SaveService",
			Handler:    _TranslateService_SaveService_Handler,
		},
		{
			MethodName: "DeleteService",
			Handler:    _TranslateService_DeleteService_Handler,
		},
		{
			MethodName: "UploadTranslationFile",
			Handler:    _TranslateService_UploadTranslationFile_Handler,
		},
		{
			MethodName: "DownloadTranslationFile",
			Handler:    _TranslateService_DownloadTranslationFile_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "translate/v1/translate.proto",
}
