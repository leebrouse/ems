package rpc

import (
	"context"
	"errors"
	"time"

	pb "github.com/leebrouse/ems/backend/common/genproto/scheduling/grpc"
	"github.com/leebrouse/ems/backend/scheduling/model"
	"github.com/leebrouse/ems/backend/scheduling/service"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SchedulingRPCServer 提供调度服务的 gRPC 接口实现
type SchedulingRPCServer struct {
	pb.UnimplementedSchedulingServiceServer
	svc service.SchedulingService
}

// NewSchedulingRPCServer 创建 SchedulingRPCServer 实例
func NewSchedulingRPCServer(svc service.SchedulingService) *SchedulingRPCServer {
	return &SchedulingRPCServer{svc: svc}
}

// CreateRequest 创建需求单
func (s *SchedulingRPCServer) CreateRequest(ctx context.Context, in *pb.CreateRequestProto) (*pb.RequestResponse, error) {
	var items []model.RequestItem
	for _, item := range in.Items {
		items = append(items, model.RequestItem{
			ItemID:   int64(item.ItemId),
			Quantity: int(item.Quantity),
		})
	}
	// For gRPC, we assume createdBy is provided or handled elsewhere.
	req, err := s.svc.CreateRequest(ctx, in.Title, in.Location, items, 0)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return s.toRequestResponse(req), nil
}

// ListRequests 分页查询需求单
func (s *SchedulingRPCServer) ListRequests(ctx context.Context, in *pb.ListRequestsProto) (*pb.ListRequestsResponse, error) {
	reqs, total, err := s.svc.ListRequests(ctx, int(in.Page), int(in.Size), in.Status)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var res []*pb.RequestResponse
	for _, req := range reqs {
		res = append(res, s.toRequestResponse(&req))
	}
	return &pb.ListRequestsResponse{
		Requests: res,
		Total:    int32(total),
	}, nil
}

// GetRequest 获取需求单详情
func (s *SchedulingRPCServer) GetRequest(ctx context.Context, in *pb.GetRequestProto) (*pb.RequestResponse, error) {
	req, err := s.svc.GetRequest(ctx, int64(in.Id))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return s.toRequestResponse(req), nil
}

// UpdateRequest 更新需求单状态或负责人
func (s *SchedulingRPCServer) UpdateRequest(ctx context.Context, in *pb.UpdateRequestProto) (*pb.RequestResponse, error) {
	var assignedTo *int64
	if in.AssignedTo != 0 {
		val := int64(in.AssignedTo)
		assignedTo = &val
	}
	req, err := s.svc.UpdateRequest(ctx, int64(in.Id), model.RequestStatus(in.Status), assignedTo)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return s.toRequestResponse(req), nil
}

// DeleteRequest 删除需求单
func (s *SchedulingRPCServer) DeleteRequest(ctx context.Context, in *pb.DeleteRequestProto) (*empty.Empty, error) {
	err := s.svc.DeleteRequest(ctx, int64(in.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &empty.Empty{}, nil
}

// CreateShipment 创建运输任务
func (s *SchedulingRPCServer) CreateShipment(ctx context.Context, in *pb.CreateShipmentProto) (*pb.ShipmentResponse, error) {
	var items []model.ShipmentItem
	for _, item := range in.Items {
		items = append(items, model.ShipmentItem{
			ItemID:   int64(item.ItemId),
			Quantity: int(item.Quantity),
		})
	}
	shipment, err := s.svc.CreateShipment(ctx, int64(in.RequestId), int64(in.FromWarehouseId), in.ToLocation, items)
	if err != nil {
		if errors.Is(err, service.ErrInsufficientStock) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		if errors.Is(err, service.ErrInvalidShipmentItem) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, service.ErrWarehouseClientUnavailable) {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return s.toShipmentResponse(shipment), nil
}

// UpdateShipmentStatus 更新运输状态
func (s *SchedulingRPCServer) UpdateShipmentStatus(ctx context.Context, in *pb.UpdateShipmentProto) (*pb.ShipmentResponse, error) {
	shipment, err := s.svc.UpdateShipmentStatus(ctx, int64(in.ShipmentId), model.ShipmentStatus(in.Status), in.Location)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return s.toShipmentResponse(shipment), nil
}

// ListShipments 分页查询运输任务
func (s *SchedulingRPCServer) ListShipments(ctx context.Context, in *pb.ListShipmentsProto) (*pb.ListShipmentsResponse, error) {
	shipments, total, err := s.svc.ListShipments(ctx, int(in.Page), int(in.Size), in.Status)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var res []*pb.ShipmentResponse
	for _, shipment := range shipments {
		res = append(res, s.toShipmentResponse(&shipment))
	}

	return &pb.ListShipmentsResponse{
		Shipments: res,
		Total:     int32(total),
	}, nil
}

// GetShipment 获取运输任务详情
func (s *SchedulingRPCServer) GetShipment(ctx context.Context, in *pb.GetShipmentProto) (*pb.ShipmentResponse, error) {
	shipment, err := s.svc.GetShipment(ctx, int64(in.ShipmentId))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return s.toShipmentResponse(shipment), nil
}

// toRequestResponse 转换为 gRPC 响应结构
func (s *SchedulingRPCServer) toRequestResponse(req *model.Request) *pb.RequestResponse {
	var items []*pb.ItemQuantity
	for _, item := range req.Items {
		items = append(items, &pb.ItemQuantity{
			ItemId:   int32(item.ItemID),
			Quantity: int32(item.Quantity),
		})
	}
	assignedTo := int32(0)
	if req.AssignedTo != nil {
		assignedTo = int32(*req.AssignedTo)
	}
	return &pb.RequestResponse{
		Id:         int32(req.ID),
		Title:      req.Title,
		Location:   req.Location,
		Status:     string(req.Status),
		Items:      items,
		AssignedTo: assignedTo,
	}
}

// toShipmentResponse 转换为 gRPC 响应结构
func (s *SchedulingRPCServer) toShipmentResponse(shipment *model.Shipment) *pb.ShipmentResponse {
	var tracking []*pb.TrackingInfo
	for _, t := range shipment.Tracking {
		tracking = append(tracking, &pb.TrackingInfo{
			Status:    string(t.Status),
			Location:  t.Location,
			Timestamp: t.RecordedAt.Format(time.RFC3339),
		})
	}

	return &pb.ShipmentResponse{
		ShipmentId:      int32(shipment.ID),
		RequestId:       int32(shipment.RequestID),
		FromWarehouseId: int32(shipment.FromWarehouseID),
		ToLocation:      shipment.ToLocation,
		Status:          string(shipment.Status),
		Tracking:        tracking,
	}
}
